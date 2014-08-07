package admin

import (
	"bytes"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/rules"
	"path"
	"reflect"
	"strings"

	"text/template"
)

type Content struct {
	Admin    *Admin
	Context  *qor.Context
	Resource *Resource
	Result   interface{}
	Action   string
}

func (content *Content) AllowedMetas(modes ...rules.PermissionMode) func(reses ...*Resource) []resource.Meta {
	return func(reses ...*Resource) []resource.Meta {
		var res = content.Resource
		if len(reses) > 0 {
			res = reses[0]
		}

		switch content.Action {
		case "index":
			return res.AllowedMetas(res.IndexAttrs(), content.Context, modes...)
		case "show":
			return res.AllowedMetas(res.ShowAttrs(), content.Context, modes...)
		case "edit":
			return res.AllowedMetas(res.EditAttrs(), content.Context, modes...)
		case "new":
			return res.AllowedMetas(res.NewAttrs(), content.Context, modes...)
		default:
			return []resource.Meta{}
		}
	}
}

func (content *Content) ValueOf(value interface{}, meta resource.Meta) interface{} {
	return meta.Value(value, content.Context)
}

func (content *Content) NewResourcePath(value interface{}) string {
	if res, ok := value.(*Resource); ok {
		return path.Join(content.Admin.Prefix, res.Name, "new")
	} else {
		return path.Join(content.Admin.Prefix, content.Resource.Name, "new")
	}
}

func (content *Content) UrlFor(value interface{}) string {
	var url string
	if admin, ok := value.(*Admin); ok {
		url = admin.Prefix
	} else if res, ok := value.(*Resource); ok {
		url = path.Join(content.Admin.Prefix, res.Name)
	} else {
		primaryKey := content.Admin.DB.NewScope(value).PrimaryKeyValue()
		url = path.Join(content.Admin.Prefix, content.Resource.Name, fmt.Sprintf("%v", primaryKey))
	}
	return url
}

func (content *Content) LinkTo(text interface{}, link interface{}) string {
	if linkStr, ok := link.(string); ok {
		return fmt.Sprintf(`<a href="%v">%v</a>`, linkStr, text)
	}
	return fmt.Sprintf(`<a href="%v">%v</a>`, content.UrlFor(link), text)
}

func (content *Content) RenderForm(value interface{}, metas []resource.Meta) string {
	var result = bytes.NewBufferString("")
	content.renderForm(result, value, metas, []string{"QorResource"})
	return result.String()
}

func (content *Content) renderForm(result *bytes.Buffer, value interface{}, metas []resource.Meta, prefix []string) {
	for _, meta := range metas {
		content.RenderMeta(result, meta, value, prefix)
	}
}

func (content *Content) RenderMeta(writer *bytes.Buffer, meta resource.Meta, value interface{}, prefix []string) {
	prefix = append(prefix, meta.Name)

	funcsMap := content.funcMap(rules.Read, rules.Update)
	funcsMap["render_form"] = func(value interface{}, metas []resource.Meta, index ...int) string {
		var result = bytes.NewBufferString("")
		newPrefix := append([]string{}, prefix...)

		if len(index) > 0 {
			last := newPrefix[len(newPrefix)-1]
			newPrefix = append(newPrefix[:len(newPrefix)-1], fmt.Sprintf("%v[%v]", last, index[0]))
		}

		content.renderForm(result, value, metas, newPrefix)
		return result.String()
	}

	var tmpl = template.New(meta.Type + ".tmpl").Funcs(funcsMap)

	if tmpl, err := content.getTemplate(tmpl, fmt.Sprintf("forms/%v.tmpl", meta.Type)); err == nil {
		data := map[string]interface{}{}
		data["InputId"] = strings.Join(prefix, "")
		data["Label"] = meta.Label
		data["InputName"] = strings.Join(prefix, ".")
		data["Value"] = meta.Value(value, content.Context)
		if meta.GetCollection != nil {
			data["CollectionValue"] = meta.GetCollection(value, content.Context)
		}
		data["Meta"] = meta

		if err := tmpl.Execute(writer, data); err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("Form type %v not supported\n", meta.Type)
	}
}

func (content *Content) HasPrimaryKey(value interface{}, primaryKey interface{}) bool {
	primaryKeys := []interface{}{}
	reflectValue := reflect.ValueOf(value)
	if reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	if reflectValue.Kind() == reflect.Slice {
		for i := 0; i < reflectValue.Len(); i++ {
			scope := &gorm.Scope{Value: reflectValue.Index(i).Interface()}
			primaryKeys = append(primaryKeys, scope.PrimaryKeyValue())
		}
	} else if reflectValue.Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: value}
		primaryKeys = append(primaryKeys, scope.PrimaryKeyValue())
	}

	for _, key := range primaryKeys {
		if fmt.Sprintf("%v", primaryKey) == fmt.Sprintf("%v", key) {
			return true
		}
	}
	return false
}

func (content *Content) funcMap(modes ...rules.PermissionMode) template.FuncMap {
	return template.FuncMap{
		"allowed_metas":     content.AllowedMetas(modes...),
		"value_of":          content.ValueOf,
		"url_for":           content.UrlFor,
		"new_resource_path": content.NewResourcePath,
		"link_to":           content.LinkTo,
		"render_form":       content.RenderForm,
		"has_primary_key":   content.HasPrimaryKey,
	}
}
