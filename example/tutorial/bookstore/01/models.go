package main

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/l10n"
	"github.com/qor/qor/media_library"
	"github.com/qor/qor/publish"
)

type Author struct {
	gorm.Model
	Name string
	publish.Status
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
	CoverImage  media_library.FileSystem
	// later
	// CoverImages []ProductImage // product image has BookId => handles relation

	publish.Status
}

type ProductImage struct {
	gorm.Model
	BookId     uint
	CoverImage media_library.FileSystem
	publish.Status
}

// step 4 - add users

type User struct {
	gorm.Model
	Name string
}

func (u User) DisplayName() string {
	return u.Name
}

var (
	db           gorm.DB
	pub          *publish.Publish
	productionDB *gorm.DB
	stagingDB    *gorm.DB
)

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

	db.AutoMigrate(&Author{}, &Book{}, &User{})
	db.LogMode(true)

	pub = publish.New(&db)
	pub.AutoMigrate(&Author{}, &Book{})

	stagingDB = pub.DraftDB()         // Draft resources are saved here
	productionDB = pub.ProductionDB() // Published resources are saved here

	// step 4
	l10n.RegisterCallbacks(&db)
}
