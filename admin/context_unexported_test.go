package admin

import (
	"fmt"
	"os"
	"path"
	"strings"
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
	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/admin/tests/views"))
	}

	Admin := New(&qor.Config{})
	context := &Context{Admin: Admin}

	customizeTheme := "theme_for_test"
	userWithCustomizeTheme := Admin.AddResource(&User{}, &Config{Theme: customizeTheme})

	context.SetResource(userWithCustomizeTheme)

	viewPaths := context.getViewPaths()
	currentFilePath, _ := os.Getwd()

	if viewPaths[0] != fmt.Sprintf("%v/tests/views/themes/%v/user", currentFilePath, customizeTheme) {
		t.Error("customize view path not returned")
	}
}
