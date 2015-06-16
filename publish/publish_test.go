package publish_test

import (
	"fmt"
	"time"

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
	pbprod.AutoMigrate(&Product{}, &Color{})
	pb.AutoMigrate(&Product{})
}

type Product struct {
	Id        int
	Name      string
	Color     Color
	ColorId   int
	Quantity  uint
	DeletedAt time.Time
	publish.Status
}

type Color struct {
	Id   int
	Name string
}
