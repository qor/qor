package database

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/i18n"
)

type Translation struct {
	gorm.Model
	Locale string `sql:"size:12;"`
	Key    string `sql:"size:4294967295;"`
	Value  string `sql:"size:4294967295"`
}

func New(db *gorm.DB) i18n.Backend {
	db.AutoMigrate(&Translation{})
	quotedKey := db.NewScope(&Translation{}).Quote("key") + "(190)"
	db.Model(&Translation{}).AddUniqueIndex("idx_translations_key_with_locale", "locale", quotedKey)
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

func (backend *Backend) SaveTranslation(t *i18n.Translation) error {
	return backend.DB.Where(Translation{Key: t.Key, Locale: t.Locale}).
		Assign(Translation{Value: t.Value}).
		FirstOrCreate(&Translation{}).Error
}

func (backend *Backend) DeleteTranslation(t *i18n.Translation) error {
	return backend.DB.Where(Translation{Key: t.Key, Locale: t.Locale}).Delete(&Translation{}).Error
}
