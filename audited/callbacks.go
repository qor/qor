package audited

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func assignCreatedBy(scope *gorm.Scope) {
	if user, ok := scope.DB().Get("audited:current_user"); ok {
		if primaryField := scope.New(user).PrimaryField(); primaryField != nil {
			scope.SetColumn("CreatedBy", fmt.Sprintf("%v", primaryField.Field.Interface()))
		} else {
			scope.SetColumn("CreatedBy", fmt.Sprintf("%v", user))
		}
	}
}

func assignUpdatedBy(scope *gorm.Scope) {
	if user, ok := scope.DB().Get("audited:current_user"); ok {
		var currentUser string
		if primaryField := scope.New(user).PrimaryField(); primaryField != nil {
			currentUser = fmt.Sprintf("%v", primaryField.Field.Interface())
		} else {
			currentUser = fmt.Sprintf("%v", user)
		}

		if attrs, ok := scope.InstanceGet("gorm:update_attrs"); ok {
			updateAttrs := attrs.(map[string]interface{})
			updateAttrs["updated_by"] = currentUser
			scope.InstanceSet("gorm:update_attrs", updateAttrs)
		} else {
			scope.SetColumn("UpdatedBy", currentUser)
		}
	}
}

func RegisterCallbacks(db *gorm.DB) {
	callback := db.Callback()
	callback.Create().After("gorm:before_create").Register("audited:assign_created_by", assignCreatedBy)
	callback.Update().After("gorm:before_update").Register("audited:assign_updated_by", assignUpdatedBy)
}
