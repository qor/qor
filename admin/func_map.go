package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/inflection"
	"github.com/qor/qor"
	"github.com/qor/qor/roles"
	"github.com/qor/qor/utils"
	"github.com/theplant/cldr"
)

func (context *Context) NewResourceContext(name ...string) *Context {
	clone := &Context{Context: context.Context, Admin: context.Admin, Result: context.Result}
	if len(name) > 0 {
		clone.setResource(context.Admin.GetResource(name[0]))
	} else {
		clone.setResource(context.Resource)
	}
	return clone
}

func (context *Context) primaryKeyOf(value interface{}) interface{} {
	return context.GetDB().NewScope(value).PrimaryKeyValue()
}

func (context *Context) isNewRecord(value interface{}) bool {
	return context.GetDB().NewRecord(value)
}

func (context *Context) ValueOf(value interface{}, meta *Meta) interface{} {
	result := meta.Valuer(value, context.Context)

	if reflectValue := reflect.ValueOf(result); reflectValue.IsValid() {
		if reflectValue.Kind() == reflect.Ptr {
			if reflectValue.IsNil() || !reflectValue.Elem().IsValid() {
				return nil
			}

			result = reflectValue.Elem().Interface()
		}

		if meta.Type == "number" || meta.Type == "float" {
			if context.isNewRecord(value) && reflect.DeepEqual(reflect.Zero(reflect.TypeOf(result)).Interface(), result) {
				return nil
			}
		}
		return result
	}

	return nil
}

func (context *Context) newResourcePath(value interface{}) string {
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
		if res == nil {
			url = ""
		} else {
			primaryKey := context.GetDB().NewScope(value).PrimaryKeyValue()
			url = path.Join(context.Admin.router.Prefix, res.ToParam(), fmt.Sprintf("%v", primaryKey))
		}
	}
	return url
}

func (context *Context) LinkTo(text interface{}, link interface{}) template.HTML {
	text = reflect.Indirect(reflect.ValueOf(text)).Interface()
	if linkStr, ok := link.(string); ok {
		return template.HTML(fmt.Sprintf(`<a href="%v">%v</a>`, linkStr, text))
	}
	return template.HTML(fmt.Sprintf(`<a href="%v">%v</a>`, context.UrlFor(link), text))
}

func (context *Context) renderIndexMeta(value interface{}, meta *Meta) template.HTML {
	var err error
	var result = bytes.NewBufferString("")
	var tmpl = template.New(meta.Type + ".tmpl").Funcs(context.FuncMap())

	if tmpl, err = context.FindTemplate(tmpl, fmt.Sprintf("metas/index/%v.tmpl", meta.Type)); err != nil {
		tmpl, _ = tmpl.Parse("{{.Value}}")
	}

	data := map[string]interface{}{"Value": context.ValueOf(value, meta), "Meta": meta}
	if err := tmpl.Execute(result, data); err != nil {
		utils.ExitWithMsg(err.Error())
	}
	return template.HTML(result.String())
}

func (context *Context) RenderForm(value interface{}, metas []*Meta) template.HTML {
	var result = bytes.NewBufferString("")
	context.renderForm(result, value, metas, []string{"QorResource"})
	return template.HTML(result.String())
}

func (context *Context) renderForm(result *bytes.Buffer, value interface{}, metas []*Meta, prefix []string) {
	for _, meta := range metas {
		context.renderMeta(result, meta, value, prefix)
	}
}

func (context *Context) renderMeta(writer *bytes.Buffer, meta *Meta, value interface{}, prefix []string) {
	prefix = append(prefix, meta.Name)

	funcsMap := context.FuncMap()
	funcsMap["render_form"] = func(value interface{}, metas []*Meta, index ...int) template.HTML {
		var result = bytes.NewBufferString("")
		newPrefix := append([]string{}, prefix...)

		if len(index) > 0 {
			last := newPrefix[len(newPrefix)-1]
			newPrefix = append(newPrefix[:len(newPrefix)-1], fmt.Sprintf("%v[%v]", last, index[0]))
		}

		context.renderForm(result, value, metas, newPrefix)
		return template.HTML(result.String())
	}

	var tmpl = template.New(meta.Type + ".tmpl").Funcs(funcsMap)

	if tmpl, err := context.FindTemplate(tmpl, fmt.Sprintf("metas/form/%v.tmpl", meta.Type)); err == nil {
		data := map[string]interface{}{}
		data["Base"] = meta.base
		data["InputId"] = strings.Join(prefix, "")
		data["Label"] = meta.Label
		data["InputName"] = strings.Join(prefix, ".")
		data["Value"] = context.ValueOf(value, meta)

		if meta.GetCollection != nil {
			data["CollectionValue"] = meta.GetCollection(value, context.Context)
		}

		data["Meta"] = meta

		if err := tmpl.Execute(writer, data); err != nil {
			panic(err)
		}
	} else {
		utils.ExitWithMsg(fmt.Sprintf("%v: form type %v not supported: got error %v", meta.Name, meta.Type, err))
	}
}

func (context *Context) isIncluded(value interface{}, hasValue interface{}) bool {
	primaryKeys := []interface{}{}
	reflectValue := reflect.Indirect(reflect.ValueOf(value))
	if reflectValue.Kind() == reflect.Slice {
		for i := 0; i < reflectValue.Len(); i++ {
			if value := reflectValue.Index(i); value.IsValid() {
				if reflect.Indirect(value).Kind() == reflect.Struct {
					scope := &gorm.Scope{Value: reflectValue.Index(i).Interface()}
					primaryKeys = append(primaryKeys, scope.PrimaryKeyValue())
				} else {
					primaryKeys = append(primaryKeys, reflect.Indirect(reflectValue.Index(i)).Interface())
				}
			}
		}
	} else if reflectValue.Kind() == reflect.Struct {
		scope := &gorm.Scope{Value: value}
		primaryKeys = append(primaryKeys, scope.PrimaryKeyValue())
	} else if reflectValue.Kind() == reflect.String {
		return strings.Contains(reflectValue.Interface().(string), fmt.Sprintf("%v", hasValue))
	} else {
		if reflectValue.IsValid() {
			primaryKeys = append(primaryKeys, reflect.Indirect(reflectValue).Interface())
		}
	}

	for _, key := range primaryKeys {
		if fmt.Sprintf("%v", hasValue) == fmt.Sprintf("%v", key) {
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

func (context *Context) allMetas(resources ...*Resource) []*Meta {
	return context.getResource(resources...).allMetas()
}

func (context *Context) indexMetas(resources ...*Resource) []*Meta {
	res := context.getResource(resources...)
	return res.allowedMetas(res.indexMetas(), context, roles.Read)
}

func (context *Context) editMetas(resources ...*Resource) []*Meta {
	res := context.getResource(resources...)
	return res.allowedMetas(res.editMetas(), context, roles.Update)
}

func (context *Context) showMetas(resources ...*Resource) []*Meta {
	res := context.getResource(resources...)
	return res.allowedMetas(res.showMetas(), context, roles.Read)
}

func (context *Context) newMetas(resources ...*Resource) []*Meta {
	res := context.getResource(resources...)
	return res.allowedMetas(res.newMetas(), context, roles.Create)
}

func (context *Context) javaScriptTag(name string) template.HTML {
	name = path.Join(context.Admin.GetRouter().Prefix, "assets", "javascripts", name+".js")
	return template.HTML(fmt.Sprintf(`<script src="%s"></script>`, name))
}

func (context *Context) styleSheetTag(name string) template.HTML {
	name = path.Join(context.Admin.GetRouter().Prefix, "assets", "stylesheets", name+".css")
	return template.HTML(fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s">`, name))
}

type scope struct {
	*Scope
	Active bool
}

type scopeMenu struct {
	Group  string
	Scopes []scope
}

func (context *Context) GetScopes() (menus []*scopeMenu) {
	scopes := context.Request.URL.Query()["scopes"]
OUT:
	for _, s := range context.Resource.scopes {
		menu := scope{Scope: s}

		for _, s := range scopes {
			if s == menu.Name {
				menu.Active = true
			}
		}

		if !menu.Default {
			if menu.Group != "" {
				for _, m := range menus {
					if m.Group == menu.Group {
						m.Scopes = append(m.Scopes, menu)
						continue OUT
					}
				}
				menus = append(menus, &scopeMenu{Group: menu.Group, Scopes: []scope{menu}})
			} else {
				menus = append(menus, &scopeMenu{Group: menu.Group, Scopes: []scope{menu}})
			}
		}
	}
	return menus
}

type HasPermissioner interface {
	HasPermission(roles.PermissionMode, *qor.Context) bool
}

func (context *Context) hasCreatePermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Create, context.Context)
}

func (context *Context) hasReadPermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Read, context.Context)
}

func (context *Context) hasUpdatePermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Update, context.Context)
}

func (context *Context) hasDeletePermission(permissioner HasPermissioner) bool {
	return permissioner.HasPermission(roles.Delete, context.Context)
}

type Page struct {
	Page       int
	Current    bool
	IsPrevious bool
	IsNext     bool
}

const (
	VISIBLE_PAGE_COUNT = 8
)

// Keep VISIBLE_PAGE_COUNT's pages visible, exclude prev and next link
// Assume there are 12 pages in total.
// When current page is 1
// [current, 2, 3, 4, 5, 6, 7, 8, next]
// When current page is 6
// [prev, 2, 3, 4, 5, current, 7, 8, 9, 10, next]
// When current page is 10
// [prev, 5, 6, 7, 8, 9, current, 11, 12]
// If total page count less than VISIBLE_PAGE_COUNT, always show all pages
func (context *Context) Pagination() *[]Page {
	pagination := context.Searcher.Pagination
	if pagination.Pages <= 1 {
		return nil
	}

	start := pagination.CurrentPage - VISIBLE_PAGE_COUNT/2
	if start < 1 {
		start = 1
	}

	end := start + VISIBLE_PAGE_COUNT - 1 // -1 for "start page" itself
	if end > pagination.Pages {
		end = pagination.Pages
	}

	if (end-start) < VISIBLE_PAGE_COUNT && start != 1 {
		start = end - VISIBLE_PAGE_COUNT + 1
	}
	if start < 1 {
		start = 1
	}

	var pages []Page
	// Append prev link
	if start > 1 {
		pages = append(pages, Page{Page: start - 1, IsPrevious: true})
	}

	for i := start; i <= end; i++ {
		pages = append(pages, Page{Page: i, Current: pagination.CurrentPage == i})
	}

	// Append next link
	if end < pagination.Pages {
		pages = append(pages, Page{Page: end + 1, IsNext: true})
	}

	return &pages
}

func equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// PatchCurrentURL is a convinent wrapper for qor/utils.PatchURL
func (context *Context) patchCurrentURL(params ...interface{}) (patchedURL string, err error) {
	return utils.PatchURL(context.Request.URL.String(), params...)
}

// PatchURL is a convinent wrapper for qor/utils.PatchURL
func (context *Context) patchURL(url string, params ...interface{}) (patchedURL string, err error) {
	return utils.PatchURL(url, params...)
}

func (context *Context) themesClass() (result string) {
	var results []string
	if context.Resource != nil {
		for _, theme := range context.Resource.Config.Themes {
			results = append(results, "qor-theme-"+theme)
		}
	}
	return strings.Join(results, " ")
}

func (context *Context) getThemes() []string {
	var themes = []string{"default"}
	if context.Resource != nil {
		themes = append(themes, context.Resource.Config.Themes...)
	}
	return themes
}

func (context *Context) loadThemeStyleSheets() template.HTML {
	var results []string
	for _, theme := range context.getThemes() {
		for _, view := range context.getViewPaths() {
			file := path.Join("assets", "stylesheets", theme+".css")
			if _, err := os.Stat(path.Join(view, file)); err == nil {
				results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s?theme=%s">`, path.Join(context.Admin.GetRouter().Prefix, file), theme))
				break
			}
		}

		for _, view := range context.getViewPaths() {
			file := path.Join("../..", "themes", theme, "assets", "stylesheets", "application.css")
			if _, err := os.Stat(path.Join(view, file)); err == nil {
				results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s?theme=%s">`, path.Join(context.Admin.GetRouter().Prefix, file), theme))
				break
			}
		}
	}

	return template.HTML(strings.Join(results, " "))
}

func (context *Context) loadThemeJavaScripts() template.HTML {
	var results []string
	var themes = []string{"default"}
	if context.Resource != nil {
		themes = append(themes, context.Resource.Config.Themes...)
	}

	for _, theme := range themes {
		for _, view := range context.getViewPaths() {
			file := path.Join("assets", "javascripts", theme+".js")
			if _, err := os.Stat(path.Join(view, file)); err == nil {
				results = append(results, fmt.Sprintf(`<script src="%s?theme=%s"></script>`, path.Join(context.Admin.GetRouter().Prefix, file), theme))
				break
			}
		}

		for _, view := range context.getViewPaths() {
			file := path.Join("../..", "themes", theme, "assets", "javascripts", "application.js")
			if _, err := os.Stat(path.Join(view, file)); err == nil {
				results = append(results, fmt.Sprintf(`<script src="%s?theme=%s"></script>`, path.Join(context.Admin.GetRouter().Prefix, file), theme))
				break
			}
		}
	}

	return template.HTML(strings.Join(results, " "))
}

func (context *Context) loadIndexActions() template.HTML {
	var actions = map[string]string{}
	var actionKeys = []string{}
	var viewPaths = context.getViewPaths()

	for j := len(viewPaths); j > 0; j-- {
		view := viewPaths[j-1]
		files, _ := filepath.Glob(path.Join(view, "actions/index", "*.tmpl"))
		for _, file := range files {
			actionKeys = append(actionKeys, path.Base(file))
			actions[path.Base(file)] = file
		}
	}

	sort.Strings(actionKeys)

	var result = bytes.NewBufferString("")
	for _, key := range actionKeys {
		file := actions[key]
		if tmpl, err := template.New(path.Base(file)).Funcs(context.FuncMap()).ParseFiles(file); err == nil {
			if err := tmpl.Execute(result, context); err != nil {
				panic(err)
			}
		}
	}
	return template.HTML(result.String())
}

func (context *Context) logoutURL() string {
	if context.Admin.auth != nil {
		return context.Admin.auth.LogoutURL(context)
	}
	return ""
}

func (context *Context) dt(key string, value string, values ...interface{}) string {
	locale := utils.GetLocale(context.Context)

	if context.Admin.I18n == nil {
		if result, err := cldr.Parse(locale, value, values...); err == nil {
			return result
		}
		return key
	} else {
		return context.Admin.I18n.Scope("qor_admin").Default(value).T(locale, key, values...)
	}
}

func (context *Context) rt(resource *Resource, key string, values ...interface{}) string {
	return context.dt(strings.Join([]string{"qor_admin", resource.ToParam(), key}, "."), key, values)
}

func (context *Context) T(key string, values ...interface{}) string {
	return context.dt(key, key, values...)
}

func (context *Context) FuncMap() template.FuncMap {
	funcMap := template.FuncMap{
		"current_user":         func() qor.CurrentUser { return context.CurrentUser },
		"get_resource":         context.GetResource,
		"new_resource_context": context.NewResourceContext,
		"is_new_record":        context.isNewRecord,
		"is_included":          context.isIncluded,
		"primary_key_of":       context.primaryKeyOf,
		"value_of":             context.ValueOf,

		"menus":      context.Admin.GetMenus,
		"get_scopes": context.GetScopes,

		"escape":    html.EscapeString,
		"raw":       func(str string) template.HTML { return template.HTML(str) },
		"equal":     equal,
		"stringify": utils.Stringify,
		"plural":    inflection.Plural,
		"singular":  inflection.Singular,

		"render":                 context.Render,
		"render_form":            context.RenderForm,
		"render_index_meta":      context.renderIndexMeta,
		"url_for":                context.UrlFor,
		"link_to":                context.LinkTo,
		"patch_current_url":      context.patchCurrentURL,
		"patch_url":              context.patchURL,
		"new_resource_path":      context.newResourcePath,
		"qor_theme_class":        context.themesClass,
		"javascript_tag":         context.javaScriptTag,
		"stylesheet_tag":         context.styleSheetTag,
		"load_theme_stylesheets": context.loadThemeStyleSheets,
		"load_theme_javascripts": context.loadThemeJavaScripts,
		"load_index_actions":     context.loadIndexActions,
		"pagination":             context.Pagination,

		"all_metas":   context.allMetas,
		"index_metas": context.indexMetas,
		"edit_metas":  context.editMetas,
		"show_metas":  context.showMetas,
		"new_metas":   context.newMetas,

		"has_create_permission": context.hasCreatePermission,
		"has_read_permission":   context.hasReadPermission,
		"has_update_permission": context.hasUpdatePermission,
		"has_delete_permission": context.hasDeletePermission,

		"logout_url": context.logoutURL,

		"marshal": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},

		"t":       context.T,
		"dt":      context.dt,
		"rt":      context.rt,
		"flashes": context.GetFlashes,
	}

	for key, value := range context.Admin.funcMaps {
		funcMap[key] = value
	}
	return funcMap
}
