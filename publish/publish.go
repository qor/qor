package publish

import (
	"strings"

	"github.com/jinzhu/gorm"

	"reflect"
)

const (
	PUBLISHED = false
	DIRTY     = true
)

type Interface interface {
	GetPublishStatus() bool
	SetPublishStatus(bool)
}

type Status struct {
	PublishStatus bool
}

func (s Status) GetPublishStatus() bool {
	return s.PublishStatus
}

func (s *Status) SetPublishStatus(status bool) {
	s.PublishStatus = status
}

type Publish struct {
	DB              *gorm.DB
	SupportedModels []interface{}
}

func modelType(value interface{}) reflect.Type {
	reflectValue := reflect.Indirect(reflect.ValueOf(value))

	if reflectValue.Kind() == reflect.Slice {
		typ := reflectValue.Type().Elem()
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		return typ
	}

	return reflectValue.Type()
}

func IsPublishableModel(model interface{}) bool {
	_, ok := reflect.New(modelType(model)).Interface().(Interface)
	return ok
}

func New(db *gorm.DB) *Publish {
	tableHandler := gorm.DefaultTableNameHandler
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		tableName := tableHandler(db, defaultTableName)

		if db != nil {
			if IsPublishableModel(db.Value) {
				var forceDraftMode = false
				if forceMode, ok := db.Get("publish:force_draft_mode"); ok {
					if forceMode, ok := forceMode.(bool); ok && forceMode {
						forceDraftMode = true
					}
				}

				if draftMode, ok := db.Get("publish:draft_mode"); ok {
					if isDraft, ok := draftMode.(bool); ok && isDraft || forceDraftMode {
						return DraftTableName(tableName)
					}
				}
			}
		}
		return tableName
	}

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

	db.Callback().RowQuery().Register("publish:set_table_in_draft_mode", SetTableAndPublishStatus(false))
	db.Callback().Query().Before("gorm:query").Register("publish:set_table_in_draft_mode", SetTableAndPublishStatus(false))
	return &Publish{DB: db}
}

func DraftTableName(table string) string {
	return OriginalTableName(table) + "_draft"
}

func OriginalTableName(table string) string {
	return strings.TrimSuffix(table, "_draft")
}

func (db *Publish) AutoMigrate(values ...interface{}) {
	for _, value := range values {
		tableName := db.DB.NewScope(value).TableName()
		db.DraftDB().Table(DraftTableName(tableName)).AutoMigrate(value)
	}
}

func (db *Publish) ProductionDB() *gorm.DB {
	return db.DB.Set("publish:draft_mode", false)
}

func (db *Publish) DraftDB() *gorm.DB {
	return db.DB.Set("publish:draft_mode", true)
}

func (db *Publish) NewResolver(records ...interface{}) *Resolver {
	return &Resolver{Records: records, DB: db, Dependencies: map[string]*Dependency{}}
}

func (db *Publish) Publish(records ...interface{}) {
	db.NewResolver(records...).Publish()
}

func (db *Publish) Discard(records ...interface{}) {
	db.NewResolver(records...).Discard()
}
