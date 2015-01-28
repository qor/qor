package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/qor/qor/responder"
)

func (admin *Admin) Dashboard(context *Context) {
	context.Execute("dashboard", nil)
}

func (admin *Admin) Index(context *Context) {
	if result, err := context.FindAll(); err == nil {
		responder.With("html", func() {
			context.Execute("index", result)
		}).With("json", func() {
			js, _ := json.Marshal(context.Resource.ConvertObjectToMap(context, result))
			context.Writer.Write(js)
		}).Respond(context.Writer, context.Request)
	} else {
		http.NotFound(context.Writer, context.Request)
	}
}

func (admin *Admin) Show(context *Context) {
	result, _ := context.FindOne()

	responder.With("html", func() {
		context.Execute("show", result)
	}).With("json", func() {
		js, _ := json.Marshal(context.Resource.ConvertObjectToMap(context, result))
		context.Writer.Write(js)
	}).Respond(context.Writer, context.Request)
}

func (admin *Admin) New(context *Context) {
	context.Execute("new", nil)
}

func (admin *Admin) Create(context *Context) {
	res := admin.GetResource(context.ResourcePath())
	result := res.NewStruct()
	if errs := res.Decode(context, result); len(errs) == 0 {
		res.CallSaver(result, context.Context)
		responder.With("html", func() {
			primaryKey := fmt.Sprintf("%v", context.GetDB().NewScope(result).PrimaryKeyValue())
			http.Redirect(context.Writer, context.Request, path.Join(context.Request.RequestURI, primaryKey), http.StatusFound)
		}).With("json", func() {
			js, _ := json.Marshal(res.ConvertObjectToMap(context, result))
			context.Writer.Write(js)
		}).Respond(context.Writer, context.Request)
	}
}

func (admin *Admin) Update(context *Context) {
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

func (admin *Admin) Delete(context *Context) {
	res := admin.GetResource(context.ResourcePath())

	responder.With("html", func() {
		if res.CallDeleter(res.NewStruct(), context.Context) == nil {
			http.Redirect(context.Writer, context.Request, path.Join(admin.router.Prefix, res.ToParam()), http.StatusFound)
		} else {
			http.Redirect(context.Writer, context.Request, path.Join(admin.router.Prefix, res.ToParam()), http.StatusNotFound)
		}
	}).With("json", func() {
		if res.CallDeleter(res.NewStruct(), context.Context) == nil {
			context.Writer.WriteHeader(http.StatusOK)
		} else {
			context.Writer.WriteHeader(http.StatusNotFound)
		}
	}).Respond(context.Writer, context.Request)
}

func (admin *Admin) Action(context *Context) {
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

func (admin *Admin) Asset(context *Context) {
	file := strings.TrimPrefix(context.Request.URL.Path, context.Admin.GetRouter().Prefix)
	if filename, err := context.findFile(file); err == nil {
		http.ServeFile(context.Writer, context.Request, filename)
	} else {
		http.NotFound(context.Writer, context.Request)
	}
}
