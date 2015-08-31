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
	Alias         string
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

func (meta *Meta) GetName() string {
	return meta.Name
}

func (meta *Meta) GetFieldName() string {
	return meta.Alias
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

func (meta *Meta) GetValuer() func(interface{}, *qor.Context) interface{} {
	return meta.Valuer
}

func (meta *Meta) GetSetter() func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
	return meta.Setter
}

func (meta *Meta) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if meta.Permission == nil {
		return true
	}
	return meta.Permission.HasPermission(mode, context.Roles...)
}

func getField(fields map[string]*gorm.Field, name string) (*gorm.Field, bool) {
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
	} else if meta.Alias == "" {
		meta.Alias = meta.Name
	}

	if meta.Label == "" {
		meta.Label = utils.HumanizeString(meta.Name)
	}

	var (
		scope       = &gorm.Scope{Value: meta.base.Value}
		nestedField = strings.Contains(meta.Alias, ".")
		field       *gorm.Field
		hasColumn   bool
	)

	if nestedField {
		subModel, name := utils.ParseNestedField(reflect.ValueOf(meta.base.Value), meta.Alias)
		subScope := &gorm.Scope{Value: subModel.Interface()}
		field, hasColumn = getField(subScope.Fields(), name)
	} else {
		if field, hasColumn = getField(scope.Fields(), meta.Alias); hasColumn {
			meta.Alias = field.Name
			meta.DBName = field.DBName
		}
	}

	var fieldType reflect.Type
	if hasColumn {
		fieldType = field.Field.Type()
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
	}

	// Set Meta Type
	if meta.Type == "" && hasColumn {
		if relationship := field.Relationship; relationship != nil {
			if relationship.Kind == "belongs_to" || relationship.Kind == "has_one" {
				meta.Type = "single_edit"
			} else if relationship.Kind == "has_many" {
				meta.Type = "collection_edit"
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
				result = reflect.New(field.Field.Type()).Interface()
			} else if fieldType.Kind().String() == "slice" {
				result = reflect.New(field.Field.Type().Elem()).Interface()
			}

			res := meta.base.GetAdmin().NewResource(result)
			res.compile()
			meta.Resource = res
		}
	}

	if meta.Type == "serialize_argument" {
		type SerializeArgumentInterface interface {
			GetSerializeArgumentResource() *Resource
			GetSerializeArgument() interface{}
			SetSerializeArgument(value interface{})
		}

		if meta.Valuer == nil {
			meta.Valuer = func(value interface{}, context *qor.Context) interface{} {
				if serializeArgument, ok := value.(SerializeArgumentInterface); ok {
					return struct {
						Value    interface{}
						Resource *Resource
					}{
						Value:    serializeArgument.GetSerializeArgument(),
						Resource: serializeArgument.GetSerializeArgumentResource(),
					}
				}
				return nil
			}
		}

		if meta.Setter == nil {
			meta.Setter = func(result interface{}, metaValue *resource.MetaValue, context *qor.Context) {
				if serializeArgument, ok := result.(SerializeArgumentInterface); ok {
					if res := serializeArgument.GetSerializeArgumentResource(); res != nil {
						value := res.NewStruct()

						for _, meta := range res.GetMetas([]string{}) {
							metaValue := metaValue.MetaValues.Get(meta.GetName())
							if setter := meta.GetSetter(); setter != nil {
								setter(value, metaValue, context)
							}
						}

						serializeArgument.SetSerializeArgument(value)
					}
				}
			}
		}
	}

	// Set Meta Valuer
	if meta.Valuer == nil {
		if hasColumn {
			meta.Valuer = func(value interface{}, context *qor.Context) interface{} {
				scope := context.GetDB().NewScope(value)
				alias := meta.Alias
				if nestedField {
					fields := strings.Split(alias, ".")
					alias = fields[len(fields)-1]
				}

				if f, ok := scope.FieldByName(alias); ok {
					if field.Relationship != nil {
						if f.Field.CanAddr() && !scope.PrimaryKeyZero() {
							context.GetDB().Model(value).Related(f.Field.Addr().Interface(), meta.Alias)
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

	scopeField, _ := scope.FieldByName(meta.Alias)

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
			meta.Setter = func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
				scope := &gorm.Scope{Value: resource}
				field := reflect.Indirect(reflect.ValueOf(resource)).FieldByName(meta.Alias)

				if field.Kind() == reflect.Ptr && field.IsNil() {
					field.Set(utils.NewValue(field.Type()).Elem())
				}

				for field.Kind() == reflect.Ptr {
					field = field.Elem()
				}

				if primaryKeys := utils.ToArray(metaValue.Value); len(primaryKeys) > 0 {
					context.GetDB().Where(primaryKeys).Find(field.Addr().Interface())
				}

				if relationship.Kind == "many_to_many" {
					// Replace many 2 many relations
					if !scope.PrimaryKeyZero() {
						context.GetDB().Model(resource).Association(meta.Alias).Replace(field.Interface())
					}
				}
			}
		} else {
			meta.Setter = func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
				if metaValue == nil {
					return
				}

				value := metaValue.Value
				alias := meta.Alias
				if nestedField {
					fields := strings.Split(alias, ".")
					alias = fields[len(fields)-1]
				}

				field := reflect.Indirect(reflect.ValueOf(resource)).FieldByName(alias)
				if field.Kind() == reflect.Ptr && field.IsNil() {
					field.Set(utils.NewValue(field.Type()).Elem())
				}

				for field.Kind() == reflect.Ptr {
					field = field.Elem()
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
			return oldvalue(utils.GetNestedModel(value, meta.Alias, context), context)
		}
		oldSetter := meta.Setter
		meta.Setter = func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			oldSetter(utils.GetNestedModel(resource, meta.Alias, context), metaValue, context)
		}
	}
}
