package admin

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/qor/qor/roles"

	"text/template"
)

var layouts = map[string]*template.Template{}
var templates = map[string]*template.Template{}
var tmplSuffix = regexp.MustCompile(`(\.tmpl)?$`)
var viewPaths = []string{}

func RegisterViewPath(path string) error {
	if isExistingDir(path) {
		viewPaths = append(viewPaths, path)
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

var Root, _ = os.Getwd()

func init() {
	if root := os.Getenv("WEB_ROOT"); root != "" {
		Root = root
	}

	for _, dir := range []string{path.Join(Root, "app/views/qor"), path.Join(Root, "templates/qor")} {
		RegisterViewPath(dir)
	}

	for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
		RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/admin/templates"))
	}
}

func (content Content) getTemplate(tmpl *template.Template, layout string) (*template.Template, error) {
	paths := []string{}
	for _, p := range []string{content.Context.ResourceName, path.Join("themes", "default"), "."} {
		for _, d := range viewPaths {
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

func (admin *Admin) RenderError(err error, code int, c *Context) {
	stacks := append([]byte(err.Error()+"\n"), debug.Stack()...)
	data := struct {
		Code int
		Body string
	}{
		Code: code,
		Body: string(bytes.Replace(stacks, []byte("\n"), []byte("<br>"), -1)),
	}
	c.Writer.WriteHeader(data.Code)
	admin.Render("error", Content{Admin: admin, Context: c, Result: data})
}
