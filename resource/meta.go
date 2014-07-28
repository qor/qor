package resource

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/rules"
	"reflect"
	"regexp"
)

type Meta struct {
	base          *Resource
	Name          string
	Type          string
	Label         string
	Value         func(interface{}, *qor.Context) interface{}
	Setter        func(resource interface{}, value interface{}, context *qor.Context)
	Collection    interface{}
	GetCollection func(interface{}, *qor.Context) [][]string
	Resource      *Resource
	Permission    *rules.Permission
}

func (meta *Meta) HasPermission(mode rules.PermissionMode, context *qor.Context) bool {
	if meta.Permission == nil {
		return true
	}
	return meta.Permission.HasPermission(mode, context)
}

func (meta *Meta) updateMeta() {
	var hasColumn bool
	var valueType string

	if meta.Name == "" {
		qor.ExitWithMsg("Meta should have name: %v", reflect.ValueOf(meta).Type())
	}

	if field, ok := gorm.FieldByName(gorm.SnakeToUpperCamel(meta.Name), meta.base.Model); ok {
		hasColumn = true
		valueType = reflect.TypeOf(field).Kind().String()
	}

	// "select_one", "select_many", "image_with_crop", "table_edit", "table_view"
	// []string, [][]string, []struct, .Resource -> Setter

	// Set Meta Type
	if meta.Type == "" {
		switch valueType {
		case "string":
			meta.Type = "string"
		case "bool":
			meta.Type = "checkbox"
		case "struct":
			meta.Type = "single_edit"
		case "slice":
			meta.Type = "collection_edit"
		default:
			if regexp.MustCompile(`^(u)?int(\d+)?`).MatchString(valueType) {
				meta.Type = "number"
			} else {
				qor.ExitWithMsg("Unsupported value type %v for meta %v", meta.Type, reflect.ValueOf(meta).Type())
			}
		}
	}

	// Set Meta Resource
	if meta.Resource == nil {
		if hasColumn && (valueType == "struct" || valueType == "slice") {
			if field, ok := gorm.FieldByName(gorm.SnakeToUpperCamel(meta.Name), meta.base.Model); ok {
				var result interface{}
				if valueType == "struct" {
					result = reflect.New(reflect.Indirect(reflect.ValueOf(field)).Type()).Interface()
				} else if valueType == "slice" {
					result = reflect.New(reflect.Indirect(reflect.ValueOf(field)).Type().Elem()).Interface()
				}
				meta.Resource = New(result)
			}
		}
	}

	// Set Meta Value
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
				return ""
			}
		} else {
			qor.ExitWithMsg("Unsupported meta name %v for resource %v", meta.Name, reflect.TypeOf(meta.base.Model))
		}
	}

	// Set Meta Collection
	switch meta.Type {
	case "select_one":
		if meta.Collection != nil {
			if maps, ok := meta.Collection.([]string); ok {
				meta.GetCollection = func(interface{}, *qor.Context) [][]string {
					var results = [][]string{}
					for _, value := range maps {
						results = append(results, []string{value, value})
					}
					return results
				}
			} else if maps, ok := meta.Collection.([][]string); ok {
				meta.GetCollection = func(interface{}, *qor.Context) [][]string {
					return maps
				}
			} else {
				qor.ExitWithMsg("Unsupported Collection format for meta %v of resource %v", meta.Name, reflect.TypeOf(meta.base.Model))
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
							qor.ExitWithMsg("Can't set value", meta, meta.base)
						}
					} else {
						qor.ExitWithMsg("Can't set value", meta, meta.base)
					}
				}
			} else if !strings.HasPrefix(meta.Name, "_") {
				qor.ExitWithMsg("Can't set value", meta, meta.base)
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
