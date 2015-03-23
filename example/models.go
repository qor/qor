package main

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/l10n"
	"github.com/qor/qor/media_library"
	"github.com/qor/qor/publish"
)

type CreditCard struct {
	Id     int
	Number string
	Issuer string
}

type Address struct {
	Id       int
	UserId   int64
	Address1 string
	Address2 string
}

type Role struct {
	Id   int
	Name string
}

type Language struct {
	Id   int
	Name string
}

type User struct {
	Id           int
	Name         string
	Gender       string
	Description  string
	File         media_library.FileSystem
	RoleId       int64
	Languages    []Language `gorm:"many2many:user_languages;"`
	CreditCard   CreditCard
	CreditCardId int64
	Addresses    []Address
	DeletedAt    time.Time
	publish.Status
}

type Product struct {
	ID int
	l10n.Locale
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	publish.Status
}

var DB gorm.DB
var Publish *publish.Publish

func init() {
	var err error
	DB, err = gorm.Open("sqlite3", "tmp/qor.db")
	if err != nil {
		panic(err)
	}

	DB.AutoMigrate(&User{}, &CreditCard{}, &Address{}, &Role{}, &Language{}, &Product{}, &admin.AssetManager{})

	Publish = publish.New(&DB)
	Publish.Support(&User{}, &Product{}).AutoMigrate()

	l10n.RegisterCallbacks(&DB)

	DB.FirstOrCreate(&Role{}, Role{Name: "admin"})
	DB.FirstOrCreate(&Role{}, Role{Name: "dev"})
	DB.FirstOrCreate(&Role{}, Role{Name: "customer_support"})

	DB.FirstOrCreate(&Language{}, Language{Name: "CN"})
	DB.FirstOrCreate(&Language{}, Language{Name: "JP"})
	DB.FirstOrCreate(&Language{}, Language{Name: "EN"})
	DB.FirstOrCreate(&Language{}, Language{Name: "DE"})

	DB.LogMode(true)
}
