package resource

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

type Resource struct {
	Name       string
	primaryKey string
	Value      interface{}
	Metas      map[string]Metaor
	Searcher   func(interface{}, *qor.Context) error
	Finder     func(interface{}, *MetaValues, *qor.Context) error
	Saver      func(interface{}, *qor.Context) error
	Deleter    func(interface{}, *qor.Context) error
	validators []func(interface{}, *MetaValues, *qor.Context) error
	processors []func(interface{}, *MetaValues, *qor.Context) error
}

type Resourcer interface {
	GetResource() *Resource
	CallSearcher(interface{}, *qor.Context) error
	CallFinder(interface{}, *MetaValues, *qor.Context) error
	CallSaver(interface{}, *qor.Context) error
	CallDeleter(interface{}, *qor.Context) error
	NewSlice() interface{}
	NewStruct() interface{}
}

// TODO: use a NewNamed method instead of a variant parameter
// would be better and clearer
func New(value interface{}, names ...string) *Resource {
	name := strings.ToLower(reflect.Indirect(reflect.ValueOf(value)).Type().Name())
	for _, n := range names {
		name = n
	}

	return &Resource{Value: value, Name: name}
}

func (res *Resource) GetResource() *Resource {
	return res
}

func (res *Resource) PrimaryKey() string {
	if res.primaryKey == "" {
		scope := gorm.Scope{Value: res.Value}
		res.primaryKey = scope.PrimaryKey()
	}
	return res.primaryKey
}

func (res *Resource) CallSearcher(result interface{}, context *qor.Context) error {
	if res.Searcher != nil {
		return res.Searcher(result, context)
	} else {
		return context.GetDB().Order(fmt.Sprintf("%v DESC", res.PrimaryKey())).Find(result).Error
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

func (res *Resource) AddValidator(fc func(interface{}, *MetaValues, *qor.Context) error) {
	res.validators = append(res.validators, fc)
}

func (res *Resource) AddProcessor(fc func(interface{}, *MetaValues, *qor.Context) error) {
	res.processors = append(res.processors, fc)
}

func (res *Resource) Meta(metaor Metaor) {
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
