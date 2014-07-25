package admin_test

import (
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"net/url"

	"testing"
)

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

func TestCreateHasOneRecord(t *testing.T) {
	name := "create_record_and_has_one"
	form := url.Values{
		"QorResource.Name":              {name},
		"QorResource.Role":              {"admin"},
		"QorResource.CreditCard.Number": {"1234567890"},
		"QorResource.CreditCard.Issuer": {"Visa"},
	}

	if req, err := http.PostForm(server.URL+"/admin/user", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name).RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		if db.Model(&user).Related(&user.CreditCard).RecordNotFound() || user.CreditCard.Number != "1234567890" {
			t.Errorf("Embedded struct should be created successfully")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestCreateHasManyRecord(t *testing.T) {
	name := "create_record_and_has_many"
	form := url.Values{
		"QorResource.Name":                  {name},
		"QorResource.Role":                  {"admin"},
		"QorResource.Addresses[0].Address1": {"address_1"},
		"QorResource.Addresses[1].Address1": {"address_2"},
		"QorResource.Addresses[2]._id":      {"0"},
		"QorResource.Addresses[2].Address1": {""},
	}

	if req, err := http.PostForm(server.URL+"/admin/user", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name).RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.Id, "address_1").RecordNotFound() {
			t.Errorf("Address 1 should be created successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.Id, "address_2").RecordNotFound() {
			t.Errorf("Address 2 should be created successfully")
		}

		var addresses []Address
		if db.Find(&addresses, "user_id = ?", user.Id); len(addresses) != 2 {
			t.Errorf("Blank address should not be created")
		}
	} else {
		t.Errorf(err.Error())
	}
}
