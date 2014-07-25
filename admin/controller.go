package admin

import (
	"fmt"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/rules"
	"net/http"
	"path"

	"reflect"
)

func (admin *Admin) Dashboard(context *qor.Context) {
	content := Content{Admin: admin, Context: context, Action: "dashboard"}
	admin.Render("dashboard", content, rules.Read)
}

func (admin *Admin) Index(context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	sliceType := reflect.SliceOf(reflect.ValueOf(resource.Model).Type())
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	result := slicePtr.Interface()
	admin.DB.Find(result)

	content := Content{Admin: admin, Context: context, Resource: resource, Result: result, Action: "index"}
	admin.Render("index", content, rules.Read)
}

func (admin *Admin) Show(context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	result := reflect.New(reflect.Indirect(reflect.ValueOf(resource.Model)).Type()).Interface()
	admin.DB.First(result, context.ResourceID)

	content := Content{Admin: admin, Context: context, Resource: resource, Result: result, Action: "edit"}
	admin.Render("show", content, rules.Read, rules.Update)
}

func (admin *Admin) New(context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	content := Content{Admin: admin, Context: context, Resource: resource, Action: "new"}
	admin.Render("new", content, rules.Create)
}

func (admin *Admin) Create(context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := reflect.New(reflect.Indirect(reflect.ValueOf(res.Model)).Type()).Interface()
	metas := res.AllowedMetas(res.EditAttrs(), context, rules.Update)
	resource.Decode(result, metas, context, "QorResource.")
	admin.DB.Save(result)
	primaryKey := fmt.Sprintf("%v", admin.DB.NewScope(result).PrimaryKeyValue())
	http.Redirect(context.Writer, context.Request, path.Join(context.Request.RequestURI, primaryKey), http.StatusFound)
}

func (admin *Admin) Update(context *qor.Context) {
	res := admin.Resources[context.ResourceName]
	result := reflect.New(reflect.Indirect(reflect.ValueOf(res.Model)).Type()).Interface()

	if !admin.DB.First(result, context.ResourceID).RecordNotFound() {
		metas := res.AllowedMetas(res.EditAttrs(), context, rules.Update)
		resource.Decode(result, metas, context, "QorResource.")
		admin.DB.Save(result)
		http.Redirect(context.Writer, context.Request, context.Request.RequestURI, http.StatusFound)
	}
}

func (admin *Admin) Delete(context *qor.Context) {
}
