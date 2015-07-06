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

type publishInterface interface {
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

func getLocaleFromContext(context *qor.Context) string {
	if locale := utils.GetLocale(context); locale != "" {
		return locale
	}

	return Global
}

var injected bool

func (l *Locale) InjectQorAdmin(res *admin.Resource) {
	Admin := res.GetAdmin()
	res.UseTheme("l10n")

	if res.Config.Permission == nil {
		res.Config.Permission = roles.NewPermission()
	}
	res.Config.Permission.Allow(roles.CRUD, "locale_admin").Allow(roles.Read, "locale_reader")

	if res.GetMeta("LanguageCode") == nil {
		res.Meta(&admin.Meta{Name: "LanguageCode", Type: "hidden"})
	}

	// Set meta permissions
	for _, field := range Admin.Config.DB.NewScope(res.Value).Fields() {
		if isSyncField(field.StructField) {
			if meta := res.GetMeta(field.Name); meta != nil {
				permission := meta.Permission
				if permission == nil {
					permission = roles.Allow(roles.CRUD, "global_admin").Allow(roles.Read, "locale_reader")
				} else {
					permission = permission.Allow(roles.CRUD, "global_admin").Allow(roles.Read, "locale_reader")
				}

				meta.Permission = permission
			} else {
				res.Meta(&admin.Meta{Name: field.Name, Permission: roles.Allow(roles.CRUD, "global_admin").Allow(roles.Read, "locale_reader")})
			}
		}
	}

	// Roles
	role := res.Config.Permission.Role
	if _, ok := role.Get("locale_admin"); !ok {
		role.Register("locale_admin", func(req *http.Request, currentUser qor.CurrentUser) bool {
			currentLocale := getLocaleFromContext(&qor.Context{Request: req})
			for _, locale := range getEditableLocales(req, currentUser) {
				if locale == currentLocale {
					return true
				}
			}
			return false
		})
	}

	if _, ok := role.Get("global_admin"); !ok {
		role.Register("global_admin", func(req *http.Request, currentUser qor.CurrentUser) bool {
			return getLocaleFromContext(&qor.Context{Request: req}) == Global
		})
	}

	if _, ok := role.Get("locale_reader"); !ok {
		role.Register("locale_reader", func(req *http.Request, currentUser qor.CurrentUser) bool {
			currentLocale := getLocaleFromContext(&qor.Context{Request: req})
			for _, locale := range getAvailableLocales(req, currentUser) {
				if locale == currentLocale {
					return true
				}
			}
			return false
		})
	}

	// Inject for l10n
	if !injected {
		injected = true
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/l10n/views"))
		}

		router := Admin.GetRouter()
		// Middleware
		router.Use(func(context *admin.Context, middleware *admin.Middleware) {
			context.SetDB(context.GetDB().Set("l10n:locale", getLocaleFromContext(context.Context)))

			middleware.Next(context)
		})

		// FunMap
		Admin.RegisterFuncMap("current_locale", func(context admin.Context) string {
			return getLocaleFromContext(context.Context)
		})

		Admin.RegisterFuncMap("viewable_locales", func(context admin.Context) []string {
			return getAvailableLocales(context.Request, context.CurrentUser)
		})

		Admin.RegisterFuncMap("editable_locales", func(context admin.Context) []string {
			return getEditableLocales(context.Request, context.CurrentUser)
		})

		Admin.RegisterFuncMap("createable_locales", func(context admin.Context) []string {
			editableLocales := getEditableLocales(context.Request, context.CurrentUser)
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

		Admin.RegisterFuncMap("locales_of_resource", func(resource interface{}, context admin.Context) []map[string]interface{} {
			scope := context.GetDB().NewScope(resource)
			var languageCodes []string
			context.GetDB().New().Set("l10n:mode", "unscoped").Model(resource).Where(fmt.Sprintf("%v = ?", scope.PrimaryKey()), scope.PrimaryKeyValue()).Pluck("language_code", &languageCodes)

			var results []map[string]interface{}
			availableLocales := getAvailableLocales(context.Request, context.CurrentUser)
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
	}
}
