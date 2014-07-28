package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"

	"net/http"
)

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

type User struct {
	Id           int64
	Name         string
	Gender       string
	Role         string
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
}

func main() {
	mux := http.NewServeMux()

	user := resource.New(&User{})
	user.Meta().Register(resource.Meta{Name: "gender", Type: "select_one", Collection: []string{"M", "F", "U"}})

	admin := admin.New(&db)
	admin.AddResource(user)
	admin.AddToMux("/admin", mux)

	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
