package main

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/l10n"
	"github.com/qor/qor/media_library"
)

type Author struct {
	gorm.Model
	Name string
	// // step 6- maybe...: l10n
	// ShortBiography string
	// l10n.Locale
}

type Book struct {
	gorm.Model
	Title       string
	Synopsis    string
	ReleaseDate time.Time
	Authors     []*Author `gorm:"many2many:book_authors"`
	Price       float64

	//	tutorial step 3
	CoverImage media_library.FileSystem

	// //	tutorial step 3
	// CoverImages []ProductImage // product image has BookId => handles relation

	// tutorial step 4 - translation

}

type ProductImage struct {
	gorm.Model
	BookId     uint
	CoverImage media_library.FileSystem
}

// step 4 - add users

type User struct {
	gorm.Model
	Name string
}

func (u User) DisplayName() string {
	return u.Name
}

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

	// step 1-3
	// DB.AutoMigrate(&Author{}, &Book{})
	// step 5:
	DB.AutoMigrate(&Author{}, &Book{}, &User{})
	DB.LogMode(true)

	// step 4
	l10n.RegisterCallbacks(&DB)
}
