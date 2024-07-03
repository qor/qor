package resource

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

// CallFindOne call find one method
func (res *Resource) CallFindOne(result interface{}, metaValues *MetaValues, context *qor.Context) error {
	return res.FindOneHandler(result, metaValues, context)
}

// CallFindMany call find many method
func (res *Resource) CallFindMany(result interface{}, context *qor.Context) error {
	return res.FindManyHandler(result, context)
}

// CallSave call save method
func (res *Resource) CallSave(result interface{}, context *qor.Context) error {
	return res.SaveHandler(result, context)
}

// CallDelete call delete method
func (res *Resource) CallDelete(result interface{}, context *qor.Context) error {
	return res.DeleteHandler(result, context)
}

// ToPrimaryQueryParams generate query params based on primary key, multiple primary value are linked with a comma
func (res *Resource) ToPrimaryQueryParams(primaryValue string, context *qor.Context) (string, []interface{}) {
	if primaryValue != "" {
		scope := context.GetDB().NewScope(res.Value)

		// multiple primary fields
		if len(res.PrimaryFields) > 1 {
			if primaryValueStrs := strings.Split(primaryValue, ","); len(primaryValueStrs) == len(res.PrimaryFields) {
				sqls := []string{}
				primaryValues := []interface{}{}
				for idx, field := range res.PrimaryFields {
					sqls = append(sqls, fmt.Sprintf("%v.%v = ?", scope.QuotedTableName(), scope.Quote(field.DBName)))
					primaryValues = append(primaryValues, primaryValueStrs[idx])
				}

				return strings.Join(sqls, " AND "), primaryValues
			}
		}

		// fallback to first configured primary field
		if len(res.PrimaryFields) > 0 {
			dbName := res.PrimaryFields[0].DBName
			if scope.HasColumn("uid") {
				if _, err := strconv.ParseUint(primaryValue, 10, 64); err != nil {
					dbName = "uid"
				}
			}

			return fmt.Sprintf("%v.%v = ?", scope.QuotedTableName(), scope.Quote(dbName)), []interface{}{primaryValue}
		}

		// if no configured primary fields found
		if primaryField := scope.PrimaryField(); primaryField != nil {
			return fmt.Sprintf("%v.%v = ?", scope.QuotedTableName(), scope.Quote(primaryField.DBName)), []interface{}{primaryValue}
		}
	}

	return "", []interface{}{}
}

// ToPrimaryQueryParamsFromMetaValue generate query params based on MetaValues
func (res *Resource) ToPrimaryQueryParamsFromMetaValue(metaValues *MetaValues, context *qor.Context) (string, []interface{}) {
	var (
		sqls          []string
		primaryValues []interface{}
		scope         = context.GetDB().NewScope(res.Value)
	)

	if metaValues != nil {
		for _, field := range res.PrimaryFields {
			if metaField := metaValues.Get(field.Name); metaField != nil {
				sqls = append(sqls, fmt.Sprintf("%v.%v = ?", scope.QuotedTableName(), scope.Quote(field.DBName)))
				primaryValues = append(primaryValues, utils.ToString(metaField.Value))
			}
		}
	}

	return strings.Join(sqls, " AND "), primaryValues
}

func (res *Resource) findOneHandler(result interface{}, metaValues *MetaValues, context *qor.Context) error {
	if res.HasPermission(roles.Read, context) {
		var (
			primaryQuerySQL string
			primaryParams   []interface{}
		)

		if metaValues == nil {
			primaryQuerySQL, primaryParams = res.ToPrimaryQueryParams(context.ResourceID, context)
		} else {
			primaryQuerySQL, primaryParams = res.ToPrimaryQueryParamsFromMetaValue(metaValues, context)
		}

		if primaryQuerySQL != "" {
			if metaValues != nil {
				if destroy := metaValues.Get("_destroy"); destroy != nil {
					if fmt.Sprint(destroy.Value) != "0" && res.HasPermission(roles.Delete, context) {
						context.GetDB().Delete(result, append([]interface{}{primaryQuerySQL}, primaryParams...)...)
						return ErrProcessorSkipLeft
					}
				}
			}
			return context.GetDB().First(result, append([]interface{}{primaryQuerySQL}, primaryParams...)...).Error
		}

		return errors.New("failed to find")
	}
	return roles.ErrPermissionDenied
}

func (res *Resource) findManyHandler(result interface{}, context *qor.Context) error {
	if res.HasPermission(roles.Read, context) {
		db := context.GetDB()
		if _, ok := db.Get("qor:getting_total_count"); ok {
			return context.GetDB().Count(result).Error
		}
		return context.GetDB().Set("gorm:order_by_primary_key", "DESC").Find(result).Error
	}

	return roles.ErrPermissionDenied
}

func (res *Resource) saveHandler(result interface{}, context *qor.Context) error {
	if (context.GetDB().NewScope(result).PrimaryKeyZero() &&
		res.HasPermission(roles.Create, context)) || // has create permission
		res.HasPermission(roles.Update, context) { // has update permission
		return context.GetDB().Save(result).Error
	}
	return roles.ErrPermissionDenied
}

func (res *Resource) deleteHandler(result interface{}, context *qor.Context) error {
	if res.HasPermission(roles.Delete, context) {
		if primaryQuerySQL, primaryParams := res.ToPrimaryQueryParams(context.ResourceID, context); primaryQuerySQL != "" {
			if !context.GetDB().First(result, append([]interface{}{primaryQuerySQL}, primaryParams...)...).RecordNotFound() {
				return context.GetDB().Delete(result).Error
			}
		}
		return gorm.ErrRecordNotFound
	}
	return roles.ErrPermissionDenied
}
