package admin_test

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"

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

type User struct {
	Id           int
	Name         string
	Role         string
	CreditCard   CreditCard
	CreditCardId int64
	Addresses    []Address
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
	db.AutoMigrate(&User{})
	db.AutoMigrate(&CreditCard{})
	db.AutoMigrate(&Address{})

	user := resource.New(&User{})
	admin := admin.New(&db)
	admin.AddResource(user)
	admin.AddToMux("/admin", mux)
	server = httptest.NewServer(mux)
}
