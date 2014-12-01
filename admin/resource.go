package admin

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type Resource struct {
	Name string
	resource.Resource
	indexAttrs  []string
	newAttrs    []string
	editAttrs   []string
	showAttrs   []string
	cachedMetas *map[string][]*resource.Meta
}

func (res *Resource) CallFinder(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
	if res.Finder != nil {
		return res.Finder(result, metaValues, context)
	} else {
		var primaryKey string
		if metaValues == nil {
			primaryKey = context.ResourceID
		} else if id := metaValues.Get("_id"); id != nil {
			primaryKey = resource.ToString(id.Value)
		}

		if primaryKey != "" {
			if metaValues != nil {
				if destroy := metaValues.Get("_destroy"); destroy != nil {
					if fmt.Sprintf("%v", destroy.Value) != "0" {
						context.DB().Delete(result, primaryKey)
						return resource.ErrProcessorSkipLeft
					}
				}
			}
			return context.DB().First(result, primaryKey).Error
		}
		return nil
	}
}

func (r *Resource) IndexAttrs(columns ...string) {
	r.indexAttrs = columns
}

func (r *Resource) NewAttrs(columns ...string) {
	r.newAttrs = columns
}

func (r *Resource) EditAttrs(columns ...string) {
	r.editAttrs = columns
}

func (r *Resource) ShowAttrs(columns ...string) {
	r.showAttrs = columns
}

func (res *Resource) getMetas(attrsSlice ...[]string) []*resource.Meta {
	var attrs []string
	for _, value := range attrsSlice {
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

	metas := []*resource.Meta{}
	for _, attr := range attrs {
		if meta, ok := res.Metas[attr]; ok {
			metas = append(metas, meta.GetMeta())
		} else {
			if strings.HasSuffix(attr, "Id") {
				continue
			}

			var _meta resource.Meta
			_meta = resource.Meta{Name: attr, Base: res}
			_meta.UpdateMeta()
			metas = append(metas, &_meta)
		}
	}

	return metas
}

func (res *Resource) getCachedMetas(cacheKey string, fc func() []*resource.Meta) []*resource.Meta {
	if res.cachedMetas == nil {
		res.cachedMetas = &map[string][]*resource.Meta{}
	}

	if values, ok := (*res.cachedMetas)[cacheKey]; ok {
		return values
	} else {
		values = fc()
		(*res.cachedMetas)[cacheKey] = values
		return values
	}
}

func (res *Resource) IndexMetas() []*resource.Meta {
	return res.getCachedMetas("index_metas", func() []*resource.Meta {
		return res.getMetas(res.indexAttrs, res.showAttrs)
	})
}

func (res *Resource) NewMetas() []*resource.Meta {
	return res.getCachedMetas("new_metas", func() []*resource.Meta {
		return res.getMetas(res.newAttrs, res.editAttrs)
	})
}

func (res *Resource) EditMetas() []*resource.Meta {
	return res.getCachedMetas("edit_metas", func() []*resource.Meta {
		return res.appendPrimaryKey(res.getMetas(res.editAttrs))
	})
}

func (res *Resource) ShowMetas() []*resource.Meta {
	return res.getCachedMetas("show_metas", func() []*resource.Meta {
		return res.getMetas(res.showAttrs, res.editAttrs)
	})
}

func (res *Resource) AllAttrs() []*resource.Meta {
	return res.getCachedMetas("all_metas", func() []*resource.Meta {
		return res.appendPrimaryKey(res.getMetas())
	})
}

func (res *Resource) appendPrimaryKey(metas []*resource.Meta) []*resource.Meta {
	primaryKeyMeta := &resource.Meta{Base: res, Name: "_id", Type: "hidden", Value: func(value interface{}, context *qor.Context) interface{} {
		return context.DB().NewScope(value).PrimaryKeyValue()
	}}
	primaryKeyMeta.UpdateMeta()

	return append(metas, primaryKeyMeta)
}

func (res *Resource) AllowedMetas(attrs []*resource.Meta, context *Context, roles ...roles.PermissionMode) []*resource.Meta {
	var metas = []*resource.Meta{}
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
