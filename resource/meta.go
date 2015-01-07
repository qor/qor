package resource

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
	"github.com/qor/qor/roles"
)

type Meta struct {
	Base          Resourcer
	Name          string
	Alias         string
	Type          string
	Label         string
	Value         func(interface{}, *qor.Context) interface{}
	Setter        func(resource interface{}, metaValues *MetaValues, context *qor.Context)
	Collection    interface{}
	GetCollection func(interface{}, *qor.Context) [][]string
	Resource      Resourcer
	Permission    *roles.Permission
}

type Metaor interface {
	GetMeta() *Meta
	HasPermission(roles.PermissionMode, *qor.Context) bool
}

func (meta *Meta) GetMeta() *Meta {
	return meta
}

func (meta *Meta) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if meta.Permission == nil {
		return true
	}
	return meta.Permission.HasPermission(mode, context.Roles...)
}

func (meta *Meta) UpdateMeta() {
	if meta.Name == "" {
		qor.ExitWithMsg("Meta should have name: %v", reflect.ValueOf(meta).Type())
	} else {
		if meta.Label == "" {
			meta.Label = HumanizeString(meta.Name)
		}
		if meta.Alias == "" {
			meta.Alias = meta.Name
		}
		meta.Alias = gorm.SnakeToUpperCamel(meta.Alias)
	}

	var (
		base        = meta.Base.GetResource()
		scope       = &gorm.Scope{Value: base.Value}
		field       *gorm.Field
		hasColumn   bool
		nestedField = strings.Contains(meta.Alias, ".")
		valueType   string
	)
	if nestedField {
		submodel, name := parseNestedField(reflect.ValueOf(base.Value), meta.Alias)
		subscope := &gorm.Scope{Value: submodel.Interface()}
		field, hasColumn = subscope.FieldByName(name)
	} else {
		field, hasColumn = scope.FieldByName(meta.Alias)
	}
	if hasColumn {
		valueType = field.Field.Type().Kind().String()
	}

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
				if regexp.MustCompile(`^(u)?(int|float)(\d+)?`).MatchString(valueType) {
					meta.Type = "number"
				} else if _, ok := field.Field.Interface().(time.Time); ok {
					meta.Type = "datetime"
				} else if _, ok := field.Field.Interface().(media_library.MediaLibrary); ok {
					meta.Type = "file"
				}
			}
		}
	}

	// Set Meta Resource
	if meta.Resource == nil {
		if hasColumn && (field.Relationship != nil) {
			var result interface{}
			if valueType == "struct" {
				result = reflect.New(reflect.Indirect(field.Field).Type()).Interface()
			} else if valueType == "slice" {
				result = reflect.New(reflect.Indirect(field.Field).Type().Elem()).Interface()
			}

			resource := reflect.New(reflect.Indirect(reflect.ValueOf(meta.Base)).Type()).Interface()
			if resourcer, ok := resource.(Resourcer); ok {
				res := resourcer.GetResource()
				res.Value = result
				meta.Resource = resourcer
			}
		}
	}

	// Set Meta Value
	if meta.Value == nil {
		if hasColumn {
			meta.Value = func(value interface{}, context *qor.Context) interface{} {
				scope := &gorm.Scope{Value: value}
				alias := meta.Alias
				if nestedField {
					fields := strings.Split(alias, ".")
					alias = fields[len(fields)-1]
				}

				if f, ok := scope.FieldByName(alias); ok {
					if field.Relationship != nil {
						if f.Field.CanAddr() {
							context.GetDB().Model(value).Related(f.Field.Addr().Interface(), meta.Alias)
						}
					}
					return f.Field.Addr().Interface()
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

	scopeField, _ := scope.FieldByName(meta.Alias)

	// TODO: also support nested fields (Phone.Num)?
	if meta.Setter == nil {
		meta.Setter = func(resource interface{}, metaValues *MetaValues, context *qor.Context) {
			metaValue := metaValues.Get(meta.Name)
			if metaValue == nil {
				return
			}

			value := metaValue.Value
			scope := &gorm.Scope{Value: resource}
			alias := meta.Alias
			if nestedField {
				fields := strings.Split(alias, ".")
				alias = fields[len(fields)-1]
			}
			field := reflect.Indirect(reflect.ValueOf(resource)).FieldByName(alias)

			if field.IsValid() && field.CanAddr() {
				var relationship string
				if scopeField != nil && scopeField.Relationship != nil {
					relationship = scopeField.Relationship.Kind
				}
				if relationship == "many_to_many" {
					context.GetDB().Where(ToArray(value)).Find(field.Addr().Interface())
					if !scope.PrimaryKeyZero() {
						context.GetDB().Model(resource).Association(meta.Alias).Replace(field.Interface())
					}
				} else {
					switch field.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						field.SetInt(ToInt(value))
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						field.SetUint(ToUint(value))
					case reflect.Float32, reflect.Float64:
						field.SetFloat(ToFloat(value))
					default:
						if scanner, ok := field.Addr().Interface().(sql.Scanner); ok {
							scanner.Scan(ToString(value))
						} else if reflect.TypeOf("").ConvertibleTo(field.Type()) {
							field.Set(reflect.ValueOf(ToString(value)).Convert(field.Type()))
						} else if reflect.TypeOf([]string{}).ConvertibleTo(field.Type()) {
							field.Set(reflect.ValueOf(ToArray(value)).Convert(field.Type()))
						} else if rvalue := reflect.ValueOf(value); reflect.TypeOf(rvalue.Type()).ConvertibleTo(field.Type()) {
							field.Set(rvalue.Convert(field.Type()))
						} else if _, ok := field.Addr().Interface().(*time.Time); ok {
							if str := ToString(value); str != "" {
								if newTime, err := now.Parse(str); err == nil {
									field.Set(reflect.ValueOf(newTime))
								}
							}
						} else {
							var buf = bytes.NewBufferString("")
							json.NewEncoder(buf).Encode(value)
							if err := json.NewDecoder(strings.NewReader(buf.String())).Decode(field.Addr().Interface()); err != nil {
								qor.ExitWithMsg("Can't set value %v to %v [meta %v]", reflect.ValueOf(value).Type(), field.Type(), meta)
							}
						}
					}
				}
			}
		}
	}

	if nestedField {
		oldvalue := meta.Value
		meta.Value = func(value interface{}, context *qor.Context) interface{} {
			return oldvalue(getNestedModel(value, meta.Alias, context), context)
		}
		oldSetter := meta.Setter
		meta.Setter = func(resource interface{}, metaValues *MetaValues, context *qor.Context) {
			oldSetter(getNestedModel(resource, meta.Alias, context), metaValues, context)
		}
	}
}

func getNestedModel(value interface{}, alias string, context *qor.Context) interface{} {
	model := reflect.Indirect(reflect.ValueOf(value))
	fields := strings.Split(alias, ".")
	for _, field := range fields[:len(fields)-1] {
		if model.CanAddr() {
			submodel := model.FieldByName(field)
			if key := submodel.FieldByName("Id"); !key.IsValid() || key.Uint() == 0 {
				if submodel.CanAddr() {
					context.GetDB().Model(model.Addr().Interface()).Related(submodel.Addr().Interface())
					model = submodel
				} else {
					break
				}
			} else {
				model = submodel
			}
		}
	}

	if model.CanAddr() {
		return model.Addr().Interface()
	} else {
		return nil
	}
}

// Profile.Name
func parseNestedField(value reflect.Value, name string) (reflect.Value, string) {
	fields := strings.Split(name, ".")
	value = reflect.Indirect(value)
	for _, field := range fields[:len(fields)-1] {
		value = value.FieldByName(field)
	}

	return value, fields[len(fields)-1]
}

func ToArray(value interface{}) (values []string) {
	if v, ok := value.([]string); ok {
		return v
	} else if v, ok := value.(string); ok {
		return []string{v}
	} else if vs, ok := value.([]interface{}); ok {
		for _, v := range vs {
			values = append(values, fmt.Sprintf("%v", v))
		}
	} else {
		panic(value)
	}
	return
}

func ToString(value interface{}) string {
	if v, ok := value.([]string); ok && len(v) > 0 {
		return v[0]
	} else if v, ok := value.(string); ok {
		return v
	} else if v, ok := value.([]interface{}); ok && len(v) > 0 {
		return fmt.Sprintf("%v", v[0])
	} else {
		panic(value)
	}
}

func ToInt(value interface{}) int64 {
	var result string
	if v, ok := value.([]string); ok && len(v) > 0 {
		result = v[0]
	} else if v, ok := value.(string); ok {
		result = v
	} else {
		return ToInt(fmt.Sprintf("%v", value))
	}

	if i, err := strconv.ParseInt(result, 10, 64); err == nil {
		return i
	} else if result == "" {
		return 0
	} else {
		panic("failed to parseint")
	}
}

func ToUint(value interface{}) uint64 {
	var result string
	if v, ok := value.([]string); ok && len(v) > 0 {
		result = v[0]
	} else if v, ok := value.(string); ok {
		result = v
	} else {
		return ToUint(fmt.Sprintf("%v", value))
	}

	if i, err := strconv.ParseUint(result, 10, 64); err == nil {
		return i
	} else if result == "" {
		return 0
	} else {
		panic("failed to parseuint")
	}
}

func ToFloat(value interface{}) float64 {
	var result string
	if v, ok := value.([]string); ok && len(v) > 0 {
		result = v[0]
	} else if v, ok := value.(string); ok {
		result = v
	} else {
		return ToFloat(fmt.Sprintf("%v", value))
	}

	if i, err := strconv.ParseFloat(result, 64); err == nil {
		return i
	} else if result == "" {
		return 0
	} else {
		panic("failed to parsefloat")
	}
}

// HumanizeString will split a string by uppercase character
// and capitablize the first character in the string
// e.g.
// 	FieldName -> Field Name
func HumanizeString(label string) string {
	var new []rune
	for _, l := range label {
		if rune('A') <= l && l <= rune('Z') {
			new = append(new, rune(' '), rune(l))
		} else {
			new = append(new, rune(l))
		}
	}

	return strings.Title(string(new))
}
