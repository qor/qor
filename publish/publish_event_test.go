package publish_test

import "github.com/qor/qor/publish"

type createResourcePublishInterface struct {
}

func (createResourcePublishInterface) Publish(event *publish.PublishEvent) error {
	return nil
}

func (createResourcePublishInterface) Discard(event *publish.PublishEvent) error {
	return nil
}

func init() {
	publish.RegisterEvent("created_resource", createResourcePublishInterface{})
}

// func TestCreateNewEvent(t *testing.T) {
// 	product1 := Product{Name: "event_1"}
// 	pbdraft.Set("publish:new_event", &publish.PublishEvent{Name: "created_resource", Argument: product1.Name}).Save(&product1)
//
// 	if !pbprod.First(&Product{}, "name = ?", product1.Name).RecordNotFound() {
// 		t.Errorf("created resource in draft db with event should not be published to production db")
// 	}
//
// 	if pbdraft.First(&Product{}, "name = ?", product1.Name).RecordNotFound() {
// 		t.Errorf("created resource in draft db with event should exist in draft db")
// 	}
//
// 	var publishEvent publish.PublishEvent
// 	if pbdraft.First(&publishEvent, "name = ?", "created_resource").RecordNotFound() {
// 		t.Errorf("created resource in draft db with event should create the event in db")
// 	}
//
// 	publishEvent.Publish()
//
// 	if pbprod.First(&Product{}, "name = ?", product1.Name).RecordNotFound() {
// 		t.Errorf("product should be published to production db after publish event")
// 	}
// }
