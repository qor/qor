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

func setTableAndPublishStatus(ensureDraftMode bool) func(*gorm.Scope) {
	return func(scope *gorm.Scope) {
		if scope.Value == nil {
			return
		}

		if isPublishableModel(scope.Value) {
			scope.InstanceSet("publish:supported_model", true)

			if ensureDraftMode {
				scope.Set("publish:force_draft_mode", true)
				scope.Search.Table(draftTableName(scope.TableName()))
			}

			if isDraftMode(scope) {
				if attrs, ok := scope.InstanceGet("gorm:update_attrs"); ok {
					updateAttrs := attrs.(map[string]interface{})
					updateAttrs["publish_status"] = DIRTY
					scope.InstanceSet("gorm:update_attrs", updateAttrs)
				} else {
					scope.SetColumn("PublishStatus", DIRTY)
				}
			}
		}
	}
}

func getModeAndNewScope(scope *gorm.Scope) (isProduction bool, clone *gorm.Scope) {
	if draftMode, ok := scope.Get("publish:draft_mode"); !ok || !draftMode.(bool) {
		if _, ok := scope.InstanceGet("publish:supported_model"); ok {
			table := originalTableName(scope.TableName())
			clone := scope.New(scope.Value)
			clone.Search.Table(table)
			return true, clone
		}
	}
	return false, nil
}

func syncToProductionAfterCreate(scope *gorm.Scope) {
	if ok, clone := getModeAndNewScope(scope); ok {
		gorm.Create(clone)
	}
}

func syncToProductionAfterUpdate(scope *gorm.Scope) {
	if ok, clone := getModeAndNewScope(scope); ok {
		if updateAttrs, ok := scope.InstanceGet("gorm:update_attrs"); ok {
			table := originalTableName(scope.TableName())
			clone.Search = scope.Search
			clone.Search.Table(table)
			clone.InstanceSet("gorm:update_attrs", updateAttrs)
		}
		gorm.Update(clone)
	}
}

func syncToProductionAfterDelete(scope *gorm.Scope) {
	if ok, clone := getModeAndNewScope(scope); ok {
		gorm.Delete(clone)
	}
}

func deleteScope(scope *gorm.Scope) {
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
