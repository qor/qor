package publish

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

func isDraftMode(scope *gorm.Scope) bool {
	if draftMode, ok := scope.Get("qor_publish:draft_mode"); ok {
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

		if update {
			scope.Set("qor_publish:force_draft_mode", true)
		}

		currentModel := scope.GetModelStruct().ModelType.String()

		var supportedModels []string
		if value, ok := scope.Get("publish:support_models"); ok {
			supportedModels = value.([]string)
		}

		for _, model := range supportedModels {
			if model == currentModel {
				scope.InstanceSet("publish:supported_model", true)
				if isDraftMode(scope) && update {
					scope.SetColumn("PublishStatus", DIRTY)
				}
				break
			}
		}
	}
}

func GetModeAndNewScope(scope *gorm.Scope) (isProduction bool, clone *gorm.Scope) {
	if draftMode, ok := scope.Get("qor_publish:draft_mode"); ok && !draftMode.(bool) {
		if table, ok := scope.InstanceGet("publish:original_table"); ok {
			clone := scope.New(scope.Value)
			scope.Search.Table(table.(string))
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
		isDraftMode, ok := scope.Get("qor_publish:draft_mode")

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
		gorm.Delete(scope)
	}
}
