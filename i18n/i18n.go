package i18n

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/utils"
	"github.com/theplant/cldr"
)

var Default = "en-US"

type I18n struct {
	scope        string
	value        string
	Backends     []Backend
	Translations map[string]map[string]*Translation
	IsInlineEdit bool
}

func (I18n) ResourceName() string {
	return "Translation"
}

type Backend interface {
	LoadTranslations() []*Translation
	SaveTranslation(*Translation) error
	DeleteTranslation(*Translation) error
}

type Translation struct {
	Key     string
	Locale  string
	Value   string
	Backend Backend
}

func New(backends ...Backend) *I18n {
	i18n := &I18n{Backends: backends, Translations: map[string]map[string]*Translation{}}
	for i := len(backends) - 1; i >= 0; i-- {
		var backend = backends[i]
		for _, translation := range backend.LoadTranslations() {
			translation.Backend = backend
			i18n.AddTranslation(translation)
		}
	}
	return i18n
}

func (i18n *I18n) AddTranslation(translation *Translation) {
	if i18n.Translations[translation.Locale] == nil {
		i18n.Translations[translation.Locale] = map[string]*Translation{}
	}
	i18n.Translations[translation.Locale][translation.Key] = translation
}

func (i18n *I18n) SaveTranslation(translation *Translation) error {
	var backends []Backend
	if backend := translation.Backend; backend != nil {
		backends = append(backends, backend)
	}

	for _, backend := range i18n.Backends {
		if backend.SaveTranslation(translation) == nil {
			i18n.AddTranslation(translation)
			return nil
		}
	}

	return errors.New("failed to save translation")
}

func (i18n *I18n) DeleteTranslation(translation *Translation) error {
	err := translation.Backend.DeleteTranslation(translation)
	if err == nil {
		delete(i18n.Translations[translation.Locale], translation.Key)
	}
	return err
}

func (i18n *I18n) EnableInlineEdit(isInlineEdit bool) admin.I18n {
	return &I18n{Translations: i18n.Translations, scope: i18n.scope, value: i18n.value, Backends: i18n.Backends, IsInlineEdit: isInlineEdit}
}

func (i18n *I18n) Scope(scope string) admin.I18n {
	return &I18n{Translations: i18n.Translations, scope: scope, value: i18n.value, Backends: i18n.Backends, IsInlineEdit: i18n.IsInlineEdit}
}

func (i18n *I18n) Default(value string) admin.I18n {
	return &I18n{Translations: i18n.Translations, scope: i18n.scope, value: value, Backends: i18n.Backends, IsInlineEdit: i18n.IsInlineEdit}
}

func (i18n *I18n) T(locale, key string, args ...interface{}) template.HTML {
	var value = i18n.value
	var translationKey = key
	if i18n.scope != "" {
		translationKey = strings.Join([]string{i18n.scope, key}, ".")
	}

	if translations := i18n.Translations[locale]; translations != nil && translations[translationKey] != nil && translations[translationKey].Value != "" {
		// Get localized translation
		value = translations[translationKey].Value
	} else if translations := i18n.Translations[Default]; translations != nil && translations[translationKey] != nil {
		// Get default translation if not translated
		value = translations[translationKey].Value
	} else {
		if value == "" {
			value = key
		}
		// Save translations
		i18n.SaveTranslation(&Translation{Key: translationKey, Value: value, Locale: Default, Backend: i18n.Backends[0]})
	}

	if value == "" {
		value = key
	}

	if str, err := cldr.Parse(locale, value, args...); err == nil {
		value = str
	}

	if i18n.IsInlineEdit {
		var editType string
		if len(value) > 25 {
			editType = "data-type=\"textarea\""
		}
		value = fmt.Sprintf("<span class=\"qor-i18n-inline\" %s data-locale=\"%s\" data-key=\"%s\">%s</span>", editType, locale, key, value)
	}

	return template.HTML(value)
}

// Using: http://vitalets.github.io/x-editable/index.html
// You could use Bootstrap or JQuery UI by set isIncludeExtendAssetLib to false and load files by yourself
func RenderInlineEditAssets(isIncludeJQuery bool, isIncludeExtendAssetLib bool) (template.HTML, error) {
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		var content string
		var hasError bool

		if isIncludeJQuery {
			content = `<script src="http://code.jquery.com/jquery-2.0.3.min.js"></script>`
		}

		if isIncludeExtendAssetLib {
			if extendLib, err := ioutil.ReadFile(path.Join(gopath, "src/github.com/qor/qor/i18n/views/themes/i18n/inline-edit-libs.tmpl")); err == nil {
				content += string(extendLib)
			} else {
				hasError = true
			}

			if css, err := ioutil.ReadFile(path.Join(gopath, "src/github.com/qor/qor/i18n/views/themes/i18n/assets/stylesheets/i18n-inline.css")); err == nil {
				content += fmt.Sprintf("<style>%s</style>", string(css))
			} else {
				hasError = true
			}

		}

		if js, err := ioutil.ReadFile(path.Join(gopath, "src/github.com/qor/qor/i18n/views/themes/i18n/assets/javascripts/i18n-inline.js")); err == nil {
			content += fmt.Sprintf("<script type=\"text/javascript\">%s</script>", string(js))
		} else {
			hasError = true
		}

		if !hasError {
			return template.HTML(content), nil
		}
	}

	return template.HTML(""), errors.New("templates not found")
}

func getLocaleFromContext(context *qor.Context) string {
	if locale := utils.GetLocale(context); locale != "" {
		return locale
	}

	return Default
}

type availableLocalesInterface interface {
	AvailableLocales() []string
}

type viewableLocalesInterface interface {
	ViewableLocales() []string
}

type editableLocalesInterface interface {
	EditableLocales() []string
}

func getAvailableLocales(req *http.Request, currentUser qor.CurrentUser) []string {
	if user, ok := currentUser.(viewableLocalesInterface); ok {
		return user.ViewableLocales()
	}

	if user, ok := currentUser.(availableLocalesInterface); ok {
		return user.AvailableLocales()
	}
	return []string{Default}
}

func getEditableLocales(req *http.Request, currentUser qor.CurrentUser) []string {
	if user, ok := currentUser.(editableLocalesInterface); ok {
		return user.EditableLocales()
	}

	if user, ok := currentUser.(availableLocalesInterface); ok {
		return user.AvailableLocales()
	}
	return []string{Default}
}

func (i18n *I18n) ConfigureQorResource(res *admin.Resource) {
	res.UseTheme("i18n")
	res.GetAdmin().I18n = i18n
	res.SearchAttrs("value") // generate search handler for i18n

	res.GetAdmin().RegisterFuncMap("lt", func(locale, key string, withDefault bool) string {
		if translations := i18n.Translations[locale]; translations != nil {
			if t := translations[key]; t != nil && t.Value != "" {
				return t.Value
			}
		}

		if withDefault {
			if translations := i18n.Translations[Default]; translations != nil {
				if t := translations[key]; t != nil {
					return t.Value
				}
			}
		}

		return ""
	})

	res.GetAdmin().RegisterFuncMap("i18n_available_keys", func(context *admin.Context) (keys []string) {
		translations := i18n.Translations[Default]
		if translations == nil {
			for _, values := range i18n.Translations {
				translations = values
				break
			}
		}

		keyword := context.Request.URL.Query().Get("keyword")

		for key, translation := range translations {
			if (keyword == "") || (strings.Index(strings.ToLower(translation.Key), strings.ToLower(keyword)) != -1 ||
				strings.Index(strings.ToLower(translation.Value), keyword) != -1) {
				keys = append(keys, key)
			}
		}

		sort.Strings(keys)

		pagination := context.Searcher.Pagination
		pagination.Total = len(keys)
		pagination.PrePage = 25
		pagination.CurrentPage, _ = strconv.Atoi(context.Request.URL.Query().Get("page"))
		if pagination.CurrentPage == 0 {
			pagination.CurrentPage = 1
		}
		if pagination.CurrentPage > 0 {
			pagination.Pages = pagination.Total / pagination.PrePage
		}
		context.Searcher.Pagination = pagination

		if pagination.CurrentPage == -1 {
			return keys
		}

		lastIndex := pagination.CurrentPage * pagination.PrePage
		if pagination.Total < lastIndex {
			lastIndex = pagination.Total
		}

		return keys[(pagination.CurrentPage-1)*pagination.PrePage : lastIndex]
	})

	res.GetAdmin().RegisterFuncMap("i18n_primary_locale", func(context admin.Context) string {
		if locale := context.Request.Form.Get("primary_locale"); locale != "" {
			return locale
		}
		return getAvailableLocales(context.Request, context.CurrentUser)[0]
	})

	res.GetAdmin().RegisterFuncMap("i18n_editing_locale", func(context admin.Context) string {
		if locale := context.Request.Form.Get("to_locale"); locale != "" {
			return locale
		}
		return getLocaleFromContext(context.Context)
	})

	res.GetAdmin().RegisterFuncMap("i18n_viewable_locales", func(context admin.Context) []string {
		return getAvailableLocales(context.Request, context.CurrentUser)
	})

	res.GetAdmin().RegisterFuncMap("i18n_editable_locales", func(context admin.Context) []string {
		return getEditableLocales(context.Request, context.CurrentUser)
	})

	controller := i18nController{i18n}
	router := res.GetAdmin().GetRouter()
	router.Get(fmt.Sprintf("^/%v", res.ToParam()), controller.Index)
	router.Post(fmt.Sprintf("^/%v", res.ToParam()), controller.Update)
	router.Put(fmt.Sprintf("^/%v", res.ToParam()), controller.Update)

	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/i18n/views"))
	}
}
