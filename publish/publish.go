package publish

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"

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
	db.Callback().Delete().Replace("gorm:delete", Delete)
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
	for _, model := range models {
		scope := gorm.Scope{Value: model}
		for _, column := range []string{"DeletedAt", "PublishStatus"} {
			if !scope.HasColumn(column) {
				qor.ExitWithMsg("%v has no %v column", model, column)
			}
		}
	}

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
