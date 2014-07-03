package admin

import (
	"github.com/qor/qor/resource"

	"net/http"
)

type Admin struct {
}

func New() *Admin {
	admin := Admin{}
	return &admin
}

func (admin *Admin) AddToMux(prefix string, mux *http.ServeMux) {

}

func (admin *Admin) AddResource(resource *resource.Resource) {

}
