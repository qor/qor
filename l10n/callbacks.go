package l10n

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func beforeQuery(scope *gorm.Scope) {
	if isLocalizable(scope) {
		quotedTableName := scope.QuotedTableName()
		locale, _ := getLocale(scope)
		switch mode, _ := scope.DB().Get("l10n:mode"); mode {
		case "locale":
			scope.Search.Where(fmt.Sprintf("%v.language_code = ?", quotedTableName), locale)
		case "global":
			scope.Search.Where(fmt.Sprintf("%v.language_code = ?", quotedTableName), Global)
		case "unscoped":
		default:
			quotedPrimaryKey := scope.Quote(scope.PrimaryKey())
			scope.Search.Unscoped = true
			if scope.Fields()["deleted_at"] != nil {
				scope.Search.Where(fmt.Sprintf("((%v NOT IN (SELECT DISTINCT(%v) FROM %v t2 WHERE t2.language_code = ? AND t2.deleted_at IS NULL) AND language_code = ?) OR language_code = ?) AND deleted_at IS NULL", quotedPrimaryKey, quotedPrimaryKey, quotedTableName), locale, Global, locale)
			} else {
				scope.Search.Where(fmt.Sprintf("(%v NOT IN (SELECT DISTINCT(%v) FROM %v t2 WHERE t2.language_code = ?) AND language_code = ?) OR (language_code = ?)", quotedPrimaryKey, quotedPrimaryKey, quotedTableName), locale, Global, locale)
			}
		}
	}
}

func beforeCreate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if locale, ok := getLocale(scope); ok { // is locale
			if isLocaleCreateable(scope) || !scope.PrimaryKeyZero() {
				setLocale(scope, locale)
			} else {
				err := fmt.Errorf("the resource %v cannot be created in %v", scope.GetModelStruct().ModelType.Name(), locale)
				scope.Err(err)
			}
		} else {
			setLocale(scope, Global)
		}
	}
}

func beforeUpdate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		locale, isLocale := getLocale(scope)

		switch mode, _ := scope.DB().Get("l10n:mode"); mode {
		case "unscoped":
		default:
			scope.Search.Unscoped = true
			scope.Search.Where("language_code = ?", locale)
			setLocale(scope, locale)
		}

		if isLocale {
			scope.Search.Omit(syncColumns(scope)...)
		}
	}
}

func afterUpdate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if locale, ok := getLocale(scope); ok {
			if scope.DB().RowsAffected == 0 { //is locale and nothing updated
				var count int
				var query = fmt.Sprintf("%v.language_code = ? AND %v.%v = ?", scope.QuotedTableName(), scope.QuotedTableName(), scope.PrimaryKey())
				if scope.NewDB().Table(scope.TableName()).Where(query, locale, scope.PrimaryKeyValue()).Count(&count); count == 0 {
					scope.DB().Create(scope.Value)
				}
			}
		} else if syncColumns := syncColumns(scope); len(syncColumns) > 0 { // is global
			if mode, _ := scope.DB().Get("l10n:mode"); mode != "unscoped" {
				if scope.DB().RowsAffected > 0 {
					primaryKey := scope.PrimaryKeyValue()

					if updateAttrs, ok := scope.InstanceGet("gorm:update_attrs"); ok {
						var syncAttrs = map[string]interface{}{}
						for key, value := range updateAttrs.(map[string]interface{}) {
							for _, syncColumn := range syncColumns {
								if syncColumn == key {
									syncAttrs[syncColumn] = value
									break
								}
							}
						}
						if len(syncAttrs) > 0 {
							scope.DB().Model(scope.Value).Set("l10n:mode", "unscoped").Where("language_code <> ?", Global).UpdateColumns(syncAttrs)
						}
					} else {
						scope.NewDB().Set("l10n:mode", "unscoped").Where(fmt.Sprintf("%v = ?", scope.PrimaryKey()), primaryKey).Select(syncColumns).Save(scope.Value)
					}
				}
			}
		}
	}
}

func beforeDelete(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if locale, ok := getLocale(scope); ok { // is locale
			scope.Search.Where(fmt.Sprintf("%v.language_code = ?", scope.QuotedTableName()), locale)
		}
	}
}

func RegisterCallbacks(db *gorm.DB) {
	callback := db.Callback()

	callback.Create().Before("gorm:before_create").Register("l10n:before_create", beforeCreate)

	callback.Update().Before("gorm:before_update").Register("l10n:before_update", beforeUpdate)
	callback.Update().After("gorm:after_update").Register("l10n:after_update", afterUpdate)

	callback.Delete().Before("gorm:before_delete").Register("l10n:before_delete", beforeDelete)

	callback.RowQuery().Register("l10n:before_query", beforeQuery)
	callback.Query().Before("gorm:query").Register("l10n:before_query", beforeQuery)
}
