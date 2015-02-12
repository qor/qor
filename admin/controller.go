package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/qor/qor/responder"
)

type controller struct {
	*Admin
}

func (ac *controller) Dashboard(context *Context) {
	context.Execute("dashboard", nil)
}

func (ac *controller) Index(context *Context) {
	if result, err := context.FindAll(); err == nil {
		responder.With("html", func() {
			context.Execute("index", result)
		}).With("json", func() {
			res := context.Resource
			js, _ := json.Marshal(res.ConvertObjectToMap(context, result, res.IndexMetas()))
			context.Writer.Write(js)
		}).Respond(context.Writer, context.Request)
	} else {
		http.NotFound(context.Writer, context.Request)
	}
}

func (ac *controller) Show(context *Context) {
	result, _ := context.FindOne()

	responder.With("html", func() {
		context.Execute("show", result)
	}).With("json", func() {
		res := context.Resource
		js, _ := json.Marshal(res.ConvertObjectToMap(context, result, res.ShowMetas()))
		context.Writer.Write(js)
	}).Respond(context.Writer, context.Request)
}

func (ac *controller) New(context *Context) {
	context.Execute("new", nil)
}

func (ac *controller) Create(context *Context) {
	res := ac.GetResource(context.ResourcePath())
	result := res.NewStruct()
	if errs := res.Decode(context, result); len(errs) == 0 {
		res.CallSaver(result, context.Context)
		responder.With("html", func() {
			primaryKey := fmt.Sprintf("%v", context.GetDB().NewScope(result).PrimaryKeyValue())
			http.Redirect(context.Writer, context.Request, path.Join(context.Request.RequestURI, primaryKey), http.StatusFound)
		}).With("json", func() {
			res := context.Resource
			js, _ := json.Marshal(res.ConvertObjectToMap(context, result, res.ShowMetas()))
			context.Writer.Write(js)
		}).Respond(context.Writer, context.Request)
	}
}

func (ac *controller) Update(context *Context) {
	if result, err := context.FindOne(); err == nil {
		if errs := context.Resource.Decode(context, result); len(errs) == 0 {
			context.Resource.CallSaver(result, context.Context)
			responder.With("html", func() {
				http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
			}).With("json", func() {
				res := context.Resource
				js, _ := json.Marshal(res.ConvertObjectToMap(context, result, res.ShowMetas()))
				context.Writer.Write(js)
			}).Respond(context.Writer, context.Request)
		}
	}
}

func (ac *controller) Delete(context *Context) {
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

func (ac *controller) Action(context *Context) {
	var err error
	name := strings.Split(context.Request.URL.Path, "/")[4]

	for _, action := range context.Resource.actions {
		if action.Name == name {
			ids := context.Request.Form.Get("ids")
			scope := context.GetDB().Where(fmt.Sprintf("%v IN (?)", context.Resource.PrimaryField().DBName), ids)
			err = action.Handle(scope, context.Context)
		}
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

func (ac *controller) Asset(context *Context) {
	file := strings.TrimPrefix(context.Request.URL.Path, ac.GetRouter().Prefix)
	if filename, err := context.findFile(file); err == nil {
		http.ServeFile(context.Writer, context.Request, filename)
	} else {
		http.NotFound(context.Writer, context.Request)
	}
}
