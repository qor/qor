package publish_test

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/publish"
)

type createResourcePublishInterface struct {
}

func (createResourcePublishInterface) Publish(db *gorm.DB, event *publish.PublishEvent) error {
	var product Product
	db.Set("publish:draft_mode", true).First(&product, event.Argument)
	pb.Publish(&product)
	return nil
}

func (createResourcePublishInterface) Discard(db *gorm.DB, event *publish.PublishEvent) error {
	var product Product
	db.Set("publish:draft_mode", true).First(&product, event.Argument)
	pb.Discard(&product)
	return nil
}

func init() {
	publish.RegisterEvent("create_product", createResourcePublishInterface{})
}

func TestCreateNewEvent(t *testing.T) {
	product1 := Product{Name: "event_1"}
	pbdraft.Set("publish:skip_publish", true).Save(&product1)
	event := publish.PublishEvent{Name: "create_product", Argument: fmt.Sprintf("%v", product1.ID)}
	db.Save(&event)

	if !pbprod.First(&Product{}, "name = ?", product1.Name).RecordNotFound() {
		t.Errorf("created resource in draft db with event should not be published to production db")
	}

	if pbdraft.First(&Product{}, "name = ?", product1.Name).RecordNotFound() {
		t.Errorf("created resource in draft db with event should exist in draft db")
	}

	var publishEvent publish.PublishEvent
	if pbdraft.First(&publishEvent, "name = ?", "create_product").Error != nil {
		t.Errorf("created resource in draft db with event should create the event in db")
	}

	publishEvent.Publish(db)

	if pbprod.First(&Product{}, "name = ?", product1.Name).RecordNotFound() {
		t.Errorf("product should be published to production db after publish event")
	}
}
