package publish

import "github.com/jinzhu/gorm"

type Publish struct {
	*gorm.DB
	SupportedModels []interface{}
}

func Open(driver, source string) (*Publish, error) {
	db, err := gorm.Open(driver, source)
	return &Publish{DB: &db}, err
}

func DraftTableName(scope *gorm.Scope) string {
	return scope.TableName() + "_draft"
}

func (publish *Publish) Support(models ...interface{}) {
	publish.SupportedModels = append(publish.SupportedModels, models...)
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
			scope.Search.TableName = DraftTableName(scope)
		}
	}
}
