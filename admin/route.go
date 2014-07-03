package admin

import (
	"github.com/julienschmidt/httprouter"
	"regexp"

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

	// format "/admin" to "/admin/"
	// the trail "/" will match under domain, refer function pathMatch in net/http/server.go
	prefix = regexp.MustCompile("//(//)*").ReplaceAllString("/"+prefix+"/", "/")
	mux.Handle(prefix, router)
}
