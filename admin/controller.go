package admin

import (
	"github.com/qor/qor"

	"reflect"
)

func (admin *Admin) Dashboard(app *qor.Context) {
	admin.Render("dashboard", app)
}

func (admin *Admin) Index(context *qor.Context) {
	resource := admin.resources[context.ResourceName]
	sliceType := reflect.SliceOf(reflect.Indirect(reflect.ValueOf(resource.Model)).Type())
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)
	admin.DB.Find(slicePtr.Interface())

	admin.Render("resources/index", context)
}

func (admin *Admin) Show(context *qor.Context) {
	resource := admin.resources[context.ResourceName]
	res := reflect.New(reflect.Indirect(reflect.ValueOf(resource.Model)).Type())
	admin.DB.First(res.Interface(), context.ResourceID)

	admin.Render("resources/show", context)
}

func (admin *Admin) New(context *qor.Context) {
	admin.Render("resources/new", context)
}

func (admin *Admin) Create(context *qor.Context) {
}

func (admin *Admin) Update(context *qor.Context) {
}

func (admin *Admin) Delete(context *qor.Context) {
}
