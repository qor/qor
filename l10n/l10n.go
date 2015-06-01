package l10n

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/roles"
	"github.com/qor/qor/utils"
)

var Global = "en-US"

type Interface interface {
	IsGlobal() bool
	SetLocale(locale string)
}

type Locale struct {
	LanguageCode string `sql:"size:6" gorm:"primary_key"`
}

func (l Locale) IsGlobal() bool {
	return l.LanguageCode == Global
}

func (l *Locale) SetLocale(locale string) {
	l.LanguageCode = locale
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

func getLocaleFromContext(context *qor.Context) string {
	if locale := utils.GetLocale(context); locale != "" {
		return locale
	}

	return Global
}

func (l *Locale) InjectQorAdmin(res *admin.Resource) {
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/l10n/views"))
	}

	if res.Config == nil {
		res.Config = &admin.Config{}
	}
	if res.Config.Permission == nil {
		res.Config.Permission = roles.NewPermission()
	}

	res.Meta(&admin.Meta{Name: "LanguageCode", Type: "hidden"})

	res.Config.Theme = "l10n"
	res.Config.Permission.Allow(roles.CRUD, "locale_admin").Allow(roles.Read, "locale_reader")

	router := res.GetAdmin().GetRouter()
	router.Use(func(context *admin.Context, middleware *admin.Middleware) {
		context.SetDB(context.GetDB().Set("l10n:locale", getLocaleFromContext(context.Context)))
		middleware.Next(context)
	})

	res.GetAdmin().RegisterFuncMap("current_locale", func(context admin.Context) string {
		return getLocaleFromContext(context.Context)
	})

	res.GetAdmin().RegisterFuncMap("viewable_locales", func(context admin.Context) []string {
		return GetAvailableLocales(context.Request, context.CurrentUser)
	})

	res.GetAdmin().RegisterFuncMap("editable_locales", func(context admin.Context) []string {
		return GetEditableLocales(context.Request, context.CurrentUser)
	})

	res.GetAdmin().RegisterFuncMap("createable_locales", func(context admin.Context) []string {
		editableLocales := GetEditableLocales(context.Request, context.CurrentUser)
		if _, ok := context.Resource.Value.(LocaleCreateableInterface); ok {
			return editableLocales
		} else {
			for _, locale := range editableLocales {
				if locale == Global {
					return []string{Global}
				}
			}
		}
		return []string{}
	})

	res.GetAdmin().RegisterFuncMap("locales_of_resource", func(resource interface{}, context admin.Context) []map[string]interface{} {
		scope := context.GetDB().NewScope(resource)
		var languageCodes []string
		context.GetDB().New().Set("l10n:mode", "unscoped").Model(resource).Where(fmt.Sprintf("%v = ?", scope.PrimaryKey()), scope.PrimaryKeyValue()).Pluck("language_code", &languageCodes)

		var results []map[string]interface{}
		availableLocales := GetAvailableLocales(context.Request, context.CurrentUser)
	OUT:
		for _, locale := range availableLocales {
			for _, localized := range languageCodes {
				if locale == localized {
					results = append(results, map[string]interface{}{"locale": locale, "localized": true})
					continue OUT
				}
			}
			results = append(results, map[string]interface{}{"locale": locale, "localized": false})
		}
		return results
	})

	role := res.Config.Permission.Role
	if _, ok := role.Get("locale_admin"); !ok {
		role.Register("locale_admin", func(req *http.Request, currentUser qor.CurrentUser) bool {
			currentLocale := getLocaleFromContext(&qor.Context{Request: req})
			for _, locale := range GetEditableLocales(req, currentUser) {
				if locale == currentLocale {
					return true
				}
			}
			return false
		})

		role.Register("locale_reader", func(req *http.Request, currentUser qor.CurrentUser) bool {
			currentLocale := getLocaleFromContext(&qor.Context{Request: req})
			for _, locale := range GetAvailableLocales(req, currentUser) {
				if locale == currentLocale {
					return true
				}
			}
			return false
		})
	}
}
