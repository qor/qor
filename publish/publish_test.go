package publish_test

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/publish"
)

var pb *publish.DB
var pbdraft *gorm.DB
var pbprod *gorm.DB

func init() {
	pb, _ = publish.Open("sqlite3", "/tmp/qor_publish_test.db")
	pb.Support(&Product{})
	pbdraft = pb.DraftMode()
	pbprod = pb.ProductionMode()

	for _, table := range []string{"products", "products_draft", "colors"} {
		pb.Exec(fmt.Sprintf("drop table %v", table))
	}
	pb.AutoMigrate(&Product{}, &Color{})
	pb.AutoMigrateDrafts()
}

type Product struct {
	Id        int
	Name      string
	Color     Color
	ColorId   int
	DeletedAt time.Time
	publish.Publish
}

type Color struct {
	Id   int
	Name string
	publish.Publish
}
