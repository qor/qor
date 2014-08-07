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
			if meta.GetMeta().Name == attr {
				metas = append(metas, *meta.GetMeta())
				continue OUT
			}
		}

		for _, meta := range res.Metas {
			if meta.GetMeta().Name == gorm.SnakeToUpperCamel(attr) {
				metas = append(metas, *meta.GetMeta())
				continue OUT
			}
		}

		if strings.HasSuffix(attr, "Id") {
			continue
		}

		var _meta resource.Meta
		_meta = resource.Meta{Name: attr, Base: res}
		_meta.UpdateMeta()
		metas = append(metas, _meta)
	}

	return metas
}

func (res *Resource) IndexAttrs() []resource.Meta {
	return res.getMetas(res.attrs.indexAttrs, res.attrs.showAttrs)
}

func (res *Resource) NewAttrs() []resource.Meta {
	return res.getMetas(res.attrs.newAttrs, res.attrs.editAttrs)
}

func (res *Resource) EditAttrs() []resource.Meta {
	return res.appendPrimaryKey(res.getMetas(res.attrs.editAttrs))
}

func (res *Resource) ShowAttrs() []resource.Meta {
	return res.getMetas(res.attrs.showAttrs, res.attrs.editAttrs)
}

func (res *Resource) AllAttrs() []resource.Meta {
	return res.appendPrimaryKey(res.getMetas())
}

func (res *Resource) appendPrimaryKey(metas []resource.Meta) []resource.Meta {
	primaryKeyMeta := resource.Meta{Base: res, Name: "_id", Type: "hidden", Value: func(value interface{}, context *qor.Context) interface{} {
		return context.DB.NewScope(value).PrimaryKeyValue()
	}}
	primaryKeyMeta.UpdateMeta()

	return append(metas, primaryKeyMeta)
}

func (res *Resource) AllowedMetas(attrs []resource.Meta, context *qor.Context, rules ...rules.PermissionMode) []resource.Meta {
	var metas = []resource.Meta{}
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
