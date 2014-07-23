package admin_test

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
	"net/http"
	"net/http/httptest"
	"net/url"

	"testing"
)

type CreditCard struct {
	Id     int64
	Number string
	Issuer string
}

type User struct {
	Id           int64
	Name         string
	Role         string
	CreditCard   CreditCard
	CreditCardId int64
}

var server *httptest.Server
var db gorm.DB

func init() {
	mux := http.NewServeMux()
	db, _ = gorm.Open("sqlite3", "/tmp/qor_test.db")
	db.LogMode(true)
	db.DropTable(&User{})
	db.DropTable(&CreditCard{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&CreditCard{})

	user := resource.New(&User{})
	admin := admin.New(&db)
	admin.AddResource(user)
	admin.AddToMux("/admin", mux)
	server = httptest.NewServer(mux)
}

func TestCreateRecord(t *testing.T) {
	form := url.Values{
		"QorResource.Name": {"create_record"},
		"QorResource.Role": {"admin"},
	}

	if req, err := http.PostForm(server.URL+"/admin/user", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		if db.First(&User{}, "name = ?", "create_record").RecordNotFound() {
			t.Errorf("User should be created successfully")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestCreateRecordWithEmbeddedStruct(t *testing.T) {
	form := url.Values{
		"QorResource.Name":              {"create_record_with_embedded_struct"},
		"QorResource.Role":              {"admin"},
		"QorResource.CreditCard.Number": {"1234567890"},
		"QorResource.CreditCard.Issuer": {"Visa"},
	}

	if req, err := http.PostForm(server.URL+"/admin/user", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", "create_record_with_embedded_struct").RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		if db.Model(&user).Related(&user.CreditCard).RecordNotFound() || user.CreditCard.Number != "1234567890" {
			t.Errorf("Embedded struct should be created successfully")
		}
	} else {
		t.Errorf(err.Error())
	}
}
