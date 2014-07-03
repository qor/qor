package admin

import (
	"fmt"
	"net/http"
)

func (admin *Admin) Dashboard(w http.ResponseWriter, r *http.Request) {
}

func (admin *Admin) Index(w http.ResponseWriter, r *http.Request, p Params) {
	fmt.Println("index")
	admin.DB.Debug().First(p.Resource.Model)
	fmt.Println(p.Resource.Name)
	fmt.Println(p.Resource.Model)
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
