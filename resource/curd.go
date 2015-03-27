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
	return context.GetDB().Set("gorm:order_by_primary_key", "DESC").Find(result).Error
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

func (res *Resource) CallFinder(result interface{}, metaValues *MetaValues, context *qor.Context) error {
	return res.Finder(result, metaValues, context)
}

func (res *Resource) CallSearcher(result interface{}, context *qor.Context) error {
	return res.Searcher(result, context)
}

func (res *Resource) CallSaver(result interface{}, context *qor.Context) error {
	return res.Saver(result, context)
}

func (res *Resource) CallDeleter(result interface{}, context *qor.Context) error {
	return res.Deleter(result, context)
}
