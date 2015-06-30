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

type MyFileSystem struct {
	media_library.FileSystem
}

func (MyFileSystem) GetSizes() map[string]media_library.Size {
	return map[string]media_library.Size{
		"small1": {20, 10},
		"small2": {20, 10},
		"square": {30, 30},
		"big":    {50, 50},
	}
}

type User struct {
	gorm.Model
	Name   string
	Avatar MyFileSystem
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
			newUser.Avatar.Scan(`{"CropOptions": {"small1": {"X": 5, "Y": 5, "Height": 10, "Width": 20}, "small2": {"X": 0, "Y": 0, "Height": 10, "Width": 20}}, "Crop": true}`)
			db.Save(&newUser)

			if newUser.Avatar.URL() == user.Avatar.URL() {
				t.Errorf("url should be different after crop")
			}

			file, err := os.Open(path.Join("public", newUser.Avatar.URL("small1")))
			if err != nil {
				t.Errorf("Failed open croped image")
			}

			if image, _, err := image.DecodeConfig(file); err == nil {
				if image.Width != 20 || image.Height != 10 {
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
