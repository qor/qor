package l10n

import (
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/utils"
)

func isLocalizable(scope *gorm.Scope) (isLocalizable bool) {
	if scope.GetModelStruct().ModelType == nil {
		return false
	}
	_, isLocalizable = reflect.New(scope.GetModelStruct().ModelType).Interface().(l10nInterface)
	return
}

type localeCreatableInterface interface {
	LocaleCreatable()
}

func isLocaleCreatable(scope *gorm.Scope) (ok bool) {
	_, ok = reflect.New(scope.GetModelStruct().ModelType).Interface().(localeCreatableInterface)
	return
}

func setLocale(scope *gorm.Scope, locale string) {
	for _, field := range scope.Fields() {
		if field.Name == "LanguageCode" {
			field.Set(locale)
		}
	}
}

func getLocale(scope *gorm.Scope) (locale string, isLocale bool) {
	if str, ok := scope.DB().Get("l10n:locale"); ok {
		if locale, ok := str.(string); ok {
			return locale, (locale != Global) && (locale != "")
		}
	}
	return Global, false
}

func isSyncField(field *gorm.StructField) bool {
	if _, ok := utils.ParseTagOption(field.Tag.Get("l10n"))["SYNC"]; ok {
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
