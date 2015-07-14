package database

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/i18n"
)

type Translation struct {
	Key    string `gorm:"primary_key"`
	Locale string `gorm:"primary_key"`
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

func (backend *Backend) SaveTranslation(t *i18n.Translation) {
	backend.DB.Where(Translation{Key: t.Key, Locale: t.Locale}).Assign(Translation{Value: t.Value}).FirstOrCreate(&Translation{})
}

func (backend *Backend) DeleteTranslation(t *i18n.Translation) {
	backend.DB.Where(Translation{Key: t.Key, Locale: t.Locale}).Delete(&Translation{})
}
