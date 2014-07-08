package admin

import (
	"fmt"

	"github.com/qor/qor"
	"reflect"
)

func (admin *Admin) Dashboard(app *qor.Context) {
}

func (admin *Admin) Index(context *qor.Context) {
	sliceType := reflect.SliceOf(reflect.Indirect(reflect.ValueOf(p.Resource.Model)).Type())
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)

	admin.DB.Find(slicePtr.Interface())
	fmt.Println(slicePtr.Interface())
}

func (admin *Admin) Show(context *qor.Context) {
	res := reflect.New(reflect.Indirect(reflect.ValueOf(p.Resource.Model)).Type())
	admin.DB.First(res.Interface(), p.Id)
	fmt.Println(res.Interface())
}

func (admin *Admin) Create(context *qor.Context) {
}

func (admin *Admin) Update(context *qor.Context) {
}

func (admin *Admin) Delete(context *qor.Context) {
}
