package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/qor/qor"
	"github.com/qor/qor/utils"
)

type Context struct {
	*qor.Context
	*Searcher
	Flashes  []Flash
	Resource *Resource
	Admin    *Admin
	Content  template.HTML
	Action   string
	Result   interface{}
}

func (admin *Admin) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	context := Context{Context: &qor.Context{Config: admin.Config, Request: r, Writer: w}, Admin: admin}

	return &context
}

func (context *Context) clone() *Context {
	return &Context{
		Context:  context.Context,
		Searcher: context.Searcher,
		Flashes:  context.Flashes,
		Resource: context.Resource,
		Admin:    context.Admin,
		Result:   context.Result,
		Content:  context.Content,
		Action:   context.Action,
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
	if res != nil {
		context.Resource = res
		context.ResourceID = res.GetPrimaryValue(context.Request)
	}
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
			if context.Action != "" {
				if isExistingDir(path.Join(d, p, context.Action)) {
					paths = append(paths, path.Join(d, p, context.Action))
				}
			}

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

func (context *Context) FindTemplate(layouts ...string) (string, error) {
	for _, layout := range layouts {
		for _, p := range context.getViewPaths() {
			if _, err := os.Stat(filepath.Join(p, layout)); !os.IsNotExist(err) {
				return filepath.Join(p, layout), nil
			}
		}
	}
	return "", errors.New("template not found")
}

func (context *Context) Render(name string, results ...interface{}) template.HTML {
	if file, err := context.FindTemplate(name + ".tmpl"); err == nil {
		var clone = context.clone()
		var result = bytes.NewBufferString("")

		if len(results) > 0 {
			clone.Result = results[0]
		}

		if tmpl, err := template.New(filepath.Base(file)).Funcs(clone.FuncMap()).ParseFiles(file); err == nil {
			if err := tmpl.Execute(result, clone); err != nil {
				utils.ExitWithMsg(err)
			}
		} else {
			utils.ExitWithMsg(err)
		}
		return template.HTML(result.String())
	}

	return ""
}

func (context *Context) Execute(name string, result interface{}) {
	var tmpl *template.Template
	var cacheKey string

	if name == "show" && !context.Resource.isSetShowAttrs {
		name = "edit"
	}

	if context.Action == "" {
		context.Action = name
	}

	if context.Resource != nil {
		cacheKey = path.Join(context.resourcePath(), name)
	} else {
		cacheKey = name
	}

	if t, ok := templates[cacheKey]; !ok || true {
		if file, err := context.FindTemplate("layout.tmpl"); err == nil {
			if tmpl, err = template.New(filepath.Base(file)).Funcs(context.FuncMap()).ParseFiles(file); err == nil {
				for _, name := range []string{"header", "footer"} {
					if tmpl.Lookup(name) == nil {
						if file, err := context.FindTemplate(name + ".tmpl"); err == nil {
							tmpl.ParseFiles(file)
						}
					} else {
						utils.ExitWithMsg(err)
					}
				}
			} else {
				utils.ExitWithMsg(err)
			}
		}
	} else {
		tmpl = t
	}

	context.Result = result
	context.Content = context.Render(name, result)
	if err := tmpl.Execute(context.Writer, context); err != nil {
		utils.ExitWithMsg(err)
	}
}

func (context *Context) JSON(name string, result interface{}) {
	if name == "show" && !context.Resource.isSetShowAttrs {
		name = "edit"
	}

	js, _ := json.MarshalIndent(context.Resource.convertObjectToJSONMap(context, result, name), "", "\t")
	context.Writer.Header().Set("Content-Type", "application/json")
	context.Writer.Write(js)
}
