package yaml_test

import (
	"testing"

	"github.com/qor/qor/i18n/backends/yaml"
)

func TestLoadTranslations(t *testing.T) {
	backend := yaml.New("tests")
	translations := backend.LoadTranslations()

	values := map[string][][]string{
		"en": [][]string{
			{"hello", "Hello"},
			{"user.name", "User Name"},
			{"user.email", "Email"},
		},
		"zh-CN": [][]string{
			{"hello", "你好"},
			{"user.name", "用户名"},
			{"user.email", "邮箱"},
		},
	}

	for locale, results := range values {
		for _, result := range results {
			var found bool
			for _, translation := range translations {
				if (translation.Locale == locale) && (translation.Key == result[0]) && (translation.Value == result[1]) {
					found = true
				}
			}
			if !found {
				t.Errorf("failed to found translation %v for %v", result[0], locale)
			}
		}
	}
}
