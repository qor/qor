package admin

import (
	"encoding/json"
	"fmt"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/responder"
	"github.com/qor/qor/roles"

	"net/http"
	"path"
)

func (admin *Admin) Dashboard(writer http.ResponseWriter, request *http.Request, context *qor.Context) {
	content := Content{Admin: admin, Context: context, Action: "dashboard", Writer: writer}
	admin.Render("dashboard", content, roles.Read)
}

func (admin *Admin) Index(writer http.ResponseWriter, request *http.Request, context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := res.NewSlice()
	res.CallSearcher(result, context)

	responder.With("html", func() {
		content := Content{Admin: admin, Context: context, Resource: res, Result: result, Action: "index", Writer: writer}
		admin.Render("index", content, roles.Read)
	}).With("json", func() {
		js, _ := json.Marshal(ConvertObjectToMap(context, result, res))
		writer.Write(js)
	}).Respond(writer, request)
}

func (admin *Admin) Show(writer http.ResponseWriter, request *http.Request, context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := res.NewStruct()
	res.CallFinder(result, nil, context)

	responder.With("html", func() {
		content := Content{Admin: admin, Context: context, Resource: res, Result: result, Action: "edit", Writer: writer}
		admin.Render("show", content, roles.Read, roles.Update)
	}).With("json", func() {
		js, _ := json.Marshal(ConvertObjectToMap(context, result, res))
		writer.Write(js)
	}).Respond(writer, request)
}

func (admin *Admin) New(writer http.ResponseWriter, request *http.Request, context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	content := Content{Admin: admin, Context: context, Resource: resource, Action: "new", Writer: writer}
	admin.Render("new", content, roles.Create)
}

func (admin *Admin) decode(result interface{}, res *Resource, context *qor.Context, writer http.ResponseWriter, request *http.Request) (errs []error) {
	responder.With("html", func() {
		errs = resource.DecodeToResource(res, result, ConvertFormToMetaValues(context, "QorResource.", res, request), context).Start()
	}).With("json", func() {
		decoder := json.NewDecoder(request.Body)
		values := map[string]interface{}{}
		if err := decoder.Decode(&values); err == nil {
			errs = resource.DecodeToResource(res, result, ConvertMapToMetaValues(values, res), context).Start()
		} else {
			errs = append(errs, err)
		}
	}).Respond(writer, request)
	return errs
}

func (admin *Admin) Create(writer http.ResponseWriter, request *http.Request, context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := res.NewStruct()
	if errs := admin.decode(result, res, context, writer, request); len(errs) == 0 {
		res.CallSaver(result, context)
		primaryKey := fmt.Sprintf("%v", context.DB().NewScope(result).PrimaryKeyValue())
		http.Redirect(writer, request, path.Join(request.RequestURI, primaryKey), http.StatusFound)
	}
}

func (admin *Admin) Update(writer http.ResponseWriter, request *http.Request, context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := res.NewStruct()
	if res.CallFinder(result, nil, context) == nil {
		if errs := admin.decode(result, res, context, writer, request); len(errs) == 0 {
			res.CallSaver(result, context)
			http.Redirect(writer, request, request.RequestURI, http.StatusFound)
		}
	}
}

func (admin *Admin) Delete(writer http.ResponseWriter, request *http.Request, context *qor.Context) {
	res := admin.Resources[context.ResourceName]

	if res.CallDeleter(res.NewStruct(), context) == nil {
		http.Redirect(writer, request, path.Join(admin.Prefix, res.Name), http.StatusFound)
	} else {
		http.Redirect(writer, request, path.Join(admin.Prefix, res.Name), http.StatusNotFound)
	}
}
