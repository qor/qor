package i18n

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/utils"
	"github.com/theplant/cldr"
)

var Default = "en-US"

type I18n struct {
	scope        string
	Backends     []Backend
	Translations map[string]map[string]*Translation
}

type Backend interface {
	LoadTranslations() []*Translation
	SaveTranslation(*Translation) error
	DeleteTranslation(*Translation)
}

type Translation struct {
	Key     string
	Locale  string
	Value   string
	Backend Backend
}

func New(backends ...Backend) *I18n {
	i18n := &I18n{Backends: backends, Translations: map[string]map[string]*Translation{}}
	for _, backend := range backends {
		for _, translation := range backend.LoadTranslations() {
			translation.Backend = backend
			i18n.AddTransaltion(translation)
		}
	}
	return i18n
}

func (i18n *I18n) AddTransaltion(translation *Translation) {
	if i18n.Translations[translation.Locale] == nil {
		i18n.Translations[translation.Locale] = map[string]*Translation{}
	}
	i18n.Translations[translation.Locale][translation.Key] = translation
}

func (i18n *I18n) SaveTranslation(translation *Translation) (err error) {
	if i18n.Translations[translation.Locale] == nil {
		i18n.Translations[translation.Locale] = map[string]*Translation{}
	}
	if backend := translation.Backend; backend != nil {
		err = backend.SaveTranslation(translation)
		if err != nil {
			return
		}
	}
	i18n.Translations[translation.Locale][translation.Key] = translation
	return
}

func (i18n *I18n) DeleteTransaltion(translation *Translation) {
	delete(i18n.Translations[translation.Locale], translation.Key)
	translation.Backend.DeleteTranslation(translation)
}

func (i18n *I18n) Scope(scope string) admin.I18n {
	return &I18n{Translations: i18n.Translations, scope: scope, Backends: i18n.Backends}
}

func (i18n *I18n) T(locale, key string, args ...interface{}) string {
	var value string
	var translationKey = key
	if i18n.scope != "" {
		translationKey = strings.Join([]string{i18n.scope, key}, ".")
	}

	if translations := i18n.Translations[locale]; translations != nil && translations[translationKey] != nil {
		value = translations[translationKey].Value
	} else {
		var value string
		if Default == locale {
			value = key
		}
		// Save translations
		err := i18n.SaveTranslation(&Translation{Key: translationKey, Locale: locale, Backend: i18n.Backends[0]})
		log.Printf("Error saving translation: [%s]: %s\n", locale, translationKey)
		if err != nil {
			return err.Error()
		}
	}

	if value == "" {
		// Get default translation if not translated
		if translations := i18n.Translations[Default]; translations != nil && translations[translationKey] != nil {
			value = translations[translationKey].Value
		}
		if value == "" {
			value = key
		}
	}

	if str, err := cldr.Parse(locale, value, args...); err == nil {
		return str
	}
	return value
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
	return []string{}
}

func getEditableLocales(req *http.Request, currentUser qor.CurrentUser) []string {
	if user, ok := currentUser.(editableLocalesInterface); ok {
		return user.EditableLocales()
	}

	if user, ok := currentUser.(availableLocalesInterface); ok {
		return user.AvailableLocales()
	}
	return []string{}
}

func (i18n *I18n) InjectQorAdmin(res *admin.Resource) {
	res.UseTheme("i18n")
	res.GetAdmin().I18n = i18n
	res.SearchHandler = func(keyword string, context *qor.Context) *gorm.DB { return context.GetDB() }

	res.GetAdmin().RegisterFuncMap("lt", func(locale, key string, withDefault bool) string {
		translations := i18n.Translations[locale]
		if (translations == nil) && withDefault {
			translations = i18n.Translations[Default]
		}

		if translation := translations[key]; translation != nil {
			return translation.Value
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
