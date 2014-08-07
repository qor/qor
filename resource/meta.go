package resource

import (
	"database/sql"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/media_library"
	"github.com/qor/qor/rules"
	"reflect"
	"regexp"
)

type Meta struct {
	Base          Resourcer
	Name          string
	Type          string
	Label         string
	Value         func(interface{}, *qor.Context) interface{}
	Setter        func(resource interface{}, value interface{}, context *qor.Context)
	Collection    interface{}
	GetCollection func(interface{}, *qor.Context) [][]string
	Resource      Resourcer
	Permission    *rules.Permission
}

type Metaor interface {
	GetMeta() *Meta
}

func (meta *Meta) GetMeta() *Meta {
	return meta
}

func (meta *Meta) HasPermission(mode rules.PermissionMode, context *qor.Context) bool {
	if meta.Permission == nil {
		return true
	}
	return meta.Permission.HasPermission(mode, context)
}

func (meta *Meta) UpdateMeta() {
	var hasColumn bool
	var valueType string

	if meta.Name == "" {
		qor.ExitWithMsg("Meta should have name: %v", reflect.ValueOf(meta).Type())
	}

	base := meta.Base.GetResource()
	scope := &gorm.Scope{Value: base.Value}
	var field *gorm.Field
	field, hasColumn = scope.FieldByName(meta.Name)
	valueType = reflect.TypeOf(field.Value).Kind().String()

	// Set Meta Type
	if meta.Type == "" {
		if relationship := field.Relationship; relationship != nil {
			if relationship.Kind == "belongs_to" || relationship.Kind == "has_one" {
				meta.Type = "single_edit"
			} else if relationship.Kind == "has_many" {
				meta.Type = "collection_edit"
			} else if relationship.Kind == "many_to_many" {
				meta.Type = "select_many"
			}
		} else {
			switch valueType {
			case "string":
				meta.Type = "string"
			case "bool":
				meta.Type = "checkbox"
			default:
				if regexp.MustCompile(`^(u)?int(\d+)?`).MatchString(valueType) {
					meta.Type = "number"
				} else if _, ok := field.Value.(media_library.MediaLibrary); ok {
					meta.Type = "file"
				} else {
					qor.ExitWithMsg("Unsupported value type %v for meta %v", valueType, meta.Name)
				}
			}
		}
	}

	// Set Meta Resource
	if meta.Resource == nil {
		if hasColumn && (field.Relationship != nil) {
			var result interface{}
			if valueType == "struct" {
				result = reflect.New(reflect.Indirect(reflect.ValueOf(field.Value)).Type()).Interface()
			} else if valueType == "slice" {
				result = reflect.New(reflect.Indirect(reflect.ValueOf(field.Value)).Type().Elem()).Interface()
			}
			meta.Resource = New(result)
		}
	}

	// Set Meta Value
	if meta.Value == nil {
		if hasColumn {
			meta.Name = gorm.SnakeToUpperCamel(meta.Name)
			meta.Value = func(value interface{}, context *qor.Context) interface{} {
				scope := &gorm.Scope{Value: value}
				if f, ok := scope.FieldByName(meta.Name); ok {
					if field.Relationship != nil {
						if !reflect.ValueOf(f.Value).CanAddr() {
							if reflect.ValueOf(f.Value).Kind() == reflect.Slice {
								sliceType := reflect.ValueOf(f.Value).Type()
								slice := reflect.MakeSlice(sliceType, 0, 0)
								slicePtr := reflect.New(sliceType)
								slicePtr.Elem().Set(slice)
								f.Value = slicePtr.Interface()
							} else if reflect.ValueOf(f.Value).Kind() == reflect.Struct {
								f.Value = reflect.New(reflect.Indirect(reflect.ValueOf(f.Value)).Type()).Interface()
							}
						}

						context.DB.Model(value).Related(f.Value, meta.Name)
					}
					return f.Value
				}
				return ""
			}
		} else {
			qor.ExitWithMsg("Unsupported meta name %v for resource %v", meta.Name, reflect.TypeOf(base.Value))
		}
	}

	// Set Meta Collection
	if meta.Collection != nil {
		if maps, ok := meta.Collection.([]string); ok {
			meta.GetCollection = func(interface{}, *qor.Context) (results [][]string) {
				for _, value := range maps {
					results = append(results, []string{value, value})
				}
				return
			}
		} else if maps, ok := meta.Collection.([][]string); ok {
			meta.GetCollection = func(interface{}, *qor.Context) [][]string {
				return maps
			}
		} else if f, ok := meta.Collection.(func(interface{}, *qor.Context) [][]string); ok {
			meta.GetCollection = f
		} else {
			qor.ExitWithMsg("Unsupported Collection format for meta %v of resource %v", meta.Name, reflect.TypeOf(base.Value))
		}
	} else if meta.Type == "select_one" || meta.Type == "select_many" {
		qor.ExitWithMsg("%v meta type %v needs Collection", meta.Name, meta.Type)
	}

	if meta.Setter == nil {
		meta.Setter = func(resource interface{}, value interface{}, context *qor.Context) {
			scope := &gorm.Scope{Value: resource}
			scopeField, _ := scope.FieldByName(meta.Name)
			field := reflect.Indirect(reflect.ValueOf(resource)).FieldByName(meta.Name)
			// fieldStruct, _ := reflect.Indirect(reflect.ValueOf(resource)).Type().FieldByName(meta.Name)

			if field.IsValid() && field.CanAddr() {
				if values, ok := context.Request.Form[value.(string)]; ok {
					relationship := scopeField.Relationship
					if relationship != nil && relationship.Kind == "many_to_many" {
						context.DB.Where(values).Find(field.Addr().Interface())
						context.DB.Model(resource).Where(values).Association(meta.Name).Replace(field.Interface())
					} else {
						switch field.Kind() {
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
							if value, err := strconv.Atoi(values[0]); err == nil {
								field.SetInt(reflect.ValueOf(value).Int())
							} else {
								qor.ExitWithMsg("Can't set value", meta, meta.Base)
							}
						case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							if value, err := strconv.Atoi(values[0]); err == nil {
								field.SetUint(reflect.ValueOf(value).Uint())
							} else {
								qor.ExitWithMsg("Can't set value", meta, meta.Base)
							}
						default:
							if scanner, ok := field.Addr().Interface().(sql.Scanner); ok {
								scanner.Scan(values[0])
							} else if reflect.TypeOf(values).ConvertibleTo(field.Type()) {
								field.Set(reflect.ValueOf(values).Convert(field.Type()))
							} else if len(values) == 1 && reflect.TypeOf(values[0]).ConvertibleTo(field.Type()) {
								field.Set(reflect.ValueOf(values[0]).Convert(field.Type()))
							} else {
								qor.ExitWithMsg("Can't set value", meta, meta.Base)
							}
						}
					}
				} else if context.Request.MultipartForm != nil {
					if headers, ok := context.Request.MultipartForm.File[value.(string)]; ok {
						for _, header := range headers {
							if media, ok := field.Interface().(media_library.MediaLibrary); ok {
								if file, err := header.Open(); err == nil {
									media.SetFile(header.Filename, file)
								}
							}
						}
					}
				} else {
					qor.ExitWithMsg("Can't set value", meta, meta.Base)
				}
			}
		}
	}

	if meta.Label == "" {
		meta.Label = meta.Name
	}
}
