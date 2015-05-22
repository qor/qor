package i18n

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/utils"
	"github.com/theplant/cldr"
)

var Default = "en-US"

type I18n struct {
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
	translation.Backend.SaveTranslation(translation)
}

func (i18n *I18n) DeleteTransaltion(translation *Translation) {
	delete(i18n.Translations[translation.Locale], translation.Key)
	translation.Backend.DeleteTranslation(translation)
}

func (i18n *I18n) T(locale, key string, args ...interface{}) string {
	var value string

	if translations := i18n.Translations[locale]; translations != nil && translations[key] != nil {
		value = translations[key].Value
	} else {
		values := strings.Split(key, ".")
		i18n.SaveTransaltion(&Translation{Key: key, Locale: locale, Value: values[len(values)-1], Backend: i18n.Backends[0]})
		value = key
	}

	if str, err := cldr.Parse(locale, value, args...); err == nil {
		return str
	}
	return key
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

	controller := I18nController{i18n}
	router := res.GetAdmin().GetRouter()
	router.Get(fmt.Sprintf("^/%v", res.ToParam()), controller.Index)

	res.GetAdmin().RegisterFuncMap("i18n_primary_locale", func(context admin.Context) string {
		return getLocaleFromContext(context.Context)
	})

	res.GetAdmin().RegisterFuncMap("i18n_edting_locale", func(context admin.Context) string {
		return getLocaleFromContext(context.Context)
	})

	res.GetAdmin().RegisterFuncMap("i18n_viewable_locales", func(context admin.Context) []string {
		return GetAvailableLocales(context.Request, context.CurrentUser)
	})

	res.GetAdmin().RegisterFuncMap("i18n_editable_locales", func(context admin.Context) []string {
		return GetEditableLocales(context.Request, context.CurrentUser)
	})

	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/i18n/views"))
	}
}
