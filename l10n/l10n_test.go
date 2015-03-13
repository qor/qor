package l10n_test

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/l10n"

	_ "github.com/go-sql-driver/mysql"
)

type Product struct {
	ID int `gorm:"primary_key"`
	l10n.Locale
	Code      string
	Name      string
	DeletedAt time.Time
}

func (Product) LocaleCreateable() {}

var dbGlobal, dbCN, dbEN *gorm.DB

func init() {
	// db, _ := gorm.Open("sqlite3", "/tmp/qor_l10n_test.db")
	db, _ := gorm.Open("mysql", "gorm:gorm@/gorm?charset=utf8&parseTime=True")
	l10n.RegisterCallbacks(&db)

	db.DropTable(&Product{})
	db.Debug().AutoMigrate(&Product{})
	db.LogMode(true)

	dbGlobal = &db
	dbCN = dbGlobal.Set("l10n:locale", "zh")
	dbEN = dbGlobal.Set("l10n:locale", "en")
}

func TestL10n(t *testing.T) {
	product := Product{Code: "L1212"}
	dbGlobal.Create(&product)

	dbCN.Create(&product)
	dbEN.Create(&product)
	dbCN.Delete(&product)
	dbCN.Unscoped().Save(&product)
	dbEN.Create(&product)
}
