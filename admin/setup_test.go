package admin_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/media_library"
	"github.com/qor/qor/test/utils"

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
	Id           int
	Name         string
	Role         string
	Active       bool
	RegisteredAt time.Time
	Avatar       media_library.FileSystem
	CreditCard   CreditCard
	CreditCardId int64
	Addresses    []Address
	Languages    []Language `gorm:"many2many:user_languages;"`

	Profile Profile
}

type Profile struct {
	Id        uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	UserId uint64
	Name   string
	Sex    string

	Phone Phone
}

type Phone struct {
	Id        uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	ProfileId uint64
	Num       string
}

var (
	server *httptest.Server
	db     gorm.DB
	Admin  *admin.Admin
)

func init() {
	mux := http.NewServeMux()
	db = utils.TestDB()
	models := []interface{}{&User{}, &CreditCard{}, &Address{}, &Language{}, &Profile{}, &Phone{}}
	for _, value := range models {
		db.DropTableIfExists(value)
		db.AutoMigrate(value)
	}

	Admin = admin.New(&qor.Config{DB: &db})
	user := Admin.AddResource(&User{})
	user.Meta(&admin.Meta{Name: "Languages", Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if languages := []Language{}; !context.GetDB().Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{strconv.Itoa(language.Id), language.Name})
				}
			}
			return
		}})

	Admin.MountTo("/admin", mux)

	server = httptest.NewServer(mux)
}
