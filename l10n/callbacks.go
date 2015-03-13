package l10n

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
)

func BeforeQuery(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if str, ok := scope.DB().Get("l10n:locale"); ok {
			if locale, ok := str.(string); ok {
				quotedTableName := scope.QuotedTableName()
				switch mode, _ := scope.DB().Get("l10n:mode"); mode {
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
		if locale, ok := getLocale(scope); ok { // is locale
			if isLocaleCreateable(scope) {
				setLocale(scope, locale)
			} else {
				scope.Err(errors.New("permission denied to create from locale"))
			}
		}
	}
}

func BeforeUpdate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if locale, ok := getLocale(scope); ok { // is locale
			setLocale(scope, locale)
			scope.Search.Omit(syncColumns(scope)...)
		}
	}
}

func AfterUpdate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if _, ok := getLocale(scope); !ok { // is global
			scope.NewDB().Where(fmt.Sprintf("%v = ?"), scope.PrimaryKeyValue).Select(syncColumns(scope)).Update(scope.Value)
		}
	}
}

func BeforeDelete(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if locale, ok := getLocale(scope); ok { // is locale
			scope.Search.Where("language_code = ?", locale)
		}
	}
}

func RegisterCallbacks(db *gorm.DB) {
	callback := db.Callback()

	callback.Create().Before("gorm:before_create").Register("l10n:before_create", BeforeCreate)

	callback.Update().Before("gorm:before_update").Register("l10n:before_update", BeforeUpdate)
	callback.Update().After("gorm:after_update").Register("l10n:after_update", AfterUpdate)

	callback.Delete().Before("gorm:before_delete").Register("l10n:before_delete", BeforeDelete)

	callback.Query().Before("gorm:query").Register("l10n:before_query", BeforeQuery)
}
