package publish_test

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/publish"

	"testing"
	"time"
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
	Id        int
	Name      string
	Color     Color
	ColorId   int
	DeletedAt time.Time
}

type Color struct {
	Id   int
	Name string
}

func TestCreateStructFromDraft(t *testing.T) {
	name := "create_product_from_draft"
	pbdraft.Create(&Product{Name: name, Color: Color{Name: name}})

	if !pbprod.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should not be found in production db")
	}

	if pbdraft.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be found in draft db")
	}

	if pb.Table("colors").First(&Color{}, "name = ?", name).Error != nil {
		t.Errorf("color should be saved")
	}

	if pb.Table("colors_draft").First(&Color{}, "name = ?", name).Error == nil {
		t.Errorf("no colors_draft table")
	}

	var product Product
	pbdraft.First(&product, "name = ?", name)
	if pbdraft.Model(&product).Related(&product.Color); product.Color.Name != name {
		t.Errorf("should be able to find related struct")
	}
}

func TestCreateStructFromProduction(t *testing.T) {
	name := "create_product_from_production"
	pbprod.Create(&Product{Name: name, Color: Color{Name: name}})

	if pbprod.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should not be found in production db")
	}

	if pbdraft.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be found in draft db")
	}

	if pb.Table("colors").First(&Color{}, "name = ?", name).Error != nil {
		t.Errorf("color should be saved")
	}

	var product Product
	pbprod.First(&product, "name = ?", name)
	if pbprod.Model(&product).Related(&product.Color); product.Color.Name != name {
		t.Errorf("should be able to find related struct")
	}
}

func TestUpdateStructFromDraft(t *testing.T) {
	name := "update_product_from_draft"
	newName := name + "_v2"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)

	pbdraft.Model(&product).Update("name", newName)

	if pbprod.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should not be changed in production db")
	}

	if pbdraft.First(&Product{}, "name = ?", newName).RecordNotFound() {
		t.Errorf("record should be changed in draft db")
	}
	if pbdraft.Model(&product).Related(&product.Color); product.Color.Name != name {
		t.Errorf("should be able to find related struct")
	}
}

func TestUpdateStructFromProduction(t *testing.T) {
	name := "update_product_from_production"
	newName := name + "_v2"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)
	pbprod.Model(&product).Update("name", newName)

	if pbprod.First(&Product{}, "name = ?", newName).RecordNotFound() {
		t.Errorf("record should be changed in production db")
	}

	if pbdraft.First(&Product{}, "name = ?", newName).RecordNotFound() {
		t.Errorf("record should be changed in draft db")
	}

	if pbprod.Model(&product).Related(&product.Color); product.Color.Name != name {
		t.Errorf("should be able to find related struct")
	}
}

func TestDeleteStructFromDraft(t *testing.T) {
	name := "delete_product_from_draft"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)
	pbdraft.Delete(&product)

	if pbprod.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should not be deleted in production db")
	}

	if !pbdraft.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in draft db")
	}

	if pbdraft.Unscoped().First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in draft db")
	}
}

func TestDeleteStructFromProduction(t *testing.T) {
	name := "delete_product_from_production"
	product := Product{Name: name, Color: Color{Name: name}}
	pbprod.Create(&product)
	pbprod.Delete(&product)

	if !pbprod.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in production db")
	}

	if pbdraft.Unscoped().First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in production db")
	}

	if !pbdraft.First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in draft db")
	}

	if pbdraft.Unscoped().First(&Product{}, "name = ?", name).RecordNotFound() {
		t.Errorf("record should be soft deleted in draft db")
	}
}
