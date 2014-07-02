package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

var db gorm.DB

type User struct {
	Id   int64
	Name string
	Role string
}

func init() {
	db, _ = gorm.Open("sqlite3", "/tmp/qor.db")
}

func main() {
	db.AutoMigrate(&User{})

	var user User
	db.FirstOrCreate(&user, User{Name: "jinzhu", Role: "admin"})
	fmt.Println(user)
}
