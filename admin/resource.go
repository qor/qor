package admin

import (
	"fmt"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type Resource struct {
	Metas      []*Meta
	actions    []*Action
	scopes     map[string]*Scope
	filters    map[string]*Filter
	indexAttrs []string
	newAttrs   []string
	editAttrs  []string
	showAttrs  []string
}

func (res *Resource) ToParam() string {
	return strings.ToLower(res.Name)
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
		return res.GetMetas(res.indexAttrs, res.showAttrs)
	})
}

func (res *Resource) NewMetas() []*resource.Meta {
	return res.getCachedMetas("new_metas", func() []*resource.Meta {
		return res.GetMetas(res.newAttrs, res.editAttrs)
	})
}

func (res *Resource) EditMetas() []*resource.Meta {
	return res.getCachedMetas("edit_metas", func() []*resource.Meta {
		return res.GetMetas(res.editAttrs)
	})
}

func (res *Resource) ShowMetas() []*resource.Meta {
	return res.getCachedMetas("show_metas", func() []*resource.Meta {
		return res.GetMetas(res.showAttrs, res.editAttrs)
	})
}

func (res *Resource) AllMetas() []*resource.Meta {
	return res.getCachedMetas("all_metas", func() []*resource.Meta {
		return res.GetMetas()
	})
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
