package admin

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	"github.com/qor/qor"
	"github.com/qor/qor/media_library"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
	"github.com/qor/qor/utils"
)

type Meta struct {
	base          *Resource
	Name          string
	DBName        string
	FieldName     string
	Label         string
	Type          string
	Valuer        func(interface{}, *qor.Context) interface{}
	Setter        func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context)
	Metas         []resource.Metaor
	Resource      resource.Resourcer
	Collection    interface{}
	GetCollection func(interface{}, *qor.Context) [][]string
	Permission    *roles.Permission
}

func (meta *Meta) GetMetas() []resource.Metaor {
	if len(meta.Metas) > 0 {
		return meta.Metas
	} else if meta.Resource == nil {
		return []resource.Metaor{}
	} else {
		return meta.Resource.GetMetas([]string{})
	}
}

func (meta *Meta) GetResource() resource.Resourcer {
	return meta.Resource
}

func getField(fields []*gorm.StructField, name string) (*gorm.StructField, bool) {
	for _, field := range fields {
		if field.Name == name || field.DBName == name {
			return field, true
		}
	}
	return nil, false
}

func (meta *Meta) updateMeta() {
	if meta.Name == "" {
		utils.ExitWithMsg("Meta should have name: %v", reflect.ValueOf(meta).Type())
	} else if meta.FieldName == "" {
		meta.FieldName = meta.Name
	}

	if meta.Label == "" {
		meta.Label = utils.HumanizeString(meta.Name)
	}

	var (
		scope       = &gorm.Scope{Value: meta.base.Value}
		nestedField = strings.Contains(meta.FieldName, ".")
		field       *gorm.StructField
		hasColumn   bool
	)

	if nestedField {
		subModel, name := utils.ParseNestedField(reflect.ValueOf(meta.base.Value), meta.FieldName)
		subScope := &gorm.Scope{Value: subModel.Interface()}
		field, hasColumn = getField(subScope.GetStructFields(), name)
	} else {
		if field, hasColumn = getField(scope.GetStructFields(), meta.FieldName); hasColumn {
			meta.FieldName = field.Name
			if field.IsNormal {
				meta.DBName = field.DBName
			}
		}
	}

	var fieldType reflect.Type
	if hasColumn {
		fieldType = field.Struct.Type
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
	}

	// Set Meta Type
	if meta.Type == "" && hasColumn {
		if relationship := field.Relationship; relationship != nil {
			if relationship.Kind == "has_one" {
				meta.Type = "single_edit"
			} else if relationship.Kind == "has_many" {
				meta.Type = "collection_edit"
			} else if relationship.Kind == "belongs_to" {
				meta.Type = "select_one"
			} else if relationship.Kind == "many_to_many" {
				meta.Type = "select_many"
			}
		} else {
			switch fieldType.Kind().String() {
			case "string":
				if size, ok := utils.ParseTagOption(field.Tag.Get("sql"))["SIZE"]; ok {
					if i, _ := strconv.Atoi(size); i > 255 {
						meta.Type = "text"
					} else {
						meta.Type = "string"
					}
				} else if text, ok := utils.ParseTagOption(field.Tag.Get("sql"))["TYPE"]; ok && text == "text" {
					meta.Type = "text"
				} else {
					meta.Type = "string"
				}
			case "bool":
				meta.Type = "checkbox"
			default:
				if regexp.MustCompile(`^(.*)?(u)?(int)(\d+)?`).MatchString(fieldType.Kind().String()) {
					meta.Type = "number"
				} else if regexp.MustCompile(`^(.*)?(float)(\d+)?`).MatchString(fieldType.Kind().String()) {
					meta.Type = "float"
				} else if _, ok := reflect.New(fieldType).Interface().(*time.Time); ok {
					meta.Type = "datetime"
				} else if _, ok := reflect.New(fieldType).Interface().(media_library.MediaLibrary); ok {
					meta.Type = "file"
				}
			}
		}
	}

	// Set Meta Resource
	if meta.Resource == nil {
		if hasColumn && (field.Relationship != nil) {
			var result interface{}
			if fieldType.Kind().String() == "struct" {
				result = reflect.New(fieldType).Interface()
			} else if fieldType.Kind().String() == "slice" {
				refelectType := fieldType.Elem()
				for refelectType.Kind() == reflect.Ptr {
					refelectType = refelectType.Elem()
				}
				result = reflect.New(refelectType).Interface()
			}

			res := meta.base.GetAdmin().NewResource(result)
			res.compile()
			meta.Resource = res
		}
	}

	// Set Meta Valuer
	if meta.Valuer == nil {
		if hasColumn {
			meta.Valuer = func(value interface{}, context *qor.Context) interface{} {
				scope := context.GetDB().NewScope(value)
				fieldName := meta.FieldName
				if nestedField {
					fields := strings.Split(fieldName, ".")
					fieldName = fields[len(fields)-1]
				}

				if f, ok := scope.FieldByName(fieldName); ok {
					if field.Relationship != nil {
						if f.Field.CanAddr() && !scope.PrimaryKeyZero() {
							context.GetDB().Model(value).Related(f.Field.Addr().Interface(), meta.FieldName)
						}
					}
					if f.Field.CanAddr() {
						return f.Field.Addr().Interface()
					} else {
						return f.Field.Interface()
					}
				}

				return ""
			}
		} else {
			utils.ExitWithMsg("Unsupported meta name %v for resource %v", meta.Name, reflect.TypeOf(meta.base.Value))
		}
	}

	scopeField, _ := scope.FieldByName(meta.FieldName)

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
			utils.ExitWithMsg("Unsupported Collection format for meta %v of resource %v", meta.Name, reflect.TypeOf(meta.base.Value))
		}
	} else if meta.Type == "select_one" || meta.Type == "select_many" {
		if scopeField.Relationship != nil {
			fieldType := scopeField.StructField.Struct.Type
			if fieldType.Kind() == reflect.Slice {
				fieldType = fieldType.Elem()
			}

			meta.GetCollection = func(value interface{}, context *qor.Context) (results [][]string) {
				values := reflect.New(reflect.SliceOf(fieldType)).Interface()
				context.GetDB().Find(values)
				reflectValues := reflect.Indirect(reflect.ValueOf(values))
				for i := 0; i < reflectValues.Len(); i++ {
					scope := scope.New(reflectValues.Index(i).Interface())
					primaryKey := fmt.Sprintf("%v", scope.PrimaryKeyValue())
					results = append(results, []string{primaryKey, utils.Stringify(reflectValues.Index(i).Interface())})
				}
				return
			}
		} else {
			utils.ExitWithMsg("%v meta type %v needs Collection", meta.Name, meta.Type)
		}
	}

	if meta.Setter == nil && hasColumn {
		if relationship := field.Relationship; relationship != nil {
			if meta.Type == "select_one" || meta.Type == "select_many" {
				meta.Setter = func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
					scope := &gorm.Scope{Value: resource}
					reflectValue := reflect.Indirect(reflect.ValueOf(resource))
					field := reflectValue.FieldByName(meta.FieldName)

					if field.Kind() == reflect.Ptr {
						if field.IsNil() {
							field.Set(utils.NewValue(field.Type()).Elem())
						}

						for field.Kind() == reflect.Ptr {
							field = field.Elem()
						}
					}

					primaryKeys := utils.ToArray(metaValue.Value)
					// associations not changed for belongs to
					if relationship.Kind == "belongs_to" && len(relationship.ForeignFieldNames) == 1 {
						oldPrimaryKeys := utils.ToArray(reflectValue.FieldByName(relationship.ForeignFieldNames[0]).Interface())
						// if not changed
						if fmt.Sprint(primaryKeys) == fmt.Sprint(oldPrimaryKeys) {
							return
						}

						// if removed
						if len(primaryKeys) == 0 {
							field := reflectValue.FieldByName(relationship.ForeignFieldNames[0])
							field.Set(reflect.Zero(field.Type()))
						}
					}

					if len(primaryKeys) > 0 {
						context.GetDB().Where(primaryKeys).Find(field.Addr().Interface())
					}

					// Replace many 2 many relations
					if relationship.Kind == "many_to_many" {
						if !scope.PrimaryKeyZero() {
							context.GetDB().Model(resource).Association(meta.FieldName).Replace(field.Interface())
							field.Set(reflect.Zero(field.Type()))
						}
					}
				}
			}
		} else {
			meta.Setter = func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
				if metaValue == nil {
					return
				}

				value := metaValue.Value
				fieldName := meta.FieldName
				if nestedField {
					fields := strings.Split(fieldName, ".")
					fieldName = fields[len(fields)-1]
				}

				field := reflect.Indirect(reflect.ValueOf(resource)).FieldByName(fieldName)
				if field.Kind() == reflect.Ptr {
					if field.IsNil() {
						field.Set(utils.NewValue(field.Type()).Elem())
					}

					for field.Kind() == reflect.Ptr {
						field = field.Elem()
					}
				}

				if field.IsValid() && field.CanAddr() {
					switch field.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						field.SetInt(utils.ToInt(value))
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						field.SetUint(utils.ToUint(value))
					case reflect.Float32, reflect.Float64:
						field.SetFloat(utils.ToFloat(value))
					case reflect.Bool:
						// TODO: add test
						if utils.ToString(value) == "true" {
							field.SetBool(true)
						} else {
							field.SetBool(false)
						}
					default:
						if scanner, ok := field.Addr().Interface().(sql.Scanner); ok {
							if scanner.Scan(value) != nil {
								scanner.Scan(utils.ToString(value))
							}
						} else if reflect.TypeOf("").ConvertibleTo(field.Type()) {
							field.Set(reflect.ValueOf(utils.ToString(value)).Convert(field.Type()))
						} else if reflect.TypeOf([]string{}).ConvertibleTo(field.Type()) {
							field.Set(reflect.ValueOf(utils.ToArray(value)).Convert(field.Type()))
						} else if rvalue := reflect.ValueOf(value); reflect.TypeOf(rvalue.Type()).ConvertibleTo(field.Type()) {
							field.Set(rvalue.Convert(field.Type()))
						} else if _, ok := field.Addr().Interface().(*time.Time); ok {
							if str := utils.ToString(value); str != "" {
								if newTime, err := now.Parse(str); err == nil {
									field.Set(reflect.ValueOf(newTime))
								}
							}
						} else {
							var buf = bytes.NewBufferString("")
							json.NewEncoder(buf).Encode(value)
							if err := json.NewDecoder(strings.NewReader(buf.String())).Decode(field.Addr().Interface()); err != nil {
								utils.ExitWithMsg("Can't set value %v to %v [meta %v]", reflect.ValueOf(value).Type(), field.Type(), meta)
							}
						}
					}
				}
			}
		}
	}

	if nestedField {
		oldvalue := meta.Valuer
		meta.Valuer = func(value interface{}, context *qor.Context) interface{} {
			return oldvalue(utils.GetNestedModel(value, meta.FieldName, context), context)
		}
		oldSetter := meta.Setter
		meta.Setter = func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			oldSetter(utils.GetNestedModel(resource, meta.FieldName, context), metaValue, context)
		}
	}
}
