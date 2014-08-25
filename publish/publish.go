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

	db.Callback().Create().Before("gorm:begin_transaction").Register("publish:set_table", SetTable)
	db.Callback().Delete().Before("gorm:begin_transaction").Register("publish:set_table", SetTable)
	db.Callback().Update().Before("gorm:begin_transaction").Register("publish:set_table", SetTable)
	db.Callback().Query().Before("gorm:query").Register("publish:set_table", SetTable)
	return &Publish{DB: &db}, err
}

func DraftTableName(scope *gorm.Scope) string {
	return scope.TableName() + "_draft"
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
		publish.Table(DraftTableName(&gorm.Scope{Value: value})).AutoMigrate(value)
	}
}

func (publish *Publish) ProductionMode() *gorm.DB {
	return publish.Set("qor_publish:draft_mode", false)
}

func (publish *Publish) DraftMode() *gorm.DB {
	return publish.Set("qor_publish:draft_mode", true)
}

func SetTable(scope *gorm.Scope) {
	if draftMode, ok := scope.Get("qor_publish:draft_mode"); ok {
		if value, ok := draftMode.(bool); ok && value {
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
					scope.Search.TableName = DraftTableName(scope)
					break
				}
			}
		}
	}
}
