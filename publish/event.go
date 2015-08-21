package publish

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/audited"
)

type EventInterface interface {
	Publish(event PublishEvent) error
	Discard(event PublishEvent) error
}

var events = map[string]EventInterface{}

func RegisterEvent(name string, event EventInterface) {
	events[name] = event
}

type PublishEvent struct {
	gorm.Model
	Name        string
	Description string
	Argument    string `sql:"size:65536"`
	PublishedAt string
	PublishedBy string
	audited.AuditedModel
}

func (publishEvent PublishEvent) Publish() error {
	if event, ok := events[publishEvent.Name]; ok {
		return event.Publish(publishEvent)
	}
	return errors.New("event not found")
}

func (publishEvent PublishEvent) Discard() error {
	if event, ok := events[publishEvent.Name]; ok {
		return event.Discard(publishEvent)
	}
	return errors.New("event not found")
}
