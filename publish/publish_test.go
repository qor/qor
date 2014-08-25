package publish_test

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/publish"

	"testing"
)

var pb *publish.Publish

func init() {
	pb, _ = publish.Open("sqlite3", "/tmp/qor_publish_test.db")
	pb.Support(&Product{})

	pb.AutoMigrate(Product{}, Color{})
	pb.AutoMigrateDrafts()
}

type Product struct {
	Id      int
	Name    string
	Color   Color
	ColorId int
}

type Color struct {
	Id   int
	Name string
}

func TestCreateStruct(t *testing.T) {
	name := "product_publish_draft"
	pb.DraftMode().Create(Product{Name: name, Color: Color{Name: "red"}})

	if !pb.ProductionMode().First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should not be found in production db")
	}

	if pb.DraftMode().First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be found in draft db")
	}

	if pb.Table("colors").First(&Color{}).Error != nil {
		t.Errorf("color should be saved")
	}

	if pb.Table("colors_draft").First(&Color{}).Error == nil {
		t.Errorf("no colors_draft table")
	}
}
