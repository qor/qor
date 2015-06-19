package publish

import "github.com/jinzhu/gorm"

type PublishJoinTableHandler struct {
	gorm.JoinTableHandler
}

func (handler PublishJoinTableHandler) Table(db *gorm.DB) string {
	if draftMode, ok := db.Get("publish:draft_mode"); ok && draftMode.(bool) {
		return handler.TableName + "_draft"
	} else {
		return handler.TableName
	}
}
