package admin

import (
	"github.com/qor/qor"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"text/template"
)

var layouts = map[string]*template.Template{}
var templates = map[string]*template.Template{}
var tmplSuffix = regexp.MustCompile(`(\.tmpl)?$`)
var viewDirs = []string{}

func isExistingDir(pth string) bool {
	fi, err := os.Stat(pth)
	if err != nil {
		return false
	}
	return fi.Mode().IsDir()
}

func init() {
	if root := os.Getenv("WEB_ROOT"); root != "" {
		if dir := path.Join(root, "templates/qor"); isExistingDir(dir) {
			viewDirs = append(viewDirs, dir)
		}
	}

	if dir, err := filepath.Abs("templates/qor"); err == nil && isExistingDir(dir) {
		viewDirs = append(viewDirs, dir)
	}

	if dir := path.Join(os.Getenv("GOROOT"), "site/src/github.com/qor/qor/admin/templates"); isExistingDir(dir) {
		viewDirs = append(viewDirs, dir)
	}
}

func (admin *Admin) Render(str string, context *qor.Context) {
	var tmpl *template.Template

	str = tmplSuffix.ReplaceAllString(str, ".tmpl")
	pathWithResourceName := path.Join(context.ResourceName, str)

	var paths []string
	for _, value := range []string{pathWithResourceName, str} {
		if root != "" {
			paths = append(paths, filepath.Join(root, value))
		}
		if p, e := filepath.Abs(value); e == nil {
			paths = append(paths, p)
		}
		paths = append(paths, filepath.Join(goroot, "site/src/github.com/qor/qor/admin", value))
	}

	for _, value := range paths {
		_, err := os.Stat(value)
		if !os.IsNotExist(err) {
			t, _ := template.ParseFiles(value)
			templates[value] = t
			tmpl = t
			break
		}
	}
	if tmpl != nil {
		tmpl.Execute(context.Writer, context)
	}
}
