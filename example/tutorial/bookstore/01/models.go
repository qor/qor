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

var db gorm.DB

func init() {
	var err error
	dbuser, dbpwd := "qor", "qor"

	db, err = gorm.Open(
		"mysql",
		fmt.Sprintf("%s:%s@/qor_bookstore?parseTime=True&loc=Local", dbuser, dbpwd),
	)
	if err != nil {
		panic(err)
	}

	// step 1-3
	// db.AutoMigrate(&Author{}, &Book{})
	// step 5:
	db.AutoMigrate(&Author{}, &Book{}, &User{})
	db.LogMode(true)

	// step 4
	l10n.RegisterCallbacks(&db)
}
