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
	product = exchange.NewResource(&Product{})
}

func TestImportCSV(t *testing.T) {
	product.Import(csv.New("fixtures/products.csv"), nil)
}
