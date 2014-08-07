package resource

import (
	"github.com/qor/qor"

	"reflect"
)

func New(value interface{}) *Resource {
	return &Resource{Value: value}
}

type Resource struct {
	Value      interface{}
	Metas      map[string]Metaor
	Finder     func(interface{}, MetaDatas, qor.Context) error
	validators []func(interface{}, MetaDatas, qor.Context) []error
	processors []func(interface{}, MetaDatas, qor.Context) []error
}

type Resourcer interface {
	GetResource() *Resource
}

func (res *Resource) GetResource() *Resource {
	return res
}

func (res *Resource) SetFinder(fc func(result interface{}, metaDatas MetaDatas, context qor.Context) error) {
	res.Finder = fc
}

func (res *Resource) AddValidator(fc func(interface{}, MetaDatas, qor.Context) []error) {
	res.validators = append(res.validators, fc)
}

func (res *Resource) AddProcessor(fc func(interface{}, MetaDatas, qor.Context) []error) {
	res.processors = append(res.processors, fc)
}

func (res *Resource) RegisterMeta(metaor Metaor) {
	meta := metaor.GetMeta()
	meta.UpdateMeta()
	meta.Base = res
	res.Metas[meta.Name] = metaor
}

func (res *Resource) Decode(result interface{}, metaDatas MetaDatas, context qor.Context) *Processor {
	return &Processor{Resource: res, Result: result, Context: context}
}

func (res *Resource) NewSlice() []interface{} {
	sliceType := reflect.SliceOf(reflect.ValueOf(res.Value).Type())
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	return slicePtr.Interface().([]interface{})
}

func (res *Resource) NewStruct() interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(res.Value)).Type()).Interface()
}
