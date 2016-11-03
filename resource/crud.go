package resource

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

func (res *Resource) findOneHandler(result interface{}, metaValues *MetaValues, context *qor.Context) error {
	if res.HasPermission(roles.Read, context) {
		var (
			scope        = context.GetDB().NewScope(res.Value)
			primaryField = res.PrimaryField()
			primaryKey   string
		)

		if metaValues == nil {
			primaryKey = context.ResourceID
		} else if primaryField == nil {
			return nil
		} else if id := metaValues.Get(primaryField.Name); id != nil {
			primaryKey = utils.ToString(id.Value)
		}

		if primaryKey != "" {
			if metaValues != nil {
				if destroy := metaValues.Get("_destroy"); destroy != nil {
					if fmt.Sprint(destroy.Value) != "0" && res.HasPermission(roles.Delete, context) {
						context.GetDB().Delete(result, fmt.Sprintf("%v = ?", scope.Quote(primaryField.DBName)), primaryKey)
						return ErrProcessorSkipLeft
					}
				}
			}
			return context.GetDB().First(result, fmt.Sprintf("%v.%v = ?", scope.QuotedTableName(), scope.Quote(primaryField.DBName)), primaryKey).Error
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
		} else {
			return context.GetDB().Set("gorm:order_by_primary_key", "DESC").Find(result).Error
		}
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
		scope := context.GetDB().NewScope(res.Value)
		if !context.GetDB().First(result, fmt.Sprintf("%v = ?", scope.Quote(res.PrimaryDBName())), context.ResourceID).RecordNotFound() {
			return context.GetDB().Delete(result).Error
		}
		return gorm.ErrRecordNotFound
	}
	return roles.ErrPermissionDenied
}

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
