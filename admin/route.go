package admin

import (
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/qor/qor"
	"github.com/qor/qor/roles"
)

type requestHandler func(c *Context)

type routeHandler struct {
	Path   string
	Handle requestHandler
}

// Middleware is a way to filter a request and response coming into your application
// Register new middlewares with `admin.GetRouter().Use(func(*Context, *Middleware))`
// It will be called in order, it need to be registered before `admin.MountTo`
type Middleware struct {
	Handler func(*Context, *Middleware)
	next    *Middleware
}

// Next will call the next middleware
func (middleware Middleware) Next(context *Context) {
	if next := middleware.next; next != nil {
		next.Handler(context, next)
	}
}

// Router contains registered routers
type Router struct {
	Prefix      string
	routers     map[string][]routeHandler
	middlewares []*Middleware
}

func newRouter() *Router {
	return &Router{routers: map[string][]routeHandler{
		"GET":    []routeHandler{},
		"PUT":    []routeHandler{},
		"POST":   []routeHandler{},
		"DELETE": []routeHandler{},
	}}
}

// Use reigster a middleware to the router
func (r *Router) Use(handler func(*Context, *Middleware)) {
	r.middlewares = append(r.middlewares, &Middleware{Handler: handler})
}

// Get register a GET request handle with the given path
func (r *Router) Get(path string, handle requestHandler) {
	r.routers["GET"] = append(r.routers["GET"], routeHandler{Path: path, Handle: handle})
}

// Post register a POST request handle with the given path
func (r *Router) Post(path string, handle requestHandler) {
	r.routers["POST"] = append(r.routers["POST"], routeHandler{Path: path, Handle: handle})
}

// Put register a PUT request handle with the given path
func (r *Router) Put(path string, handle requestHandler) {
	r.routers["PUT"] = append(r.routers["PUT"], routeHandler{Path: path, Handle: handle})
}

// Delete register a DELETE request handle with the given path
func (r *Router) Delete(path string, handle requestHandler) {
	r.routers["DELETE"] = append(r.routers["DELETE"], routeHandler{Path: path, Handle: handle})
}

// MountTo mount the service into mux (HTTP request multiplexer) with given path
func (admin *Admin) MountTo(mountTo string, mux *http.ServeMux) {
	prefix := "/" + strings.Trim(mountTo, "/")
	router := admin.router
	router.Prefix = prefix

	admin.compile()

	controller := &controller{admin}
	router.Get("", controller.Dashboard)
	router.Get("/!search", controller.SearchCenter)

	var registerResourceToRouter func(*Resource, ...string)
	registerResourceToRouter = func(res *Resource, modes ...string) {
		var prefix = func(r *Resource) string {
			p := r.ToParam()
			for r.base != nil {
				p = path.Join(r.base.ToParam(), ":id", p)
				r = r.base
			}
			return "/" + strings.Trim(p, "/")
		}(res)

		for _, mode := range modes {
			if mode == "create" {
				// New
				router.Get(path.Join(prefix, "new"), controller.New)

				// Create
				router.Post(prefix, controller.Create)
			}

			if mode == "read" {
				// Index
				router.Get(prefix, controller.Index)

				// Show
				router.Get(path.Join(prefix, ":id"), controller.Show)
			}

			if mode == "update" {
				// Edit
				router.Get(path.Join(prefix, ":id", "edit"), controller.Edit)

				// Update
				router.Put(path.Join(prefix, ":id"), controller.Update)
				router.Post(path.Join(prefix, ":id"), controller.Update)

				// Action
				for _, action := range res.actions {
					router.Post(path.Join(prefix, ":id", action.Name), controller.Action)
				}
			}

			if mode == "delete" {
				// Delete
				router.Delete(path.Join(prefix, ":id"), controller.Delete)
			}
		}

		// Sub Resources
		for _, meta := range res.ConvertSectionToMetas(res.NewAttrs()) {
			if meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil {
				registerResourceToRouter(meta.Resource, "create")
			}
		}

		for _, meta := range res.ConvertSectionToMetas(res.ShowAttrs()) {
			if meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil {
				registerResourceToRouter(meta.Resource, "read")
			}
		}

		for _, meta := range res.ConvertSectionToMetas(res.EditAttrs()) {
			if meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil {
				registerResourceToRouter(meta.Resource, "update", "delete")
			}
		}
	}

	for _, res := range admin.resources {
		if !res.Config.Invisible {
			registerResourceToRouter(res, "create", "read", "update", "delete")
		}
	}

	mux.Handle(prefix, admin)     // /:prefix
	mux.Handle(prefix+"/", admin) // /:prefix/:xxx
}

func (admin *Admin) compile() {
	admin.generateMenuLinks()

	for _, res := range admin.resources {
		res.configure()
	}

	router := admin.GetRouter()
	router.Use(func(context *Context, middleware *Middleware) {
		writer := context.Writer
		request := context.Request

		// 128 MB
		request.ParseMultipartForm(32 << 22)
		if len(request.Form["_method"]) > 0 {
			request.Method = strings.ToUpper(request.Form["_method"][0])
		}

		var pathMatch = regexp.MustCompile(path.Join(router.Prefix, `(\w+)(?:/(\w+))?[^/]*`))
		var matches = pathMatch.FindStringSubmatch(request.URL.Path)
		if len(matches) > 1 {
			context.setResource(admin.GetResource(matches[1]))
			if len(matches) > 2 {
				context.ResourceID = matches[2]
			}
		}

		handlers := router.routers[strings.ToUpper(request.Method)]
		relativePath := strings.TrimPrefix(request.URL.Path, router.Prefix)
		for _, handler := range handlers {
			if handler.Path == relativePath {
				handler.Handle(context)
				return
			}
		}

		http.NotFound(writer, request)
	})

	for index, middleware := range router.middlewares {
		if len(router.middlewares) > index+1 {
			middleware.next = router.middlewares[index+1]
		}
	}
}

// ServeHTTP dispatches the handler registered in the matched route
func (admin *Admin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var relativePath = strings.TrimPrefix(req.URL.Path, admin.router.Prefix)
	var context = admin.NewContext(w, req)

	if regexp.MustCompile("^/assets/.*$").MatchString(relativePath) {
		(&controller{admin}).Asset(context)
		return
	}

	defer func() func() {
		begin := time.Now()
		log.Printf("Start [%s] %s\n", req.Method, req.RequestURI)

		return func() {
			log.Printf("Finish [%s] %s Took %.2fms\n", req.Method, req.RequestURI, time.Now().Sub(begin).Seconds()*1000)
		}
	}()()

	var currentUser qor.CurrentUser
	if admin.auth != nil {
		if currentUser = admin.auth.GetCurrentUser(context); currentUser == nil {
			http.Redirect(w, req, admin.auth.LoginURL(context), http.StatusSeeOther)
			return
		} else {
			context.CurrentUser = currentUser
			context.SetDB(context.GetDB().Set("qor:current_user", context.CurrentUser))
		}
	}
	context.Roles = roles.MatchedRoles(req, currentUser)

	// Call first middleware
	for _, middleware := range admin.router.middlewares {
		middleware.Handler(context, middleware)
		break
	}
}
