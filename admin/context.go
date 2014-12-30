package admin

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"text/template"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

type Context struct {
	*qor.Context
	*Resource
	*Searcher
	Admin   *Admin
	Writer  http.ResponseWriter
	Result  interface{}
	Content string
}

// Resource
func (context *Context) SetResource(res *Resource) *Context {
	context.Resource = res
	context.Searcher = &Searcher{Resource: res, Admin: context.Admin}
	return context
}

func (context *Context) GetResource(name string) *Context {
	context.SetResource(context.Admin.GetResource(name))
	return context
}

// Template
func (context *Context) findTemplate(tmpl *template.Template, layout string) (*template.Template, error) {
	paths := []string{}
	for _, p := range []string{context.Name, path.Join("themes", "default"), "."} {
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

func (context *Context) Render(name string) string {
	var err error
	var tmpl = template.New(name + ".tmpl")

	if tmpl, err = context.findTemplate(tmpl, name+".tmpl"); err == nil {
		var result = bytes.NewBufferString("")
		if err := tmpl.Execute(result, context); err != nil {
			fmt.Println(err)
		}
		return result.String()
	}

	return ""
}

func (context *Context) Execute(name string, result interface{}) {
	var tmpl *template.Template

	cacheKey := path.Join(context.Name, name)
	if t, ok := templates[cacheKey]; !ok || true {
		tmpl, _ = context.findTemplate(tmpl, "layout.tmpl")

		for _, name := range []string{"header", "footer"} {
			if tmpl.Lookup(name) == nil {
				tmpl, _ = context.findTemplate(tmpl, name+".tmpl")
			}
		}
	} else {
		tmpl = t
	}

	context.Content = context.Render(name)
	context.Result = result
	if err := tmpl.Execute(context.Writer, context); err != nil {
		fmt.Println(err)
	}
}

// Function Maps
func (context *Context) ValueOf(value interface{}, meta *resource.Meta) interface{} {
	return meta.Value(value, context.Context)
}

// context.NewSearcher().FindAll
// context.NewSearcher().FindOne

//// Controller
// results := context.FindAll
// result := context.FindOne
// context.Execute("show", result)
// context.Render("show", result)
//// VIEW
// results := GetResource("order").FindAll
// GetResource("order").Render "index", result
// admin.GetResource("order")

// $order := admin.NewContext(request, writer).GetResource("order").Scope("today").FindAll
// admin.NewContext(request, writer).GetResource("order").Render("index", $order)
// admin.NewContext(request, writer).Render("dashboard")
