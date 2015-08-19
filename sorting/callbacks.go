package sorting

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/jinzhu/gorm"
)

func initalizePosition(scope *gorm.Scope) {
	if !scope.HasError() {
		if position, ok := scope.Value.(sortingInterface); ok {
			if pos, err := strconv.Atoi(fmt.Sprintf("%v", scope.PrimaryKeyValue())); err == nil {
				if scope.DB().UpdateColumn("position", pos).Error == nil {
					position.SetPosition(pos)
				}
			}
		}
	}
}

func modelValue(value interface{}) interface{} {
	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	typ := reflectValue.Type()

	if reflectValue.Kind() == reflect.Slice {
		typ = reflectValue.Type().Elem()
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
	}

	return reflect.New(typ).Interface()
}

func beforeQuery(scope *gorm.Scope) {
	modelValue := modelValue(scope.Value)
	if _, ok := modelValue.(sortingDescInterface); ok {
		scope.Search.Order("position desc")
	} else if _, ok := modelValue.(sortingInterface); ok {
		scope.Search.Order("position")
	}
}

func RegisterCallbacks(db *gorm.DB) {
	db.Callback().Query().Before("gorm:query").Register("sorting:sort_by_position", beforeQuery)

	db.Callback().Create().Before("gorm:commit_or_rollback_transaction").
		Register("sorting:initalize_position", initalizePosition)
}
