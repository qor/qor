package sorting_test

import (
	"fmt"
	"reflect"
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

func checkProductPositionInDB(db *gorm.DB, names ...string) bool {
	var products []Product
	var positions []string

	db.Find(&products)
	for _, product := range products {
		positions = append(positions, product.Name)
	}

	if reflect.DeepEqual(positions, names) {
		return true
	} else {
		fmt.Printf("Expect %v, got %v\n", names, positions)
		return false
	}
}

func TestSortingWithPublishStatus(t *testing.T) {
	prepareProducts()
	var count1 int
	pb.DraftDB().Model(&Product{}).Where("publish_status = ?", publish.DIRTY).Count(&count1)
	if count1 != 0 {
		t.Errorf("Should no draft products after create product")
	}

	sorting.MoveTo(db, getProduct("product5"), 2)
	if !checkProductPositionInDB(pb.DraftDB(), "product1", "product5", "product2", "product3", "product4") {
		t.Errorf("Order in draft db should be correct when change order in production db")
	}
	if !checkProductPositionInDB(pb.ProductionDB(), "product1", "product5", "product2", "product3", "product4") {
		t.Errorf("Order in production db should be correct when change order in production db")
	}

	var count2 int
	pb.DraftDB().Model(&Product{}).Where("publish_status = ?", publish.DIRTY).Count(&count2)
	if count2 != 0 {
		t.Errorf("Should no draft products after sorting in production db")
	}

	sorting.MoveTo(pb.DraftDB(), getProduct("product5"), 4)
	if !checkProductPositionInDB(pb.DraftDB(), "product1", "product2", "product3", "product5", "product4") {
		t.Errorf("Order in draft db should be correct when change order in draft db")
	}
	if !checkProductPositionInDB(pb.ProductionDB(), "product1", "product5", "product2", "product3", "product4") {
		t.Errorf("Order in production db should be correct when change order in draft db")
	}

	var count3 int
	pb.DraftDB().Model(&Product{}).Where("publish_status = ?", publish.DIRTY).Count(&count3)
	if count3 != 0 {
		t.Errorf("Should no draft products after sorting in draft db")
	}

	var publishEvents []publish.PublishEvent
	db.Where("published_at IS NULL AND discarded_at IS NULL AND name = ?", "changed_sorting").Find(&publishEvents)
	for _, publishEvent := range publishEvents {
		publishEvent.Publish(pb.ProductionDB())
	}
	if !checkProductPositionInDB(pb.DraftDB(), "product1", "product2", "product3", "product5", "product4") {
		t.Errorf("Order in draft db should be correct after publish event")
	}
	if !checkProductPositionInDB(pb.ProductionDB(), "product1", "product2", "product3", "product5", "product4") {
		t.Errorf("Order in production db should be correct after publish event")
	}

	sorting.MoveTo(pb.DraftDB(), getProduct("product5"), 1)
	if !checkProductPositionInDB(pb.DraftDB(), "product5", "product1", "product2", "product3", "product4") {
		t.Errorf("Order in draft db should be correct when change order in draft db")
	}
	if !checkProductPositionInDB(pb.ProductionDB(), "product1", "product2", "product3", "product5", "product4") {
		t.Errorf("Order in production db should be correct when change order in draft db")
	}

	var publishDiscardEvents []publish.PublishEvent
	db.Where("published_at IS NULL AND discarded_at IS NULL AND name = ?", "changed_sorting").Find(&publishDiscardEvents)
	if len(publishDiscardEvents) > 1 {
		t.Errorf("Should only found one publish event for changed_sorting, but got %v", len(publishDiscardEvents))
	}
	for _, publishEvent := range publishDiscardEvents {
		publishEvent.Discard(pb.ProductionDB())
	}

	if !checkProductPositionInDB(pb.ProductionDB(), "product1", "product2", "product3", "product5", "product4") {
		t.Errorf("Order in draft db should be correct after discard event")
	}
	if !checkProductPositionInDB(pb.ProductionDB(), "product1", "product2", "product3", "product5", "product4") {
		t.Errorf("Order in production db should be correct after discard event")
	}
}
