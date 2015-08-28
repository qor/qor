package sorting

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/publish"
)

type changedSortingPublishEvent struct {
}

func (changedSortingPublishEvent) Publish(db *gorm.DB, event *publish.PublishEvent) error {
	scope := db.NewScope("")
	originalTable := scope.Quote(publish.OriginalTableName(event.Argument))
	draftTable := scope.Quote(publish.DraftTableName(event.Argument))
	sql := fmt.Sprintf("UPDATE %v SET position = (select position FROM %v WHERE %v.id = %v.id);", originalTable, draftTable, originalTable, draftTable)
	return db.Exec(sql).Error
}

func (changedSortingPublishEvent) Discard(db *gorm.DB, event *publish.PublishEvent) error {
	scope := db.NewScope("")
	originalTable := scope.Quote(publish.OriginalTableName(event.Argument))
	draftTable := scope.Quote(publish.DraftTableName(event.Argument))
	sql := fmt.Sprintf("UPDATE %v SET position = (select position FROM %v WHERE %v.id = %v.id);", draftTable, originalTable, draftTable, originalTable)
	return db.Exec(sql).Error
}

func init() {
	publish.RegisterEvent("changed_sorting", changedSortingPublishEvent{})
}
