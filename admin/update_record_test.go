package admin_test

import (
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"net/url"
	"strconv"

	"testing"
)

func TestUpdateRecord(t *testing.T) {
	user := User{Name: "update_record", Role: "admin"}
	db.Save(&user)

	form := url.Values{
		"QorResource.Name": {user.Name + "_new"},
		"QorResource.Role": {"admin"},
	}

	if req, err := http.PostForm(server.URL+"/admin/user/"+strconv.Itoa(user.Id), form); err == nil {
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
		"QorResource.CreditCard.Id":     {strconv.Itoa(user.CreditCard.Id)},
		"QorResource.CreditCard.Number": {"1234567890"},
		"QorResource.CreditCard.Issuer": {"UnionPay"},
	}

	if req, err := http.PostForm(server.URL+"/admin/user/"+strconv.Itoa(user.Id), form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("User request should be processed successfully")
		}

		if db.First(&User{}, "name = ?", user.Name+"_new").RecordNotFound() {
			t.Errorf("User should be updated successfully")
		}

		var creditCard CreditCard
		if db.Model(&user).Related(&creditCard).RecordNotFound() ||
			creditCard.Issuer != "UnionPay" || creditCard.Id != user.CreditCard.Id {
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
	user := User{Name: "update_record_and_has_many", Role: "admin", Addresses: []Address{{Address1: "address 1"}, {Address1: "address 2"}}}
	db.Save(&user)

	form := url.Values{
		"QorResource.Name":                  {user.Name},
		"QorResource.Role":                  {"admin"},
		"QorResource.Addresses[0].Id":       {strconv.Itoa(user.Addresses[0].Id)},
		"QorResource.Addresses[0].Address1": {"address 1 new"},
		"QorResource.Addresses[1].Id":       {strconv.Itoa(user.Addresses[1].Id)},
		"QorResource.Addresses[1].Address1": {"address 2 new"},
		"QorResource.Addresses[2].Address1": {"address 3"},
	}

	if req, err := http.PostForm(server.URL+"/admin/user/"+strconv.Itoa(user.Id), form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.Id, "address 1 new").RecordNotFound() {
			t.Errorf("Address 1 should be updated successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.Id, "address 2 new").RecordNotFound() {
			t.Errorf("Address 2 should be updated successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.Id, "address 3").RecordNotFound() {
			t.Errorf("Address 3 should be created successfully")
		}
		var addresses []Address
		if db.Find(&addresses, "user_id = ?", user.Id); len(addresses) != 3 {
			t.Errorf("Addresses's count should be updated after update")
		}
	} else {
		t.Errorf(err.Error())
	}
}
