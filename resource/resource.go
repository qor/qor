package resource

import (
	"github.com/qor/qor"
	"github.com/qor/qor/rules"

	"reflect"
)

func New(value interface{}) *Resource {
	Resource{value: value}
}

type Resource struct {
	value      interface{}
	Metas      []Metaor
	Finder     func()
	validators []func()
	processors []func()
}

func (r *Resource) SetFinder() {
}

func (r *Resource) AddValidator() {
}

func (r *Resource) AddProcessor() {
}

func (r *Resource) RegisterMeta(metaor Metaor) {
	r.Metas = append(r.Metas, metaor)
}

func (r *Resource) NewSlice() []interface{} {
	sliceType := reflect.SliceOf(reflect.ValueOf(r.Model).Type())
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	return slicePtr.Interface()
}

func (r *Resource) NewStruct() interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(r.Model)).Type()).Interface()
}

func (r *Resource) Decode(result interface{}, metaDatas MetaDatas) Processor {
}

type Processor struct {
}

func (p *Processor) Validate() []error {
}

func (p *Processor) Commit() []error {
}

type Meta struct {
	Name       string
	Type       string
	Label      string
	Value      func(interface{}, *qor.Context) interface{}
	Setter     func(resource interface{}, value interface{}, context *qor.Context)
	Collection interface{}
	Resource   *Resource
	Permission *rules.Permission
}

func (m *Meta) GetMeta() *Meta {
	return m
}

type Metaor interface {
	GetMeta() *Meta
}

type MetaData struct {
	Name  string
	Value interface{}
	Metaor
}

type MetaDatas []MetaData

func (m MetaDatas) Get(name string) Metaor {
}
