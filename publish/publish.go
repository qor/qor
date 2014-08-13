package publish

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
)

type Publish struct {
	*gorm.DB
}

func Open(driver, source string) (Publish, error) {
	db, err := gorm.Open(driver, source)
	return Publish{DB: &db}, err
}

func (publish *Publish) GormDB() *gorm.DB {
	return publish.DB
}

func (publish *Publish) Support(models ...interface{}) {
}

func (publish *Publish) ProductionMode() {
}

func (publish *Publish) DraftMode() {
}

func SetTable(scope *gorm.Scope) {
	tableName := scope.TableName()
	inDraft := true

	if inDraft {
		tableName = fmt.Sprintf("%v_draft", tableName)
		scope.Search.TableName = tableName
	}
}

// Auto Migration

type Product struct {
	Title     string
	ColorCode string
	Price     float64
	Ext       string
	PublishAt time.Time
	Image     MediaLibrary `media_library:"path:/system/:table_name/:id/:filename;"`
}

Product{Title: "product A", Image: os.Open("xxxx")}
db.Save(&product)

db, err := publish.Open("sqlite", "/tmp/qor.db")
user := db.NewResource(&Product{})
user.InstantPublishAttrs("title", "color_code", "price", "colorA", "colorB")
user.IgnoredAttrs("ext")

// /system_draft/products/xxx.png
// /system/products/xxx.png

// -> s3
