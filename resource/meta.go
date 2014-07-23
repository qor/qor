package resource

import (
	"database/sql"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/rules"

	"os"
	"reflect"
	"regexp"
)

type Meta struct {
	base       *Resource
	Name       string
	Type       string
	Label      string
	Value      interface{}
	GetValue   func(interface{}, *qor.Context) interface{}
	Setter     func(resource interface{}, value interface{}, context *qor.Context)
	Collection []Meta
	Resource   *Resource
	Permission *rules.Permission
}

func (meta *Meta) HasPermission(mode rules.PermissionMode, context *qor.Context) bool {
	if meta.Permission == nil {
		return true
	}
	return meta.Permission.HasPermission(mode, context)
}

func (meta *Meta) updateMeta() {
	var typ = "string"
	if meta.Name == "" {
		fmt.Printf("Meta should have name: %v\n", reflect.ValueOf(meta).Type())
		os.Exit(1)
	}

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
				typ = "bool"
			} else {
				fmt.Printf("Unsupported value type %v for meta: %v\n", reflect.ValueOf(meta.Value).Type(), meta.Name)
				os.Exit(1)
			}
		} else {
			if field, ok := gorm.FieldByName(gorm.SnakeToUpperCamel(meta.Name), meta.base.Model); ok {
				typ = reflect.TypeOf(field).Kind().String()
				meta.Name = gorm.SnakeToUpperCamel(meta.Name)
				meta.GetValue = func(value interface{}, context *qor.Context) interface{} {
					if v, ok := gorm.FieldByName(meta.Name, value, true); ok {
						if typ == "struct" {
							context.DB.Model(value).Related(v)
						}
						return reflect.Indirect(reflect.ValueOf(v)).Interface()
					}
					return ""
				}
			} else {
				fmt.Printf("Unsupported meta name %v for resource: %v\n", meta.Name, reflect.TypeOf(meta.base.Model))
				os.Exit(1)
			}
		}
	}

	// "single_edit", "collection_edit", "select_one", "select_many", "image_with_crop", "table_edit", "table_view"
	if meta.Type == "" {
		if regexp.MustCompile(`^(u)?int(\d+)?`).MatchString(typ) {
			meta.Type = "number"
		} else if typ == "string" {
			meta.Type = "string"
		} else if typ == "bool" {
			meta.Type = "checkbox"
		} else if typ == "struct" {
			meta.Type = "single_edit"
			if meta.Resource == nil {
				if field, ok := gorm.FieldByName(gorm.SnakeToUpperCamel(meta.Name), meta.base.Model); ok {
					result := reflect.New(reflect.Indirect(reflect.ValueOf(field)).Type()).Interface()
					meta.Resource = New(result)
				}
			}
		} else if typ == "slice" {
			meta.Type = "collection_edit"
		}
	}

	if meta.Setter == nil {
		if typ == "slice" {
		} else if typ == "struct" {
		} else {
			meta.Setter = func(resource interface{}, value interface{}, context *qor.Context) {
				field := reflect.Indirect(reflect.ValueOf(resource)).FieldByName(meta.Name)
				if field.IsValid() && field.CanAddr() {
					if scanner, ok := field.Addr().Interface().(sql.Scanner); ok {
						scanner.Scan(value)
					} else if reflect.TypeOf(value).ConvertibleTo(field.Type()) {
						field.Set(reflect.ValueOf(value).Convert(field.Type()))
					}
				} else {
					fmt.Println("Can't set value")
				}
			}
		}
	}

	if meta.Label == "" {
		meta.Label = meta.Name
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
