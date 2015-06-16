package publish

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func isDraftMode(scope *gorm.Scope) bool {
	if draftMode, ok := scope.Get("publish:draft_mode"); ok {
		if isDraft, ok := draftMode.(bool); ok && isDraft {
			return true
		}
	}
	return false
}

func SetTableAndPublishStatus(update bool) func(*gorm.Scope) {
	return func(scope *gorm.Scope) {
		if scope.Value == nil {
			return
		}

		if IsPublishableModel(scope.Value) {
			scope.InstanceSet("publish:supported_model", true)

			if update {
				scope.Set("publish:force_draft_mode", true)
				scope.Search.Table(DraftTableName(scope.TableName()))
			}

			if isDraftMode(scope) && update {
				scope.SetColumn("PublishStatus", DIRTY)
			}
		}
	}
}

func GetModeAndNewScope(scope *gorm.Scope) (isProduction bool, clone *gorm.Scope) {
	if draftMode, ok := scope.Get("publish:draft_mode"); ok && !draftMode.(bool) {
		if _, ok := scope.InstanceGet("publish:supported_model"); ok {
			table := OriginalTableName(scope.TableName())
			clone := scope.New(scope.Value)
			clone.Search.Table(table)
			return true, clone
		}
	}
	return false, nil
}

func SyncToProductionAfterCreate(scope *gorm.Scope) {
	if ok, clone := GetModeAndNewScope(scope); ok {
		gorm.Create(clone)
	}
}

func SyncToProductionAfterUpdate(scope *gorm.Scope) {
	if ok, clone := GetModeAndNewScope(scope); ok {
		if updateAttrs, ok := scope.InstanceGet("gorm:update_attrs"); ok {
			clone.InstanceSet("gorm:update_attrs", updateAttrs)
		}
		gorm.Update(clone)
	}
}

func SyncToProductionAfterDelete(scope *gorm.Scope) {
	if ok, clone := GetModeAndNewScope(scope); ok {
		gorm.Delete(clone)
	}
}

func Delete(scope *gorm.Scope) {
	if !scope.HasError() {
		_, supportedModel := scope.InstanceGet("publish:supported_model")
		isDraftMode, ok := scope.Get("publish:draft_mode")

		if supportedModel && (ok && isDraftMode.(bool)) {
			scope.Raw(
				fmt.Sprintf("UPDATE %v SET deleted_at=%v, publish_status=%v %v",
					scope.QuotedTableName(),
					scope.AddToVars(gorm.NowFunc()),
					scope.AddToVars(DIRTY),
					scope.CombinedConditionSql(),
				))
			scope.Exec()
		} else {
			gorm.Delete(scope)
		}
	}
}
