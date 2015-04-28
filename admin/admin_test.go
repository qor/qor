package admin

import (
	"testing"

	"github.com/qor/qor"
)

type User struct {
	Name string
}

func TestAddResource(t *testing.T) {
	admin := New(&qor.Config{})
	user := admin.AddResource(&User{})

	if user != admin.resources[0] {
		t.Error("resource not added")
	}

	if admin.GetMenus()[0].Name != "User" {
		t.Error("resource not added to menu")
	}
}

func TestAddResourceWithInvisibleOption(t *testing.T) {
	admin := New(&qor.Config{})
	user := admin.AddResource(&User{}, &Config{Invisible: true})

	if user != admin.resources[0] {
		t.Error("resource not added")
	}

	if len(admin.GetMenus()) != 0 {
		t.Error("invisible resource registered in menu")
	}
}

func TestGetResource(t *testing.T) {
	admin := New(&qor.Config{})
	user := admin.AddResource(&User{})

	if admin.GetResource("user") != user {
		t.Error("resource not returned")
	}
}

func TestNewResource(t *testing.T) {
	admin := New(&qor.Config{})
	user := admin.NewResource(&User{})

	if user.Name != "User" {
		t.Error("default resource name didn't set")
	}

	if user.Config.PageCount != DEFAULT_PAGE_COUNT {
		t.Error("default page count didn't set")
	}
}

type UserWithCustomizedName struct{}

func (u *UserWithCustomizedName) ResourceName() string {
	return "CustomizedName"
}

func TestNewResourceWithCustomizedName(t *testing.T) {
	admin := New(&qor.Config{})
	user := admin.NewResource(&UserWithCustomizedName{})

	if user.Name != "CustomizedName" {
		t.Error("customize resource name didn't set")
	}
}

type UserWithInjector struct{}

func (u *UserWithInjector) InjectQorAdmin(res *Resource) {
	res.Config.PageCount = DEFAULT_PAGE_COUNT + 10
}

func TestNewResourceWithInjector(t *testing.T) {
	admin := New(&qor.Config{})
	user := admin.NewResource(&UserWithInjector{})

	if user.Config.PageCount != DEFAULT_PAGE_COUNT+10 {
		t.Error("injected page count didn't set")
	}
}
