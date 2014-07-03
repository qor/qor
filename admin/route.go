package admin

import (
	"github.com/julienschmidt/httprouter"

	"net/http"
	"path"
)

func (admin *Admin) AddToMux(prefix string, mux *http.ServeMux) {
	router := httprouter.New()
	router.HandlerFunc("GET", prefix, admin.Dashboard)
	router.HandlerFunc("GET", path.Join(prefix, ":resource"), admin.Index)
	router.HandlerFunc("POST", path.Join(prefix, ":resource"), admin.Create)
	router.HandlerFunc("PUT", path.Join(prefix, ":resource", ":id"), admin.Update)
	router.HandlerFunc("GET", path.Join(prefix, ":resource", ":id"), admin.Show)
	mux.Handle("/", router)
}
