package resource

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
)

func (res *Resource) findOneHandler(result interface{}, metaValues *MetaValues, context *qor.Context) error {
	var primaryKey string
	if metaValues == nil {
		primaryKey = context.ResourceID
	} else if id := metaValues.Get(res.PrimaryFieldName()); id != nil {
		primaryKey = utils.ToString(id.Value)
	}

	if primaryKey != "" {
		if metaValues != nil {
			if destroy := metaValues.Get("_destroy"); destroy != nil {
				if fmt.Sprint(destroy.Value) != "0" {
					context.GetDB().Delete(result, primaryKey)
					return ErrProcessorSkipLeft
				}
			}
		}
		return context.GetDB().First(result, primaryKey).Error
	}
	return errors.New("failed to find")
}

func (res *Resource) findManyHandler(result interface{}, context *qor.Context) error {
	return context.GetDB().Set("gorm:order_by_primary_key", "DESC").Find(result).Error
}

func (res *Resource) saveHandler(result interface{}, context *qor.Context) error {
	return context.GetDB().Save(result).Error
}

func (res *Resource) deleteHandler(result interface{}, context *qor.Context) error {
	if !context.GetDB().First(result, context.ResourceID).RecordNotFound() {
		return context.GetDB().Delete(result).Error
	} else {
		return gorm.RecordNotFound
	}
}

func (res *Resource) CallFindOne(result interface{}, metaValues *MetaValues, context *qor.Context) error {
	return res.FindOneHandler(result, metaValues, context)
}

func (res *Resource) CallFindMany(result interface{}, context *qor.Context) error {
	return res.FindManyHandler(result, context)
}

func (res *Resource) CallSaver(result interface{}, context *qor.Context) error {
	return res.SaveHandler(result, context)
}

func (res *Resource) CallDeleter(result interface{}, context *qor.Context) error {
	return res.DeleteHandler(result, context)
}
