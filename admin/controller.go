package admin

import (
	"fmt"
	"net/http"
	"reflect"
)

func (admin *Admin) Dashboard(w http.ResponseWriter, r *http.Request) {
}

func (admin *Admin) Index(w http.ResponseWriter, r *http.Request, p Params) {
	sliceType := reflect.SliceOf(reflect.Indirect(reflect.ValueOf(p.Resource.Model)).Type())
	slice := reflect.MakeSlice(sliceType, 0, 0)
	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)

	admin.DB.Find(slicePtr.Interface())
	fmt.Println(slicePtr.Interface())
}

func (admin *Admin) Show(w http.ResponseWriter, r *http.Request, p Params) {
	fmt.Println("show")
	fmt.Println(p)
}

func (admin *Admin) Create(w http.ResponseWriter, r *http.Request, p Params) {
}

func (admin *Admin) Update(w http.ResponseWriter, r *http.Request, p Params) {
}

func (admin *Admin) Delete(w http.ResponseWriter, r *http.Request, p Params) {
}
