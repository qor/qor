package l10n

import (
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

func isLocalizable(scope *gorm.Scope) (isLocalizable bool) {
	if scope.GetModelStruct().ModelType == nil {
		return false
	}
	_, isLocalizable = reflect.New(scope.GetModelStruct().ModelType).Interface().(Interface)
	return
}

type LocaleCreateableInterface interface {
	LocaleCreateable()
}

func isLocaleCreateable(scope *gorm.Scope) (ok bool) {
	_, ok = reflect.New(scope.GetModelStruct().ModelType).Interface().(LocaleCreateableInterface)
	return
}

func setLocale(scope *gorm.Scope, locale string) {
	for _, field := range scope.Fields() {
		if field.Name == "LanguageCode" {
			field.Set(locale)
		}
	}
}

func getLocale(scope *gorm.Scope) (locale string, ok bool) {
	if str, ok := scope.DB().Get("l10n:locale"); ok {
		if locale, ok := str.(string); ok {
			return locale, (locale != Global) && (locale != "")
		}
	}
	return Global, false
}

func parseTagOption(str string) map[string]string {
	tags := strings.Split(str, ";")
	setting := map[string]string{}
	for _, value := range tags {
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) == 2 {
			setting[k] = v[1]
		} else {
			setting[k] = k
		}
	}
	return setting
}

func isSyncField(field *gorm.StructField) bool {
	if _, ok := parseTagOption(field.Tag.Get("l10n"))["SYNC"]; ok {
		return true
	}
	return false
}

func syncColumns(scope *gorm.Scope) (columns []string) {
	for _, field := range scope.GetModelStruct().StructFields {
		if isSyncField(field) {
			columns = append(columns, field.DBName)
		}
	}
	return
}
