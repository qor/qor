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
var root = os.Getenv("WEB_ROOT")
var goroot = os.Getenv("GOROOT")
var htmlSuffix = regexp.MustCompile(`(\.html)?$`)

func (admin *Admin) Render(str string, context *qor.Context) {
	var tmpl *template.Template

	str = path.Join("templates", htmlSuffix.ReplaceAllString(str, ".html"))
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
		if os.IsExist(err) {
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
