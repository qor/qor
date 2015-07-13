package database

import (
	"fmt"
	"unicode/utf8"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/i18n"
)

/* when using a database with DEFAULT CHARACTER SET utf8mb4 on mysql the table
creation will fail when not limited in size for key,locale: the compound primary
key would exceed the size limit:
Error 1071: Specified key was too long; max key length is 767 bytes
ideally we'd want to only limit the length to which the index uses the fields,
but 202/8 should be sufficient for keys and locales.
*/
type Translation struct {
	Key    string `gorm:"primary_key" sql:"size:202"`
	Locale string `gorm:"primary_key" sql:"size:8"`
	Value  string `sql:"size:4294967295"`
}

func New(db *gorm.DB) i18n.Backend {
	db.AutoMigrate(&Translation{})
	return &Backend{DB: db}
}

type Backend struct {
	DB *gorm.DB
}

func (backend *Backend) LoadTranslations() []*i18n.Translation {
	var translations []*i18n.Translation
	backend.DB.Find(&translations)
	return translations
}

func (backend *Backend) SaveTranslation(t *i18n.Translation) (err error) {
	if utf8.RuneCountInString(t.Key) > 202 {
		err = fmt.Errorf("Translation key is too long: %v. Maximum key length is 202: %s", len(t.Key), t.Key)
		return
	}
	if len(t.Locale) > 8 {
		err = fmt.Errorf("Translation locale is too long: %v. Maximum locale length is 8: %s", len(t.Locale), t.Locale)
		return
	}
	backend.DB.Where(Translation{Key: t.Key, Locale: t.Locale}).Assign(Translation{Value: t.Value}).FirstOrCreate(&Translation{})
	return
}

func (backend *Backend) DeleteTranslation(t *i18n.Translation) {
	backend.DB.Where(Translation{Key: t.Key, Locale: t.Locale}).Delete(&Translation{})
}
