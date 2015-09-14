package admin_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"testing"
)

func TestCreateRecord(t *testing.T) {
	form := url.Values{
		"QorResource.Name": {"create_record"},
		"QorResource.Role": {"admin"},
	}

	if req, err := http.PostForm(server.URL+"/admin/users", form); err == nil {
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

// func TestCreateRecordWithXML(t *testing.T) {
// 	xml := []byte(`<?xml version="1.0" encoding="utf-8"?>
// <soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
//   <soap:Body>
//     <ClientGetByGuid xmlns="http://tempuri.org/">
//       <guid>fc40a874-2902-4539-b8e7-6aa7084644ec</guid>
//     </ClientGetByGuid>
//   </soap:Body>
// </soap:Envelope>`)
// 	buf := bytes.NewBuffer(xml)

// 	if req, err := http.Post(server.URL+"/admin/user", "application/xml", buf); err == nil {
// 		fmt.Println(req)
// 		fmt.Println(err)
// 		t.Errorf("sss")
// 	}
// }

func TestCreateHasOneRecord(t *testing.T) {
	name := "create_record_and_has_one"
	form := url.Values{
		"QorResource.Name":              {name},
		"QorResource.Role":              {"admin"},
		"QorResource.CreditCard.Number": {"1234567890"},
		"QorResource.CreditCard.Issuer": {"Visa"},
	}

	if req, err := http.PostForm(server.URL+"/admin/users", form); err == nil {
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
		"QorResource.Addresses[2].ID":       {"0"},
		"QorResource.Addresses[2].Address1": {""},
	}

	if req, err := http.PostForm(server.URL+"/admin/users", form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name).RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address_1").RecordNotFound() {
			t.Errorf("Address 1 should be created successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address_2").RecordNotFound() {
			t.Errorf("Address 2 should be created successfully")
		}

		var addresses []Address
		if db.Find(&addresses, "user_id = ?", user.ID); len(addresses) != 2 {
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
		"QorResource.Languages": {fmt.Sprintf("%d", languageCN.ID), fmt.Sprintf("%d", languageEN.ID)},
	}

	if req, err := http.PostForm(server.URL+"/admin/users", form); err == nil {
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

		if req, err := http.Post(server.URL+"/admin/users", writer.FormDataContentType(), body); err == nil {
			if req.StatusCode != 200 {
				t.Errorf("Create request should be processed successfully")
			}

			var user User
			if db.First(&user, "name = ?", name).RecordNotFound() {
				t.Errorf("User should be created successfully")
			}

			// fmt.Println(user.Avatar.Path)
			// if user.Avatar.Path == "" {
			// 	t.Errorf("Avatar should be saved")
			// }
		}
	}
}

func TestCreateRecordWithJSON(t *testing.T) {
	name := "api_create_record"

	var languageCN Language
	var languageEN Language
	db.FirstOrCreate(&languageCN, Language{Name: "CN"})
	db.FirstOrCreate(&languageEN, Language{Name: "EN"})

	json := fmt.Sprintf(`{"Name":"api_create_record",
                        "Role":"admin",
                          "CreditCard": {"Number": "987654321", "Issuer": "Visa"},
                          "Addresses": [{"Address1": "address_1"}, {"Address1": "address_2"}, {"_id": "0"}],
                          "Languages": [%v, %v]
                       }`, languageCN.ID, languageEN.ID)

	buf := strings.NewReader(json)

	if req, err := http.Post(server.URL+"/admin/users", "application/json", buf); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Create request should be processed successfully")
		}

		var user User
		if db.First(&user, "name = ?", name).RecordNotFound() {
			t.Errorf("User should be created successfully")
		}

		if db.Model(&user).Related(&user.CreditCard).RecordNotFound() || user.CreditCard.Number != "987654321" {
			t.Errorf("Embedded struct should be created successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address_1").RecordNotFound() {
			t.Errorf("Address 1 should be created successfully")
		}

		if db.First(&Address{}, "user_id = ? and address1 = ?", user.ID, "address_2").RecordNotFound() {
			t.Errorf("Address 2 should be created successfully")
		}

		var addresses []Address
		if db.Find(&addresses, "user_id = ?", user.ID); len(addresses) != 2 {
			t.Errorf("Blank address should not be created")
		}

		var languages []Language
		db.Model(&user).Related(&languages, "Languages")

		if len(languages) != 2 {
			t.Errorf("User should have two languages after create")
		}
	}
}
