package resource

import (
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

type Resource struct {
	Value      interface{}
	Metas      map[string]Metaor
	Searcher   func(*qor.Context) (interface{}, error)
	Finder     func(*MetaValues, *qor.Context) (interface{}, error)
	Saver      func(interface{}, *qor.Context) error
	Deleter    func(interface{}, *qor.Context) error
	validators []func(interface{}, *MetaValues, *qor.Context) error
	processors []func(interface{}, *MetaValues, *qor.Context) error
}

type Resourcer interface {
	GetResource() *Resource
	CallSearcher(*qor.Context) (interface{}, error)
	CallFinder(*MetaValues, *qor.Context) (interface{}, error)
	CallSaver(interface{}, *qor.Context) error
	CallDeleter(interface{}, *qor.Context) error
	NewSlice() interface{}
	NewStruct() interface{}
}

func New(value interface{}) *Resource {
	return &Resource{Value: value}
}

func (res *Resource) GetResource() *Resource {
	return res
}

func (res *Resource) CallSearcher(context *qor.Context) (interface{}, error) {
	result := res.NewSlice()
	if res.Searcher != nil {
		return res.Searcher(context)
	} else {
		err := context.DB().Find(result).Error
		return result, err
	}
}

func (res *Resource) CallSaver(result interface{}, context *qor.Context) error {
	if res.Saver != nil {
		return res.Saver(result, context)
	} else {
		return context.DB().Save(result).Error
	}
}

func (res *Resource) CallDeleter(result interface{}, context *qor.Context) error {
	if res.Deleter != nil {
		return res.Deleter(result, context)
	} else {
		db := context.DB().Delete(result, context.ResourceID)
		if db.Error != nil {
			return db.Error
		} else if db.RowsAffected == 0 {
			return gorm.RecordNotFound
		}
		return nil
	}
}

func (res *Resource) CallFinder(metaValues *MetaValues, context *qor.Context) (result interface{}, err error) {
	if res.Finder != nil {
		return res.Finder(metaValues, context)
	} else {
		result := res.NewStruct()
		if metaValues == nil {
			err = context.DB().First(result, context.ResourceID).Error
		}
		return result, err
	}
}

func (res *Resource) AddValidator(fc func(interface{}, *MetaValues, *qor.Context) error) {
	res.validators = append(res.validators, fc)
}

func (res *Resource) AddProcessor(fc func(interface{}, *MetaValues, *qor.Context) error) {
	res.processors = append(res.processors, fc)
}

func (res *Resource) RegisterMeta(metaor Metaor) {
	if res.Metas == nil {
		res.Metas = make(map[string]Metaor)
	}

	meta := metaor.GetMeta()
	meta.Base = res
	meta.UpdateMeta()
	res.Metas[meta.Name] = metaor
}

func (res *Resource) NewSlice() interface{} {
	sliceType := reflect.SliceOf(reflect.ValueOf(res.Value).Type())
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	return slicePtr.Interface()
}

func (res *Resource) NewStruct() interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(res.Value)).Type()).Interface()
}
