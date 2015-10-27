package exchange_test

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/exchange"
	csv_adaptor "github.com/qor/qor/exchange/backends/csv"
	"github.com/qor/qor/test/utils"
)

var db *gorm.DB
var product *exchange.Resource

func init() {
	db = utils.TestDB()

	db.DropTable(&Product{})
	db.AutoMigrate(&Product{})

	product = exchange.NewResource(&Product{}, exchange.Config{PrimaryField: "Code"})
	product.Meta(exchange.Meta{Name: "Code"})
	product.Meta(exchange.Meta{Name: "Name"})
	product.Meta(exchange.Meta{Name: "Price"})
}

func checkProduct(t *testing.T, filename string) {
	csvfile, _ := os.Open(filename)
	defer csvfile.Close()
	reader := csv.NewReader(csvfile)
	reader.TrimLeadingSpace = true
	params, _ := reader.ReadAll()

	for index, param := range params {
		if index == 0 {
			continue
		}
		var count int
		if db.Model(&Product{}).Where("code = ?", param[0]).Count(&count); count != 1 {
			t.Errorf("Found %v with code %v, but should find one (%v)", count, param[0], filename)
		} else if db.Model(&Product{}).Where("code = ? AND name = ? AND price = ?", param[0], param[1], param[2]).Count(&count); count != 1 {
			t.Errorf("Found %v with params %v, but should find one (%v)", count, param, filename)
		}
	}
}

func TestImportCSV(t *testing.T) {
	context := &qor.Context{DB: db}
	if err := product.Import(csv_adaptor.New("fixtures/products.csv"), context); err != nil {
		t.Fatalf("Failed to import csv, get error %v", err)
	}

	checkProduct(t, "fixtures/products.csv")

	if err := product.Import(csv_adaptor.New("fixtures/products_update.csv"), context); err != nil {
		t.Fatalf("Failed to import csv, get error %v", err)
	}

	checkProduct(t, "fixtures/products_update.csv")
}

func TestExportCSV(t *testing.T) {
	context := &qor.Context{DB: db}
	product.Import(csv_adaptor.New("fixtures/products.csv"), context)

	if err := product.Export(csv_adaptor.New("fixtures/products2.csv"), context); err != nil {
		t.Fatalf("Failed to export csv, get error %v", err)
	}
}
