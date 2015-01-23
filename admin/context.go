package admin

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type Context struct {
	*qor.Context
	*Searcher
	Resource *Resource
	Admin    *Admin
	Writer   http.ResponseWriter
	Result   interface{}
	Content  string
}

// Resource
func (context *Context) ResourcePath() string {
	if context.Resource == nil {
		return ""
	}
	return context.Resource.ToParam()
}

func (context *Context) SetResource(res *Resource) *Context {
	context.Resource = res
	context.Searcher = &Searcher{Context: context}
	return context
}

func (context *Context) NewResource(name ...string) *Context {
	clone := &Context{Context: context.Context, Admin: context.Admin, Writer: context.Writer, Result: context.Result}
	if len(name) > 0 {
		clone.SetResource(context.Admin.GetResource(name[0]))
	} else {
		clone.SetResource(context.Resource)
	}
	return clone
}

// Template
func (context *Context) findFile(layout string) (string, error) {
	paths := []string{}
	for _, p := range []string{context.ResourcePath(), path.Join("themes", "default"), "."} {
		for _, d := range viewPaths {
			if isExistingDir(path.Join(d, p)) {
				paths = append(paths, path.Join(d, p))
			}
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(path.Join(p, layout)); !os.IsNotExist(err) {
			return path.Join(p, layout), nil
		}
	}
	return "", errors.New("file not found")
}

func (context *Context) findTemplate(tmpl *template.Template, layout string) (*template.Template, error) {
	paths := []string{}
	for _, p := range []string{context.ResourcePath(), path.Join("themes", "default"), "."} {
		for _, d := range viewPaths {
			if isExistingDir(path.Join(d, p)) {
				paths = append(paths, path.Join(d, p))
			}
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(path.Join(p, layout)); !os.IsNotExist(err) {
			if tmpl, err = tmpl.ParseFiles(path.Join(p, layout)); err != nil {
				fmt.Println(err)
			} else {
				return tmpl, nil
			}
		}
	}
	return tmpl, errors.New("template not found")
}

func (context *Context) Render(name string, result interface{}) string {
	var err error
	names := strings.Split(name, "/")
	tmpl := template.New(names[len(names)-1] + ".tmpl").Funcs(context.funcMap())
	context.Result = result

	if tmpl, err = context.findTemplate(tmpl, name+".tmpl"); err == nil {
		var result = bytes.NewBufferString("")
		if err := tmpl.Execute(result, context); err != nil {
			fmt.Println(err)
		}
		return result.String()
	}

	return ""
}

func (context *Context) Execute(name string, result interface{}) {
	var tmpl *template.Template
	var cacheKey string

	if context.Resource != nil {
		cacheKey = path.Join(context.ResourcePath(), name)
	} else {
		cacheKey = name
	}

	if t, ok := templates[cacheKey]; !ok || true {
		tmpl, _ = context.findTemplate(tmpl, "layout.tmpl")
		tmpl = tmpl.Funcs(context.funcMap())

		for _, name := range []string{"header", "footer"} {
			if tmpl.Lookup(name) == nil {
				tmpl, _ = context.findTemplate(tmpl, name+".tmpl")
			}
		}
	} else {
		tmpl = t
	}

	context.Content = context.Render(name, result)
	if err := tmpl.Execute(context.Writer, context); err != nil {
		fmt.Println(err)
	}
}

// Function Maps
func (context *Context) ValueOf(value interface{}, meta *resource.Meta) interface{} {
	return meta.Value(value, context.Context)
}

func (context *Context) NewResourcePath(value interface{}) string {
	if res, ok := value.(*Resource); ok {
		return path.Join(context.Admin.router.Prefix, res.Name, "new")
	} else {
		return path.Join(context.Admin.router.Prefix, context.Resource.Name, "new")
	}
}

func (context *Context) UrlFor(value interface{}, resources ...*Resource) string {
	var url string
	if admin, ok := value.(*Admin); ok {
		url = admin.router.Prefix
	} else if resource, ok := value.(*Resource); ok {
		url = path.Join(context.Admin.router.Prefix, resource.ToParam())
	} else {
		primaryKey := context.GetDB().NewScope(value).PrimaryKeyValue()
		name := NewResource(value).ToParam()
		url = path.Join(context.Admin.router.Prefix, name, fmt.Sprintf("%v", primaryKey))
	}
	return url
}

func (context *Context) LinkTo(text interface{}, link interface{}) string {
	if linkStr, ok := link.(string); ok {
		return fmt.Sprintf(`<a href="%v">%v</a>`, linkStr, text)
	}
	return fmt.Sprintf(`<a href="%v">%v</a>`, context.UrlFor(link), text)
}

func (context *Context) RenderForm(value interface{}, metas []*resource.Meta) string {
	var result = bytes.NewBufferString("")
	context.renderForm(result, value, metas, []string{"QorResource"})
	return result.String()
}

func (context *Context) renderForm(result *bytes.Buffer, value interface{}, metas []*resource.Meta, prefix []string) {
	for _, meta := range metas {
		context.RenderMeta(result, meta, value, prefix)
	}
}

func (context *Context) RenderMeta(writer *bytes.Buffer, meta *resource.Meta, value interface{}, prefix []string) {
	prefix = append(prefix, meta.Name)

	funcsMap := context.funcMap()
	funcsMap["render_form"] = func(value interface{}, metas []*resource.Meta, index ...int) string {
		var result = bytes.NewBufferString("")
		newPrefix := append([]string{}, prefix...)

		if len(index) > 0 {
			last := newPrefix[len(newPrefix)-1]
			newPrefix = append(newPrefix[:len(newPrefix)-1], fmt.Sprintf("%v[%v]", last, index[0]))
		}

		context.renderForm(result, value, metas, newPrefix)
		return result.String()
	}

	var tmpl = template.New(meta.Type + ".tmpl").Funcs(funcsMap)

	if tmpl, err := context.findTemplate(tmpl, fmt.Sprintf("forms/%v.tmpl", meta.Type)); err == nil {
		data := map[string]interface{}{}
		data["InputId"] = strings.Join(prefix, "")
		data["Label"] = meta.Label
		data["InputName"] = strings.Join(prefix, ".")
		data["Value"] = meta.Value(value, context.Context)
		if meta.GetCollection != nil {
			data["CollectionValue"] = meta.GetCollection(value, context.Context)
		}
		data["Meta"] = meta

		if err := tmpl.Execute(writer, data); err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("%v: form type %v not supported\n", meta.Name, meta.Type)
	}
}

func (context *Context) HasPrimaryKey(value interface{}, primaryKey interface{}) bool {
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

func (context *Context) getResource(resources ...*Resource) *Resource {
	for _, res := range resources {
		return res
	}
	return context.Resource
}

func (context *Context) AllMetas(resources ...*Resource) []*resource.Meta {
	return context.getResource(resources...).AllMetas()
}

func (context *Context) IndexMetas(resources ...*Resource) []*resource.Meta {
	res := context.getResource(resources...)
	return res.AllowedMetas(res.IndexMetas(), context, roles.Read)
}

func (context *Context) EditMetas(resources ...*Resource) []*resource.Meta {
	res := context.getResource(resources...)
	return res.AllowedMetas(res.EditMetas(), context, roles.Update)
}

func (context *Context) ShowMetas(resources ...*Resource) []*resource.Meta {
	res := context.getResource(resources...)
	return res.AllowedMetas(res.ShowMetas(), context, roles.Read)
}

func (context *Context) NewMetas(resources ...*Resource) []*resource.Meta {
	res := context.getResource(resources...)
	return res.AllowedMetas(res.NewMetas(), context, roles.Create)
}

func (context *Context) JavaScriptTag(name string) string {
	name = path.Join(context.Admin.GetRouter().Prefix, "assets", "javascripts", name+".js")
	return fmt.Sprintf(`<script src="%s"></script>`, name)
}

func (context *Context) StyleSheetTag(name string) string {
	name = path.Join(context.Admin.GetRouter().Prefix, "assets", "stylesheets", name+".css")
	return fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s"%s>`, name)
}

func (context *Context) funcMap() template.FuncMap {
	funcMap := template.FuncMap{
		"value_of":          context.ValueOf,
		"url_for":           context.UrlFor,
		"new_resource_path": context.NewResourcePath,
		"link_to":           context.LinkTo,
		"render_form":       context.RenderForm,
		"has_primary_key":   context.HasPrimaryKey,
		"new_resource":      context.NewResource,
		"all_metas":         context.AllMetas, // Resource Metas
		"index_metas":       context.IndexMetas,
		"edit_metas":        context.EditMetas,
		"show_metas":        context.ShowMetas,
		"new_metas":         context.NewMetas,
		"javascript_tag":    context.JavaScriptTag,
		"stylesheet_tag":    context.StyleSheetTag,
	}
	for key, value := range context.Admin.funcMaps {
		funcMap[key] = value
	}
	return funcMap
}
