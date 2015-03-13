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
			if isLocaleCreateable(scope) {
				setLocale(scope, str.(string))
			} else {
				scope.Err(errors.New("permission denied to create from locale"))
			}
		}
	}
}

func BeforeUpdate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if str, ok := scope.DB().Get("l10n:locale"); ok {
			setLocale(scope, str.(string))
			scope.Search.Omit(syncColumns(scope)...)
		}
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
