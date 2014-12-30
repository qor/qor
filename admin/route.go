package admin

import (
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
)

type Handle func(c *Context)
type Handler struct {
	Path   *regexp.Regexp
	Handle Handle
}

type Router struct {
	Prefix  string
	routers map[string][]Handler
}

func newRouter() *Router {
	return &Router{routers: map[string][]Handler{
		"GET":    []Handler{},
		"PUT":    []Handler{},
		"POST":   []Handler{},
		"DELETE": []Handler{},
	}}
}

func (r *Router) Get(path string, handle Handle) {
	r.routers["GET"] = append(r.routers["GET"], Handler{Path: regexp.MustCompile(path), Handle: handle})
}

func (r *Router) Post(path string, handle Handle) {
	r.routers["POST"] = append(r.routers["POST"], Handler{Path: regexp.MustCompile(path), Handle: handle})
}

func (r *Router) Put(path string, handle Handle) {
	r.routers["PUT"] = append(r.routers["PUT"], Handler{Path: regexp.MustCompile(path), Handle: handle})
}

func (r *Router) Delete(path string, handle Handle) {
	r.routers["DELETE"] = append(r.routers["DELETE"], Handler{Path: regexp.MustCompile(path), Handle: handle})
}

func (admin *Admin) MountTo(prefix string, mux *http.ServeMux) {
	prefix = "/" + strings.Trim(prefix, "/")
	router := admin.router
	router.Prefix = prefix

	router.Get("^/$", admin.Dashboard)
	router.Get("^$", admin.Dashboard)
	router.Get("^/[^/]+/new$", admin.New)
	router.Put("^/[^/]+/new$", admin.Create)
	router.Post("^/[^/]+/new$", admin.Create)
	router.Post("^/[^/]+$", admin.Create)
	router.Get("^/[^/]+/.*$", admin.Show)
	router.Put("^/[^/]+/.*$", admin.Update)
	router.Post("^/[^/]+/.*$", admin.Update)
	router.Delete("^/[^/]+/.*$", admin.Delete)
	router.Get("^/[^/]+$", admin.Index)

	mux.Handle(prefix, admin)     // /:prefix
	mux.Handle(prefix+"/", admin) // /:prefix/:xxx
}

func (admin *Admin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() func() {
		begin := time.Now()
		log.Printf("Start [%s] %s\n", req.Method, req.RequestURI)
		return func() {
			log.Printf("Finish [%s] %s Took %.2fms\n", req.Method, req.RequestURI, time.Now().Sub(begin).Seconds()*1000)
		}
	}()()

	// 128 MB
	req.ParseMultipartForm(32 << 22)
	if len(req.Form["_method"]) > 0 {
		req.Method = strings.ToUpper(req.Form["_method"][0])
	}

	var router = admin.router
	var context = admin.NewContext(w, req)

	var pathMatch = regexp.MustCompile(path.Join(router.Prefix, `(\w+)(?:/(\w+))?[^/]*/?$`))
	var matches = pathMatch.FindStringSubmatch(req.URL.Path)
	if len(matches) > 1 {
		context.SetResource(admin.GetResource(matches[1]))
		if len(matches) > 2 {
			context.ResourceID = matches[2]
		}
	}

	routers := router.routers[strings.ToUpper(req.Method)]
	relativePath := strings.TrimPrefix(req.URL.Path, router.Prefix)
	for _, handler := range routers {
		if handler.Path.MatchString(relativePath) {
			handler.Handle(context)
			return
		}
	}
	http.NotFound(w, req)
}
