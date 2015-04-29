package admin_test

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
)

func TestTextInput(t *testing.T) {
	user := Admin.AddResource(&User{})
	meta := &admin.Meta{Name: "Name"}
	user.Meta(meta)

	if meta.Label != "Name" {
		t.Error("default label not set")
	}

	if meta.Alias != "Name" {
		t.Error("default Alias is not same as field Name")
	}

	if meta.Type != "string" {
		t.Error("default Type is not string")
	}

	oldUserName := "old name"
	newUserName := "new name"

	// Valuer
	userRecord := &User{Name: oldUserName}
	db.Create(&userRecord)
	value := meta.Valuer(userRecord, &qor.Context{Config: &qor.Config{DB: &db}})

	if *value.(*string) != oldUserName {
		t.Error("resource's value doesn't get")
	}

	// Setter
	metaValue := &resource.MetaValue{
		Name:  "User.Name",
		Value: newUserName,
		Meta:  meta,
	}

	meta.Setter(userRecord, metaValue, &qor.Context{Config: &qor.Config{DB: &db}})
	if userRecord.Name != newUserName {
		t.Error("resource's value doesn't set")
	}
}

// TODO: there maybe a bug of file Type meta generation meta.go#L142
func TestDefaultMetaType(t *testing.T) {
	user := Admin.AddResource(&User{})
	booleanMeta := &admin.Meta{Name: "Active"}
	timeMeta := &admin.Meta{Name: "RegisteredAt"}
	numberMeta := &admin.Meta{Name: "CreditCardId"}
	// fileMeta := &admin.Meta{Name: "Avatar"}
	user.Meta(booleanMeta)
	user.Meta(timeMeta)
	user.Meta(numberMeta)
	// user.Meta(fileMeta)

	if booleanMeta.Type != "checkbox" {
		t.Error("boolean field doesn't set as checkbox")
	}

	if timeMeta.Type != "datetime" {
		t.Error("time field doesn't set as datetime")
	}

	if numberMeta.Type != "number" {
		t.Error("number field doesn't set as number")
	}

	// if fileMeta.Type != "file" {
	// 	t.Error("file field doesn't set as file")
	// }
}

func TestMeta(t *testing.T) {
	profileModel := Profile{
		Name:  "Qor",
		Sex:   "Female",
		Phone: Phone{Num: "1024"},
	}
	userModel := &User{Profile: profileModel}
	db.Create(userModel)

	user := Admin.AddResource(&User{})
	profileNameMeta := &admin.Meta{Name: "Profile.Name"}
	user.Meta(profileNameMeta)
	profileSexMeta := &admin.Meta{Name: "Profile.Sex"}
	user.Meta(profileSexMeta)
	phoneNumMeta := &admin.Meta{Name: "Profile.Phone.Num"}
	user.Meta(phoneNumMeta)

	userModel.Profile = Profile{}
	valx := phoneNumMeta.Valuer(userModel, &qor.Context{Config: &qor.Config{DB: &db}})
	if val, ok := valx.(*string); !ok || *val != profileModel.Phone.Num {
		t.Errorf("Profile.Phone.Num: got %q; expect %q", *val, profileModel.Phone.Num)
	}
	if userModel.Profile.Name != profileModel.Name {
		t.Errorf("Profile.Name: got %q; expect %q", userModel.Profile.Name, profileModel.Name)
	}
	if userModel.Profile.Sex != profileModel.Sex {
		t.Errorf("Profile.Sex: got %q; expect %q", userModel.Profile.Sex, profileModel.Sex)
	}
	if userModel.Profile.Phone.Num != profileModel.Phone.Num {
		t.Errorf("Profile.Phone.Num: got %q; expect %q", userModel.Profile.Phone.Num, profileModel.Phone.Num)
	}

	mvs := &resource.MetaValues{
		Values: []*resource.MetaValue{
			&resource.MetaValue{
				Name:  "Profile.Name",
				Value: "Qor III",
				Meta:  profileNameMeta,
			},
			&resource.MetaValue{
				Name:  "Profile.Sex",
				Value: "Male",
				Meta:  profileSexMeta,
			},
			&resource.MetaValue{
				Name:  "Profile.Phone.Num",
				Value: "2048",
				Meta:  phoneNumMeta,
			},
		},
	}
	profileNameMeta.Setter(userModel, mvs.Values[0], &qor.Context{Config: &qor.Config{DB: &db}})
	if userModel.Profile.Name != mvs.Values[0].Value {
		t.Errorf("Profile.Name: got %q; expect %q", userModel.Profile.Name, mvs.Values[0].Value)
	}
	profileSexMeta.Setter(userModel, mvs.Values[1], &qor.Context{Config: &qor.Config{DB: &db}})
	if userModel.Profile.Sex != mvs.Values[1].Value {
		t.Errorf("Profile.Sex: got %q; expect %q", userModel.Profile.Sex, mvs.Values[1].Value)
	}
	phoneNumMeta.Setter(userModel, mvs.Values[2], &qor.Context{Config: &qor.Config{DB: &db}})
	if userModel.Profile.Phone.Num != mvs.Values[2].Value {
		t.Errorf("Profile.Phone.Num: got %q; expect %q", userModel.Profile.Phone.Num, mvs.Values[2].Value)
	}
}
