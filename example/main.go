package main

import (
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor"
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

type Role struct {
	Id   int
	Name string
}

type Language struct {
	Id   int
	Name string
}

type User struct {
	Id        int64
	Name      string
	Gender    string
	RoleId    int64
	Languages []Language `orm:"many2many:user_languages;"`
	// db.Model(User).Related(Languages, "Languages")
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

func main() {
	mux := http.NewServeMux()

	user := resource.New(&User{})
	user.Attrs().Index("name", "gender")
	user.Meta().Register(resource.Meta{Name: "gender", Type: "select_one", Collection: []string{"M", "F", "U"}})
	user.Meta().Register(resource.Meta{Name: "RoleId", Label: "Role", Type: "select_one",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			var roles []Role
			context.DB.Find(&roles)
			for _, role := range roles {
				results = append(results, []string{strconv.Itoa(role.Id), role.Name})
			}
			return
		},
	})

	// db.Model(&User).Relate(&Languages{})
	user.Meta().Register(resource.Meta{Name: "Languages", Type: "select_many",
		Value: func(interface{}, *qor.Context) interface{} {
			languages := []Language{}
			db.Find(&languages)
			return languages
		},
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			languages := []Language{}
			db.Find(&languages)
			for _, language := range languages {
				results = append(results, []string{strconv.Itoa(language.Id), language.Name})
			}
			return results
		},
		Setter: func(resource interface{}, value interface{}, context *qor.Context) {
		},
	})

	role := resource.New(&Role{})
	language := resource.New(&Language{})

	admin := admin.New(&db)
	admin.AddResource(user)
	admin.AddResource(role)
	admin.AddResource(language)
	admin.AddToMux("/admin", mux)

	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
