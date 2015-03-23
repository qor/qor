package l10n

import (
	"os"
	"path"
	"strings"

	"github.com/qor/qor/admin"
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

func (l *Locale) InjectQorAdmin(res *admin.Resource) {
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/l10n/views"))
	}

	res.GetAdmin().RegisterFuncMap("viewable_locales", func(context admin.Context) []string {
		if user, ok := context.CurrentUser.(ViewableLocalesInterface); ok {
			return user.ViewableLocales()
		}

		if user, ok := context.CurrentUser.(AvailableLocalesInterface); ok {
			return user.AvailableLocales()
		}
		return []string{}
	})

	res.GetAdmin().RegisterFuncMap("editable_locales", func(context admin.Context) []string {
		if user, ok := context.CurrentUser.(EditableLocalesInterface); ok {
			return user.EditableLocales()
		}

		if user, ok := context.CurrentUser.(AvailableLocalesInterface); ok {
			return user.AvailableLocales()
		}
		return []string{}
	})
}
