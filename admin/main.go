package admin

import (
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
