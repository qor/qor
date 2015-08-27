package publish

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/audited"
)

type EventInterface interface {
	Publish(db *gorm.DB, event *PublishEvent) error
	Discard(db *gorm.DB, event *PublishEvent) error
}

var events = map[string]EventInterface{}

func RegisterEvent(name string, event EventInterface) {
	events[name] = event
}

type PublishEvent struct {
	gorm.Model
	Name        string
	Description string
	Argument    string `sql:"size:65532"`
	PublishedAt string
	PublishedBy string
	audited.AuditedModel
}

func (publishEvent *PublishEvent) Publish(db *gorm.DB) error {
	if event, ok := events[publishEvent.Name]; ok {
		return event.Publish(db, publishEvent)
	}
	return errors.New("event not found")
}

func (publishEvent *PublishEvent) Discard(db *gorm.DB) error {
	if event, ok := events[publishEvent.Name]; ok {
		return event.Discard(db, publishEvent)
	}
	return errors.New("event not found")
}
