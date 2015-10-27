package exchange

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
	"github.com/qor/qor/utils"
)

type Meta struct {
	base       *Resource
	Name       string
	Header     string
	Valuer     func(interface{}, *qor.Context) interface{}
	Setter     func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context)
	Permission *roles.Permission
}

func (meta *Meta) GetName() string {
	return meta.Name
}

func (meta *Meta) GetFieldName() string {
	return meta.Name
}

func (meta *Meta) GetMetas() []resource.Metaor {
	return []resource.Metaor{}
}

func (meta *Meta) GetResource() resource.Resourcer {
	return nil
}

func (meta *Meta) GetValuer() func(interface{}, *qor.Context) interface{} {
	return func(record interface{}, context *qor.Context) interface{} {
		if valuer := meta.Valuer; valuer != nil {
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

func (meta *Meta) GetSetter() func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
	return meta.Setter
}

func (meta *Meta) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if meta.Permission == nil {
		return true
	}
	return meta.Permission.HasPermission(mode, context.Roles...)
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
	}

	var (
		scope       = &gorm.Scope{Value: meta.base.Value}
		nestedField = strings.Contains(meta.Name, ".")
		field       *gorm.StructField
		hasColumn   bool
	)

	if nestedField {
		subModel, name := utils.ParseNestedField(reflect.ValueOf(meta.base.Value), meta.Name)
		subScope := &gorm.Scope{Value: subModel.Interface()}
		field, hasColumn = getField(subScope.GetStructFields(), name)
	} else {
		field, hasColumn = getField(scope.GetStructFields(), meta.Name)
	}

	var fieldType reflect.Type
	if hasColumn {
		fieldType = field.Struct.Type
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
	}

	// Set Meta Valuer
	if meta.Valuer == nil {
		if hasColumn {
			meta.Valuer = func(value interface{}, context *qor.Context) interface{} {
				scope := context.GetDB().NewScope(value)
				alias := meta.Name
				if nestedField {
					fields := strings.Split(alias, ".")
					alias = fields[len(fields)-1]
				}

				if f, ok := scope.FieldByName(alias); ok {
					if field.Relationship != nil {
						if f.Field.CanAddr() && !scope.PrimaryKeyZero() {
							context.GetDB().Model(value).Related(f.Field.Addr().Interface(), meta.Name)
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

	if meta.Setter == nil && hasColumn {
		if relationship := field.Relationship; relationship != nil {
			if relationship.Kind == "belongs_to" || relationship.Kind == "many_to_many" {
				meta.Setter = func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
					scope := &gorm.Scope{Value: resource}
					reflectValue := reflect.Indirect(reflect.ValueOf(resource))
					field := reflectValue.FieldByName(meta.Name)

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
							context.GetDB().Model(resource).Association(meta.Name).Replace(field.Interface())
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
				alias := meta.Name
				if nestedField {
					fields := strings.Split(alias, ".")
					alias = fields[len(fields)-1]
				}

				field := reflect.Indirect(reflect.ValueOf(resource)).FieldByName(alias)
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
			return oldvalue(utils.GetNestedModel(value, meta.Name, context), context)
		}
		oldSetter := meta.Setter
		meta.Setter = func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
			oldSetter(utils.GetNestedModel(resource, meta.Name, context), metaValue, context)
		}
	}
}
