package media_library_test

import (
	"os"
	"path"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/media_library"
	"github.com/qor/qor/test/utils"
)

var db = utils.TestDB()

type User struct {
	gorm.Model
	Name   string
	Avatar media_library.FileSystem
}

func init() {
	db.AutoMigrate(&User{})
}

func TestSaveIntoFileSystem(t *testing.T) {
	var user = User{Name: "jinzhu"}
	if avatar, err := os.Open("test/logo.png"); err == nil {
		user.Avatar.Scan(avatar)
		if err := db.Save(&user).Error; err == nil {
			if _, err := os.Stat(path.Join("public", user.Avatar.URL())); err != nil {
				t.Errorf("should find saved user avatar")
			}
		} else {
			t.Errorf("should saved user successfully")
		}
	} else {
		panic("file doesn't exist")
	}
}
