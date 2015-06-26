package aliyun_test

import (
	"os"
	"testing"

	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/media_library/aliyun"
	"github.com/qor/qor/test/utils"
)

var db = utils.TestDB()

type User struct {
	gorm.Model
	Name   string
	Avatar aliyun.OSS
}

func init() {
	db.AutoMigrate(&User{})
}

func TestSaveIntoAliyun(t *testing.T) {
	var user = User{Name: "jinzhu"}
	avatar, err := os.Open("../test/logo.png")
	if err != nil {
		t.Error(err)
	}

	user.Avatar.Scan(avatar)
	err1 := db.Save(&user).Error

	if err1 != nil {
		t.Error(err1)
	}

	resp, err2 := http.Get("http:" + user.Avatar.URL())

	if err2 != nil {
		t.Error(err2)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Status code is not 200, is %+v", resp.Status)
	}

}
