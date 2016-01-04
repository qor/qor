package publish_test

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/l10n"
	"github.com/qor/qor/publish"
	"github.com/qor/qor/test/utils"
)

var pb *publish.Publish
var pbdraft *gorm.DB
var pbprod *gorm.DB
var db *gorm.DB

func init() {
	db = utils.TestDB()
	l10n.RegisterCallbacks(db)

	pb = publish.New(db)
	pbdraft = pb.DraftDB()
	pbprod = pb.ProductionDB()

	for _, table := range []string{"product_categories", "product_categories_draft", "product_languages", "product_languages_draft", "author_books", "author_books_draft"} {
		pbprod.Exec(fmt.Sprintf("drop table %v", table))
	}

	for _, value := range []interface{}{&Product{}, &Color{}, &Category{}, &Language{}, &Book{}, &Publisher{}, &Comment{}, &Author{}} {
		pbprod.DropTable(value)
		pbdraft.DropTable(value)

		pbprod.AutoMigrate(value)
		pb.AutoMigrate(value)
	}
}

type Product struct {
	gorm.Model
	Name       string
	Quantity   uint
	Color      Color
	ColorId    int
	Categories []Category `gorm:"many2many:product_categories"`
	Languages  []Language `gorm:"many2many:product_languages"`
	publish.Status
}

type Color struct {
	gorm.Model
	Name string
}

type Language struct {
	gorm.Model
	Name string
}

type Category struct {
	gorm.Model
	Name string
	publish.Status
}
