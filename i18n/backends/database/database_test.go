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

	longText := "Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

	backend.SaveTranslation(&i18n.Translation{Key: longText + "1", Value: longText, Locale: "zh-CN"})
	backend.SaveTranslation(&i18n.Translation{Key: longText + "2", Value: longText, Locale: "zh-CN"})

	if len(backend.LoadTranslations()) != 2 {
		t.Errorf("should has two translations")
	}

	backend.DeleteTranslation(&i18n.Translation{Key: longText + "1", Value: longText, Locale: "zh-CN"})
	if len(backend.LoadTranslations()) != 1 {
		t.Errorf("should has one translation left")
	}
}
