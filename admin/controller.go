package admin

import (
	"encoding/json"
	"fmt"

	"github.com/qor/qor/resource"
	"github.com/qor/qor/responder"

	"net/http"
	"path"
)

func (admin *Admin) Dashboard(context *Context) {
	content := Content{Context: context}
	admin.Render("dashboard", content)
}

func (admin *Admin) Index(context *Context) {
	res := admin.Resources[context.ResourceName]
	if res == nil {
		http.NotFound(context.Writer, context.Request)
		return
	}

	result, _ := admin.NewSearcher(res).FindAll(context)

	responder.With("html", func() {
		content := Content{Context: context, Resource: res, Result: result}
		admin.Render("index", content)
	}).With("json", func() {
		js, _ := json.Marshal(ConvertObjectToMap(context, result, res))
		context.Writer.Write(js)
	}).Respond(context.Writer, context.Request)
}

func (admin *Admin) Show(context *Context) {
	res := admin.Resources[context.ResourceName]
	result, _ := admin.NewSearcher(res).FindOne(context)

	responder.With("html", func() {
		content := Content{Context: context, Resource: res, Result: result}
		admin.Render("show", content)
	}).With("json", func() {
		js, _ := json.Marshal(ConvertObjectToMap(context, result, res))
		context.Writer.Write(js)
	}).Respond(context.Writer, context.Request)
}

func (admin *Admin) New(context *Context) {
	resource := admin.Resources[context.ResourceName]
	content := Content{Context: context, Resource: resource}
	admin.Render("new", content)
}

func (admin *Admin) decode(result interface{}, res *Resource, context *Context) (errs []error) {
	responder.With("html", func() {
		errs = resource.DecodeToResource(res, result, ConvertFormToMetaValues(context, "QorResource.", res), context.Context).Start()
	}).With("json", func() {
		decoder := json.NewDecoder(context.Request.Body)
		values := map[string]interface{}{}
		if err := decoder.Decode(&values); err == nil {
			errs = resource.DecodeToResource(res, result, ConvertMapToMetaValues(values, res), context.Context).Start()
		} else {
			errs = append(errs, err)
		}
	}).Respond(context.Writer, context.Request)
	return errs
}

func (admin *Admin) Create(context *Context) {
	res := admin.Resources[context.ResourceName]
	result := res.NewStruct()
	if errs := admin.decode(result, res, context); len(errs) == 0 {
		res.CallSaver(result, context.Context)
		responder.With("html", func() {
			primaryKey := fmt.Sprintf("%v", context.GetDB().NewScope(result).PrimaryKeyValue())
			http.Redirect(context.Writer, context.Request, path.Join(context.Request.RequestURI, primaryKey), http.StatusFound)
		}).With("json", func() {
			js, _ := json.Marshal(ConvertObjectToMap(context, result, res))
			context.Writer.Write(js)
		}).Respond(context.Writer, context.Request)
	}
}

func (admin *Admin) Update(context *Context) {
	res := admin.Resources[context.ResourceName]
	if result, err := admin.NewSearcher(res).FindOne(context); err == nil {
		if errs := admin.decode(result, res, context); len(errs) == 0 {
			res.CallSaver(result, context.Context)
			responder.With("html", func() {
				http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
			}).With("json", func() {
				js, _ := json.Marshal(ConvertObjectToMap(context, result, res))
				context.Writer.Write(js)
			}).Respond(context.Writer, context.Request)
		}
	}
}

func (admin *Admin) Delete(context *Context) {
	res := admin.Resources[context.ResourceName]

	if res.CallDeleter(res.NewStruct(), context.Context) == nil {
		http.Redirect(context.Writer, context.Request, path.Join(admin.router.Prefix, res.Name), http.StatusFound)
	} else {
		http.Redirect(context.Writer, context.Request, path.Join(admin.router.Prefix, res.Name), http.StatusNotFound)
	}
}
