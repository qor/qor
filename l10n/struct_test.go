package l10n_test

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/l10n"
)

type Product struct {
	ID              int    `gorm:"primary_key"`
	Code            string `l10n:"sync"`
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
	// CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor';
	// CREATE DATABASE qor_l10n;
	// GRANT ALL ON qor_l10n.* TO 'gorm'@'localhost';
	dbuser, dbpwd := "qor", "qor"
	if os.Getenv("TEST_ENV") == "CI" {
		dbuser, dbpwd = os.Getenv("DB_USER"), os.Getenv("DB_PWD")
	}
	db, _ := gorm.Open("mysql", fmt.Sprintf("%s:%s@/qor_test?charset=utf8&parseTime=True&loc=Local", dbuser, dbpwd))
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
