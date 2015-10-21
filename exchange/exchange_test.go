package exchange_test

import (
	"testing"

	"github.com/jinzhu/gorm"
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
}

func TestImportCSV(t *testing.T) {
	if err := product.Import(csv.New("fixtures/products.csv"), nil); err != nil {
		t.Fatalf("Failed to import csv, get error %v", err)
	}

	var products []Product
	db.Find(&products)
	if len(products) != 3 {
		t.Errorf("Failed to find importted products, got %v", len(products))
	}
}
