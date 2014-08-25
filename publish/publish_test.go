package publish_test

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/publish"

	"testing"
)

var pb *publish.Publish
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

func TestCreateStructFromDraft(t *testing.T) {
	name := "product_publish_draft"
	pbdraft.Create(&Product{Name: name, Color: Color{Name: "red"}})

	if !pbprod.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should not be found in production db")
	}

	if pbdraft.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be found in draft db")
	}

	if pb.Table("colors").First(&Color{}).Error != nil {
		t.Errorf("color should be saved")
	}

	if pb.Table("colors_draft").First(&Color{}).Error == nil {
		t.Errorf("no colors_draft table")
	}

	var product Product
	pbdraft.First(&product, "name = ?", name)
	if pbdraft.Model(&product).Related(&product.Color); product.Color.Name != "red" {
		t.Errorf("should be able to find related struct")
	}
}
