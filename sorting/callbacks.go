package sorting

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

func initalizePosition(scope *gorm.Scope) {
	if !scope.HasError() {
		if _, ok := scope.Value.(sortingInterface); ok {
			var lastPosition int
			scope.NewDB().Model(modelValue(scope.Value)).Select("position").Order("position DESC").Limit(1).Row().Scan(&lastPosition)
			scope.SetColumn("Position", lastPosition+1)
		}
	}
}

func reorderPositions(scope *gorm.Scope) {
	if !scope.HasError() {
		if _, ok := scope.Value.(sortingInterface); ok {
			table := scope.TableName()
			var sql string
			if scope.HasColumn("DeletedAt") {
				sql = fmt.Sprintf("UPDATE %v SET position = (SELECT COUNT(pos) + 1 FROM (SELECT DISTINCT(position) AS pos FROM %v WHERE deleted_at IS NULL) AS t2 WHERE t2.pos < %v.position)", table, table, table)
			} else {
				sql = fmt.Sprintf("UPDATE %v SET position = (SELECT COUNT(pos) + 1 FROM (SELECT DISTINCT(position) AS pos FROM %v) AS t2 WHERE t2.pos < %v.position)", table, table, table)
			}
			scope.NewDB().Exec(sql)
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
	db.Callback().Create().Before("gorm:create").Register("sorting:initalize_position", initalizePosition)
	db.Callback().Delete().After("gorm:after_delete").Register("sorting:reorder_positions", reorderPositions)
	db.Callback().Query().Before("gorm:query").Register("sorting:sort_by_position", beforeQuery)
}
