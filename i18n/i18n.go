package i18n

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

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
	SaveTranslation(*Translation)
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

func (i18n *I18n) SaveTransaltion(translation *Translation) {
	if i18n.Translations[translation.Locale] == nil {
		i18n.Translations[translation.Locale] = map[string]*Translation{}
	}
	i18n.Translations[translation.Locale][translation.Key] = translation
	if backend := translation.Backend; backend != nil {
		backend.SaveTranslation(translation)
	}
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
		// Save translations
		i18n.SaveTransaltion(&Translation{Key: translationKey, Locale: locale, Backend: i18n.Backends[0]})
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

type AvailableLocalesInterface interface {
	AvailableLocales() []string
}

type ViewableLocalesInterface interface {
	ViewableLocales() []string
}

type EditableLocalesInterface interface {
	EditableLocales() []string
}

func GetAvailableLocales(req *http.Request, currentUser qor.CurrentUser) []string {
	if user, ok := currentUser.(ViewableLocalesInterface); ok {
		return user.ViewableLocales()
	}

	if user, ok := currentUser.(AvailableLocalesInterface); ok {
		return user.AvailableLocales()
	}
	return []string{}
}

func GetEditableLocales(req *http.Request, currentUser qor.CurrentUser) []string {
	if user, ok := currentUser.(EditableLocalesInterface); ok {
		return user.EditableLocales()
	}

	if user, ok := currentUser.(AvailableLocalesInterface); ok {
		return user.AvailableLocales()
	}
	return []string{}
}

func (i18n *I18n) InjectQorAdmin(res *admin.Resource) {
	res.Config.Theme = "i18n"
	res.GetAdmin().I18n = i18n

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

	res.GetAdmin().RegisterFuncMap("i18n_available_keys", func() (keys []string) {
		translations := i18n.Translations[Default]
		if translations == nil {
			for _, values := range i18n.Translations {
				translations = values
				break
			}
		}

		for key := range translations {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return keys
	})

	res.GetAdmin().RegisterFuncMap("i18n_primary_locale", func(context admin.Context) string {
		if locale := context.Request.Form.Get("primary_locale"); locale != "" {
			return locale
		}
		return getLocaleFromContext(context.Context)
	})

	res.GetAdmin().RegisterFuncMap("i18n_editing_locale", func(context admin.Context) string {
		if locale := context.Request.Form.Get("to_locale"); locale != "" {
			return locale
		}
		return getLocaleFromContext(context.Context)
	})

	res.GetAdmin().RegisterFuncMap("i18n_viewable_locales", func(context admin.Context) []string {
		return GetAvailableLocales(context.Request, context.CurrentUser)
	})

	res.GetAdmin().RegisterFuncMap("i18n_editable_locales", func(context admin.Context) []string {
		return GetEditableLocales(context.Request, context.CurrentUser)
	})

	controller := I18nController{i18n}
	router := res.GetAdmin().GetRouter()
	router.Get(fmt.Sprintf("^/%v", res.ToParam()), controller.Index)
	router.Post(fmt.Sprintf("^/%v", res.ToParam()), controller.Update)
	router.Put(fmt.Sprintf("^/%v", res.ToParam()), controller.Update)

	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/i18n/views"))
	}
}
