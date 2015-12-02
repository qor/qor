package exchange

import (
	"reflect"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type Meta struct {
	base *Resource
	resource.Meta
	Name       string
	Header     string
	Valuer     func(interface{}, *qor.Context) interface{}
	Setter     func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context)
	Permission *roles.Permission
}

func (meta *Meta) GetMetas() []resource.Metaor {
	return []resource.Metaor{}
}

func (meta *Meta) GetResource() resource.Resourcer {
	return nil
}

func (meta *Meta) GetValuer() func(interface{}, *qor.Context) interface{} {
	return func(record interface{}, context *qor.Context) interface{} {
		if valuer := meta.Meta.Valuer; valuer != nil {
			result := valuer(record, context)

			if reflectValue := reflect.ValueOf(result); reflectValue.IsValid() {
				if reflectValue.Kind() == reflect.Ptr {
					if reflectValue.IsNil() || !reflectValue.Elem().IsValid() {
						return nil
					}

					result = reflectValue.Elem().Interface()
				}

				return result
			}
		}
		return nil
	}
}

func (meta *Meta) updateMeta() {
	meta.Meta = resource.Meta{
		Name:          meta.Name,
		FieldName:     meta.FieldName,
		Setter:        meta.Setter,
		Valuer:        meta.Valuer,
		Permission:    meta.Permission,
		ResourceValue: meta.base.Value,
	}

	meta.PreInitialize()
	if meta.FieldStruct != nil {
		if injector, ok := reflect.New(meta.FieldStruct.Struct.Type).Interface().(resource.ConfigureMetaBeforeInitializeInterface); ok {
			injector.ConfigureQorMetaBeforeInitialize(meta)
		}
	}

	meta.Initialize()

	if meta.FieldStruct != nil {
		if injector, ok := reflect.New(meta.FieldStruct.Struct.Type).Interface().(resource.ConfigureMetaInterface); ok {
			injector.ConfigureQorMeta(meta)
		}
	}
}
