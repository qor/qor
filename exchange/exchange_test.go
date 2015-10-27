package exchange_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/exchange"
	"github.com/qor/qor/exchange/backends/csv"
	"github.com/qor/qor/test/utils"
)

var db *gorm.DB
var product *exchange.Resource

func init() {
	db = utils.TestDB()

	db.DropTable(&Product{})
	db.AutoMigrate(&Product{})

	product = exchange.NewResource(&Product{})
	product.Meta(exchange.Meta{Name: "Code"})
	product.Meta(exchange.Meta{Name: "Name"})
	product.Meta(exchange.Meta{Name: "Price"})
}

func TestImportCSV(t *testing.T) {
	context := &qor.Context{DB: db}
	if err := product.Import(csv.New("fixtures/products.csv"), context); err != nil {
		t.Fatalf("Failed to import csv, get error %v", err)
	}

	params := [][]interface{}{{"P001", "Product P001", 100}, {"P002", "Product P002", 200}, {"P003", "Product P003", 300}}
	for _, param := range params {
		if db.Where("code = ? AND name = ? AND price = ?", param...).First(&Product{}).RecordNotFound() {
			t.Errorf("Failed to find product", params)
		}
	}
}

func TestExportCSV(t *testing.T) {
	context := &qor.Context{DB: db}
	product.Import(csv.New("fixtures/products.csv"), context)

	if err := product.Export(csv.New("fixtures/products2.csv"), context); err != nil {
		t.Fatalf("Failed to export csv, get error %v", err)
	}
}
