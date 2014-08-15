package admin

import (
	"encoding/json"
	"fmt"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/responder"
	"github.com/qor/qor/rules"

	"net/http"
	"path"
)

func (admin *Admin) Dashboard(context *qor.Context) {
	content := Content{Admin: admin, Context: context, Action: "dashboard"}
	admin.Render("dashboard", content, rules.Read)
}

func (admin *Admin) Index(context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := res.NewSlice()
	admin.DB.Find(result)

	responder.With("html", func() {
		content := Content{Admin: admin, Context: context, Resource: res, Result: result, Action: "index"}
		admin.Render("index", content, rules.Read)
	}).With("json", func() {
		js, _ := json.Marshal(ConvertObjectToMap(context, result, res))
		context.Writer.Write(js)
	}).Respond(context.Writer, context.Request)
}

func (admin *Admin) Show(context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := res.NewStruct()
	admin.DB.First(result, context.ResourceID)

	responder.With("html", func() {
		content := Content{Admin: admin, Context: context, Resource: res, Result: result, Action: "edit"}
		admin.Render("show", content, rules.Read, rules.Update)
	}).With("json", func() {
		js, _ := json.Marshal(ConvertObjectToMap(context, result, res))
		context.Writer.Write(js)
	}).Respond(context.Writer, context.Request)
}

func (admin *Admin) New(context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	content := Content{Admin: admin, Context: context, Resource: resource, Action: "new"}
	admin.Render("new", content, rules.Create)
}

func (admin *Admin) decode(result interface{}, res *Resource, context *qor.Context) (errs []error) {
	responder.With("html", func() {
		errs = resource.DecodeToResource(res, result, ConvertFormToMetaValues(context, "QorResource.", res), context).Start()
	}).With("json", func() {
		decoder := json.NewDecoder(context.Request.Body)
		values := map[string]interface{}{}
		if err := decoder.Decode(&values); err == nil {
			errs = resource.DecodeToResource(res, result, ConvertMapToMetaValues(context, values, res), context).Start()
		} else {
			errs = append(errs, err)
		}
	}).Respond(context.Writer, context.Request)
	return errs
}

func (admin *Admin) Create(context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := res.NewStruct()
	if errs := admin.decode(result, res, context); len(errs) == 0 {
		admin.DB.Save(result)
		primaryKey := fmt.Sprintf("%v", admin.DB.NewScope(result).PrimaryKeyValue())
		http.Redirect(context.Writer, context.Request, path.Join(context.Request.RequestURI, primaryKey), http.StatusFound)
	}
}

func (admin *Admin) Update(context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := res.NewStruct()
	if !admin.DB.First(result, context.ResourceID).RecordNotFound() {
		if errs := admin.decode(result, res, context); len(errs) == 0 {
			admin.DB.Save(result)
			http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
		}
	}
}

func (admin *Admin) Delete(context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := res.NewStruct()

	if admin.DB.Delete(result, context.ResourceID).RowsAffected > 0 {
		http.Redirect(context.Writer, context.Request, path.Join(admin.Prefix, res.Name), http.StatusFound)
	} else {
		http.Redirect(context.Writer, context.Request, path.Join(admin.Prefix, res.Name), http.StatusNotFound)
	}
}
