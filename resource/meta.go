package resource

import (
	"github.com/qor/qor"
	"github.com/qor/qor/rules"
)

type Meta struct {
	Name       string
	Type       string
	Label      string
	Value      interface{}
	GetValue   func(interface{}, *qor.Context) string
	Collection []Meta
	Resource   interface{}
	Permission *rules.Permission
}

func (meta *Meta) HasPermission(mode rules.PermissionMode, context *qor.Context) bool {
	if meta.Permission == nil {
		return true
	}
	return meta.Permission.HasPermission(mode, context)
}

type meta struct {
	resource *Resource
	metas    []Meta
}

func (m *meta) Register(meta Meta) {
	m.metas = append(m.metas, meta)

	// data := reflect.Indirect(reflect.ValueOf(value))
	// metaValue := meta.Value
	// if str, ok := metaValue.(string); ok {
	// 	return str
	// } else if f, ok := metaValue.(func() string); ok {
	// 	return f()
	// } else if f, ok := metaValue.(func(*qor.Context) string); ok {
	// 	return f(content.Context)
	// } else if f, ok := metaValue.(func(interface{}) string); ok {
	// 	return f(value)
	// } else if f, ok := metaValue.(func(interface{}, *qor.Context) string); ok {
	// 	return f(value, content.Context)
	// } else if data.Kind() == reflect.Struct {
	// 	if field := data.FieldByName(meta.Name); field.IsValid() {
	// 		return field.Interface()
	// 	}
	// }
}
