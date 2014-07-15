package resource

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/rules"

	"os"
	"reflect"
)

type Meta struct {
	base       *Resource
	Name       string
	Type       string
	Label      string
	Value      interface{}
	GetValue   func(interface{}, *qor.Context) interface{}
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

func (meta *Meta) updateMeta() {
	if meta.GetValue == nil {
		if meta.Value != nil {
			if f, ok := meta.Value.(func() string); ok {
				meta.GetValue = func(interface{}, *qor.Context) interface{} { return f() }
			} else if f, ok := meta.Value.(func(*qor.Context) string); ok {
				meta.GetValue = func(value interface{}, context *qor.Context) interface{} { return f(context) }
			} else if f, ok := meta.Value.(func(interface{}) string); ok {
				meta.GetValue = func(value interface{}, context *qor.Context) interface{} { return f(value) }
			} else if f, ok := meta.Value.(func(interface{}, *qor.Context) string); ok {
				meta.GetValue = func(value interface{}, context *qor.Context) interface{} { return f(value, context) }
			} else if str, ok := meta.Value.(string); ok {
				meta.GetValue = func(interface{}, *qor.Context) interface{} { return str }
			} else if v, ok := meta.Value.(bool); ok {
				meta.GetValue = func(interface{}, *qor.Context) interface{} { return v }
			} else {
				fmt.Printf("Unsupported value type %v for meta: %v\n", reflect.ValueOf(meta.Value).Type(), meta.Name)
				os.Exit(1)
			}
		} else {
			if _, ok := gorm.FieldByName(meta.Name, meta.base.Model); ok {
				meta.GetValue = func(value interface{}, context *qor.Context) interface{} {
					if v, ok := gorm.FieldByName(meta.Name, value); ok {
						return v
					}
					return ""
				}
			} else {
				fmt.Printf("Unsupported meta name %v for resource: %v\n", meta.Name, reflect.TypeOf(meta.base.Model))
				os.Exit(1)
			}
		}
	}
}

type meta struct {
	resource *Resource
	metas    []Meta
}

func (m *meta) Register(meta Meta) {
	meta.base = m.resource
	meta.updateMeta()
	m.metas = append(m.metas, meta)
}
