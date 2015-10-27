package exchange_test

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/exchange"
	csv_adaptor "github.com/qor/qor/exchange/backends/csv"
	"github.com/qor/qor/resource"
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

func newContext() *qor.Context {
	return &qor.Context{DB: db}
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
	if err := product.Import(csv_adaptor.New("fixtures/products.csv"), newContext()); err != nil {
		t.Fatalf("Failed to import csv, get error %v", err)
	}

	checkProduct(t, "fixtures/products.csv")

	if err := product.Import(csv_adaptor.New("fixtures/products_update.csv"), newContext()); err != nil {
		t.Fatalf("Failed to import csv, get error %v", err)
	}

	checkProduct(t, "fixtures/products_update.csv")
}

func TestExportCSV(t *testing.T) {
	product.Import(csv_adaptor.New("fixtures/products.csv"), newContext())

	if err := product.Export(csv_adaptor.New("fixtures/products2.csv"), newContext()); err != nil {
		t.Fatalf("Failed to export csv, get error %v", err)
	}
}

func TestImportWithInvalidData(t *testing.T) {
	product = exchange.NewResource(&Product{}, exchange.Config{PrimaryField: "Code"})
	product.Meta(exchange.Meta{Name: "Code"})
	product.Meta(exchange.Meta{Name: "Name"})
	product.Meta(exchange.Meta{Name: "Price"})

	product.AddValidator(func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
		if f, err := strconv.ParseFloat(fmt.Sprint(metaValues.Get("Price").Value), 64); err == nil {
			if f == 0 {
				return errors.New("product's price can't be env")
			}
			return nil
		} else {
			return err
		}
	})

	if err := product.Import(csv_adaptor.New("fixtures/products.csv"), newContext()); err != nil {
		t.Errorf("Failed to import product, get error", err)
	}

	if err := product.Import(csv_adaptor.New("fixtures/invalid_price_products.csv"), newContext()); err == nil {
		t.Errorf("should get error when import products with invalid price")
	}
}

func TestProcessImportedData(t *testing.T) {
	product = exchange.NewResource(&Product{}, exchange.Config{PrimaryField: "Code"})
	product.Meta(exchange.Meta{Name: "Code"})
	product.Meta(exchange.Meta{Name: "Name"})
	product.Meta(exchange.Meta{Name: "Price"})

	product.AddProcessor(func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
		product := result.(*Product)
		product.Price = float64(int(product.Price * 1.1)) // Add 10% Tax
		return nil
	})

	if err := product.Import(csv_adaptor.New("fixtures/products.csv"), newContext()); err != nil {
		t.Errorf("Failed to import product, get error", err)
	}

	checkProduct(t, "fixtures/products_with_tax.csv")
}
