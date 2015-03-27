package l10n

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/roles"
)

type Interface interface {
	IsGlobal() bool
	SetLocale(locale string)
}

type Locale struct {
	LanguageCode string `sql:"size:6" gorm:"primary_key"`
}

func (l Locale) IsGlobal() bool {
	return l.LanguageCode == ""
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

func GetCurrentLocale(req *http.Request) string {
	return req.Form.Get("locale")
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

	res.Config.Theme = "l10n"
	res.Config.Permission.Allow(roles.CRUD, "locale_admin").Allow(roles.Read, "locale_reader")

	res.GetAdmin().RegisterFuncMap("viewable_locales", func(context admin.Context) []string {
		return GetAvailableLocales(context.Request, context.CurrentUser)
	})

	res.GetAdmin().RegisterFuncMap("editable_locales", func(context admin.Context) []string {
		return GetEditableLocales(context.Request, context.CurrentUser)
	})

	role := res.Config.Permission.Role
	if _, ok := role.Get("locale_admin"); !ok {
		role.Register("locale_admin", func(req *http.Request, currentUser qor.CurrentUser) bool {
			currentLocale := GetCurrentLocale(req)
			for _, locale := range GetEditableLocales(req, currentUser) {
				if locale == currentLocale {
					return true
				}
			}
			return false
		})

		role.Register("locale_reader", func(req *http.Request, currentUser qor.CurrentUser) bool {
			currentLocale := GetCurrentLocale(req)
			for _, locale := range GetAvailableLocales(req, currentUser) {
				if locale == currentLocale {
					return true
				}
			}
			return false
		})
	}
}
