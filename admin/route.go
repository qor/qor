package admin

import (
	"net/http"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/roles"
)

type Handle func(c *Context)

type Router struct {
	Prefix  string
	gets    map[string]Handle
	posts   map[string]Handle
	deletes map[string]Handle
	puts    map[string]Handle
}

func newRouter() *Router {
	return &Router{
		gets:    map[string]Handle{},
		posts:   map[string]Handle{},
		deletes: map[string]Handle{},
		puts:    map[string]Handle{},
	}
}

// Possible path types
// /admin/orders
// /admin/orders/new
// /admin/orders/123
func (r *Router) parsePath(path string) (res, id string) {
	parts := strings.Split(strings.TrimLeft(path, r.Prefix), "/")
	// fmt.Printf("--> %+v\n", parts)
	for _, part := range parts {
		if part != "" && res == "" {
			res = part
			continue
		}

		if part != "" && id == "" {
			id = part
			break
		}
	}

	return
}

func (r *Router) Get(path string, handle Handle) {
	r.gets[path] = handle
}

func (r *Router) Post(path string, handle Handle) {
	r.posts[path] = handle
}

func (r *Router) Put(path string, handle Handle) {
	r.puts[path] = handle
}

func (r *Router) Delete(path string, handle Handle) {
	r.deletes[path] = handle
}

func (admin *Admin) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	var currentUser *qor.CurrentUser
	context := Context{Context: &qor.Context{Config: admin.Config, Request: r}, Writer: w}
	if admin.auth != nil {
		currentUser = admin.auth.GetCurrentUser(&context)
	}
	context.Roles = roles.MatchedRoles(r, currentUser)
	return &context
}

// TODO: to extend this api
func (admin *Admin) MountTo(prefix string, mux *http.ServeMux) {
	prefix = "/" + strings.Trim(prefix, "/")
	router := admin.router
	router.Prefix = prefix + "/"

	mux.Handle(prefix, admin)        // /:prefix
	mux.Handle(router.Prefix, admin) // /:prefix/:xxx
}

func (admin *Admin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 128 MB
	req.ParseMultipartForm(32 << 22)
	if len(req.Form["_method"]) > 0 {
		req.Method = strings.ToUpper(req.Form["_method"][0])
	}

	var (
		router          = admin.router
		context         = admin.NewContext(w, req)
		builtin, custom Handle
	)

	context.ResourceName, context.ResourceID = router.parsePath(req.URL.Path)
	// fmt.Printf("--> %+v\n", req.URL.Path)
	// fmt.Printf("--> %+v\n", context.ResourceName, context.ResourceID)
	if context.ResourceName == "" && context.ResourceID == "" {
		builtin = admin.Dashboard
	} else if context.ResourceID == "new" {
		// /admin/:ressource/new
		switch req.Method {
		case "GET":
			custom = router.gets["/"+context.ResourceName+"/new"]
			builtin = admin.New
		}
	} else if context.ResourceName != "" && context.ResourceID == "" {
		// /admin/:ressource
		switch req.Method {
		case "GET":
			custom = router.gets["/"+context.ResourceName]
			builtin = admin.Index
		case "POST":
			custom = router.posts["/"+context.ResourceName]
			builtin = admin.Create
		case "PUT":
			custom = router.puts["/"+context.ResourceName+"/new"]
			builtin = admin.Create
		}
	} else if context.ResourceName != "" && context.ResourceID != "" {
		// /admin/:ressource/:id
		switch req.Method {
		case "GET":
			custom = router.gets["/"+context.ResourceName+"/:id"]
			builtin = admin.Show
		case "POST":
			custom = router.posts["/"+context.ResourceName+"/:id"]
			builtin = admin.Update
		case "PUT":
			custom = router.puts["/"+context.ResourceName+"/:id"]
			builtin = admin.Update
		case "DELETE":
			custom = router.deletes["/"+context.ResourceName+"/:id"]
			builtin = admin.Delete
		}
	}

	if custom != nil {
		custom(context)
	} else if builtin != nil {
		builtin(context)
	} else {
		http.NotFound(w, req)
	}
}
