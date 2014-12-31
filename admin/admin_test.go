package admin_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"

	_ "github.com/mattn/go-sqlite3"
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
	Id   int
	Name string
	Role string
	// Avatar       media_library.FileSystem
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
	db.AutoMigrate(&User{}, &CreditCard{}, &Address{}, &Language{})

	admin := admin.New(&qor.Config{DB: &db})
	user := admin.NewResource(User{})
	user.Meta(&resource.Meta{Name: "Languages", Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if languages := []Language{}; !context.GetDB().Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{strconv.Itoa(language.Id), language.Name})
				}
			}
			return
		}})

	admin.MountTo("/admin", mux)

	server = httptest.NewServer(mux)
}
