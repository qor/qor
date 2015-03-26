package admin

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

func (res *Resource) CallFinder(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
	if res.Finder != nil {
		return res.Finder(result, metaValues, context)
	} else {
		var primaryKey string
		if metaValues == nil {
			primaryKey = context.ResourceID
		} else if id := metaValues.Get(res.PrimaryFieldName()); id != nil {
			primaryKey = utils.ToString(id.Value)
		}

		if primaryKey != "" {
			if metaValues != nil {
				if destroy := metaValues.Get("_destroy"); destroy != nil {
					if fmt.Sprintf("%v", destroy.Value) != "0" {
						context.GetDB().Delete(result, primaryKey)
						return resource.ErrProcessorSkipLeft
					}
				}
			}
			return context.GetDB().First(result, primaryKey).Error
		}
		return nil
	}
}

func (res *Resource) CallSearcher(result interface{}, context *qor.Context) error {
	if res.Searcher != nil {
		return res.Searcher(result, context)
	} else {
		return context.GetDB().Set("gorm:order_by_primary_key", "DESC").Find(result).Error
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
