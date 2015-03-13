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
	db.AutoMigrate(&Product{})
	db.LogMode(true)

	dbGlobal = &db
	dbCN = dbGlobal.Set("l10n:locale", "zh")
	dbEN = dbGlobal.Set("l10n:locale", "en")
}

func checkHasErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func checkHasProductInLocale(db *gorm.DB, locale string, t *testing.T) {
	var count int
	if db.Where("language_code = ?", locale).Count(&count); count != 1 {
		t.Errorf("should create one product for locale %v", locale)
	}
}

func checkHasProductInAllLocales(db *gorm.DB, t *testing.T) {
	checkHasProductInLocale(db, "", t)
	checkHasProductInLocale(db, "zh", t)
	checkHasProductInLocale(db, "en", t)
}

func TestCreateLocalesWithCreate(t *testing.T) {
	product := Product{Code: "CreateLocalesWithCreate"}
	checkHasErr(t, dbGlobal.Create(&product).Error)
	checkHasErr(t, dbCN.Create(&product).Error)
	checkHasErr(t, dbEN.Create(&product).Error)

	checkHasProductInAllLocales(dbGlobal.Model(&Product{}).Where("id = ? AND code = ?", product.ID, "CreateLocalesWithCreate"), t)
}

func TestCreateLocalesWithSave(t *testing.T) {
	product := Product{Code: "CreateLocalesWithSave"}
	checkHasErr(t, dbGlobal.Create(&product).Error)
	checkHasErr(t, dbCN.Create(&product).Error)
	checkHasErr(t, dbEN.Create(&product).Error)

	checkHasProductInAllLocales(dbGlobal.Model(&Product{}).Where("id = ? AND code = ?", product.ID, "CreateLocalesWithSave"), t)
}

func TestUpdateLocales(t *testing.T) {
	product := Product{Code: "CreateUpdateLocales", Name: "global"}
	checkHasErr(t, dbGlobal.Create(&product).Error)
	product.Name = "中文名"
	checkHasErr(t, dbCN.Create(&product).Error)
	checkHasProductInLocale(dbGlobal.Model(&Product{}).Where("id = ? AND code = ?", product.ID, "中文名"), "zh", t)

	product.Name = "English Name"
	checkHasErr(t, dbEN.Create(&product).Error)
	checkHasProductInLocale(dbGlobal.Model(&Product{}).Where("id = ? AND code = ?", product.ID, "English Name"), "en", t)

}
