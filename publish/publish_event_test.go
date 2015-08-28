package publish_test

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/publish"
)

type createResourcePublishInterface struct {
}

func (createResourcePublishInterface) Publish(db *gorm.DB, event publish.PublishEventInterface) error {
	if event, ok := event.(*publish.PublishEvent); ok {
		var product Product
		db.Set("publish:draft_mode", true).First(&product, event.Argument)
		pb.Publish(&product)
	}
	return nil
}

func (createResourcePublishInterface) Discard(db *gorm.DB, event publish.PublishEventInterface) error {
	if event, ok := event.(*publish.PublishEvent); ok {
		var product Product
		db.Set("publish:draft_mode", true).First(&product, event.Argument)
		pb.Discard(&product)
	}
	return nil
}

type publishAllResourcesInterface struct {
}

func (publishAllResourcesInterface) Publish(db *gorm.DB, event publish.PublishEventInterface) error {
	return nil
}

func (publishAllResourcesInterface) Discard(db *gorm.DB, event publish.PublishEventInterface) error {
	return nil
}

func init() {
	publish.RegisterEvent("create_product", createResourcePublishInterface{})
	publish.RegisterEvent("publish_all_resources", publishAllResourcesInterface{})
}

func TestCreateNewEvent(t *testing.T) {
	product1 := Product{Name: "event_1"}
	pbdraft.Set("publish:publish_event", true).Save(&product1)
	event := publish.PublishEvent{Name: "create_product", Argument: fmt.Sprintf("%v", product1.ID)}
	db.Save(&event)

	if !pbprod.First(&Product{}, "name = ?", product1.Name).RecordNotFound() {
		t.Errorf("created resource in draft db with event should not be published to production db")
	}

	var productDraft Product
	if pbdraft.First(&productDraft, "name = ?", product1.Name).RecordNotFound() {
		t.Errorf("created resource in draft db with event should exist in draft db")
	}

	if productDraft.PublishStatus == publish.DIRTY {
		t.Errorf("product's publish status should not be DIRTY before publish event")
	}

	var publishEvent publish.PublishEvent
	if pbdraft.First(&publishEvent, "name = ?", "create_product").Error != nil {
		t.Errorf("created resource in draft db with event should create the event in db")
	}

	if !pbprod.First(&Product{}, "name = ?", product1.Name).RecordNotFound() {
		t.Errorf("product should not be published to production db before publish event")
	}

	publishEvent.Publish(db)

	if pbprod.First(&Product{}, "name = ?", product1.Name).RecordNotFound() {
		t.Errorf("product should be published to production db after publish event")
	}
}

func TestCreateProductWithPublishAllEvent(t *testing.T) {
	product1 := Product{Name: "event_1"}
	event := &publish.PublishEvent{Name: "publish_all_resources", Argument: "products"}
	pbdraft.Set("publish:publish_event", event).Save(&product1)
}
