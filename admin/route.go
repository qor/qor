package admin

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"net/http"
	"path"
	"regexp"
	"strings"
)

type Params struct {
	Resource *resource.Resource
	Id       string
}

func (admin *Admin) generateApp(w http.ResponseWriter, r *http.Request) *qor.App {
	app := qor.App{ResponseWriter: w, Request: r}
	app.CurrentUser = admin.auth.GetCurrentUser(&app)
	return &app
}

func (admin *Admin) AddToMux(prefix string, mux *http.ServeMux) {
	// format "/admin" to "/admin/"
	// the trail "/" will match under domain, refer function pathMatch in net/http/server.go
	prefix = regexp.MustCompile("//(//)*").ReplaceAllString("/"+prefix+"/", "/")
	mux.HandleFunc(strings.TrimRight(prefix, "/"), func(w http.ResponseWriter, r *http.Request) {
		admin.Dashboard(admin.generateApp(w, r))
	})

	pathMatch := regexp.MustCompile(path.Join(prefix, `(\w+)(?:/(\w+))?[^/]*/?$`))
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		var isIndexURL, isShowURL bool
		var params Params

		matches := pathMatch.FindStringSubmatch(r.URL.Path)
		if resource := admin.resources[matches[1]]; matches[1] != "" && resource != nil {
			isIndexURL = true
			params = Params{Resource: resource}

			if matches[2] != "" { // "/admin/user/1234"
				isIndexURL = false
				isShowURL = true
				params.Id = matches[2]
			}
		}

		switch {
		case r.Method == "GET" && isIndexURL:
			admin.Index(w, r, params)
		case r.Method == "GET" && isShowURL:
			admin.Show(w, r, params)
		case r.Method == "PUT" && isShowURL:
			admin.Update(w, r, params)
		case r.Method == "POST" && isIndexURL:
			admin.Create(w, r, params)
		case r.Method == "DELETE" && isShowURL:
			admin.Delete(w, r, params)
		default:
			http.NotFound(w, r)
		}
	})
}
