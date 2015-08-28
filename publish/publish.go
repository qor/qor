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

	publishDraftMode = "publish:draft_mode"
	publishEventMode = "publish:publish_event"
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

func IsDraftMode(db *gorm.DB) bool {
	if draftMode, ok := db.Get(publishDraftMode); ok {
		if isDraft, ok := draftMode.(bool); ok && isDraft {
			return true
		}
	}
	return false
}

func isPublishableModel(model interface{}) (ok bool) {
	if model != nil {
		_, ok = reflect.New(utils.ModelType(model)).Interface().(publishInterface)
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
				typ := utils.ModelType(db.Value)
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

				var forceDraftTable bool
				if forceDraftTable, ok := db.Get("publish:force_draft_table"); ok {
					if forceMode, ok := forceDraftTable.(bool); ok && forceMode {
						forceDraftTable = true
					}
				}

				if IsDraftMode(db) || forceDraftTable {
					return draftTableName(tableName)
				}
			}
		}
		return tableName
	}

	db.AutoMigrate(&PublishEvent{})

	db.Callback().Create().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", setTableAndPublishStatus(true))
	db.Callback().Create().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production_after_create", syncCreateFromProductionToDraft)

	db.Callback().Delete().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", setTableAndPublishStatus(true))
	db.Callback().Delete().Replace("gorm:delete", deleteScope)
	db.Callback().Delete().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production_after_delete", syncDeleteFromProductionToDraft)

	db.Callback().Update().Before("gorm:begin_transaction").Register("publish:set_table_to_draft", setTableAndPublishStatus(true))
	db.Callback().Update().Before("gorm:commit_or_rollback_transaction").
		Register("publish:sync_to_production", syncUpdateFromProductionToDraft)

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
	return db.DB.Set(publishDraftMode, false)
}

func (db Publish) DraftDB() *gorm.DB {
	return db.DB.Set(publishDraftMode, true)
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
