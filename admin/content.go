package admin

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"

	"text/template"
)

type Content struct {
	Content    string
	Context    *Context
	Resource   *Resource
	Result     interface{}
	Action     string
	Permission map[string]roles.PermissionMode
}

func (content *Content) Admin() *Admin {
	return content.Context.Admin
}

func (content *Content) Render(name string) string {
	var err error
	var tmpl = template.New(name + ".tmpl").Funcs(content.funcMap())

	if tmpl, err = content.getTemplate(tmpl, name+".tmpl"); err == nil {
		var result = bytes.NewBufferString("")
		if err := tmpl.Execute(result, content); err != nil {
			fmt.Println(err)
		}
		return result.String()
	}

	return ""
}

func (content *Content) Execute(name string) {
	var tmpl *template.Template

	cacheKey := path.Join(content.Context.ResourceName, name)
	if t, ok := templates[cacheKey]; !ok || true {
		tmpl, _ = content.getTemplate(tmpl, "layout.tmpl")
		tmpl = tmpl.Funcs(content.funcMap())

		for _, name := range []string{"header", "footer"} {
			if tmpl.Lookup(name) == nil {
				tmpl, _ = content.getTemplate(tmpl, name+".tmpl")
			}
		}
	} else {
		tmpl = t
	}

	content.Content = content.Render(name)
	if err := tmpl.Execute(content.Context.Writer, content); err != nil {
		fmt.Println(err)
	}
}

func (content *Content) RenderIndex() string {
	return ""
}

func (content *Content) AllowedMetas(modes ...roles.PermissionMode) func(reses ...*Resource) []*resource.Meta {
	return func(reses ...*Resource) []*resource.Meta {
		var res = content.Resource
		if len(reses) > 0 {
			res = reses[0]
		}

		switch content.Action {
		case "index":
			return res.AllowedMetas(res.IndexMetas(), content.Context, modes...)
		case "show":
			return res.AllowedMetas(res.ShowMetas(), content.Context, modes...)
		case "edit":
			return res.AllowedMetas(res.EditMetas(), content.Context, modes...)
		case "new":
			return res.AllowedMetas(res.NewMetas(), content.Context, modes...)
		default:
			return []*resource.Meta{}
		}
	}
}

func (content *Content) ValueOf(value interface{}, meta *resource.Meta) interface{} {
	return meta.Value(value, content.Context.Context)
}

func (content *Content) NewResourcePath(value interface{}) string {
	if res, ok := value.(*Resource); ok {
		return path.Join(content.Admin().router.Prefix, res.Name, "new")
	} else {
		return path.Join(content.Admin().router.Prefix, content.Resource.Name, "new")
	}
}

func (content *Content) UrlFor(value interface{}, resources ...*Resource) string {
	var url string
	if admin, ok := value.(*Admin); ok {
		url = admin.router.Prefix
	} else if resource, ok := value.(*Resource); ok {
		url = path.Join(content.Admin().router.Prefix, resource.Name)
	} else {
		primaryKey := content.Context.GetDB().NewScope(value).PrimaryKeyValue()
		name := strings.ToLower(reflect.Indirect(reflect.ValueOf(value)).Type().Name())
		url = path.Join(content.Admin().router.Prefix, name, fmt.Sprintf("%v", primaryKey))
	}
	return url
}

func (content *Content) LinkTo(text interface{}, link interface{}) string {
	if linkStr, ok := link.(string); ok {
		return fmt.Sprintf(`<a href="%v">%v</a>`, linkStr, text)
	}
	return fmt.Sprintf(`<a href="%v">%v</a>`, content.UrlFor(link), text)
}

func (content *Content) RenderForm(value interface{}, metas []*resource.Meta) string {
	var result = bytes.NewBufferString("")
	content.renderForm(result, value, metas, []string{"QorResource"})
	return result.String()
}

func (content *Content) renderForm(result *bytes.Buffer, value interface{}, metas []*resource.Meta, prefix []string) {
	for _, meta := range metas {
		content.RenderMeta(result, meta, value, prefix)
	}
}

func (content *Content) RenderMeta(writer *bytes.Buffer, meta *resource.Meta, value interface{}, prefix []string) {
	prefix = append(prefix, meta.Name)

	funcsMap := content.funcMap(roles.Read, roles.Update)
	funcsMap["render_form"] = func(value interface{}, metas []*resource.Meta, index ...int) string {
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
		data["Value"] = meta.Value(value, content.Context.Context)
		if meta.GetCollection != nil {
			data["CollectionValue"] = meta.GetCollection(value, content.Context.Context)
		}
		data["Meta"] = meta

		if err := tmpl.Execute(writer, data); err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("%v: form type %v not supported\n", meta.Name, meta.Type)
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

func (content *Content) funcMap(modes ...roles.PermissionMode) template.FuncMap {
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
