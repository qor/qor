package publish

import (
	"github.com/jinzhu/gorm"

	"reflect"
)

const (
	PUBLISHED = false
	DIRTY     = true
)

type Publish struct {
	PublishStatus bool
}

type DB struct {
	*gorm.DB
	SupportedModels []interface{}
}

func Open(driver, source string) (*DB, error) {
	db, err := gorm.Open(driver, source)

	db.Callback().Create().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", SetTableAndPublishStatus(true))
	db.Callback().Create().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production_after_create", SyncToProductionAfterCreate)

	db.Callback().Delete().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", SetTableAndPublishStatus(true))
	db.Callback().Delete().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production_after_delete", SyncToProductionAfterDelete)

	db.Callback().Update().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", SetTableAndPublishStatus(true))
	db.Callback().Update().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production", SyncToProductionAfterUpdate)

	db.Callback().Query().Before("gorm:query").Register("publish:set_table_in_draft_mode", SetTableAndPublishStatus(false))
	return &DB{DB: &db}, err
}

func DraftTableName(table string) string {
	return table + "_draft"
}

func (db *DB) Support(models ...interface{}) {
	db.SupportedModels = append(db.SupportedModels, models...)

	var supportedModels []string
	for _, model := range db.SupportedModels {
		supportedModels = append(supportedModels, reflect.Indirect(reflect.ValueOf(model)).Type().String())
	}
	db.InstantSet("publish:support_models", supportedModels)
}

func (db *DB) AutoMigrateDrafts() {
	for _, value := range db.SupportedModels {
		table := (&gorm.Scope{Value: value}).TableName()
		db.Table(DraftTableName(table)).AutoMigrate(value)
	}
}

func (db *DB) ProductionMode() *gorm.DB {
	return db.Set("qor_publish:draft_mode", false)
}

func (db *DB) DraftMode() *gorm.DB {
	return db.Set("qor_publish:draft_mode", true)
}

func SetTableAndPublishStatus(force bool) func(*gorm.Scope) {
	return func(scope *gorm.Scope) {
		if draftMode, ok := scope.Get("qor_publish:draft_mode"); force || ok {
			if isDraft, ok := draftMode.(bool); force || ok && isDraft {
				data := scope.IndirectValue()
				if data.Kind() == reflect.Slice {
					elem := data.Type().Elem()
					if elem.Kind() == reflect.Ptr {
						elem = elem.Elem()
					}
					data = reflect.New(elem).Elem()
				}
				currentModel := data.Type().String()

				var supportedModels []string
				if value, ok := scope.Get("publish:support_models"); ok {
					supportedModels = value.([]string)
				}

				for _, model := range supportedModels {
					if model == currentModel {
						table := scope.TableName()
						scope.InstanceSet("publish:original_table", table)
						scope.Search.TableName = DraftTableName(table)
						if isDraft {
							scope.SetColumn("PublishStatus", DIRTY)
						}
						break
					}
				}
			}
		}
	}
}

func GetModeAndNewScope(scope *gorm.Scope) (isProduction bool, clone *gorm.Scope) {
	if draftMode, ok := scope.Get("qor_publish:draft_mode"); ok && !draftMode.(bool) {
		if table, ok := scope.InstanceGet("publish:original_table"); ok {
			clone := scope.New(scope.Value)
			clone.Search.TableName = table.(string)
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
