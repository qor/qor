package publish

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
)

type EventInterface interface {
	Publish(db *gorm.DB, event PublishEventInterface) error
	Discard(db *gorm.DB, event PublishEventInterface) error
}

var events = map[string]EventInterface{}

func RegisterEvent(name string, event EventInterface) {
	events[name] = event
}

type PublishEvent struct {
	gorm.Model
	Name          string
	Description   string
	Argument      string `sql:"size:65532"`
	PublishStatus bool
	PublishedBy   string
}

func getCurrentUser(db *gorm.DB) (string, bool) {
	if user, hasUser := db.Get("qor:current_user"); hasUser {
		var currentUser string
		if primaryField := db.NewScope(user).PrimaryField(); primaryField != nil {
			currentUser = fmt.Sprintf("%v", primaryField.Field.Interface())
		} else {
			currentUser = fmt.Sprintf("%v", user)
		}

		return currentUser, true
	}

	return "", false
}

func (publishEvent *PublishEvent) Publish(db *gorm.DB) error {
	if event, ok := events[publishEvent.Name]; ok {
		err := event.Publish(db, publishEvent)
		if err == nil {
			var updateAttrs = map[string]interface{}{"PublishStatus": PUBLISHED}
			if user, hasUser := getCurrentUser(db); hasUser {
				updateAttrs["PublishedBy"] = user
			}
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
			var updateAttrs = map[string]interface{}{"PublishStatus": PUBLISHED}
			if user, hasUser := getCurrentUser(db); hasUser {
				updateAttrs["PublishedBy"] = user
			}
			err = db.Model(publishEvent).Update(updateAttrs).Error
		}
		return err
	}
	return errors.New("event not found")
}

func (PublishEvent) VisiblePublishResource() bool {
	return true
}
