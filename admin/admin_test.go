package admin_test

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/media_library"
	"github.com/qor/qor/resource"
	"strconv"

	"net/http"
	"net/http/httptest"
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

type Language struct {
	Id   int
	Name string
}

type User struct {
	Id           int
	Name         string
	Role         string
	Avatar       media_library.FileSystem
	CreditCard   CreditCard
	CreditCardId int64
	Addresses    []Address
	Languages    []Language `gorm:"many2many:user_languages;"`
}

var server *httptest.Server
var db gorm.DB

func init() {
	mux := http.NewServeMux()
	db, _ = gorm.Open("sqlite3", "/tmp/qor_test.db")
	// db.LogMode(true)
	db.DropTable(&User{})
	db.DropTable(&CreditCard{})
	db.DropTable(&Address{})
	db.DropTable(&Language{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&CreditCard{})
	db.AutoMigrate(&Address{})
	db.AutoMigrate(&Language{})

	user := resource.New(&User{})
	user.Meta().Register(resource.Meta{Name: "Languages", Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if languages := []Language{}; !context.DB.Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{strconv.Itoa(language.Id), language.Name})
				}
			}
			return
		}})

	admin := admin.New(&db)
	admin.AddResource(user)
	admin.AddToMux("/admin", mux)
	server = httptest.NewServer(mux)
}
