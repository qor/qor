package admin

import (
	"errors"
	"os"
	"path"
	"regexp"
	"strings"

	"html/template"
)

var (
	layouts    = map[string]*template.Template{}
	templates  = map[string]*template.Template{}
	tmplSuffix = regexp.MustCompile(`(\.tmpl)?$`)
	viewPaths  = []string{}
	root, _    = os.Getwd()
)

func init() {
	if path := os.Getenv("WEB_ROOT"); path != "" {
		root = path
	}

	registerViewPath(path.Join(root, "app/views/qor"))
	RegisterViewPath("github.com/qor/qor/admin/views")
}

// RegisterViewPath register views directory
func RegisterViewPath(p string) {
	if registerViewPath(path.Join(root, "vendor", p)) != nil {
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			if registerViewPath(path.Join(gopath, "src", p)) == nil {
				return
			}
		}
	}
}

func isExistingDir(pth string) bool {
	fi, err := os.Stat(pth)
	if err != nil {
		return false
	}
	return fi.Mode().IsDir()
}

func registerViewPath(path string) error {
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
