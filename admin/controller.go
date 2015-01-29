package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/qor/qor/responder"
)

type AdminController struct {
	*Admin
}

func (ac *AdminController) Dashboard(context *Context) {
	context.Execute("dashboard", nil)
}

func (ac *AdminController) Index(context *Context) {
	if result, err := context.FindAll(); err == nil {
		responder.With("html", func() {
			context.Execute("index", result)
		}).With("json", func() {
			js, _ := json.Marshal(context.Resource.ConvertObjectToMap(context, result))
			context.Writer.Write(js)
		}).Respond(context.Writer, context.Request)
	}
}

func (ac *AdminController) Update(context *Context) {
	if result, err := context.FindOne(); err == nil {
		if errs := context.Resource.Decode(context, result); len(errs) == 0 {
			context.Resource.CallSaver(result, context.Context)
			responder.With("html", func() {
				http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
			}).With("json", func() {
				js, _ := json.Marshal(context.Resource.ConvertObjectToMap(context, result))
				context.Writer.Write(js)
			}).Respond(context.Writer, context.Request)
		}
	}
}

func (ac *AdminController) Delete(context *Context) {
	res := ac.GetResource(context.ResourcePath())

	responder.With("html", func() {
		if res.CallDeleter(res.NewStruct(), context.Context) == nil {
			http.Redirect(context.Writer, context.Request, path.Join(ac.GetRouter().Prefix, res.ToParam()), http.StatusFound)
		} else {
			http.Redirect(context.Writer, context.Request, path.Join(ac.GetRouter().Prefix, res.ToParam()), http.StatusNotFound)
		}
	}).With("json", func() {
		if res.CallDeleter(res.NewStruct(), context.Context) == nil {
			context.Writer.WriteHeader(http.StatusOK)
		} else {
			context.Writer.WriteHeader(http.StatusNotFound)
		}
	}).Respond(context.Writer, context.Request)
}

func (ac *AdminController) Action(context *Context) {
	var err error
	name := strings.Split(context.Request.URL.Path, "/")[4]
	if action := context.Resource.actions[name]; action != nil {
		ids := context.Request.Form.Get("ids")
		scope := context.GetDB().Where(fmt.Sprintf("%v IN (?)", context.Resource.PrimaryKey()), ids)
		err = action.Handle(scope, context.Context)
	}

	responder.With("html", func() {
		http.Redirect(context.Writer, context.Request, context.Request.Referer(), http.StatusFound)
	}).With("json", func() {
		if err == nil {
			context.Writer.Write([]byte("OK"))
		} else {
			context.Writer.Write([]byte(err.Error()))
		}
	}).Respond(context.Writer, context.Request)
}

func (ac *AdminController) Asset(context *Context) {
	file := strings.TrimPrefix(context.Request.URL.Path, ac.GetRouter().Prefix)
	if filename, err := context.findFile(file); err == nil {
		http.ServeFile(context.Writer, context.Request, filename)
	} else {
		http.NotFound(context.Writer, context.Request)
	}
}
