package admin

import (
	"github.com/qor/qor"

	"net/http"
	"path"
	"regexp"
	"strings"
)

func (admin *Admin) generateContext(w http.ResponseWriter, r *http.Request) *qor.Context {
	context := qor.Context{Writer: w, Request: r}
	context.CurrentUser = admin.auth.GetCurrentUser(&context)
	return &context
}

func (admin *Admin) AddToMux(prefix string, mux *http.ServeMux) {
	// format "/admin" to "/admin/"
	// the trail "/" will match under domain, refer function pathMatch in net/http/server.go
	prefix = regexp.MustCompile("//(//)*").ReplaceAllString("/"+prefix+"/", "/")
	mux.HandleFunc(strings.TrimRight(prefix, "/"), func(w http.ResponseWriter, r *http.Request) {
		admin.Dashboard(admin.generateContext(w, r))
	})

	pathMatch := regexp.MustCompile(path.Join(prefix, `(\w+)(?:/(\w+))?[^/]*/?$`))
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		var isIndexURL, isShowURL bool
		context := admin.generateContext(w, r)

		matches := pathMatch.FindStringSubmatch(r.URL.Path)
		if resource := admin.resources[matches[1]]; matches[1] != "" && resource != nil {
			isIndexURL = true
			context.ResourceName = matches[1]

			if matches[2] != "" { // "/admin/user/1234"
				context.ResourceID = matches[2]
				isIndexURL = false
				isShowURL = true
			}
		}

		switch {
		case r.Method == "GET" && isIndexURL:
			admin.Index(context)
		case r.Method == "GET" && isShowURL:
			admin.Show(context)
		case r.Method == "PUT" && isShowURL:
			admin.Update(context)
		case r.Method == "POST" && isIndexURL:
			admin.Create(context)
		case r.Method == "DELETE" && isShowURL:
			admin.Delete(context)
		default:
			http.NotFound(w, r)
		}
	})
}
