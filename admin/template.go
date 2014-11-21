package admin

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/qor/qor/roles"

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

var Root, _ = os.Getwd()

func init() {
	if root := os.Getenv("WEB_ROOT"); root != "" {
		Root = root
	}

	for _, dir := range []string{path.Join(Root, "app/views/qor"), path.Join(Root, "templates/qor")} {
		if isExistingDir(dir) {
			viewDirs = append(viewDirs, dir)
		}
	}

	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		if dir := path.Join(gopath, "src/github.com/qor/qor/admin/templates"); isExistingDir(dir) {
			viewDirs = append(viewDirs, dir)
		}
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

	if err := tmpl.Execute(content.Context.Writer, content); err != nil {
		fmt.Println(err)
	}
}
