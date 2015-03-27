package l10n

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
)

func BeforeQuery(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if locale, ok := getLocale(scope); ok { // is locale
			quotedTableName := scope.QuotedTableName()
			switch mode, _ := scope.DB().Get("l10n:mode"); mode {
			case "locale":
				scope.Search.Where(fmt.Sprintf("%v.language_code = ?", quotedTableName), locale)
			case "global":
				scope.Search.Where(fmt.Sprintf("%v.language_code = ?", quotedTableName), "")
			default:
				scope.Search.Where(fmt.Sprintf("%v.language_code = ? OR %v.language_code = ?", quotedTableName, quotedTableName), locale, "").Order("language_code DESC")
			}
		}
	}
}

func BeforeCreate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if locale, ok := getLocale(scope); ok { // is locale
			if isLocaleCreateable(scope) || !scope.PrimaryKeyZero() {
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
			scope.Search.Unscoped = true
			scope.Search.Where(fmt.Sprintf("%v.language_code = ?", scope.QuotedTableName()), locale)
			setLocale(scope, locale)
			scope.Search.Omit(syncColumns(scope)...)
		}
	}
}

func AfterUpdate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if locale, ok := getLocale(scope); ok {
			if scope.DB().RowsAffected == 0 { //is locale and nothing updated
				var count int
				var query = fmt.Sprintf("language_code = ? AND %v = ?", scope.PrimaryKey())
				if scope.NewDB().Table(scope.TableName()).Where(query, locale, scope.PrimaryKeyValue()).Count(&count); count == 0 {
					scope.DB().Create(scope.Value)
				}
			}
		} else if syncColumns := syncColumns(scope); len(syncColumns) > 0 { // is global
			scope.NewDB().Where(fmt.Sprintf("%v = ?"), scope.PrimaryKeyValue).Select(syncColumns).Update(scope.Value)
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
