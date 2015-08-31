package sorting

import (
	"reflect"

	"github.com/jinzhu/gorm"
)

func initalizePosition(scope *gorm.Scope) {
	if !scope.HasError() {
		if _, ok := scope.Value.(sortingInterface); ok {
			var lastPosition int
			scope.NewDB().Table(scope.TableName()).Select("position").Order("position DESC").Limit(1).Row().Scan(&lastPosition)
			scope.SetColumn("Position", lastPosition+1)
		}
	}
}

func modelValue(value interface{}) interface{} {
	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	if reflectValue.IsValid() {
		typ := reflectValue.Type()

		if reflectValue.Kind() == reflect.Slice {
			typ = reflectValue.Type().Elem()
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
		}

		return reflect.New(typ).Interface()
	} else {
		return nil
	}
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

	db.Callback().Create().Before("gorm:create").
		Register("sorting:initalize_position", initalizePosition)
}
