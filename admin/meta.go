package admin

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

type Meta struct {
	Name            string
	FieldName       string
	Label           string
	Type            string
	FormattedValuer func(interface{}, *qor.Context) interface{}
	Valuer          func(interface{}, *qor.Context) interface{}
	Setter          func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context)
	Metas           []resource.Metaor
	Resource        *Resource
	Collection      interface{}
	GetCollection   func(interface{}, *qor.Context) [][]string
	Permission      *roles.Permission
	resource.Meta
	baseResource *Resource
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

func (meta *Meta) DBName() string {
	if meta.FieldStruct != nil {
		return meta.FieldStruct.DBName
	}
	return ""
}

func getField(fields []*gorm.StructField, name string) (*gorm.StructField, bool) {
	for _, field := range fields {
		if field.Name == name || field.DBName == name {
			return field, true
		}
	}
	return nil, false
}

func (meta *Meta) setBaseResource(base *Resource) {
	res := meta.Resource
	res.base = base

	findOneHandle := res.FindOneHandler
	res.FindOneHandler = func(value interface{}, metaValues *resource.MetaValues, context *qor.Context) (err error) {
		if metaValues != nil {
			return findOneHandle(value, metaValues, context)
		}

		if primaryKey := res.GetPrimaryValue(context.Request); primaryKey != "" {
			clone := context.Clone()
			baseValue := base.NewStruct()
			if err = base.FindOneHandler(baseValue, nil, clone); err == nil {
				scope := clone.GetDB().NewScope(nil)
				sql := fmt.Sprintf("%v = ?", scope.Quote(res.PrimaryDBName()))
				err = context.GetDB().Model(baseValue).Where(sql, primaryKey).Related(value).Error
			}
		}
		return
	}

	res.FindManyHandler = func(value interface{}, context *qor.Context) error {
		clone := context.Clone()
		baseValue := base.NewStruct()
		if err := base.FindOneHandler(baseValue, nil, clone); err == nil {
			base.FindOneHandler(baseValue, nil, clone)
			return context.GetDB().Model(baseValue).Related(value).Error
		} else {
			return err
		}
	}

	res.SaveHandler = func(value interface{}, context *qor.Context) error {
		clone := context.Clone()
		baseValue := base.NewStruct()
		if err := base.FindOneHandler(baseValue, nil, clone); err == nil {
			base.FindOneHandler(baseValue, nil, clone)
			return context.GetDB().Model(baseValue).Association(meta.FieldName).Append(value).Error
		} else {
			return err
		}
	}

	res.DeleteHandler = func(value interface{}, context *qor.Context) (err error) {
		var clone = context.Clone()
		var baseValue = base.NewStruct()
		if primryKey := res.GetPrimaryValue(context.Request); primryKey != "" {
			var scope = clone.GetDB().NewScope(nil)
			var sql = fmt.Sprintf("%v = ?", scope.Quote(res.PrimaryDBName()))
			if err = context.GetDB().First(value, sql, primryKey).Error; err == nil {
				if err = base.FindOneHandler(baseValue, nil, clone); err == nil {
					base.FindOneHandler(baseValue, nil, clone)
					return context.GetDB().Model(baseValue).Association(meta.FieldName).Delete(value).Error
				}
			}
		}
		return
	}
}

func (meta *Meta) SetPermission(permission *roles.Permission) {
	meta.Permission = permission
	meta.Meta.Permission = permission
	if meta.Resource != nil {
		meta.Resource.Permission = permission
		for _, meta := range meta.Resource.Metas {
			meta.SetPermission(permission.Concat(meta.Meta.Permission))
		}
	}
}

func (meta *Meta) updateMeta() {
	meta.Meta = resource.Meta{
		Name:            meta.Name,
		FieldName:       meta.FieldName,
		Setter:          meta.Setter,
		FormattedValuer: meta.FormattedValuer,
		Valuer:          meta.Valuer,
		Permission:      meta.Permission.Concat(meta.baseResource.Permission),
		Resource:        meta.baseResource,
	}

	meta.PreInitialize()
	if meta.FieldStruct != nil {
		if injector, ok := reflect.New(meta.FieldStruct.Struct.Type).Interface().(resource.ConfigureMetaBeforeInitializeInterface); ok {
			injector.ConfigureQorMetaBeforeInitialize(meta)
		}
	}

	meta.Initialize()

	if meta.Label == "" {
		meta.Label = utils.HumanizeString(meta.Name)
	}

	var fieldType reflect.Type
	var hasColumn = meta.FieldStruct != nil

	if hasColumn {
		fieldType = meta.FieldStruct.Struct.Type
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
	}

	// Set Meta Type
	if meta.Type == "" && hasColumn {
		if relationship := meta.FieldStruct.Relationship; relationship != nil {
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
			switch fieldType.Kind() {
			case reflect.String:
				var tag = meta.FieldStruct.Tag
				if size, ok := utils.ParseTagOption(tag.Get("sql"))["SIZE"]; ok {
					if i, _ := strconv.Atoi(size); i > 255 {
						meta.Type = "text"
					} else {
						meta.Type = "string"
					}
				} else if text, ok := utils.ParseTagOption(tag.Get("sql"))["TYPE"]; ok && text == "text" {
					meta.Type = "text"
				} else {
					meta.Type = "string"
				}
			case reflect.Bool:
				meta.Type = "checkbox"
			default:
				if regexp.MustCompile(`^(.*)?(u)?(int)(\d+)?`).MatchString(fieldType.Kind().String()) {
					meta.Type = "number"
				} else if regexp.MustCompile(`^(.*)?(float)(\d+)?`).MatchString(fieldType.Kind().String()) {
					meta.Type = "float"
				} else if _, ok := reflect.New(fieldType).Interface().(*time.Time); ok {
					meta.Type = "datetime"
				}
			}
		}
	}

	{ // Set Meta Resource
		if hasColumn && (meta.FieldStruct.Relationship != nil) {
			if meta.Resource == nil {
				var result interface{}
				if fieldType.Kind() == reflect.Struct {
					result = reflect.New(fieldType).Interface()
				} else if fieldType.Kind() == reflect.Slice {
					refelectType := fieldType.Elem()
					for refelectType.Kind() == reflect.Ptr {
						refelectType = refelectType.Elem()
					}
					result = reflect.New(refelectType).Interface()
				}

				res := meta.baseResource.GetAdmin().NewResource(result)
				res.configure()
				meta.Resource = res
			}

			if meta.Resource != nil {
				permission := meta.Resource.Permission.Concat(meta.Meta.Permission)
				meta.Resource.Permission = permission
				meta.SetPermission(permission)
				meta.setBaseResource(meta.baseResource)
			}
		}
	}

	scope := &gorm.Scope{Value: meta.baseResource.Value}
	scopeField, _ := scope.FieldByName(meta.GetFieldName())

	{ // Format Meta FormattedValueOf
		if meta.FormattedValuer == nil {
			if meta.Type == "select_one" {
				meta.SetFormattedValuer(func(value interface{}, context *qor.Context) interface{} {
					return utils.Stringify(meta.GetValuer()(value, context))
				})
			} else if meta.Type == "select_many" {
				meta.SetFormattedValuer(func(value interface{}, context *qor.Context) interface{} {
					reflectValue := reflect.Indirect(reflect.ValueOf(meta.GetValuer()(value, context)))
					var results []string
					for i := 0; i < reflectValue.Len(); i++ {
						results = append(results, utils.Stringify(reflectValue.Index(i).Interface()))
					}
					return results
				})
			}
		}
	}

	{ // Format Meta Collection
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
				utils.ExitWithMsg("Unsupported Collection format for meta %v of resource %v", meta.Name, reflect.TypeOf(meta.baseResource.Value))
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

	meta.FieldName = meta.GetFieldName()

	// call ConfigureMetaInterface
	if meta.FieldStruct != nil {
		if injector, ok := reflect.New(meta.FieldStruct.Struct.Type).Interface().(resource.ConfigureMetaInterface); ok {
			injector.ConfigureQorMeta(meta)
		}
	}

	// run meta configors
	if baseResource := meta.baseResource; baseResource != nil {
		for key, fc := range baseResource.GetAdmin().metaConfigorMaps {
			if key == meta.Type {
				fc(meta)
			}
		}
	}
}
