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
	Metas      []Metaor
	Finder     func(interface{}, MetaDatas, qor.Context) error
	validators []func(interface{}, MetaDatas, qor.Context) []error
	processors []func(interface{}, MetaDatas, qor.Context) []error
}

func (resource *Resource) SetFinder(fc func(result interface{}, metaDatas MetaDatas, context qor.Context) error) {
	resource.Finder = fc
}

func (resource *Resource) AddValidator(fc func(interface{}, MetaDatas, qor.Context) []error) {
	resource.validators = append(resource.validators, fc)
}

func (resource *Resource) AddProcessor(fc func(interface{}, MetaDatas, qor.Context) []error) {
	resource.processors = append(resource.processors, fc)
}

func (resource *Resource) RegisterMeta(metaor Metaor) {
	meta := metaor.GetMeta()
	meta.updateMeta()
	meta.Base = resource
	resource.Metas = append(resource.Metas, metaor)
}

func (resource *Resource) Decode(result interface{}, metaDatas MetaDatas, context qor.Context) *Processor {
	return &Processor{Resource: resource, Result: result, Context: context}
}

func (resource *Resource) NewSlice() []interface{} {
	sliceType := reflect.SliceOf(reflect.ValueOf(resource.Value).Type())
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	return slicePtr.Interface().([]interface{})
}

func (resource *Resource) NewStruct() interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(resource.Value)).Type()).Interface()
}
