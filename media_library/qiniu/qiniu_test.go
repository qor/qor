package qiniu

import (
	"net/http"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	tu "github.com/qor/qor/test/utils"
)

var db = tu.TestDB()

type User struct {
	gorm.Model
	Name   string
	Avatar Qiniu
}

func init() {
	db.AutoMigrate(&User{})
}

func TestUploadToQiniu(t *testing.T) {
	var user = User{Name: "jinzhu"}
	if avatar, err := os.Open("../test/logo.png"); err == nil {
		user.Avatar.Scan(avatar)
		if err := db.Save(&user).Error; err == nil {
			// if _, err := os.Stat(path.Join("public", user.Avatar.URL())); err != nil {
			// 	t.Errorf("should find saved user avatar")
			// }
		} else {
			t.Errorf("should saved user successfully")
		}
	} else {
		panic("file doesn't exist")
	}

	resp, err2 := http.Get("http:" + user.Avatar.URL())
	if err2 != nil {
		t.Error(err2)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Status code is not 200, is %+v", resp.Status)
	}

	var newUser User
	db.First(&newUser, user.ID)
	newUser.Avatar.Scan(`{"CropOption": {"X": 5, "Y": 5, "Height": 50, "Width": 50}, "Crop": true}`)
	db.Save(&newUser)

	if newUser.Avatar.URL() == user.Avatar.URL() {
		t.Errorf("url should be different after crop")
	}
}
