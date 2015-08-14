package admin

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	"github.com/qor/inflection"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
	"github.com/qor/qor/utils"
)

type Resource struct {
	resource.Resource
	admin         *Admin
	Config        *Config
	Metas         []*Meta
	actions       []*Action
	scopes        []*Scope
	filters       map[string]*Filter
	searchAttrs   []string
	sortableAttrs []string
	indexAttrs    []string
	newAttrs      []string
	editAttrs     []string
	showAttrs     []string
	cachedMetas   *map[string][]*Meta
	SearchHandler func(keyword string, context *qor.Context) *gorm.DB
}

func (res *Resource) Meta(meta *Meta) {
	if res.GetMeta(meta.Name) != nil {
		utils.ExitWithMsg("Duplicated meta %v defined for resource %v", meta.Name, res.Name)
	}

	meta.base = res
	meta.updateMeta()
	res.Metas = append(res.Metas, meta)
}

func (res Resource) GetAdmin() *Admin {
	return res.admin
}

func (res Resource) ToParam() string {
	return utils.ToParamString(inflection.Plural(res.Name))
}

func (res Resource) UseTheme(theme string) []string {
	if res.Config != nil {
		res.Config.Themes = append(res.Config.Themes, theme)
		return res.Config.Themes
	}
	return []string{}
}

func (res *Resource) convertObjectToMap(context *Context, value interface{}, kind string) interface{} {
	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	switch reflectValue.Kind() {
	case reflect.Slice:
		values := []interface{}{}
		for i := 0; i < reflectValue.Len(); i++ {
			values = append(values, res.convertObjectToMap(context, reflectValue.Index(i).Interface(), kind))
		}
		return values
	case reflect.Struct:
		var metas []*Meta
		if kind == "index" {
			metas = res.indexMetas()
		} else if kind == "show" {
			metas = res.showMetas()
		}

		values := map[string]interface{}{}
		for _, meta := range metas {
			if meta.HasPermission(roles.Read, context.Context) {
				value := meta.GetValuer()(value, context.Context)
				if meta.Resource != nil {
					value = meta.Resource.(*Resource).convertObjectToMap(context, value, kind)
				}
				values[meta.GetName()] = value
			}
		}
		return values
	default:
		panic(fmt.Sprintf("Can't convert %v (%v) to map", reflectValue, reflectValue.Kind()))
	}
}

func (res *Resource) Decode(context *qor.Context, value interface{}) error {
	return resource.Decode(context, value, res)
}

func (res *Resource) allAttrs() []string {
	var attrs []string
	scope := &gorm.Scope{Value: res.Value}

Fields:
	for _, field := range scope.GetModelStruct().StructFields {
		for _, meta := range res.Metas {
			if field.Name == meta.Alias {
				attrs = append(attrs, meta.Name)
				continue Fields
			}
		}

		if field.IsForeignKey {
			continue
		}

		for _, value := range []string{"CreatedAt", "UpdatedAt", "DeletedAt"} {
			if value == field.Name {
				continue Fields
			}
		}

		attrs = append(attrs, field.Name)
	}

MetaIncluded:
	for _, meta := range res.Metas {
		for _, attr := range attrs {
			if attr == meta.Alias || attr == meta.Name {
				continue MetaIncluded
			}
		}
		attrs = append(attrs, meta.Name)
	}

	return attrs
}

func (res *Resource) getAttrs(attrs []string) []string {
	if len(attrs) == 0 {
		return res.allAttrs()
	} else {
		var onlyExcludeAttrs = true
		for _, attr := range attrs {
			if !strings.HasPrefix(attr, "-") {
				onlyExcludeAttrs = false
				break
			}
		}
		if onlyExcludeAttrs {
			return append(res.allAttrs(), attrs...)
		}
		return attrs
	}
}

func (res *Resource) IndexAttrs(columns ...string) []string {
	if len(columns) > 0 {
		res.indexAttrs = columns
		if len(res.SortableAttrs()) == 0 {
			res.SortableAttrs(columns...)
		}
		if len(res.SearchAttrs()) == 0 {
			res.SearchAttrs(columns...)
		}
	}
	return res.getAttrs(res.indexAttrs)
}

func (res *Resource) NewAttrs(columns ...string) []string {
	if len(columns) > 0 {
		res.newAttrs = columns
	}
	return res.getAttrs(res.newAttrs)
}

func (res *Resource) EditAttrs(columns ...string) []string {
	if len(columns) > 0 {
		res.editAttrs = columns
	}
	return res.getAttrs(res.editAttrs)
}

func (res *Resource) ShowAttrs(columns ...string) []string {
	if len(columns) > 0 {
		res.showAttrs = columns
	}
	return res.getAttrs(res.showAttrs)
}

func (res *Resource) SortableAttrs(columns ...string) []string {
	if len(columns) > 0 {
		res.sortableAttrs = []string{}
		scope := res.GetAdmin().Config.DB.NewScope(res.Value)
		for _, column := range columns {
			if field, ok := scope.FieldByName(column); ok && field.DBName != "" {
				res.sortableAttrs = append(res.sortableAttrs, column)
			}
		}
	}
	return res.sortableAttrs
}

func (res *Resource) SearchAttrs(columns ...string) []string {
	if len(columns) > 0 {
		res.searchAttrs = columns
		res.SearchHandler = func(keyword string, context *qor.Context) *gorm.DB {
			db := context.GetDB()
			var conditions []string
			var keywords []interface{}
			scope := db.NewScope(res.Value)

			for _, column := range columns {
				if field, ok := scope.FieldByName(column); ok {
					switch field.Field.Kind() {
					case reflect.String:
						conditions = append(conditions, fmt.Sprintf("upper(%v) like upper(?)", scope.Quote(field.DBName)))
						keywords = append(keywords, "%"+keyword+"%")
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						if _, err := strconv.Atoi(keyword); err == nil {
							conditions = append(conditions, fmt.Sprintf("%v = ?", scope.Quote(field.DBName)))
							keywords = append(keywords, keyword)
						}
					case reflect.Float32, reflect.Float64:
						if _, err := strconv.ParseFloat(keyword, 64); err == nil {
							conditions = append(conditions, fmt.Sprintf("%v = ?", scope.Quote(field.DBName)))
							keywords = append(keywords, keyword)
						}
					case reflect.Struct:
						// time ?
						if _, ok := field.Field.Interface().(time.Time); ok {
							if parsedTime, err := now.Parse(keyword); err == nil {
								conditions = append(conditions, fmt.Sprintf("%v = ?", scope.Quote(field.DBName)))
								keywords = append(keywords, parsedTime)
							}
						}
					case reflect.Ptr:
						// time ?
						if _, ok := field.Field.Interface().(*time.Time); ok {
							if parsedTime, err := now.Parse(keyword); err == nil {
								conditions = append(conditions, fmt.Sprintf("%v = ?", scope.Quote(field.DBName)))
								keywords = append(keywords, parsedTime)
							}
						}
					default:
						conditions = append(conditions, fmt.Sprintf("%v = ?", scope.Quote(field.DBName)))
						keywords = append(keywords, keyword)
					}
				}
			}

			if len(conditions) > 0 {
				return context.GetDB().Where(strings.Join(conditions, " OR "), keywords...)
			} else {
				return context.GetDB()
			}
		}
	}

	return res.searchAttrs
}

func (res *Resource) getCachedMetas(cacheKey string, fc func() []resource.Metaor) []*Meta {
	if res.cachedMetas == nil {
		res.cachedMetas = &map[string][]*Meta{}
	}

	if values, ok := (*res.cachedMetas)[cacheKey]; ok {
		return values
	} else {
		values := fc()
		var metas []*Meta
		for _, value := range values {
			metas = append(metas, value.(*Meta))
		}
		(*res.cachedMetas)[cacheKey] = metas
		return metas
	}
}

func (res *Resource) GetMetas(attrs []string) []resource.Metaor {
	if len(attrs) == 0 {
		attrs = res.allAttrs()
	}
	var showAttrs, ignoredAttrs []string
	for _, attr := range attrs {
		if strings.HasPrefix(attr, "-") {
			ignoredAttrs = append(ignoredAttrs, strings.TrimLeft(attr, "-"))
		} else {
			showAttrs = append(showAttrs, attr)
		}
	}

	primaryKey := res.PrimaryFieldName()

	metas := []resource.Metaor{}
Attrs:
	for _, attr := range showAttrs {
		for _, a := range ignoredAttrs {
			if attr == a {
				continue Attrs
			}
		}

		var meta *Meta
		for _, m := range res.Metas {
			if m.GetName() == attr {
				meta = m
				break
			}
		}

		if meta == nil {
			meta = &Meta{}
			meta.Name = attr
			meta.base = res
			if attr == primaryKey {
				meta.Type = "hidden"
			}
			meta.updateMeta()
		}

		metas = append(metas, meta)
	}

	return metas
}

func (res *Resource) GetMeta(name string) *Meta {
	for _, meta := range res.Metas {
		if meta.Name == name || meta.GetFieldName() == name {
			return meta
		}
	}
	return nil
}

func (res *Resource) indexMetas() []*Meta {
	return res.getCachedMetas("index_metas", func() []resource.Metaor {
		return res.GetMetas(res.IndexAttrs())
	})
}

func (res *Resource) newMetas() []*Meta {
	return res.getCachedMetas("new_metas", func() []resource.Metaor {
		return res.GetMetas(res.NewAttrs())
	})
}

func (res *Resource) editMetas() []*Meta {
	return res.getCachedMetas("edit_metas", func() []resource.Metaor {
		return res.GetMetas(res.EditAttrs())
	})
}

func (res *Resource) showMetas() []*Meta {
	return res.getCachedMetas("show_metas", func() []resource.Metaor {
		return res.GetMetas(res.ShowAttrs())
	})
}

func (res *Resource) allMetas() []*Meta {
	return res.getCachedMetas("all_metas", func() []resource.Metaor {
		return res.GetMetas([]string{})
	})
}

func (res *Resource) allowedMetas(attrs []*Meta, context *Context, roles ...roles.PermissionMode) []*Meta {
	var metas = []*Meta{}
	for _, meta := range attrs {
		for _, role := range roles {
			if meta.HasPermission(role, context.Context) {
				metas = append(metas, meta)
				break
			}
		}
	}
	return metas
}

func (res *Resource) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if res.Config == nil || res.Config.Permission == nil {
		return true
	}
	return res.Config.Permission.HasPermission(mode, context.Roles...)
}
