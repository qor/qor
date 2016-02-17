package resource

import (
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/roles"
)

type Resourcer interface {
	GetResource() *Resource
	GetMetas([]string) []Metaor
	CallFindMany(interface{}, *qor.Context) error
	CallFindOne(interface{}, *MetaValues, *qor.Context) error
	CallSave(interface{}, *qor.Context) error
	CallDelete(interface{}, *qor.Context) error
	NewSlice() interface{}
	NewStruct() interface{}
}

// ConfigureResourcerBeforeInitializeInterface if a struct implemented this interface, it will be called before everything when create a resource with the struct
type ConfigureResourceBeforeInitializeInterface interface {
	ConfigureQorResourceBeforeInitialize(Resourcer)
}

// ConfigureResourcerInterface if a struct implemented this interface, it will be called after configured by user
type ConfigureResourceInterface interface {
	ConfigureQorResource(Resourcer)
}

type Resource struct {
	Name            string
	Value           interface{}
	FindManyHandler func(interface{}, *qor.Context) error
	FindOneHandler  func(interface{}, *MetaValues, *qor.Context) error
	SaveHandler     func(interface{}, *qor.Context) error
	DeleteHandler   func(interface{}, *qor.Context) error
	Permission      *roles.Permission
	validators      []func(interface{}, *MetaValues, *qor.Context) error
	processors      []func(interface{}, *MetaValues, *qor.Context) error
	primaryField    *gorm.Field
}

func New(value interface{}) *Resource {
	name := reflect.Indirect(reflect.ValueOf(value)).Type().Name()
	res := &Resource{Value: value, Name: name}
	res.FindOneHandler = res.findOneHandler
	res.FindManyHandler = res.findManyHandler
	res.SaveHandler = res.saveHandler
	res.DeleteHandler = res.deleteHandler

	return res
}

func (res *Resource) GetResource() *Resource {
	return res
}

func (res *Resource) AddValidator(fc func(interface{}, *MetaValues, *qor.Context) error) {
	res.validators = append(res.validators, fc)
}

func (res *Resource) AddProcessor(fc func(interface{}, *MetaValues, *qor.Context) error) {
	res.processors = append(res.processors, fc)
}

func (res *Resource) NewStruct() interface{} {
	return reflect.New(reflect.Indirect(reflect.ValueOf(res.Value)).Type()).Interface()
}

func (res *Resource) NewSlice() interface{} {
	sliceType := reflect.SliceOf(reflect.TypeOf(res.Value))
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	return slicePtr.Interface()
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

// PrimaryField return gorm's primary field
func (res *Resource) PrimaryField() *gorm.Field {
	if res.primaryField == nil {
		scope := gorm.Scope{Value: res.Value}
		res.primaryField = scope.PrimaryField()
	}
	return res.primaryField
}

// PrimaryDBName return db column name of the resource's primary field
func (res *Resource) PrimaryDBName() (name string) {
	field := res.PrimaryField()
	if field != nil {
		name = field.DBName
	}
	return
}

// PrimaryFieldName return struct column name of the resource's primary field
func (res *Resource) PrimaryFieldName() (name string) {
	field := res.PrimaryField()
	if field != nil {
		name = field.Name
	}
	return
}
