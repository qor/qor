package l10n

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

func isLocalizable(scope *gorm.Scope) (isLocalizable bool) {
	_, isLocalizable = reflect.New(scope.GetModelStruct().ModelType).Interface().(Interface)
	return
}

type LocalizationLocaleCreateable interface {
	LocalizationLocaleCreateable()
}

func isLocalizationLocaleCreateable(scope *gorm.Scope) (ok bool) {
	_, ok = reflect.New(scope.GetModelStruct().ModelType).Interface().(LocalizationLocaleCreateable)
	return
}

func setScopeLocale(scope *gorm.Scope, locale string) {
	method := func(value interface{}) {
		if model, ok := value.(Interface); ok {
			model.SetLocale(locale)
		}
	}

	if values := scope.IndirectValue(); values.Kind() == reflect.Slice {
		for i := 0; i < values.Len(); i++ {
			method(values.Index(i).Addr().Interface())
		}
	} else {
		method(scope.Value)
	}
}

func BeforeQuery(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if str, ok := scope.DB().Get("l10n:locale"); ok {
			if locale, ok := str.(string); ok {
				quotedTableName := scope.QuotedTableName()
				switch mode, _ := scope.DB().Get("l10n"); mode {
				case "locale":
					scope.Search.Where(fmt.Sprintf("%v.language_code = ?", quotedTableName), locale)
				case "global":
					scope.Search.Where(fmt.Sprintf("%v.language_code IS NULL", quotedTableName))
				default:
					scope.Search.Where(fmt.Sprintf("%v.language_code = ? OR %v.language_code IS NULL", quotedTableName), locale)
				}
			}
		}
	}
}

func BeforeCreate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if str, ok := scope.DB().Get("l10n:locale"); ok {
			if isLocalizationLocaleCreateable(scope) {
				setScopeLocale(scope, str.(string))
			} else {
				scope.Err(errors.New("permission denied to create from locale"))
			}
		}
	}
}

func BeforeUpdate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		// is locale -> update localized columns
	}
}

func AfterUpdate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		// is global -> sync colums that need sync
	}
}

func BeforeDelete(scope *gorm.Scope) {
	if isLocalizable(scope) {
		// is locale -> scope.Search.Where("language_code = ?", locale)
	}
}
