package admin

import (
	"fmt"
	"os"
	"testing"

	"github.com/qor/qor"
)

func TestGetDefaultViewPaths(t *testing.T) {
	Admin := New(&qor.Config{})
	context := &Context{Admin: Admin}

	viewPaths := context.getViewPaths()
	currentFilePath, _ := os.Getwd()

	if viewPaths[0] != fmt.Sprintf("%v/views", currentFilePath) {
		t.Error("default view path not returned")
	}

	// TODO: this may hide a logic problem
	if viewPaths[0] != viewPaths[1] {
		t.Error("Why getViewPaths() generate 2 same paths ?")
	}
}

func TestGetCustomizeViewPaths(t *testing.T) {
	Admin := New(&qor.Config{})
	context := &Context{Admin: Admin}
	customizeTheme := "theme_for_test"
	userWithCustomizeTheme := Admin.AddResource(&User{}, &Config{Theme: customizeTheme})

	context.SetResource(userWithCustomizeTheme)

	viewPaths := context.getViewPaths()
	currentFilePath, _ := os.Getwd()

	if viewPaths[0] != fmt.Sprintf("%v/views/themes/%v/user", currentFilePath, customizeTheme) {
		t.Error("customize view path not returned")
	}
}
