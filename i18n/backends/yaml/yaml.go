package yaml

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/qor/qor/i18n"
)

func New(paths ...string) i18n.Backend {
	backend := &Backend{}

	for _, p := range paths {
		if file, err := os.Open(p); err == nil {
			defer file.Close()
			if fileInfo, err := file.Stat(); err == nil {
				if fileInfo.IsDir() {
					yamlFiles, _ := filepath.Glob(path.Join(p, "*.yaml"))
					backend.files = append(backend.files, yamlFiles...)

					ymlFiles, _ := filepath.Glob(path.Join(p, "*.yml"))
					backend.files = append(backend.files, ymlFiles...)
				} else if fileInfo.Mode().IsRegular() {
					backend.files = append(backend.files, p)
				}
			}
		}
	}

	return backend
}

type Backend struct {
	files []string
}

func loadTranslationsFromYaml(locale string, value interface{}, scopes []string) (translations []*i18n.Translation) {
	switch v := value.(type) {
	case yaml.MapSlice:
		for _, s := range v {
			results := loadTranslationsFromYaml(locale, s.Value, append(scopes, fmt.Sprintf("%v", s.Key)))
			translations = append(translations, results...)
		}
	default:
		var translation = &i18n.Translation{
			Locale: locale,
			Key:    strings.Join(scopes, "."),
			Value:  fmt.Sprintf("%v", v),
		}
		translations = append(translations, translation)
	}
	return
}

func (backend *Backend) LoadTranslations() (translations []*i18n.Translation) {
	for _, file := range backend.files {
		if content, err := ioutil.ReadFile(file); err == nil {
			var slice yaml.MapSlice
			if err := yaml.Unmarshal(content, &slice); err == nil {
				for _, item := range slice {
					translations = append(translations, loadTranslationsFromYaml(item.Key.(string) /* locale */, item.Value, []string{})...)
				}
			}
		}
	}
	return translations
}

func (backend *Backend) SaveTranslation(t *i18n.Translation) error {
	return errors.New("not implemented")
}

func (backend *Backend) DeleteTranslation(t *i18n.Translation) error {
	return errors.New("not implemented")
}
