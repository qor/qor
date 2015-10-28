package admin

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/media_library"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
	"github.com/qor/qor/utils"
)

type Meta struct {
	base          *Resource
	Name          string
	FieldName     string
	DBName        string
	Label         string
	Type          string
	Valuer        func(interface{}, *qor.Context) interface{}
	Setter        func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context)
	Metas         []resource.Metaor
	Resource      resource.Resourcer
	Collection    interface{}
	GetCollection func(interface{}, *qor.Context) [][]string
	Permission    *roles.Permission
	resource.Meta
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
	meta.Meta = resource.Meta{
		Name:          meta.Name,
		FieldName:     meta.FieldName,
		Setter:        meta.Setter,
		Valuer:        meta.Valuer,
		Permission:    meta.Permission,
		ResourceValue: meta.base.Value,
	}
	meta.UpdateMeta()

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
		subModel, name := parseNestedField(reflect.ValueOf(meta.base.Value), meta.FieldName)
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
}

func parseNestedField(value reflect.Value, name string) (reflect.Value, string) {
	fields := strings.Split(name, ".")
	value = reflect.Indirect(value)
	for _, field := range fields[:len(fields)-1] {
		value = value.FieldByName(field)
	}

	return value, fields[len(fields)-1]
}
