package sorting_test

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/publish"
	"github.com/qor/qor/sorting"
)

type Product struct {
	gorm.Model
	Name string
	sorting.Sorting
	publish.Status
}

func prepareProducts() {
	db.Delete(&Product{})

	for i := 1; i <= 5; i++ {
		product := Product{Name: fmt.Sprintf("product%v", i)}
		db.Save(&product)
	}
}

func getProduct(name string) *Product {
	var product Product
	db.First(&product, "name = ?", name)
	return &product
}

func TestMoveUpPositionWithPublishStatus(t *testing.T) {
	prepareProducts()
}
