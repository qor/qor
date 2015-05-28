package main

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type Author struct {
	gorm.Model
	// Id   int64 // gets created by gorm automatically
	Name string
}

type Book struct {
	gorm.Model
	// Id          int64 // gets created by gorm automatically
	Title       string
	Synopsis    string
	ReleaseDate time.Time
	Authors     []*Author `gorm:"many2many:book_authors"`
	Price       float64
	// add locale stuff here...
}

// type User struct {
// 	gorm.Model
// 	Name string
// }

var DB gorm.DB

func init() {
	var err error
	dbuser, dbpwd := "qor", "qor"

	// if os.Getenv("WEB_ENV") == "online" {
	// 	dbuser, dbpwd = os.Getenv("DB_USER"), os.Getenv("DB_PWD")
	// }

	DB, err = gorm.Open(
		"mysql",
		fmt.Sprintf("%s:%s@/qor_bookstore?parseTime=True&loc=Local", dbuser, dbpwd),
	)
	if err != nil {
		panic(err)
	}

	// DB.AutoMigrate(&Author{}, &Book{}, &User{})
	DB.AutoMigrate(&Author{}, &Book{})
	DB.LogMode(true)
}
