package admin

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/qor/qor"
)

type Context struct {
	*qor.Context
	*Searcher
	Flashs   []Flash
	Resource *Resource
	Admin    *Admin
	Result   interface{}
	Content  template.HTML
}

func (context *Context) clone() *Context {
	return &Context{
		Context:  context.Context,
		Searcher: context.Searcher,
		Flashs:   context.Flashs,
		Resource: context.Resource,
		Admin:    context.Admin,
		Result:   context.Result,
		Content:  context.Content,
	}
}

// Resource
func (context *Context) resourcePath() string {
	if context.Resource == nil {
		return ""
	}
	return context.Resource.ToParam()
}

func (context *Context) setResource(res *Resource) *Context {
	context.Resource = res
	context.Searcher = &Searcher{Context: context}
	return context
}

func (context *Context) GetResource(name string) *Resource {
	return context.Admin.GetResource(name)
}

// Template
func (context *Context) getViewPaths() (paths []string) {
	var dirs = []string{context.resourcePath(), path.Join("themes", "default"), "."}
	var themes []string

	if context.Request != nil {
		if theme := context.Request.URL.Query().Get("theme"); theme != "" {
			themePath := path.Join("themes", theme)
			themes = append(themes, []string{path.Join(themePath, context.resourcePath()), themePath}...)
		}
	}

	if context.Resource != nil {
		for _, theme := range context.Resource.Config.Themes {
			themePath := path.Join("themes", theme)
			themes = append(themes, []string{path.Join(themePath, context.resourcePath()), themePath}...)
		}
	}

	for _, p := range append(themes, dirs...) {
		for _, d := range viewPaths {
			if isExistingDir(path.Join(d, p)) {
				paths = append(paths, path.Join(d, p))
			}
		}
	}
	return paths
}

func (context *Context) findFile(layout string) (string, error) {
	for _, p := range context.getViewPaths() {
		if _, err := os.Stat(path.Join(p, layout)); !os.IsNotExist(err) {
			return path.Join(p, layout), nil
		}
	}
	return "", errors.New("file not found")
}

func (context *Context) FindTemplate(tmpl *template.Template, layout string) (*template.Template, error) {
	for _, p := range context.getViewPaths() {
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

func (context *Context) Render(name string, results ...interface{}) template.HTML {
	var clone = context.clone()
	var err error

	if len(results) > 0 {
		clone.Result = results[0]
	}
	names := strings.Split(name, "/")
	tmpl := template.New(names[len(names)-1] + ".tmpl").Funcs(clone.FuncMap())

	if tmpl, err = context.FindTemplate(tmpl, name+".tmpl"); err == nil {

		var result = bytes.NewBufferString("")
		if err := tmpl.Execute(result, clone); err != nil {
			fmt.Println(err)
		}
		return template.HTML(result.String())
	}

	return ""
}

func (context *Context) Execute(name string, result interface{}) {
	var tmpl *template.Template
	var cacheKey string

	if context.Resource != nil {
		cacheKey = path.Join(context.resourcePath(), name)
	} else {
		cacheKey = name
	}

	if t, ok := templates[cacheKey]; !ok || true {
		var err error
		tmpl = template.New("layout.tmpl").Funcs(context.FuncMap())
		if tmpl, err = context.FindTemplate(tmpl, "layout.tmpl"); err == nil {
			for _, name := range []string{"header", "footer"} {
				if tmpl.Lookup(name) == nil {
					tmpl, _ = context.FindTemplate(tmpl, name+".tmpl")
				}
			}
		}
	} else {
		tmpl = t
	}

	context.Content = context.Render(name, result)
	if err := tmpl.Execute(context.Writer, context); err != nil {
		fmt.Println(err)
	}
}
