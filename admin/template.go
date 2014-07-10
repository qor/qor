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

	cacheKey := path.Join(context.ResourceName, str)
	if t, ok := templates[cacheKey]; !ok && false {
		str = tmplSuffix.ReplaceAllString(str, ".tmpl")

		tmpl = template.New("template")

		// parse layout
		paths := []string{}
		for _, p := range []string{path.Join("resources", context.ResourceName), path.Join("themes", "default")} {
			for _, d := range viewDirs {
				if isExistingDir(path.Join(d, p)) {
					paths = append(paths, path.Join(d, p))
				}
			}
		}

		for _, f := range []string{"layout.tmpl", str} {
			for _, p := range paths {
				if _, err := os.Stat(path.Join(p, f)); !os.IsNotExist(err) {
					tmpl, err = tmpl.ParseFiles(path.Join(p, f))
					break
				}
			}
		}

		for _, name := range []string{"header", "footer"} {
			if tmpl.Lookup(name) == nil {
				for _, p := range paths {
					if _, err := os.Stat(path.Join(p, name+".tmpl")); !os.IsNotExist(err) {
						tmpl, err = tmpl.ParseFiles(path.Join(p, name+".tmpl"))
						break
					}
				}
			}
		}

		templates[cacheKey] = tmpl
	} else {
		tmpl = t
	}

	if tmpl != nil {
		tmpl.Execute(context.Writer, context)
	}
}
