package admin

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/rules"

	"reflect"
)

type Content struct {
	Admin    *Admin
	Context  *qor.Context
	Resource *resource.Resource
	Result   interface{}
	Action   string
}

func (content *Content) Metas() []resource.Meta {
	return content.Resource.IndexAttrs()
}

func (content *Content) HasPermission(mode rules.PermissionMode, meta resource.Meta) bool {
	return meta.Permission.HasPermission(mode, content.Context)
}

func (content *Content) ValueOf(meta resource.Meta, value interface{}) interface{} {
	data := reflect.Indirect(reflect.ValueOf(value))
	metaValue := meta.Value

	if str, ok := metaValue.(string); ok {
		return str
	} else if f, ok := metaValue.(func() string); ok {
		return f()
	} else if f, ok := metaValue.(func(*qor.Context) string); ok {
		return f(content.Context)
	} else if f, ok := metaValue.(func(interface{}) string); ok {
		return f(value)
	} else if f, ok := metaValue.(func(interface{}, *qor.Context) string); ok {
		return f(value, content.Context)
	} else if data.Kind() == reflect.Struct {
		if field := data.FieldByName(meta.Name); field.IsValid() {
			return field.Interface()
		}
	}
	return ""
}
