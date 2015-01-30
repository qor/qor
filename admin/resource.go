package admin

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type Resource struct {
	resource.Resource
	Config      *Config
	Metas       []*Meta
	actions     []*Action
	scopes      map[string]*Scope
	filters     map[string]*Filter
	indexAttrs  []string
	newAttrs    []string
	editAttrs   []string
	showAttrs   []string
	cachedMetas *map[string][]*Meta
}

func (res Resource) ToParam() string {
	return strings.ToLower(res.Name)
}

func (res *Resource) ConvertObjectToMap(context qor.Contextor, value interface{}) interface{} {
	return resource.ConvertObjectToMap(context, value, res.GetMetas())
}

func (res *Resource) Decode(contextor qor.Contextor, value interface{}) (errs []error) {
	return resource.Decode(contextor, value, res)
}

func (res *Resource) CallFinder(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
	if res.Finder != nil {
		return res.Finder(result, metaValues, context)
	} else {
		var primaryKey string
		if metaValues == nil {
			primaryKey = context.ResourceID
		} else if id := metaValues.Get(res.PrimaryKey()); id != nil {
			primaryKey = resource.ToString(id.Value)
		}

		if primaryKey != "" {
			if metaValues != nil {
				if destroy := metaValues.Get("_destroy"); destroy != nil {
					if fmt.Sprintf("%v", destroy.Value) != "0" {
						context.GetDB().Delete(result, primaryKey)
						return resource.ErrProcessorSkipLeft
					}
				}
			}
			return context.GetDB().First(result, primaryKey).Error
		}
		return nil
	}
}

func (res *Resource) CallSearcher(result interface{}, context *qor.Context) error {
	if res.Searcher != nil {
		return res.Searcher(result, context)
	} else {
		return context.GetDB().Set("gorm:order_by_primary_key", "ASC").Find(result).Error
	}
}

func (res *Resource) CallSaver(result interface{}, context *qor.Context) error {
	if res.Saver != nil {
		return res.Saver(result, context)
	} else {
		return context.GetDB().Save(result).Error
	}
}

func (res *Resource) IndexAttrs(columns ...string) {
	res.indexAttrs = columns
}

func (res *Resource) NewAttrs(columns ...string) {
	res.newAttrs = columns
}

func (res *Resource) EditAttrs(columns ...string) {
	res.editAttrs = columns
}

func (res *Resource) ShowAttrs(columns ...string) {
	res.showAttrs = columns
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

func (res *Resource) GetMetas(_attrs ...[]string) []resource.Metaor {
	var attrs []string
	for _, value := range _attrs {
		if value != nil {
			attrs = value
			break
		}
	}

	if attrs == nil {
		scope := &gorm.Scope{Value: res.Value}
		attrs = []string{}
		fields := scope.Fields()

		includedMeta := map[string]bool{}
		for _, meta := range res.Metas {
			meta := meta.GetMeta()
			if _, ok := fields[meta.Name]; !ok {
				includedMeta[meta.Alias] = true
				attrs = append(attrs, meta.Name)
			}
		}

		for _, field := range fields {
			if _, ok := includedMeta[field.Name]; ok {
				continue
			}
			attrs = append(attrs, field.Name)
		}
	}

	primaryKey := res.PrimaryKey()

	metas := []resource.Metaor{}
	for _, attr := range attrs {
		var meta *Meta
		for _, m := range res.Metas {
			if m.GetMeta().Name == attr {
				meta = m
				break
			}
		}

		if meta == nil {
			meta = &Meta{}
			meta.Name = attr
			meta.Base = res
			if attr == primaryKey {
				meta.Type = "hidden"
			}
			meta.UpdateMeta()
		}

		metas = append(metas, meta)
	}

	return metas
}

func (res *Resource) IndexMetas() []*Meta {
	return res.getCachedMetas("index_metas", func() []resource.Metaor {
		return res.GetMetas(res.indexAttrs, res.showAttrs)
	})
}

func (res *Resource) NewMetas() []*Meta {
	return res.getCachedMetas("new_metas", func() []resource.Metaor {
		return res.GetMetas(res.newAttrs, res.editAttrs)
	})
}

func (res *Resource) EditMetas() []*Meta {
	return res.getCachedMetas("edit_metas", func() []resource.Metaor {
		return res.GetMetas(res.editAttrs)
	})
}

func (res *Resource) ShowMetas() []*Meta {
	return res.getCachedMetas("show_metas", func() []resource.Metaor {
		return res.GetMetas(res.showAttrs, res.editAttrs)
	})
}

func (res *Resource) AllMetas() []*Meta {
	return res.getCachedMetas("all_metas", func() []resource.Metaor {
		return res.GetMetas()
	})
}

func (res *Resource) AllowedMetas(attrs []*Meta, context *Context, roles ...roles.PermissionMode) []*Meta {
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
