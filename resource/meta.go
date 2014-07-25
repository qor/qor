package resource

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/rules"
	"reflect"
	"regexp"
)

type Meta struct {
	base       *Resource
	Name       string
	Type       string
	Label      string
	Value      func(interface{}, *qor.Context) interface{}
	Setter     func(resource interface{}, value interface{}, context *qor.Context)
	Collection func(interface{}, *qor.Context) []interface{}
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
	if meta.Name == "" {
		qor.ExitWithMsg("Meta should have name: %v", reflect.ValueOf(meta).Type())
	}
	var hasColumn bool
	var valueType = "string"

	if field, ok := gorm.FieldByName(gorm.SnakeToUpperCamel(meta.Name), meta.base.Model); ok {
		hasColumn = true
		valueType = reflect.TypeOf(field).Kind().String()
	}

	if meta.Value == nil {
		if hasColumn {
			meta.Name = gorm.SnakeToUpperCamel(meta.Name)
			meta.Value = func(value interface{}, context *qor.Context) interface{} {
				if v, ok := gorm.FieldByName(meta.Name, value, true); ok {
					if valueType == "struct" || valueType == "slice" {
						context.DB.Model(value).Related(v)
					}
					return reflect.Indirect(reflect.ValueOf(v)).Interface()
				}

				// qor.ExitWithMsg("Can't get value from meta: %v.%v", meta.base.Name, meta.Name)
				return ""
			}
		} else {
			qor.ExitWithMsg("Unsupported meta name %v for resource %v", meta.Name, reflect.TypeOf(meta.base.Model))
		}
	}

	// "select_one", "select_many", "image_with_crop", "table_edit", "table_view"
	// []string, [][]string, []struct, .Resource -> Setter
	if meta.Type == "" {
		switch valueType {
		case "string":
			meta.Type = "string"
		case "bool":
			meta.Type = "checkbox"
		case "struct":
			meta.Type = "single_edit"
			if meta.Resource == nil {
				if field, ok := gorm.FieldByName(gorm.SnakeToUpperCamel(meta.Name), meta.base.Model); ok {
					result := reflect.New(reflect.Indirect(reflect.ValueOf(field)).Type()).Interface()
					meta.Resource = New(result)
				}
			}
		case "slice":
			meta.Type = "collection_edit"
			if field, ok := gorm.FieldByName(gorm.SnakeToUpperCamel(meta.Name), meta.base.Model); ok {
				result := reflect.New(reflect.Indirect(reflect.ValueOf(field)).Type().Elem()).Interface()
				meta.Resource = New(result)
			}
		default:
			if regexp.MustCompile(`^(u)?int(\d+)?`).MatchString(valueType) {
				meta.Type = "number"
			} else {
				qor.ExitWithMsg("Unsupported value type %v for meta %v", meta.Type, reflect.ValueOf(meta).Type())
			}
		}
	}

	if meta.Setter == nil {
		meta.Setter = func(resource interface{}, value interface{}, context *qor.Context) {
			field := reflect.Indirect(reflect.ValueOf(resource)).FieldByName(meta.Name)
			if field.IsValid() && field.CanAddr() {
				if scanner, ok := field.Addr().Interface().(sql.Scanner); ok {
					scanner.Scan(value)
				} else if reflect.TypeOf(value).ConvertibleTo(field.Type()) {
					field.Set(reflect.ValueOf(value).Convert(field.Type()))
				} else {
					if str, ok := value.(string); ok {
						switch field.Kind() {
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
							value, _ = strconv.Atoi(str)
							field.SetInt(reflect.ValueOf(value).Int())
						case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							value, _ = strconv.Atoi(str)
							field.SetUint(reflect.ValueOf(value).Uint())
						default:
							fmt.Println("Can't set value", meta, meta.base)
						}
					} else {
						fmt.Println("Can't set value", meta, meta.base)
					}
				}
			} else {
				fmt.Println("Can't set value", meta, meta.base)
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
