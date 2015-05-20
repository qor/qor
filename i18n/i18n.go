package i18n

import (
	"os"
	"path"
	"strings"

	"github.com/qor/qor/admin"
)

type I18n struct {
	Backends     []Backend
	Translations map[string]map[string]Translation
}

type Backend interface {
	LoadTransations() []Translation
}

type Translation struct {
	Key    string
	Locale string
	Value  string
}

func New(backends ...Backend) *I18n {
	i18n := &I18n{Backends: backends, Translations: map[string]map[string]Translation{}}
	for _, backend := range backends {
		for _, translation := range backend.LoadTransations() {
			i18n.AddTransaltion(translation)
		}
	}
	return i18n
}

func (i18n *I18n) AddTransaltion(translation Translation) {
	if i18n.Translations[translation.Locale] == nil {
		i18n.Translations[translation.Locale] = map[string]Translation{}
	}
	i18n.Translations[translation.Locale][translation.Key] = translation
}

func (i18n *I18n) UpdateTransaltion(translation Translation) {
	i18n.Translations[translation.Locale][translation.Key] = translation
}

func (i18n *I18n) DeleteTransaltion(translation Translation) {
	delete(i18n.Translations[translation.Locale], translation.Key)
}

func (i18n *I18n) T(locale, key string, args ...interface{}) string {
	return key // TODO cldr
}

func (i18n *I18n) InjectQorAdmin(res *admin.Resource) {
	res.Config.Theme = "i18n"
	res.GetAdmin().I18n = i18n

	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/i18n/views"))
	}
}
