package admin_test

import (
	"bytes"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

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
			fmt.Println(addresses)
			t.Errorf("Blank address should not be created")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestCreateManyToManyRecord(t *testing.T) {
	name := "create_record_many_to_many"
	var languageCN Language
	var languageEN Language
	db.FirstOrCreate(&languageCN, Language{Name: "CN"})
	db.FirstOrCreate(&languageEN, Language{Name: "EN"})

	form := url.Values{
		"QorResource.Name":      {name},
		"QorResource.Role":      {"admin"},
		"QorResource.Languages": {strconv.Itoa(languageCN.Id), strconv.Itoa(languageEN.Id)},
	}

	if req, err := http.PostForm(server.URL+"/admin/user", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name).RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		var languages []Language
		db.Model(&user).Related(&languages, "Languages")

		if len(languages) != 2 {
			t.Errorf("User should have two languages after create")
		}
	} else {
		t.Errorf(err.Error())
	}
}

func TestUploadAttachment(t *testing.T) {
	name := "create_record_upload_attachment"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if attachment, err := filepath.Abs("tests/qor.png"); err == nil {
		if part, err := writer.CreateFormFile("QorResource.Avatar", filepath.Base(attachment)); err == nil {
			if file, err := os.Open(attachment); err == nil {
				io.Copy(part, file)
			}
		}
		form := url.Values{
			"QorResource.Name": {name},
			"QorResource.Role": {"admin"},
		}
		for key, val := range form {
			_ = writer.WriteField(key, val[0])
		}
		writer.Close()

		if req, err := http.Post(server.URL+"/admin/user", writer.FormDataContentType(), body); err == nil {
			if req.StatusCode != 200 {
				t.Errorf("Create request should be processed successfully")
			}

			var user User
			if db.First(&user, "name = ?", name).RecordNotFound() {
				t.Errorf("User should be created successfully")
			}

			fmt.Println(user)
			// fmt.Println(user.Avatar.Path)
			// if user.Avatar.Path == "" {
			// 	t.Errorf("Avatar should be saved")
			// }
		}
	}
}
