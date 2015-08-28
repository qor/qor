package sorting

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/publish"
)

type changedSortingPublishEvent struct {
	Table       string
	PrimaryKeys []string
}

func (e changedSortingPublishEvent) Publish(db *gorm.DB, event publish.PublishEventInterface) error {
	if event, ok := event.(*publish.PublishEvent); ok {
		scope := db.NewScope("")
		if err := json.Unmarshal([]byte(event.Argument), &e); err == nil {
			var conditions []string
			originalTable := scope.Quote(publish.OriginalTableName(e.Table))
			draftTable := scope.Quote(publish.DraftTableName(e.Table))
			for _, primaryKey := range e.PrimaryKeys {
				conditions = append(conditions, fmt.Sprintf("%v.%v = %v.%v", originalTable, primaryKey, draftTable, primaryKey))
			}
			sql := fmt.Sprintf("UPDATE %v SET position = (select position FROM %v WHERE %v);", originalTable, draftTable, strings.Join(conditions, " AND "))
			return db.Exec(sql).Error
		} else {
			return err
		}
	}
	return errors.New("invalid publish event")
}

func (e changedSortingPublishEvent) Discard(db *gorm.DB, event publish.PublishEventInterface) error {
	if event, ok := event.(*publish.PublishEvent); ok {
		scope := db.NewScope("")
		if err := json.Unmarshal([]byte(event.Argument), &e); err == nil {
			var conditions []string
			originalTable := scope.Quote(publish.OriginalTableName(e.Table))
			draftTable := scope.Quote(publish.DraftTableName(e.Table))
			for _, primaryKey := range e.PrimaryKeys {
				conditions = append(conditions, fmt.Sprintf("%v.%v = %v.%v", originalTable, primaryKey, draftTable, primaryKey))
			}
			sql := fmt.Sprintf("UPDATE %v SET position = (select position FROM %v WHERE %v);", draftTable, originalTable, strings.Join(conditions, " AND "))
			return db.Exec(sql).Error
		} else {
			return err
		}
	}
	return errors.New("invalid publish event")
}

func init() {
	publish.RegisterEvent("changed_sorting", changedSortingPublishEvent{})
}
