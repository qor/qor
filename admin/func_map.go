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
	"regexp"
	"sort"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

func (context *Context) NewResourceContext(name ...interface{}) *Context {
	clone := &Context{Context: context.Context, Admin: context.Admin, Result: context.Result, Action: context.Action}
	if len(name) > 0 {
		if str, ok := name[0].(string); ok {
			clone.setResource(context.Admin.GetResource(str))
		} else if res, ok := name[0].(*Resource); ok {
			clone.setResource(res)
		}
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

func (context *Context) newResourcePath(res *Resource) string {
	return path.Join(context.UrlFor(res), "new")
}

func (context *Context) UrlFor(value interface{}, resources ...*Resource) string {
	getPrefix := func(res *Resource) string {
		var params string
		for res.base != nil {
			params = path.Join(res.base.ToParam(), res.base.GetPrimaryValue(context.Request), params)
			res = res.base
		}
		return path.Join(res.GetAdmin().router.Prefix, params)
	}

	if admin, ok := value.(*Admin); ok {
		return admin.router.Prefix
	} else if res, ok := value.(*Resource); ok {
		return path.Join(getPrefix(res), res.ToParam())
	} else {
		var res *Resource

		if len(resources) > 0 {
			res = resources[0]
		}

		if res == nil {
			res = context.Admin.GetResource(reflect.Indirect(reflect.ValueOf(value)).Type().String())
		}

		if res != nil {
			if res.Config.Singleton {
				return path.Join(getPrefix(res), res.ToParam())
			} else {
				primaryKey := fmt.Sprint(context.GetDB().NewScope(value).PrimaryKeyValue())
				return path.Join(getPrefix(res), res.ToParam(), primaryKey)
			}
		}
	}
	return ""
}

func (context *Context) LinkTo(text interface{}, link interface{}) template.HTML {
	text = reflect.Indirect(reflect.ValueOf(text)).Interface()
	if linkStr, ok := link.(string); ok {
		return template.HTML(fmt.Sprintf(`<a href="%v">%v</a>`, linkStr, text))
	}
	return template.HTML(fmt.Sprintf(`<a href="%v">%v</a>`, context.UrlFor(link), text))
}

func (context *Context) valueOf(valuer func(interface{}, *qor.Context) interface{}, value interface{}, meta *Meta) interface{} {
	if valuer != nil {
		reflectValue := reflect.ValueOf(value)
		if reflectValue.Kind() != reflect.Ptr {
			reflectPtr := reflect.New(reflectValue.Type())
			reflectPtr.Elem().Set(reflectValue)
			value = reflectPtr.Interface()
		}

		result := valuer(value, context.Context)

		if reflectValue := reflect.ValueOf(result); reflectValue.IsValid() {
			if reflectValue.Kind() == reflect.Ptr {
				if reflectValue.IsNil() || !reflectValue.Elem().IsValid() {
					return nil
				}

				result = reflectValue.Elem().Interface()
			}

			if meta.Type == "number" || meta.Type == "float" {
				if context.isNewRecord(value) && equal(reflect.Zero(reflect.TypeOf(result)).Interface(), result) {
					return nil
				}
			}
			return result
		} else {
			return nil
		}
	}

	utils.ExitWithMsg(fmt.Sprintf("No valuer found for meta %v of resource %v", meta.Name, meta.baseResource.Name))
	return nil
}

func (context *Context) RawValueOf(value interface{}, meta *Meta) interface{} {
	return context.valueOf(meta.GetValuer(), value, meta)
}

func (context *Context) FormattedValueOf(value interface{}, meta *Meta) interface{} {
	return context.valueOf(meta.GetFormattedValuer(), value, meta)
}

func (context *Context) renderForm(value interface{}, sections []*Section) template.HTML {
	var result = bytes.NewBufferString("")
	context.renderSections(value, sections, []string{"QorResource"}, result, "form")
	return template.HTML(result.String())
}

func (context *Context) renderSections(value interface{}, sections []*Section, prefix []string, writer *bytes.Buffer, kind string) {
	for _, section := range sections {
		var rows []struct {
			Length      int
			ColumnsHTML template.HTML
		}

		for _, column := range section.Rows {
			columnsHTML := bytes.NewBufferString("")
			for _, col := range column {
				meta := section.Resource.GetMetaOrNew(col)
				if meta != nil {
					context.RenderMeta(meta, value, prefix, kind, columnsHTML)
				}
			}

			rows = append(rows, struct {
				Length      int
				ColumnsHTML template.HTML
			}{
				Length:      len(column),
				ColumnsHTML: template.HTML(string(columnsHTML.Bytes())),
			})
		}

		var data = map[string]interface{}{
			"Title": template.HTML(section.Title),
			"Rows":  rows,
		}
		if file, err := context.FindTemplate("metas/section.tmpl"); err == nil {
			if tmpl, err := template.New(filepath.Base(file)).Funcs(context.FuncMap()).ParseFiles(file); err == nil {
				tmpl.Execute(writer, data)
			}
		}
	}
}

func (context *Context) RenderMeta(meta *Meta, value interface{}, prefix []string, metaType string, writer *bytes.Buffer) {
	var (
		tmpl     *template.Template
		err      error
		funcsMap = context.FuncMap()
	)
	prefix = append(prefix, meta.Name)

	var generateNestedRenderSections = func(kind string) func(interface{}, []*Section, ...int) template.HTML {
		return func(value interface{}, sections []*Section, index ...int) template.HTML {
			var result = bytes.NewBufferString("")
			var newPrefix = append([]string{}, prefix...)

			if len(index) > 0 {
				last := newPrefix[len(newPrefix)-1]
				newPrefix = append(newPrefix[:len(newPrefix)-1], fmt.Sprintf("%v[%v]", last, index[0]))
			}

			if len(sections) > 0 {
				for _, field := range context.GetDB().NewScope(value).PrimaryFields() {
					if meta := sections[0].Resource.GetMetaOrNew(field.Name); meta != nil {
						context.RenderMeta(meta, value, newPrefix, kind, result)
					}
				}

				context.renderSections(value, sections, newPrefix, result, kind)
			}

			return template.HTML(result.String())
		}
	}

	funcsMap["render_form"] = generateNestedRenderSections("form")

	if file, err := context.FindTemplate(fmt.Sprintf("metas/%v/%v.tmpl", metaType, meta.Name), fmt.Sprintf("metas/%v/%v.tmpl", metaType, meta.Type)); err == nil {
		defer func() {
			if r := recover(); r != nil {
				writer.Write([]byte(fmt.Sprintf("Get error when render template %v meta %v: %v", file, meta.Name, r)))
			}
		}()

		tmpl, err = template.New(filepath.Base(file)).Funcs(funcsMap).ParseFiles(file)
	} else {
		tmpl, err = template.New(meta.Type + ".tmpl").Funcs(funcsMap).Parse("{{.Value}}")
	}

	if err == nil {
		var scope = context.GetDB().NewScope(value)
		var data = map[string]interface{}{
			"Context":       context,
			"BaseResource":  meta.baseResource,
			"ResourceValue": value,
			"InputId":       fmt.Sprintf("%v_%v_%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue(), meta.Name),
			"Label":         meta.Label,
			"InputName":     strings.Join(prefix, "."),
			"Value":         context.FormattedValueOf(value, meta),
			"Meta":          meta,
		}

		if meta.GetCollection != nil {
			data["CollectionValue"] = meta.GetCollection(value, context.Context)
		}

		err = tmpl.Execute(writer, data)
	}

	if err != nil {
		utils.ExitWithMsg(fmt.Sprintf("got error when render %v template for %v(%v):%v", metaType, meta.Name, meta.Type, err))
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

func (context *Context) indexSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.IndexAttrs(), context, roles.Read)
}

func (context *Context) editSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.EditAttrs(), context, roles.Read)
}

func (context *Context) newSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.NewAttrs(), context, roles.Create)
}

func (context *Context) showSections(resources ...*Resource) []*Section {
	res := context.getResource(resources...)
	return res.allowedSections(res.ShowAttrs(), context, roles.Read)
}

type menu struct {
	*Menu
	Active   bool
	SubMenus []*menu
}

func (context *Context) getMenus() (menus []*menu) {
	var globalMenu = &menu{}
	var mostMatchedMenu *menu
	var mostMatchedLength int

	var addMenu func(parent *menu, menus []*Menu)
	addMenu = func(parent *menu, menus []*Menu) {
		for _, m := range menus {
			menu := &menu{Menu: m}
			if strings.Contains(context.Request.URL.Path, m.Link) && len(m.Link) > mostMatchedLength {
				mostMatchedMenu = menu
				mostMatchedLength = len(m.Link)
			}
			addMenu(menu, menu.GetSubMenus())
			parent.SubMenus = append(parent.SubMenus, menu)
		}
	}

	addMenu(globalMenu, context.Admin.GetMenus())

	if mostMatchedMenu != nil {
		mostMatchedMenu.Active = true
	}

	return globalMenu.SubMenus
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
	if context.Resource == nil {
		return
	}

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

func (context *Context) javaScriptTag(names ...string) template.HTML {
	var results []string
	for _, name := range names {
		name = path.Join(context.Admin.GetRouter().Prefix, "assets", "javascripts", name+".js")
		results = append(results, fmt.Sprintf(`<script src="%s"></script>`, name))
	}
	return template.HTML(strings.Join(results, ""))
}

func (context *Context) styleSheetTag(names ...string) template.HTML {
	var results []string
	for _, name := range names {
		name = path.Join(context.Admin.GetRouter().Prefix, "assets", "stylesheets", name+".css")
		results = append(results, fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s">`, name))
	}
	return template.HTML(strings.Join(results, ""))
}

func (context *Context) getThemes() (themes []string) {
	if context.Resource != nil {
		themes = append(themes, context.Resource.Config.Themes...)
	}
	return
}

func (context *Context) loadThemeStyleSheets() template.HTML {
	var results []string
	for _, theme := range context.getThemes() {
		var file = path.Join("assets", "stylesheets", theme+".css")
		for _, view := range context.getViewPaths() {
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
	for _, theme := range context.getThemes() {
		var file = path.Join("assets", "javascripts", theme+".js")
		for _, view := range context.getViewPaths() {
			if _, err := os.Stat(path.Join(view, file)); err == nil {
				results = append(results, fmt.Sprintf(`<script src="%s?theme=%s"></script>`, path.Join(context.Admin.GetRouter().Prefix, file), theme))
				break
			}
		}
	}

	return template.HTML(strings.Join(results, " "))
}

func (context *Context) loadAdminJavaScripts() template.HTML {
	var siteName = context.Admin.SiteName
	if siteName == "" {
		siteName = "application"
	}

	var file = path.Join("assets", "javascripts", strings.ToLower(strings.Replace(siteName, " ", "_", -1))+".js")
	for _, view := range context.getViewPaths() {
		if _, err := os.Stat(path.Join(view, file)); err == nil {
			return template.HTML(fmt.Sprintf(`<script src="%s"></script>`, path.Join(context.Admin.GetRouter().Prefix, file)))
		}
	}
	return ""
}

func (context *Context) loadAdminStyleSheets() template.HTML {
	var siteName = context.Admin.SiteName
	if siteName == "" {
		siteName = "application"
	}

	var file = path.Join("assets", "stylesheets", strings.ToLower(strings.Replace(siteName, " ", "_", -1))+".css")
	for _, view := range context.getViewPaths() {
		if _, err := os.Stat(path.Join(view, file)); err == nil {
			return template.HTML(fmt.Sprintf(`<link type="text/css" rel="stylesheet" href="%s">`, path.Join(context.Admin.GetRouter().Prefix, file)))
		}
	}
	return ""
}

func (context *Context) loadActions(action string) template.HTML {
	var actions = map[string]string{}
	var actionKeys = []string{}
	var viewPaths = context.getViewPaths()

	for j := len(viewPaths); j > 0; j-- {
		view := viewPaths[j-1]
		globalfiles, _ := filepath.Glob(path.Join(view, "actions/*.tmpl"))
		files, _ := filepath.Glob(path.Join(view, "actions", action, "*.tmpl"))

		for _, file := range append(globalfiles, files...) {
			base := regexp.MustCompile("^\\d+\\.").ReplaceAllString(path.Base(file), "")
			if _, ok := actions[base]; !ok {
				actionKeys = append(actionKeys, path.Base(file))
			}
			actions[base] = file
		}
	}

	sort.Strings(actionKeys)

	var result = bytes.NewBufferString("")
	for _, key := range actionKeys {
		base := regexp.MustCompile("^\\d+\\.").ReplaceAllString(key, "")
		file := actions[base]
		if tmpl, err := template.New(filepath.Base(file)).Funcs(context.FuncMap()).ParseFiles(file); err == nil {
			if err := tmpl.Execute(result, context); err != nil {
				panic(err)
			}
		}
	}
	return template.HTML(strings.TrimSpace(result.String()))
}

func (context *Context) logoutURL() string {
	if context.Admin.auth != nil {
		return context.Admin.auth.LogoutURL(context)
	}
	return ""
}

func (context *Context) t(values ...interface{}) template.HTML {
	switch len(values) {
	case 1:
		return context.Admin.T(context.Context, values[0].(string), "")
	case 2:
		return context.Admin.T(context.Context, values[0].(string), values[1].(string))
	case 3:
		return context.Admin.T(context.Context, values[0].(string), values[1].(string), values[2:len(values)]...)
	default:
		utils.ExitWithMsg("T have a wrong params")
	}
	return ""
}

func (context *Context) isSortableMeta(meta *Meta) bool {
	for _, attr := range context.Resource.SortableAttrs() {
		if attr == meta.Name && meta.FieldStruct != nil && meta.FieldStruct.IsNormal && meta.FieldStruct.DBName != "" {
			return true
		}
	}
	return false
}

func (context *Context) convertSectionToMetas(res *Resource, sections []*Section) []*Meta {
	return res.ConvertSectionToMetas(sections)
}

type formatedError struct {
	Label  string
	Errors []string
}

func (context *Context) getFormattedErrors() (formatedErrors []formatedError) {
	type labelInterface interface {
		Label() string
	}

	for _, err := range context.GetErrors() {
		if labelErr, ok := err.(labelInterface); ok {
			var found bool
			label := labelErr.Label()
			for _, formatedError := range formatedErrors {
				if formatedError.Label == label {
					formatedError.Errors = append(formatedError.Errors, err.Error())
				}
			}
			if !found {
				formatedErrors = append(formatedErrors, formatedError{Label: label, Errors: []string{err.Error()}})
			}
		} else {
			formatedErrors = append(formatedErrors, formatedError{Errors: []string{err.Error()}})
		}
	}
	return
}

func (context *Context) AllowedActions(actions []*Action, mode string) []*Action {
	var allowedActions []*Action
	for _, action := range actions {
		if action.HasPermission(roles.Update, context.Context) {
			if len(action.Visibles) == 0 {
				allowedActions = append(allowedActions, action)
			} else {
				for _, v := range action.Visibles {
					if v == mode {
						allowedActions = append(allowedActions, action)
						break
					}
				}
			}
		}
	}
	return allowedActions
}

func (context *Context) pageTitle() template.HTML {
	if context.Resource == nil {
		return context.t("qor_admin.layout.title", "Admin")
	}
	if context.Action == "action" {
		return context.t(fmt.Sprintf("%v.actions.%v", context.Resource.ToParam(), context.Result.(*Action).Label), context.Result.(*Action).Label)
	}
	var value string
	var resourceKey, resourceName string
	titleKey := fmt.Sprintf("qor_admin.form.%v_title", context.Action)
	if context.Resource.Config.Singleton {
		resourceKey = fmt.Sprintf("%v.name", context.Resource.ToParam())
		resourceName = string(context.t(resourceKey, context.Resource.Name))
	} else {
		resourceKey = fmt.Sprintf("%v.name.plural", context.Resource.ToParam())
		resourceName = string(context.t(resourceKey, inflection.Plural(context.Resource.Name)))
	}
	switch context.Action {
	case "new":
		value = "Add {{$1}}"
	case "edit":
		value = "Edit {{$1}}"
	case "show":
		value = "{{$1}} Details"
	default:
		value = "{{$1}}"
	}
	return context.t(titleKey, value, resourceName)
}

func (context *Context) FuncMap() template.FuncMap {
	funcMap := template.FuncMap{
		"current_user":         func() qor.CurrentUser { return context.CurrentUser },
		"get_resource":         context.GetResource,
		"new_resource_context": context.NewResourceContext,
		"is_new_record":        context.isNewRecord,
		"is_included":          context.isIncluded,
		"primary_key_of":       context.primaryKeyOf,
		"formatted_value_of":   context.FormattedValueOf,
		"raw_value_of":         context.RawValueOf,

		"get_menus":            context.getMenus,
		"get_scopes":           context.GetScopes,
		"get_formatted_errors": context.getFormattedErrors,

		"escape":    html.EscapeString,
		"raw":       func(str string) template.HTML { return template.HTML(str) },
		"equal":     equal,
		"stringify": utils.Stringify,
		"plural":    inflection.Plural,
		"singular":  inflection.Singular,

		"render":      context.Render,
		"render_form": context.renderForm,
		"render_index_meta": func(value interface{}, meta *Meta) template.HTML {
			var result = bytes.NewBufferString("")
			context.RenderMeta(meta, value, []string{}, "index", result)
			return template.HTML(result.String())
		},
		"url_for":                context.UrlFor,
		"link_to":                context.LinkTo,
		"new_resource_path":      context.newResourcePath,
		"search_center_path":     func() string { return path.Join(context.Admin.router.Prefix, "!search") },
		"patch_current_url":      context.patchCurrentURL,
		"patch_url":              context.patchURL,
		"qor_theme_class":        context.themesClass,
		"javascript_tag":         context.javaScriptTag,
		"stylesheet_tag":         context.styleSheetTag,
		"load_theme_stylesheets": context.loadThemeStyleSheets,
		"load_theme_javascripts": context.loadThemeJavaScripts,
		"load_admin_stylesheets": context.loadAdminStyleSheets,
		"load_admin_javascripts": context.loadAdminJavaScripts,
		"load_actions":           context.loadActions,
		"allowed_actions":        context.AllowedActions,
		"pagination":             context.Pagination,

		"page_title": context.pageTitle,
		"meta_label": func(meta *Meta) template.HTML {
			key := fmt.Sprintf("%v.attributes.%v", meta.baseResource.ToParam(), meta.Label)
			return context.Admin.T(context.Context, key, meta.Label)
		},
		"all_metas":                 context.allMetas,
		"index_sections":            context.indexSections,
		"show_sections":             context.showSections,
		"new_sections":              context.newSections,
		"edit_sections":             context.editSections,
		"is_sortable_meta":          context.isSortableMeta,
		"convert_sections_to_metas": context.convertSectionToMetas,

		"has_create_permission": context.hasCreatePermission,
		"has_read_permission":   context.hasReadPermission,
		"has_update_permission": context.hasUpdatePermission,
		"has_delete_permission": context.hasDeletePermission,

		"logout_url": context.logoutURL,

		"marshal": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},

		"t":       context.t,
		"flashes": context.GetFlashes,
	}

	for key, value := range context.Admin.funcMaps {
		funcMap[key] = value
	}
	return funcMap
}
