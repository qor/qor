package publish

import "github.com/jinzhu/gorm"

type publishJoinTableHandler struct {
	gorm.JoinTableHandler
}

func (handler publishJoinTableHandler) Table(db *gorm.DB) string {
	if draftMode, ok := db.Get("publish:draft_mode"); ok && draftMode.(bool) {
		return handler.TableName + "_draft"
	} else {
		return handler.TableName
	}
}

func (handler publishJoinTableHandler) Add(h gorm.JoinTableHandlerInterface, db *gorm.DB, source1 interface{}, source2 interface{}) error {
	// production mode
	if draftMode, ok := db.Get("publish:draft_mode"); !ok || !draftMode.(bool) {
		if err := handler.JoinTableHandler.Add(h, db.Set("publish:draft_mode", true), source1, source2); err != nil {
			return err
		}
	}
	return handler.JoinTableHandler.Add(h, db, source1, source2)
}

func (handler publishJoinTableHandler) Delete(h gorm.JoinTableHandlerInterface, db *gorm.DB, sources ...interface{}) error {
	// production mode
	if draftMode, ok := db.Get("publish:draft_mode"); !ok || !draftMode.(bool) {
		if err := handler.JoinTableHandler.Delete(h, db.Set("publish:draft_mode", true), sources...); err != nil {
			return err
		}
	}
	return handler.JoinTableHandler.Delete(h, db, sources...)
}
