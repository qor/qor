package publish_test

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/publish"
)

var pb *publish.Publish
var pbdraft *gorm.DB
var pbprod *gorm.DB

func init() {
	db, _ := gorm.Open("sqlite3", "/tmp/qor_publish_test.db")
	pb = publish.New(&db)
	pb.Support(&Product{})
	pbdraft = pb.DraftDB()
	pbprod = pb.ProductionDB()

	for _, table := range []string{"products", "products_draft", "colors"} {
		pbprod.Exec(fmt.Sprintf("drop table %v", table))
	}
	pbprod.AutoMigrate(&Product{}, &Color{})
	pb.AutoMigrate()
}

type Product struct {
	Id        int
	Name      string
	Color     Color
	ColorId   int
	DeletedAt time.Time
	publish.Status
}

type Color struct {
	Id   int
	Name string
	publish.Status
}
