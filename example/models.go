package main

import "github.com/jinzhu/gorm"

type CreditCard struct {
	Id     int64
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
	Id           int64
	Name         string
	Gender       string
	RoleId       int64
	Languages    []Language `gorm:"many2many:user_languages;"`
	CreditCard   CreditCard
	CreditCardId int64
	Addresses    []Address
}

var db gorm.DB

func init() {
	db, _ = gorm.Open("sqlite3", "/tmp/qor.db")
	db.LogMode(true)
	db.AutoMigrate(&User{})
	db.AutoMigrate(&CreditCard{})
	db.AutoMigrate(&Address{})
	db.AutoMigrate(&Role{})
	db.AutoMigrate(&Language{})

	db.FirstOrCreate(&Role{}, Role{Name: "admin"})
	db.FirstOrCreate(&Role{}, Role{Name: "dev"})
	db.FirstOrCreate(&Role{}, Role{Name: "customer_support"})

	db.FirstOrCreate(&Language{}, Role{Name: "CN"})
	db.FirstOrCreate(&Language{}, Role{Name: "JP"})
	db.FirstOrCreate(&Language{}, Role{Name: "EN"})
	db.FirstOrCreate(&Language{}, Role{Name: "DE"})
}
