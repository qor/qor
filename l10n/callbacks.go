package l10n

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/utils"
)

func beforeQuery(scope *gorm.Scope) {
	if isLocalizable(scope) {
		quotedTableName := scope.QuotedTableName()
		quotedPrimaryKey := scope.Quote(scope.PrimaryKey())

		locale, isLocale := getLocale(scope)
		switch mode, _ := scope.DB().Get("l10n:mode"); mode {
		case "unscoped":
		case "global":
			scope.Search.Where(fmt.Sprintf("%v.language_code = ?", quotedTableName), Global)
		case "locale":
			scope.Search.Where(fmt.Sprintf("%v.language_code = ?", quotedTableName), locale)
		case "reverse":
			if !scope.Search.Unscoped && scope.Fields()["deleted_at"] != nil {
				scope.Search.Where(fmt.Sprintf("(%v NOT IN (SELECT DISTINCT(%v) FROM %v t2 WHERE t2.language_code = ? AND t2.deleted_at IS NULL) AND language_code = ?)", quotedPrimaryKey, quotedPrimaryKey, quotedTableName), locale, Global)
			} else {
				scope.Search.Where(fmt.Sprintf("(%v NOT IN (SELECT DISTINCT(%v) FROM %v t2 WHERE t2.language_code = ?) AND language_code = ?)", quotedPrimaryKey, quotedPrimaryKey, quotedTableName), locale, Global)
			}
		default:
			if isLocale {
				if !scope.Search.Unscoped && scope.Fields()["deleted_at"] != nil {
					scope.Search.Where(fmt.Sprintf("((%v NOT IN (SELECT DISTINCT(%v) FROM %v t2 WHERE t2.language_code = ? AND t2.deleted_at IS NULL) AND language_code = ?) OR language_code = ?) AND deleted_at IS NULL", quotedPrimaryKey, quotedPrimaryKey, quotedTableName), locale, Global, locale)
				} else {
					scope.Search.Where(fmt.Sprintf("(%v NOT IN (SELECT DISTINCT(%v) FROM %v t2 WHERE t2.language_code = ?) AND language_code = ?) OR (language_code = ?)", quotedPrimaryKey, quotedPrimaryKey, quotedTableName), locale, Global, locale)
				}
			} else {
				scope.Search.Where(fmt.Sprintf("%v.language_code = ?", quotedTableName), Global)
			}
		}
	}
}

func beforeCreate(scope *gorm.Scope) {
	if isLocalizable(scope) {
		if locale, ok := getLocale(scope); ok { // is locale
			if isLocaleCreatable(scope) || !scope.PrimaryKeyZero() {
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
			scope.Search.Where("language_code = ?", locale)
			setLocale(scope, locale)
		}

		if isLocale {
			scope.Search.Omit(syncColumns(scope)...)
		}
	}
}

func afterUpdate(scope *gorm.Scope) {
	if !scope.HasError() {
		if isLocalizable(scope) {
			if locale, ok := getLocale(scope); ok {
				if scope.DB().RowsAffected == 0 && !scope.PrimaryKeyZero() { //is locale and nothing updated
					var count int
					var query = fmt.Sprintf("%v.language_code = ? AND %v.%v = ?", scope.QuotedTableName(), scope.QuotedTableName(), scope.PrimaryKey())

					// if enabled soft delete, delete soft deleted records
					if scope.HasColumn("DeletedAt") {
						scope.NewDB().Unscoped().Where("deleted_at is not null").Where(query, locale, scope.PrimaryKeyValue()).Delete(scope.Value)
					}

					// if no localized records exist, localize it
					if scope.NewDB().Table(scope.TableName()).Where(query, locale, scope.PrimaryKeyValue()).Count(&count); count == 0 {
						scope.DB().Create(scope.Value)
					}
				}
			} else if syncColumns := syncColumns(scope); len(syncColumns) > 0 { // is global
				if mode, _ := scope.DB().Get("l10n:mode"); mode != "unscoped" {
					if scope.DB().RowsAffected > 0 {
						var primaryField = scope.PrimaryField()
						var syncAttrs = map[string]interface{}{}

						if updateAttrs, ok := scope.InstanceGet("gorm:update_attrs"); ok {
							for key, value := range updateAttrs.(map[string]interface{}) {
								for _, syncColumn := range syncColumns {
									if syncColumn == key {
										syncAttrs[syncColumn] = value
										break
									}
								}
							}
						} else {
							var fields = scope.Fields()
							for _, syncColumn := range syncColumns {
								if field, ok := fields[syncColumn]; ok && field.IsNormal {
									syncAttrs[syncColumn] = field.Field.Interface()
								}
							}
						}

						if len(syncAttrs) > 0 {
							db := scope.DB().Model(reflect.New(utils.ModelType(scope.Value)).Interface()).Set("l10n:mode", "unscoped").Where("language_code <> ?", Global)
							if !primaryField.IsBlank {
								db = db.Where(fmt.Sprintf("%v = ?", primaryField.DBName), primaryField.Field.Interface())
							}
							scope.Err(db.UpdateColumns(syncAttrs).Error)
						}
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
