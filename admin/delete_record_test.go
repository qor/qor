package admin_test

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

func TestDeleteRecord(t *testing.T) {
	user := User{Name: "delete_record", Role: "admin"}
	db.Save(&user)
	form := url.Values{
		"_method": {"delete"},
	}

	if req, err := http.PostForm(server.URL+"/admin/users/"+strconv.Itoa(user.Id), form); err == nil {
		if req.StatusCode != 200 {
			t.Errorf("Delete request should be processed successfully")
		}

		if !db.First(&User{}, "name = ?", "delete_record").RecordNotFound() {
			t.Errorf("User should be deleted successfully")
		}
	} else {
		t.Errorf(err.Error())
	}
}
