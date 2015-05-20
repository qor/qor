package database

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/i18n"
)

type Translation struct {
	Key          string `gorm:"primary_key"`
	LanguageCode string `gorm:"primary_key"`
	Value        string `sql"size:4294967295"`
}

func New(db *gorm.DB) i18n.Backend {
	return &Backend{DB: db}
}

type Backend struct {
	DB *gorm.DB
}

func (*Backend) LoadTransations() []i18n.Translation {
	return []i18n.Translation{}
}

func (*Backend) UpdateTranslation(i18n.Translation) {
}

func (*Backend) DeleteTranslation(i18n.Translation) {
}
