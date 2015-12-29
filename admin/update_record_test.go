package admin_test

import (
	"fmt"
	"net/http"
	"net/url"

	_ "github.com/mattn/go-sqlite3"

	"testing"
)

func TestUpdateRecord(t *testing.T) {
	user := User{Name: "update_record", Role: "admin"}
	db.Save(&user)

	form := url.Values{
		"QorResource.Name": {user.Name + "_new"},
		"QorResource.Role": {"admin"},
	}

	if req, err := http.PostForm(server.URL+"/admin/users/"+fmt.Sprint(user.ID), form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		if db.First(&User{}, "name = ?", user.Name+"_new").RecordNotFound() {
			t.Errorf("User should be updated successfully")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestUpdateHasOneRecord(t *testing.T) {
	user := User{Name: "update_record_and_has_one", Role: "admin", CreditCard: CreditCard{Number: "1234567890", Issuer: "JCB"}}
	db.Save(&user)

	form := url.Values{
		"QorResource.Name":              {user.Name + "_new"},
		"QorResource.Role":              {"admin"},
		"QorResource.CreditCard.ID":     {fmt.Sprint(user.CreditCard.ID)},
		"QorResource.CreditCard.Number": {"1234567890"},
		"QorResource.CreditCard.Issuer": {"UnionPay"},
	}

	if req, err := http.PostForm(server.URL+"/admin/users/"+fmt.Sprint(user.ID), form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("User request should be processed successfully")
		}

		if db.First(&User{}, "name = ?", user.Name+"_new").RecordNotFound() {
			t.Errorf("User should be updated successfully")
		}

		var creditCard CreditCard
		if db.Model(&user).Related(&creditCard).RecordNotFound() ||
			creditCard.Issuer != "UnionPay" || creditCard.ID != user.CreditCard.ID {
			t.Errorf("Embedded struct should be updated successfully")
		}

		if !db.First(&CreditCard{}, "number = ? and issuer = ?", "1234567890", "JCB").RecordNotFound() {
			t.Errorf("Old embedded struct should be updated")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestUpdateHasManyRecord(t *testing.T) {
	user := User{Name: "update_record_and_has_many", Role: "admin", Addresses: []Address{{Address1: "address 1.1", Address2: "address 1.2"}, {Address1: "address 2.1"}, {Address1: "address 3.1"}}}
	db.Save(&user)

	form := url.Values{
		"QorResource.Name":                  {user.Name},
		"QorResource.Role":                  {"admin"},
		"QorResource.Addresses[0].ID":       {fmt.Sprint(user.Addresses[0].ID)},
		"QorResource.Addresses[0].Address1": {"address 1.1 new"},
		"QorResource.Addresses[1].ID":       {fmt.Sprint(user.Addresses[1].ID)},
		"QorResource.Addresses[1].Address1": {"address 2.1 new"},
		"QorResource.Addresses[2].ID":       {fmt.Sprint(user.Addresses[2].ID)},
		"QorResource.Addresses[2]._destroy": {"1"},
		"QorResource.Addresses[2].Address1": {"address 3.1"},
		"QorResource.Addresses[3].Address1": {"address 4.1"},
	}

	if req, err := http.PostForm(server.URL+"/admin/users/"+fmt.Sprint(user.ID), form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var address1 Address
		if db.First(&address1, "user_id = ? and address1 = ?", user.ID, "address 1.1 new").RecordNotFound() {
			t.Errorf("Address 1 should be updated successfully")
		} else if address1.Address2 != "address 1.2" {
			t.Errorf("Address 1's Address 2 should not be updated")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address 2.1 new").RecordNotFound() {
			t.Errorf("Address 2 should be updated successfully")
		}

		if !db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address 3.1").RecordNotFound() {
			t.Errorf("Address 3 should be destroyed successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address 4.1").RecordNotFound() {
			t.Errorf("Address 4 should be created successfully")
		}

		var addresses []Address
		if db.Find(&addresses, "user_id = ?", user.ID); len(addresses) != 3 {
			t.Errorf("Addresses's count should be updated after update")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestDestroyEmbeddedHasOneRecord(t *testing.T) {
	user := User{Name: "destroy_embedded_has_one_record", Role: "admin", CreditCard: CreditCard{Number: "1234567890", Issuer: "JCB"}}
	db.Save(&user)

	form := url.Values{
		"QorResource.Name":                {user.Name + "_new"},
		"QorResource.Role":                {"admin"},
		"QorResource.CreditCard.ID":       {fmt.Sprint(user.CreditCard.ID)},
		"QorResource.CreditCard._destroy": {"1"},
		"QorResource.CreditCard.Number":   {"1234567890"},
		"QorResource.CreditCard.Issuer":   {"UnionPay"},
	}

	if req, err := http.PostForm(server.URL+"/admin/users/"+fmt.Sprint(user.ID), form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("User request should be processed successfully")
		}

		var newUser User
		if db.First(&newUser, "name = ?", user.Name+"_new").RecordNotFound() {
			t.Errorf("User should be updated successfully")
		}

		if !db.Model(&newUser).Related(&CreditCard{}).RecordNotFound() {
			t.Errorf("Embedded struct should be destroyed successfully")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestUpdateManyToManyRecord(t *testing.T) {
	name := "update_record_many_to_many"
	var languageCN Language
	var languageEN Language
	db.FirstOrCreate(&languageCN, Language{Name: "CN"})
	db.FirstOrCreate(&languageEN, Language{Name: "EN"})
	user := User{Name: name, Role: "admin", Languages: []Language{languageCN, languageEN}}
	db.Save(&user)

	form := url.Values{
		"QorResource.Name":      {name + "_new"},
		"QorResource.Role":      {"admin"},
		"QorResource.Languages": {fmt.Sprint(languageCN.ID)},
	}

	if req, err := http.PostForm(server.URL+"/admin/users/"+fmt.Sprint(user.ID), form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Update request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name+"_new").RecordNotFound() {
			t.Errorf("User should be updated successfully")
		}

		var languages []Language
		db.Model(&user).Related(&languages, "Languages")

		if len(languages) != 1 {
			t.Errorf("User should have one languages after update")
		}
	} else {
		t.Errorf(err.Error())
	}
}
