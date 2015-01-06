package admin

import (
	"encoding/json"
	"fmt"

	"github.com/qor/qor/responder"

	"net/http"
	"path"
)

func (admin *Admin) Dashboard(context *Context) {
	context.Execute("dashboard", nil)
}

func (admin *Admin) Index(context *Context) {
	if result, err := context.FindAll(); err == nil {
		responder.With("html", func() {
			context.Execute("index", result)
		}).With("json", func() {
			js, _ := json.Marshal(context.Resource.ConvertObjectToMap(context.Context, result))
			context.Writer.Write(js)
		}).Respond(context.Writer, context.Request)
	} else {
		fmt.Println(err)
		http.NotFound(context.Writer, context.Request)
	}
}

func (admin *Admin) Show(context *Context) {
	result, _ := context.FindOne()

	responder.With("html", func() {
		context.Execute("show", result)
	}).With("json", func() {
		js, _ := json.Marshal(context.Resource.ConvertObjectToMap(context.Context, result))
		context.Writer.Write(js)
	}).Respond(context.Writer, context.Request)
}

func (admin *Admin) New(context *Context) {
	context.Execute("new", nil)
}

func (admin *Admin) Create(context *Context) {
	res := admin.GetResource(context.ResourceName())
	result := res.NewStruct()
	if errs := res.Decode(context.Context, result); len(errs) == 0 {
		res.CallSaver(result, context.Context)
		responder.With("html", func() {
			primaryKey := fmt.Sprintf("%v", context.GetDB().NewScope(result).PrimaryKeyValue())
			http.Redirect(context.Writer, context.Request, path.Join(context.Request.RequestURI, primaryKey), http.StatusFound)
		}).With("json", func() {
			js, _ := json.Marshal(res.ConvertObjectToMap(context.Context, result))
			context.Writer.Write(js)
		}).Respond(context.Writer, context.Request)
	}
}

func (admin *Admin) Update(context *Context) {
	if result, err := context.FindOne(); err == nil {
		if errs := context.Resource.Decode(context.Context, result); len(errs) == 0 {
			context.Resource.CallSaver(result, context.Context)
			responder.With("html", func() {
				http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
			}).With("json", func() {
				js, _ := json.Marshal(context.Resource.ConvertObjectToMap(context.Context, result))
				context.Writer.Write(js)
			}).Respond(context.Writer, context.Request)
		}
	}
}

func (admin *Admin) Delete(context *Context) {
	res := admin.GetResource(context.ResourceName())

	if res.CallDeleter(res.NewStruct(), context.Context) == nil {
		http.Redirect(context.Writer, context.Request, path.Join(admin.router.Prefix, res.Name), http.StatusFound)
	} else {
		http.Redirect(context.Writer, context.Request, path.Join(admin.router.Prefix, res.Name), http.StatusNotFound)
	}
}
