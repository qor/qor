package resource

import (
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/roles"
)

type Resource struct {
	Name            string
	StructType      string
	Permission      *roles.Permission
	primaryField    *gorm.Field
	Value           interface{}
	FindManyHandler func(interface{}, *qor.Context) error
	FindOneHandler  func(interface{}, *MetaValues, *qor.Context) error
	Saver           func(interface{}, *qor.Context) error
	Deleter         func(interface{}, *qor.Context) error
	validators      []func(interface{}, *MetaValues, *qor.Context) error
	processors      []func(interface{}, *MetaValues, *qor.Context) error
}

type Resourcer interface {
	GetResource() *Resource
	GetMetas([]string) []Metaor
	CallFindMany(interface{}, *qor.Context) error
	CallFindOne(interface{}, *MetaValues, *qor.Context) error
	CallSaver(interface{}, *qor.Context) error
	CallDeleter(interface{}, *qor.Context) error
	NewSlice() interface{}
	NewStruct() interface{}
}

func New(value interface{}) *Resource {
	structType := reflect.Indirect(reflect.ValueOf(value)).Type()
	typeName := structType.String()
	name := structType.Name()

	res := &Resource{Value: value, Name: name, StructType: typeName, Saver: DefaultSaver, FindManyHandler: DefaultSearcher, Deleter: DefaultDeleter}
	res.FindOneHandler = res.finder

	return res
}

func (res *Resource) GetResource() *Resource {
	return res
}

func (res *Resource) PrimaryField() *gorm.Field {
	if res.primaryField == nil {
		scope := gorm.Scope{Value: res.Value}
		res.primaryField = scope.PrimaryField()
	}
	return res.primaryField
}

func (res *Resource) PrimaryDBName() (name string) {
	field := res.PrimaryField()
	if field != nil {
		name = field.DBName
	}
	return
}

func (res *Resource) PrimaryFieldName() (name string) {
	field := res.PrimaryField()
	if field != nil {
		name = field.Name
	}
	return
}

func (res *Resource) AddValidator(fc func(interface{}, *MetaValues, *qor.Context) error) {
	res.validators = append(res.validators, fc)
}

func (res *Resource) AddProcessor(fc func(interface{}, *MetaValues, *qor.Context) error) {
	res.processors = append(res.processors, fc)
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

func (res *Resource) GetMetas([]string) []Metaor {
	panic("not defined")
}

func (res *Resource) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if res == nil || res.Permission == nil {
		return true
	}
	return res.Permission.HasPermission(mode, context.Roles...)
}
