package admin

import (
	"errors"
	"os"
	"path"
	"regexp"
	"strings"

	"html/template"
)

var layouts = map[string]*template.Template{}
var templates = map[string]*template.Template{}
var tmplSuffix = regexp.MustCompile(`(\.tmpl)?$`)
var viewPaths = []string{}
var Root, _ = os.Getwd()

func init() {
	if root := os.Getenv("WEB_ROOT"); root != "" {
		Root = root
	}

	for _, dir := range []string{path.Join(Root, "app/views/qor")} {
		RegisterViewPath(dir)
	}

	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/admin/views"))
	}
}

func RegisterViewPath(path string) error {
	if isExistingDir(path) {
		var found bool

		for _, viewPath := range viewPaths {
			if path == viewPath {
				found = true
				break
			}
		}

		if !found {
			viewPaths = append(viewPaths, path)
		}
		return nil
	}
	return errors.New("path not found")
}

func isExistingDir(pth string) bool {
	fi, err := os.Stat(pth)
	if err != nil {
		return false
	}
	return fi.Mode().IsDir()
}
