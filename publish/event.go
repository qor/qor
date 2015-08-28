package publish

import (
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
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
	PublishedAt *time.Time
	DiscardedAt *time.Time
	PublishedBy string
}

func (publishEvent *PublishEvent) Publish(db *gorm.DB) error {
	if event, ok := events[publishEvent.Name]; ok {
		err := event.Publish(db, publishEvent)
		if err == nil {
			now := time.Now()
			var updateAttrs = map[string]interface{}{"PublishedAt": &now}
			if user, ok := db.Get("qor:current_user"); ok {
				if primaryField := db.NewScope(user).PrimaryField(); primaryField != nil {
					updateAttrs["PublishedBy"] = fmt.Sprintf("%v", primaryField.Field.Interface())
				} else {
					updateAttrs["PublishedBy"] = fmt.Sprintf("%v", user)
				}
			}
			fmt.Println(updateAttrs)
			err = db.Model(publishEvent).Update(updateAttrs).Error
		}
		return err
	}
	return errors.New("event not found")
}

func (publishEvent *PublishEvent) Discard(db *gorm.DB) error {
	if event, ok := events[publishEvent.Name]; ok {
		err := event.Discard(db, publishEvent)
		if err == nil {
			now := time.Now()
			var updateAttrs = map[string]interface{}{"DiscardedAt": &now}
			if user, ok := db.Get("qor:current_user"); ok {
				if primaryField := db.NewScope(user).PrimaryField(); primaryField != nil {
					updateAttrs["PublishedBy"] = fmt.Sprintf("%v", primaryField.Field.Interface())
				} else {
					updateAttrs["PublishedBy"] = fmt.Sprintf("%v", user)
				}
			}
			err = db.Model(publishEvent).Update(updateAttrs).Error
		}
		return err
	}
	return errors.New("event not found")
}
