package admin

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"

	"github.com/qor/qor/utils"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/roles"
)

type Context struct {
	*qor.Context
	*Searcher
	Resource    *Resource
	Admin       *Admin
	CurrentUser qor.CurrentUser
	Result      interface{}
	Content     string
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
	clone := &Context{Context: context.Context, Admin: context.Admin, Result: context.Result}
	if len(name) > 0 {
		clone.SetResource(context.Admin.GetResource(name[0]))
	} else {
		clone.SetResource(context.Resource)
	}
	return clone
}

// Template
func (context *Context) getViewPaths() (paths []string) {
	dirs := []string{context.ResourcePath(), path.Join("themes", "default"), "."}
	if context.Resource != nil && context.Resource.Config != nil && context.Resource.Config.Theme != "" {
		themePath := path.Join("themes", context.Resource.Config.Theme)
		dirs = append([]string{path.Join(themePath, context.ResourcePath()), themePath}, dirs...)
	}

	for _, p := range dirs {
		for _, d := range viewPaths {
			if isExistingDir(path.Join(d, p)) {
				paths = append(paths, path.Join(d, p))
			}
		}
	}
	return paths
}

func (context *Context) findFile(layout string) (string, error) {
	for _, p := range context.getViewPaths() {
		if _, err := os.Stat(path.Join(p, layout)); !os.IsNotExist(err) {
			return path.Join(p, layout), nil
		}
	}
	return "", errors.New("file not found")
}

func (context *Context) findTemplate(tmpl *template.Template, layout string) (*template.Template, error) {
	for _, p := range context.getViewPaths() {
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

func (context *Context) Render(name string, results ...interface{}) string {
	var err error
	names := strings.Split(name, "/")
	tmpl := template.New(names[len(names)-1] + ".tmpl").Funcs(context.funcMap())
	if len(results) > 0 {
		context.Result = results[0]
	}

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
		var err error
		tmpl = template.New("layout.tmpl").Funcs(context.funcMap())
		if tmpl, err = context.findTemplate(tmpl, "layout.tmpl"); err == nil {
			for _, name := range []string{"header", "footer"} {
				if tmpl.Lookup(name) == nil {
					tmpl, _ = context.findTemplate(tmpl, name+".tmpl")
				}
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
func (context *Context) PrimaryKeyOf(value interface{}) interface{} {
	return context.GetDB().NewScope(value).PrimaryKeyValue()
}

func (context *Context) NewRecord(value interface{}) interface{} {
	return context.GetDB().NewRecord(value)
}

func (context *Context) ValueOf(value interface{}, meta *Meta) interface{} {
	return meta.Valuer(value, context.Context)
}

func (context *Context) NewResourcePath(value interface{}) string {
	if res, ok := value.(*Resource); ok {
		return path.Join(context.Admin.router.Prefix, res.ToParam(), "new")
	} else {
		return path.Join(context.Admin.router.Prefix, context.Resource.ToParam(), "new")
	}
}

func (context *Context) UrlFor(value interface{}, resources ...*Resource) string {
	var url string
	if admin, ok := value.(*Admin); ok {
		url = admin.router.Prefix
	} else if res, ok := value.(*Resource); ok {
		url = path.Join(context.Admin.router.Prefix, res.ToParam())
	} else {
		structType := reflect.Indirect(reflect.ValueOf(value)).Type().String()
		res := context.Admin.GetResource(structType)
		primaryKey := context.GetDB().NewScope(value).PrimaryKeyValue()
		url = path.Join(context.Admin.router.Prefix, res.ToParam(), fmt.Sprintf("%v", primaryKey))
	}
	return url
}

func (context *Context) LinkTo(text interface{}, link interface{}) string {
	text = reflect.Indirect(reflect.ValueOf(text)).Interface()
	if linkStr, ok := link.(string); ok {
		return fmt.Sprintf(`<a href="%v">%v</a>`, linkStr, text)
	}
	return fmt.Sprintf(`<a href="%v">%v</a>`, context.UrlFor(link), text)
}

func (context *Context) RenderForm(value interface{}, metas []*Meta) string {
	var result = bytes.NewBufferString("")
	context.renderForm(result, value, metas, []string{"QorResource"})
	return result.String()
}

func (context *Context) renderForm(result *bytes.Buffer, value interface{}, metas []*Meta, prefix []string) {
	for _, meta := range metas {
		context.RenderMeta(result, meta, value, prefix)
	}
}

func (context *Context) RenderMeta(writer *bytes.Buffer, meta *Meta, value interface{}, prefix []string) {
	prefix = append(prefix, meta.Name)

	funcsMap := context.funcMap()
	funcsMap["render_form"] = func(value interface{}, metas []*Meta, index ...int) string {
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
		value := meta.Valuer(value, context.Context)
		if rv := reflect.ValueOf(value); rv.Kind() == reflect.Ptr && !rv.IsNil() {
			value = rv.Elem().Interface()
		}
		data["Value"] = value
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

func (context *Context) AllMetas(resources ...*Resource) []*Meta {
	return context.getResource(resources...).AllMetas()
}

func (context *Context) IndexMetas(resources ...*Resource) []*Meta {
	res := context.getResource(resources...)
	return res.AllowedMetas(res.IndexMetas(), context, roles.Read)
}

func (context *Context) EditMetas(resources ...*Resource) []*Meta {
	res := context.getResource(resources...)
	return res.AllowedMetas(res.EditMetas(), context, roles.Update)
}

func (context *Context) ShowMetas(resources ...*Resource) []*Meta {
	res := context.getResource(resources...)
	return res.AllowedMetas(res.ShowMetas(), context, roles.Read)
}

func (context *Context) NewMetas(resources ...*Resource) []*Meta {
	res := context.getResource(resources...)
	return res.AllowedMetas(res.NewMetas(), context, roles.Create)
}

func (context *Context) JavaScriptTag(name string) string {
	name = path.Join(context.Admin.GetRouter().Prefix, "assets", "javascripts", name+".js")
	return fmt.Sprintf(`<script src="%s"></script>`, name)
}

func (context *Context) StyleSheetTag(name string) string {
	name = path.Join(context.Admin.GetRouter().Prefix, "assets", "stylesheets", name+".css")
	return fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s">`, name)
}

func (context *Context) GetScopes() (scopes []string) {
	for scope, _ := range context.Resource.scopes {
		scopes = append(scopes, scope)
	}
	return
}

func (context *Context) HasUpdatePermission(meta *Meta) bool {
	return meta.HasPermission(roles.Update, context.GetContext())
}

func (context *Context) HasDeletePermission(meta *Meta) bool {
	return meta.HasPermission(roles.Delete, context.GetContext())
}

type Page struct {
	Page       int
	Current    bool
	IsPrevious bool
	IsNext     bool
}

func (context *Context) Pagination() []Page {
	pagination := context.Searcher.Pagination
	start := pagination.CurrentPage
	if start-5 < 1 {
		start = 1
	}

	end := start + 9
	if end > pagination.Pages {
		end = pagination.Pages
	}

	var pages []Page
	if start > 1 {
		pages = append(pages, Page{Page: start - 1, IsPrevious: true})
	}

	for i := start; i <= end; i++ {
		pages = append(pages, Page{Page: i, Current: pagination.CurrentPage == i})
	}

	if end < pagination.Pages {
		pages = append(pages, Page{Page: end + 1, IsNext: true})
	}

	return pages
}

func Equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func (context *Context) funcMap() template.FuncMap {
	funcMap := template.FuncMap{
		"menus":                 context.Admin.GetMenus,
		"current_user":          func() qor.CurrentUser { return context.CurrentUser },
		"render":                context.Render,
		"render_form":           context.RenderForm,
		"url_for":               context.UrlFor,
		"link_to":               context.LinkTo,
		"new_resource_path":     context.NewResourcePath,
		"new_resource":          context.NewResource,
		"is_new_record":         context.NewRecord,
		"value_of":              context.ValueOf,
		"primary_key_of":        context.PrimaryKeyOf,
		"get_scopes":            context.GetScopes,
		"has_primary_key":       context.HasPrimaryKey,
		"all_metas":             context.AllMetas,
		"index_metas":           context.IndexMetas,
		"edit_metas":            context.EditMetas,
		"show_metas":            context.ShowMetas,
		"new_metas":             context.NewMetas,
		"pagination":            context.Pagination,
		"javascript_tag":        context.JavaScriptTag,
		"stylesheet_tag":        context.StyleSheetTag,
		"equal":                 Equal,
		"patch_current_url":     context.PatchCurrentURL,
		"has_update_permission": context.HasUpdatePermission,
		"has_delete_permission": context.HasDeletePermission,
	}

	for key, value := range context.Admin.funcMaps {
		funcMap[key] = value
	}
	return funcMap
}

// PatchCurrentURL is a convinent wrapper for qor/utils.PatchCurrentURL
func (context *Context) PatchCurrentURL(params ...interface{}) (patchedURL string, err error) {
	return utils.PatchURL(context.Request.URL.String(), params...)
}
