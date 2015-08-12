package audited

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

type auditableInterface interface {
	SetCreatedBy(createdBy interface{})
	GetCreatedBy() string
	SetUpdatedBy(updatedBy interface{})
	GetUpdatedBy() string
}

func isAuditable(scope *gorm.Scope) (isAuditable bool) {
	if scope.GetModelStruct().ModelType == nil {
		return false
	}
	_, isAuditable = reflect.New(scope.GetModelStruct().ModelType).Interface().(auditableInterface)
	return
}

func assignCreatedBy(scope *gorm.Scope) {
	if isAuditable(scope) {
		if user, ok := scope.DB().Get("audited:current_user"); ok {
			var currentUser string
			if primaryField := scope.New(user).PrimaryField(); primaryField != nil {
				currentUser = fmt.Sprintf("%v", primaryField.Field.Interface())
			} else {
				currentUser = fmt.Sprintf("%v", user)
			}

			scope.SetColumn("CreatedBy", currentUser)
		}
	}
}

func assignUpdatedBy(scope *gorm.Scope) {
	if isAuditable(scope) {
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
}

func RegisterCallbacks(db *gorm.DB) {
	callback := db.Callback()
	callback.Create().After("gorm:before_create").Register("audited:assign_created_by", assignCreatedBy)
	callback.Update().After("gorm:before_update").Register("audited:assign_updated_by", assignUpdatedBy)
}
