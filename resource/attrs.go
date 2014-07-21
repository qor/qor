package resource

import (
	"fmt"
	"github.com/jinzhu/gorm"

	"go/ast"
	"reflect"
)

type attrs struct {
	indexAttrs []string
	newAttrs   []string
	editAttrs  []string
	showAttrs  []string
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

func (resource *Resource) getMetas(attrsSlice ...[]string) []Meta {
	var attrs []string
	for _, value := range attrsSlice {
		if value != nil {
			attrs = value
			break
		}
	}

	if attrs == nil {
		attrs = []string{}
		indirectValue := reflect.Indirect(reflect.ValueOf(resource.Model))
		scopeTyp := indirectValue.Type()
		for i := 0; i < scopeTyp.NumField(); i++ {
			fieldStruct := scopeTyp.Field(i)
			if !ast.IsExported(fieldStruct.Name) {
				continue
			}
			attrs = append(attrs, fieldStruct.Name)
		}
	}

	metas := []Meta{}
	for _, attr := range attrs {
		metaFound := false
		for _, meta := range resource.meta.metas {
			if meta.Name == attr {
				metas = append(metas, meta)
				metaFound = true
				break
			}
		}
		if !metaFound {
			var _meta Meta
			if _, ok := gorm.FieldByName(gorm.SnakeToUpperCamel(attr), resource.Model); ok {
				_meta = Meta{Name: gorm.SnakeToUpperCamel(attr), base: resource}
				_meta.updateMeta()
				metas = append(metas, _meta)
			} else {
				fmt.Printf("%v is not existing for %v\n", attr, resource.Model)
			}
		}
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
	return resource.getMetas(resource.attrs.editAttrs)
}

func (resource *Resource) ShowAttrs() []Meta {
	return resource.getMetas(resource.attrs.showAttrs, resource.attrs.editAttrs)
}
