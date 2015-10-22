package exchange_test

import (
	"fmt"
	"strconv"
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
		if value, err := strconv.ParseFloat(metaValue.Value.(string), 64); err == nil {
			resource.(*Product).Price = value
		} else {
			fmt.Println(err)
		}
	}})
}

func TestImportCSV(t *testing.T) {
	context := &qor.Context{DB: db}
	if err := product.Import(csv.New("fixtures/products.csv"), context); err != nil {
		t.Fatalf("Failed to import csv, get error %v", err)
	}

	var products []Product
	db.Find(&products)
	fmt.Printf("%+v\n", products[0])
	if len(products) != 3 {
		t.Errorf("Failed to find importted products, got %v", len(products))
	}
}
