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
	ID     int
	Number string
	Issuer string
}

type Address struct {
	ID       int
	UserId   int64
	Address1 string
	Address2 string
}

type Role struct {
	ID   int
	Name string
}

type Language struct {
	ID   int
	Name string
}

type User struct {
	ID           int
	Name         string
	Gender       string
	Description  string
	File         media_library.FileSystem
	RoleID       int
	Languages    []Language `gorm:"many2many:user_languages;"`
	CreditCard   CreditCard
	CreditCardID int
	Addresses    []Address
	DeletedAt    time.Time
	publish.Status
}

func (User) ViewableLocales() []string {
	return []string{"zh-CN", "JP", "EN", "DE"}
}

func (User) EditableLocales() []string {
	return []string{"zh-CN", "EN"}
}

func (u User) DisplayName() string {
	return u.Name
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
	Publish.Support(&Product{}).AutoMigrate()

	l10n.RegisterCallbacks(&DB)

	var AdminRole Role
	DB.FirstOrCreate(&AdminRole, Role{Name: "admin"})
	DB.FirstOrCreate(&Role{}, Role{Name: "dev"})
	DB.FirstOrCreate(&Role{}, Role{Name: "customer_support"})

	DB.FirstOrCreate(&User{}, User{Name: "admin", RoleID: AdminRole.ID})

	DB.FirstOrCreate(&Language{}, Language{Name: "CN"})
	DB.FirstOrCreate(&Language{}, Language{Name: "JP"})
	DB.FirstOrCreate(&Language{}, Language{Name: "EN"})
	DB.FirstOrCreate(&Language{}, Language{Name: "DE"})

	DB.LogMode(true)
}
