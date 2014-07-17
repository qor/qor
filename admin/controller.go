package admin

import (
	"fmt"
	"github.com/qor/qor"
	"github.com/qor/qor/rules"

	"reflect"
)

func (admin *Admin) Dashboard(context *qor.Context) {
	content := Content{Admin: admin, Context: context, Action: "dashboard"}
	admin.Render("dashboard", content, rules.Read)
}

func (admin *Admin) Index(context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	sliceType := reflect.SliceOf(reflect.Indirect(reflect.ValueOf(resource.Model)).Type())
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	admin.DB.Find(slicePtr.Interface())

	content := Content{Admin: admin, Context: context, Resource: resource, Result: slicePtr.Interface(), Action: "index"}
	admin.Render("index", content, rules.Read)
}

func (admin *Admin) Show(context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	res := reflect.New(reflect.Indirect(reflect.ValueOf(resource.Model)).Type())
	admin.DB.First(res.Interface(), context.ResourceID)

	content := Content{Admin: admin, Context: context, Resource: resource, Result: res.Interface(), Action: "show"}
	admin.Render("show", content, rules.Read, rules.Update)
}

func (admin *Admin) New(context *qor.Context) {
	resource := admin.Resources[context.ResourceName]
	content := Content{Admin: admin, Context: context, Resource: resource, Action: "new"}
	admin.Render("new", content, rules.Create)
}

func (admin *Admin) Create(context *qor.Context) {
}

func (admin *Admin) Update(context *qor.Context) {
	context.Request.ParseMultipartForm(32 << 22)
	for key, value := range context.Request.Form {
		fmt.Println(key)
		fmt.Println(value)
		fmt.Println(reflect.TypeOf(value))
	}
}

func (admin *Admin) Delete(context *qor.Context) {
}
