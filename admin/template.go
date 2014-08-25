package admin

import (
	"errors"
	"fmt"
	"github.com/qor/qor/roles"
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

	if dir, err := filepath.Abs(filepath.Dir(os.Args[0])); err == nil {
		if dir := path.Join(dir, "src/github.com/qor/qor/admin/templates"); isExistingDir(dir) {
			viewDirs = append(viewDirs, dir)
		}
	}

	if dir := path.Join(os.Getenv("GOPATH"), "src/github.com/qor/qor/admin/templates"); isExistingDir(dir) {
		viewDirs = append(viewDirs, dir)
	}
}

func (content Content) getTemplate(tmpl *template.Template, layout string) (*template.Template, error) {
	paths := []string{}
	for _, p := range []string{path.Join("resources", content.Context.ResourceName), path.Join("themes", "default"), "."} {
		for _, d := range viewDirs {
			if isExistingDir(path.Join(d, p)) {
				paths = append(paths, path.Join(d, p))
			}
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(path.Join(p, layout)); !os.IsNotExist(err) {
			if tmpl, err = tmpl.ParseFiles(path.Join(p, layout)); err != nil {
				fmt.Println(err)
			} else {
				return tmpl, nil
			}
		}
	}
	return tmpl, errors.New("template not found")
}

func (admin *Admin) Render(str string, content Content, modes ...roles.PermissionMode) {
	var tmpl *template.Template

	cacheKey := path.Join(content.Context.ResourceName, str)
	if t, ok := templates[cacheKey]; !ok || true {
		str = tmplSuffix.ReplaceAllString(str, ".tmpl")

		tmpl, _ = content.getTemplate(tmpl, "layout.tmpl")
		tmpl, _ = content.getTemplate(tmpl.Funcs(content.funcMap(modes...)), str)

		for _, name := range []string{"header", "footer"} {
			if tmpl.Lookup(name) == nil {
				tmpl, _ = content.getTemplate(tmpl, name+".tmpl")
			}
		}

		templates[cacheKey] = tmpl
	} else {
		tmpl = t
	}

	if err := tmpl.Execute(content.Writer, content); err != nil {
		fmt.Println(err)
	}
}
