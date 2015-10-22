package exchange_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/exchange"
	"github.com/qor/qor/exchange/backends/csv"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/test/utils"
)

var db *gorm.DB
var product *exchange.Resource

func init() {
	db = utils.TestDB()

	db.DropTable(&Product{})
	db.AutoMigrate(&Product{})

	product = exchange.NewResource(&Product{})
	product.Meta(exchange.Meta{Name: "Code", Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
		resource.(*Product).Code = metaValue.Value.(string)
	}})
	product.Meta(exchange.Meta{Name: "Name", Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
		resource.(*Product).Name = metaValue.Value.(string)
	}})
	product.Meta(exchange.Meta{Name: "Price", Setter: func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
		resource.(*Product).Price = metaValue.Value.(float64)
	}})
}

func TestImportCSV(t *testing.T) {
	context := &qor.Context{DB: db}
	if err := product.Import(csv.New("fixtures/products.csv"), context); err != nil {
		t.Fatalf("Failed to import csv, get error %v", err)
	}

	var products []Product
	db.Find(&products)
	if len(products) != 3 {
		t.Errorf("Failed to find importted products, got %v", len(products))
	}
}
