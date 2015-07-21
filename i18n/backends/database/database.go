package database

import (
	"sync"

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
	db.Model(&Translation{}).AddIndex("idx_translations_key_with_locale", "locale", "`key`(190)")
	return &Backend{DB: db, mutex: &sync.Mutex{}}
}

type Backend struct {
	DB    *gorm.DB
	mutex *sync.Mutex
}

func (backend *Backend) LoadTranslations() []*i18n.Translation {
	var translations []*i18n.Translation
	backend.DB.Find(&translations)
	return translations
}

func (backend *Backend) SaveTranslation(t *i18n.Translation) {
	backend.mutex.Lock()
	backend.DB.Where(Translation{Key: t.Key, Locale: t.Locale}).Assign(Translation{Value: t.Value}).FirstOrCreate(&Translation{})
	backend.mutex.Unlock()
}

func (backend *Backend) DeleteTranslation(t *i18n.Translation) {
	backend.mutex.Lock()
	backend.DB.Where(Translation{Key: t.Key, Locale: t.Locale}).Delete(&Translation{})
	backend.mutex.Unlock()
}
