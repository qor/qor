package main

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type Author struct {
	gorm.Model
	Id   int64
	Name string
}

type Book struct {
	gorm.Model
	Id          int64
	Title       string
	Synopsis    string
	ReleaseDate time.Time
	Authors     []*Author `gorm:one2many:authors`
	Price       float64
	// add locale stuff here...
}

var DB gorm.DB

func init() {
	var err error
	dbuser, dbpwd := "qor_tutorial", "qor_tutorial"

	// if os.Getenv("WEB_ENV") == "online" {
	// 	dbuser, dbpwd = os.Getenv("DB_USER"), os.Getenv("DB_PWD")
	// }

	DB, err = gorm.Open(
		"mysql",
		fmt.Sprintf("%s:%s@/qor_bookstore?charset=utf8&parseTime=True&loc=Local", dbuser, dbpwd),
	)
	if err != nil {
		panic(err)
	}

	DB.AutoMigrate(&Author{}, &Book{})
	DB.LogMode(true)
}
