package resource

import (
	"flag"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Id        uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	Profile Profile
}

type Profile struct {
	Id        uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	UserId uint64
	Name   string
	Sex    string

	Phone Phone
}

type Phone struct {
	Id        uint64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time

	ProfileId uint64
	Num       string
}

var db, _ = gorm.Open("sqlite3", "/tmp/qor_resource_test.db")
var debug bool

func init() {
	flag.BoolVar(&debug, "qor.debug", false, "print out debug information")
	flag.Parse()
	if debug {
		db = *db.Debug()
	}

	db.DropTable(&User{})
	db.DropTable(&Profile{})
	db.DropTable(&Phone{})
	db.CreateTable(&User{})
	db.CreateTable(&Profile{})
	db.CreateTable(&Phone{})
}

func TestMeta(t *testing.T) {
	profileModel := Profile{
		Name:  "Qor",
		Sex:   "Female",
		Phone: Phone{Num: "1024"},
	}
	userModel := &User{Profile: profileModel}
	db.Create(userModel)

	user := New(&User{})
	user.RegisterMeta(&Meta{Name: "Profile.Name"})
	user.RegisterMeta(&Meta{Name: "Profile.Sex"})
	user.RegisterMeta(&Meta{Name: "Profile.Phone.Num"})

	userModel.Profile = Profile{}
	phoneNumMeta := user.Metas["Profile.Phone.Num"].(*Meta)
	valx := phoneNumMeta.Value(userModel, &qor.Context{Config: &qor.Config{DB: &db}})
	if val, ok := valx.(string); !ok || val != profileModel.Phone.Num {
		t.Errorf("Profile.Phone.Num: expect %q got %q", profileModel.Phone.Num, val)
	}
	if userModel.Profile.Name != profileModel.Name {
		t.Errorf("Profile.Name: expect %q got %q", profileModel.Name, userModel.Profile.Name)
	}
	if userModel.Profile.Sex != profileModel.Sex {
		t.Errorf("Profile.Sex: expect %q got %q", profileModel.Sex, userModel.Profile.Sex)
	}
	if userModel.Profile.Phone.Num != profileModel.Phone.Num {
		t.Errorf("Profile.Phone.Num: expect %q got %q", profileModel.Phone.Num, userModel.Profile.Phone.Num)
	}

	profileNameMeta := user.Metas["Profile.Name"].(*Meta)
	mvs := &MetaValues{
		Values: []*MetaValue{
			&MetaValue{
				Name:  "Profile.Name",
				Value: "Qor III",
				Meta:  profileNameMeta,
			},
			&MetaValue{
				Name:  "Profile.Phone.Num",
				Value: "2048",
				Meta:  phoneNumMeta,
			},
		},
	}
	profileNameMeta.Setter(userModel, mvs, &qor.Context{Config: &qor.Config{DB: &db}})
	if userModel.Profile.Name != mvs.Values[0].Value {
		t.Errorf("Profile.Name: expect %q got %q", mvs.Values[0].Value, userModel.Profile.Name)
	}
	phoneNumMeta.Setter(userModel, mvs, &qor.Context{Config: &qor.Config{DB: &db}})
	if userModel.Profile.Phone.Num != mvs.Values[1].Value {
		t.Errorf("Profile.Phone.Num: expect %q got %q", mvs.Values[1].Value, userModel.Profile.Phone.Num)
	}
}
