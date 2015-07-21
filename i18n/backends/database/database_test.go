package database_test

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/i18n"
	"github.com/qor/qor/i18n/backends/database"
	"github.com/qor/qor/test/utils"
)

var db *gorm.DB
var backend i18n.Backend

func init() {
	db = utils.TestDB()
	db.DropTable(&database.Translation{})
	backend = database.New(db)
}

func TestTranslations(t *testing.T) {
	translation := i18n.Translation{Key: "hello_world", Value: "Hello World", Locale: "zh-CN"}

	backend.SaveTranslation(&translation)
	if len(backend.LoadTranslations()) != 1 {
		t.Errorf("should has only one translation")
	}

	backend.DeleteTranslation(&translation)
	if len(backend.LoadTranslations()) != 0 {
		t.Errorf("should has none translation")
	}
}
