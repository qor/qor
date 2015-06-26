package publish_test

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/publish"
	"github.com/qor/qor/test/utils"
)

var pb *publish.Publish
var pbdraft *gorm.DB
var pbprod *gorm.DB
var db *gorm.DB

func init() {
	db = utils.TestDB()
	pb = publish.New(db)
	pbdraft = pb.DraftDB()
	pbprod = pb.ProductionDB()

	for _, table := range []string{"products", "products_draft", "colors"} {
		pbprod.Exec(fmt.Sprintf("drop table %v", table))
	}
	pbprod.AutoMigrate(&Product{}, &Color{}, &Category{})
	pb.AutoMigrate(&Product{})
}

type Product struct {
	gorm.Model
	Name       string
	Quantity   uint
	Color      Color
	ColorId    int
	Categories []Category `gorm:"many2many:product_categories"`
	publish.Status
}

type Color struct {
	gorm.Model
	Name string
}

type Category struct {
	gorm.Model
	Name string
}
