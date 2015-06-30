package media_library_test

import (
	"image"
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

			var newUser User
			db.First(&newUser, user.ID)
			newUser.Avatar.Scan(`{"CropOptions": {"original": {"X": 5, "Y": 5, "Height": 10, "Width": 10}}, "Crop": true}`)
			db.Save(&newUser)

			if newUser.Avatar.URL() == user.Avatar.URL() {
				t.Errorf("url should be different after crop")
			}

			file, err := os.Open(path.Join("public", newUser.Avatar.URL()))
			if err != nil {
				t.Errorf("Failed open croped image")
			}

			if image, _, err := image.DecodeConfig(file); err == nil {
				if image.Width != 10 || image.Height != 10 {
					t.Errorf("image should be croped successfully")
				}
			} else {
				t.Errorf("Failed to decode croped image")
			}
		} else {
			t.Errorf("should saved user successfully")
		}
	} else {
		panic("file doesn't exist")
	}
}
