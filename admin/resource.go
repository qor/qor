package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/rules"

	"go/ast"
	"reflect"
	"strings"
)

type Resource struct {
	Name  string
	attrs *attrs
	resource.Resource
}

type attrs struct {
	indexAttrs []string
	newAttrs   []string
	editAttrs  []string
	showAttrs  []string
}

func (r *Resource) Attrs() *attrs {
	if r.attrs == nil {
		r.attrs = &attrs{}
	}
	return r.attrs
}

func (a *attrs) Index(columns ...string) {
	a.indexAttrs = columns
}

func (a *attrs) New(columns ...string) {
	a.newAttrs = columns
}

func (a *attrs) Edit(columns ...string) {
	a.editAttrs = columns
}

func (a *attrs) Show(columns ...string) {
	a.showAttrs = columns
}

func (res *Resource) getMetas(attrsSlice ...[]string) []resource.Meta {
	var attrs []string
	for _, value := range attrsSlice {
		if value != nil {
			attrs = value
			break
		}
	}

	if attrs == nil {
		attrs = []string{}
		indirectValue := reflect.Indirect(reflect.ValueOf(res.Value))
		scopeTyp := indirectValue.Type()
		for i := 0; i < scopeTyp.NumField(); i++ {
			fieldStruct := scopeTyp.Field(i)
			if !ast.IsExported(fieldStruct.Name) {
				continue
			}
			attrs = append(attrs, fieldStruct.Name)
		}
	}

	metas := []resource.Meta{}
OUT:
	for _, attr := range attrs {
		for _, meta := range res.Metas {
			meta := meta.(Meta)
			if meta.Name == attr {
				metas = append(metas, meta)
				continue OUT
			}
		}

		for _, meta := range res.Metas {
			if meta.Name == gorm.SnakeToUpperCamel(attr) {
				metas = append(metas, meta)
				continue OUT
			}
		}

		if strings.HasSuffix(attr, "Id") {
			continue
		}

		var _meta Meta
		_meta = Meta{Name: attr, base: resource}
		_meta.updateMeta()
		metas = append(metas, _meta)
	}

	return metas
}

func (resource *Resource) IndexAttrs() []Meta {
	return resource.getMetas(resource.attrs.indexAttrs, resource.attrs.showAttrs)
}

func (resource *Resource) NewAttrs() []Meta {
	return resource.getMetas(resource.attrs.newAttrs, resource.attrs.editAttrs)
}

func (resource *Resource) EditAttrs() []Meta {
	return resource.appendPrimaryKey(resource.getMetas(resource.attrs.editAttrs))
}

func (resource *Resource) ShowAttrs() []Meta {
	return resource.getMetas(resource.attrs.showAttrs, resource.attrs.editAttrs)
}

func (resource *Resource) AllAttrs() []Meta {
	return resource.appendPrimaryKey(resource.getMetas())
}

func (resource *Resource) appendPrimaryKey(metas []Meta) []Meta {
	primaryKeyMeta := Meta{base: resource, Name: "_id", Type: "hidden", Value: func(value interface{}, context *qor.Context) interface{} {
		return context.DB.NewScope(value).PrimaryKeyValue()
	}}
	primaryKeyMeta.updateMeta()

	return append(metas, primaryKeyMeta)
}

func (resource *Resource) AllowedMetas(attrs []Meta, context *qor.Context, rules ...rules.PermissionMode) []Meta {
	var metas = []Meta{}
	for _, meta := range attrs {
		for _, rule := range rules {
			if meta.HasPermission(rule, context) {
				metas = append(metas, meta)
				break
			}
		}
	}
	return metas
}
