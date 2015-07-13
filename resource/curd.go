package resource

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

var DefaultFinder = func(result interface{}, metaValues *MetaValues, context *qor.Context) error {
	if metaValues == nil && context.ResourceID != "" {
		return context.GetDB().First(result, context.ResourceID).Error
	}
	return errors.New("failed to find")
}

var DefaultSearcher = func(result interface{}, context *qor.Context) error {
	if noOrdering, ok := context.GetDB().Get("qor:no_ordering"); ok && noOrdering.(bool) {
		return context.GetDB().Find(result).Error
	} else {
		return context.GetDB().Set("gorm:order_by_primary_key", "DESC").Find(result).Error
	}
}

var DefaultSaver = func(result interface{}, context *qor.Context) error {
	return context.GetDB().Save(result).Error
}

var DefaultDeleter = func(result interface{}, context *qor.Context) error {
	db := context.GetDB().Delete(result, context.ResourceID)
	if db.Error != nil {
		return db.Error
	} else if db.RowsAffected == 0 {
		return gorm.RecordNotFound
	}
	return nil
}

func (res *Resource) CallFindOne(result interface{}, metaValues *MetaValues, context *qor.Context) error {
	return res.FindOneHandler(result, metaValues, context)
}

func (res *Resource) CallFindMany(result interface{}, context *qor.Context) error {
	return res.FindManyHandler(result, context)
}

func (res *Resource) CallSaver(result interface{}, context *qor.Context) error {
	return res.Saver(result, context)
}

func (res *Resource) CallDeleter(result interface{}, context *qor.Context) error {
	return res.Deleter(result, context)
}
