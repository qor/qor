package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/admin"

	"net/http"
)

type User struct {
	Id   int64
	Name string
	Role string
}

var db gorm.DB

func init() {
	db, _ = gorm.Open("sqlite3", "/tmp/qor.db")
	db.AutoMigrate(&User{})

	var user User
	db.FirstOrCreate(&user, User{Name: "jinzhu", Role: "admin"})
}

func main() {
	mux := http.NewServeMux()

	admin := admin.New()
	admin.AddToMux("/admin", mux)

	http.ListenAndServe(":8080", mux)
}
