package l10n_test

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/l10n"
	"github.com/qor/qor/test/utils"
)

type Product struct {
	ID              int    `gorm:"primary_key"`
	Code            string `l10n:"sync"`
	Quantity        uint   `l10n:"sync"`
	Name            string
	DeletedAt       *time.Time
	ColorVariations []ColorVariation
	Brand           Brand
	Tags            []Tag `gorm:"many2many:product_tags"`
	l10n.Locale
}

// func (Product) LocaleCreateable() {}

type ColorVariation struct {
	ID       int `gorm:"primary_key"`
	Quantity int
	Color    Color
}

type Color struct {
	ID   int `gorm:"primary_key"`
	Code string
	Name string
	l10n.Locale
}

type Brand struct {
	ID int `gorm:"primary_key"`
	l10n.Locale
}

type Tag struct {
	ID   int `gorm:"primary_key"`
	Name string
	l10n.Locale
}

var dbGlobal, dbCN, dbEN *gorm.DB

func init() {
	db := utils.TestDB()
	l10n.RegisterCallbacks(&db)

	db.DropTableIfExists(&Product{})
	db.DropTableIfExists(&Tag{})
	db.Exec("drop table product_tags;")
	db.AutoMigrate(&Product{})
	db.AutoMigrate(&Tag{})

	dbGlobal = &db
	dbCN = dbGlobal.Set("l10n:locale", "zh")
	dbEN = dbGlobal.Set("l10n:locale", "en")
}
