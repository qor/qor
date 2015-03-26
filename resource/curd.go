package resource

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

func (res *Resource) CallFinder(result interface{}, metaValues *MetaValues, context *qor.Context) error {
	if res.Finder != nil {
		return res.Finder(result, metaValues, context)
	} else {
		if metaValues == nil {
			return context.GetDB().First(result, context.ResourceID).Error
		}
		return nil
	}
}

func (res *Resource) CallSearcher(result interface{}, context *qor.Context) error {
	if res.Searcher != nil {
		return res.Searcher(result, context)
	} else {
		return context.GetDB().Order(fmt.Sprintf("%v DESC", res.PrimaryFieldDBName())).Find(result).Error
	}
}

func (res *Resource) CallSaver(result interface{}, context *qor.Context) error {
	if res.Saver != nil {
		return res.Saver(result, context)
	} else {
		return context.GetDB().Save(result).Error
	}
}

func (res *Resource) CallDeleter(result interface{}, context *qor.Context) error {
	if res.Deleter != nil {
		return res.Deleter(result, context)
	} else {
		db := context.GetDB().Delete(result, context.ResourceID)
		if db.Error != nil {
			return db.Error
		} else if db.RowsAffected == 0 {
			return gorm.RecordNotFound
		}
		return nil
	}
}
