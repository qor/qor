package publish

import (
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/utils"

	"reflect"
)

const (
	PUBLISHED = false
	DIRTY     = true
)

type publishInterface interface {
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

func (s Status) InjectQorAdmin(res *admin.Resource) {
	if res.GetMeta("PublishStatus") == nil {
		res.IndexAttrs(append(res.IndexAttrs(), "-PublishStatus")...)
		res.ShowAttrs(append(res.ShowAttrs(), "-PublishStatus")...)
		res.EditAttrs(append(res.EditAttrs(), "-PublishStatus")...)
		res.NewAttrs(append(res.NewAttrs(), "-PublishStatus")...)
	}
}

type Publish struct {
	DB *gorm.DB
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

func isPublishableModel(model interface{}) (ok bool) {
	if model != nil {
		_, ok = reflect.New(modelType(model)).Interface().(publishInterface)
	}
	return
}

var injectedJoinTableHandler = map[reflect.Type]bool{}

func New(db *gorm.DB) *Publish {
	tableHandler := gorm.DefaultTableNameHandler
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		tableName := tableHandler(db, defaultTableName)

		if db != nil {
			if isPublishableModel(db.Value) {
				// Set join table handler
				typ := modelType(db.Value)
				if !injectedJoinTableHandler[typ] {
					injectedJoinTableHandler[typ] = true
					scope := db.NewScope(db.Value)
					for _, field := range scope.GetModelStruct().StructFields {
						if many2many := utils.ParseTagOption(field.Tag.Get("gorm"))["MANY2MANY"]; many2many != "" {
							db.SetJoinTableHandler(db.Value, field.Name, &publishJoinTableHandler{})
							db.AutoMigrate(db.Value)
						}
					}
				}

				var forceDraftMode = false
				if forceMode, ok := db.Get("publish:force_draft_mode"); ok {
					if forceMode, ok := forceMode.(bool); ok && forceMode {
						forceDraftMode = true
					}
				}

				if draftMode, ok := db.Get("publish:draft_mode"); ok {
					if isDraft, ok := draftMode.(bool); ok && isDraft || forceDraftMode {
						return draftTableName(tableName)
					}
				}
			}
		}
		return tableName
	}

	db.Callback().Create().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", setTableAndPublishStatus(true))
	db.Callback().Create().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production_after_create", syncToProductionAfterCreate)

	db.Callback().Delete().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", setTableAndPublishStatus(true))
	db.Callback().Delete().Replace("gorm:delete", deleteScope)
	db.Callback().Delete().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production_after_delete", syncToProductionAfterDelete)

	db.Callback().Update().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", setTableAndPublishStatus(true))
	db.Callback().Update().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production", syncToProductionAfterUpdate)

	db.Callback().RowQuery().Register("publish:set_table_in_draft_mode", setTableAndPublishStatus(false))
	db.Callback().Query().Before("gorm:query").Register("publish:set_table_in_draft_mode", setTableAndPublishStatus(false))
	return &Publish{DB: db}
}

func draftTableName(table string) string {
	return originalTableName(table) + "_draft"
}

func originalTableName(table string) string {
	return strings.TrimSuffix(table, "_draft")
}

func (db *Publish) AutoMigrate(values ...interface{}) {
	for _, value := range values {
		tableName := db.DB.NewScope(value).TableName()
		db.DraftDB().Table(draftTableName(tableName)).AutoMigrate(value)
	}
}

func (db Publish) ProductionDB() *gorm.DB {
	return db.DB.Set("publish:draft_mode", false)
}

func (db Publish) DraftDB() *gorm.DB {
	return db.DB.Set("publish:draft_mode", true)
}

func (db Publish) newResolver(records ...interface{}) *resolver {
	return &resolver{Records: records, DB: db.DB, Dependencies: map[string]*dependency{}}
}

func (db Publish) Publish(records ...interface{}) {
	db.newResolver(records...).Publish()
}

func (db Publish) Discard(records ...interface{}) {
	db.newResolver(records...).Discard()
}
