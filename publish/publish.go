package publish

import (
	"github.com/jinzhu/gorm"

	"reflect"
)

type Publish struct {
	*gorm.DB
	SupportedModels []interface{}
}

func Open(driver, source string) (*Publish, error) {
	db, err := gorm.Open(driver, source)

	db.Callback().Create().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", SetTable(true))
	db.Callback().Create().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production_after_create", SyncToProductionAfterCreate)

	db.Callback().Delete().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", SetTable(true))
	db.Callback().Delete().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production_after_delete", SyncToProductionAfterDelete)

	db.Callback().Update().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", SetTable(true))
	db.Callback().Update().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production", SyncToProductionAfterUpdate)

	db.Callback().Query().Before("gorm:query").Register("publish:set_table_in_draft_mode", SetTable(false))
	return &Publish{DB: &db}, err
}

func DraftTableName(table string) string {
	return table + "_draft"
}

func (publish *Publish) Support(models ...interface{}) {
	publish.SupportedModels = append(publish.SupportedModels, models...)

	var supportedModels []string
	for _, model := range publish.SupportedModels {
		supportedModels = append(supportedModels, reflect.Indirect(reflect.ValueOf(model)).Type().String())
	}
	publish.InstantSet("publish:support_models", supportedModels)
}

func (publish *Publish) AutoMigrateDrafts() {
	for _, value := range publish.SupportedModels {
		table := (&gorm.Scope{Value: value}).TableName()
		publish.Table(DraftTableName(table)).AutoMigrate(value)
	}
}

func (publish *Publish) ProductionMode() *gorm.DB {
	return publish.Set("qor_publish:draft_mode", false)
}

func (publish *Publish) DraftMode() *gorm.DB {
	return publish.Set("qor_publish:draft_mode", true)
}

func SetTable(force bool) func(*gorm.Scope) {
	return func(scope *gorm.Scope) {
		if draftMode, ok := scope.Get("qor_publish:draft_mode"); force || ok {
			if value, ok := draftMode.(bool); force || ok && value {
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
						break
					}
				}
			}
		}
	}
}

func SyncToProductionAfterCreate(scope *gorm.Scope) {
	if draftMode, ok := scope.Get("qor_publish:draft_mode"); ok && !draftMode.(bool) {
		if table, ok := scope.InstanceGet("publish:original_table"); ok {
			clone := scope.New(scope.Value)
			clone.Search.TableName = table.(string)
			gorm.Create(clone)
		}
	}
}

func SyncToProductionAfterUpdate(scope *gorm.Scope) {
	if draftMode, ok := scope.Get("qor_publish:draft_mode"); ok && !draftMode.(bool) {
		if table, ok := scope.InstanceGet("publish:original_table"); ok {
			clone := scope.New(scope.Value)
			clone.Search.TableName = table.(string)
			gorm.Update(clone)
		}
	}
}

func SyncToProductionAfterDelete(scope *gorm.Scope) {
	if draftMode, ok := scope.Get("qor_publish:draft_mode"); ok && !draftMode.(bool) {
		if table, ok := scope.InstanceGet("publish:original_table"); ok {
			clone := scope.New(scope.Value)
			clone.Search.TableName = table.(string)
			gorm.Delete(clone)
		}
	}
}
