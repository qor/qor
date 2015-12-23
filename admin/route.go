package admin

import (
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/qor/qor"
	"github.com/qor/qor/roles"
)

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
func (r *Router) Get(path string, handle requestHandler, config ...RouteConfig) {
	r.routers["GET"] = append(r.routers["GET"], newRouteHandler(path, handle, config...))
}

// Post register a POST request handle with the given path
func (r *Router) Post(path string, handle requestHandler, config ...RouteConfig) {
	r.routers["POST"] = append(r.routers["POST"], newRouteHandler(path, handle, config...))
}

// Put register a PUT request handle with the given path
func (r *Router) Put(path string, handle requestHandler, config ...RouteConfig) {
	r.routers["PUT"] = append(r.routers["PUT"], newRouteHandler(path, handle, config...))
}

// Delete register a DELETE request handle with the given path
func (r *Router) Delete(path string, handle requestHandler, config ...RouteConfig) {
	r.routers["DELETE"] = append(r.routers["DELETE"], newRouteHandler(path, handle, config...))
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
		var prefix string
		if prefix = func(r *Resource) string {
			cp := r.ToParam()
			p := cp

			for r.base != nil {
				bp := r.base.ToParam()
				if bp == cp {
					return ""
				}
				p = path.Join(bp, ":id", p)
				r = r.base
			}
			return "/" + strings.Trim(p, "/")
		}(res); prefix == "" {
			return
		}

		for _, mode := range modes {
			if mode == "create" {
				if !res.Config.Singleton {
					// New
					router.Get(path.Join(prefix, "new"), controller.New, RouteConfig{
						PermissionMode: roles.Create,
						Resource:       res,
					})

					// Create
					router.Post(prefix, controller.Create, RouteConfig{
						PermissionMode: roles.Create,
						Resource:       res,
					})
				}
			}

			if mode == "read" {
				if res.Config.Singleton {
					// Index
					router.Get(prefix, controller.Show, RouteConfig{
						PermissionMode: roles.Read,
						Resource:       res,
					})
				} else {
					// Index
					router.Get(prefix, controller.Index, RouteConfig{
						PermissionMode: roles.Read,
						Resource:       res,
					})

					// Show
					router.Get(path.Join(prefix, ":id"), controller.Show, RouteConfig{
						PermissionMode: roles.Read,
						Resource:       res,
					})
				}
			}

			if mode == "update" {
				if res.Config.Singleton {
					// Update
					router.Put(prefix, controller.Update, RouteConfig{
						PermissionMode: roles.Update,
						Resource:       res,
					})

					// Action
					for _, action := range res.actions {
						router.Post(path.Join(prefix, action.Name), controller.Action, RouteConfig{
							PermissionMode: roles.Update,
							Resource:       res,
						})
					}
				} else {
					// Edit
					router.Get(path.Join(prefix, ":id", "edit"), controller.Edit, RouteConfig{
						PermissionMode: roles.Update,
						Resource:       res,
					})

					// Update
					router.Put(path.Join(prefix, ":id"), controller.Update, RouteConfig{
						PermissionMode: roles.Update,
						Resource:       res,
					})

					// Action
					for _, action := range res.actions {
						router.Post(path.Join(prefix, ":id", action.Name), controller.Action, RouteConfig{
							PermissionMode: roles.Update,
							Resource:       res,
						})
					}
				}
			}

			if mode == "delete" {
				if !res.Config.Singleton {
					// Delete
					router.Delete(path.Join(prefix, ":id"), controller.Delete, RouteConfig{
						PermissionMode: roles.Delete,
						Resource:       res,
					})
				}
			}
		}

		// Sub Resources
		for _, meta := range res.ConvertSectionToMetas(res.NewAttrs()) {
			if meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil && meta.Resource.base != nil {
				registerResourceToRouter(meta.Resource, "create")
			}
		}

		for _, meta := range res.ConvertSectionToMetas(res.ShowAttrs()) {
			if meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil && meta.Resource.base != nil {
				registerResourceToRouter(meta.Resource, "read")
			}
		}

		for _, meta := range res.ConvertSectionToMetas(res.EditAttrs()) {
			if meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil && meta.Resource.base != nil {
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
		request := context.Request

		// 128 MB
		request.ParseMultipartForm(32 << 22)

		// set request method
		if len(request.Form["_method"]) > 0 {
			request.Method = strings.ToUpper(request.Form["_method"][0])
		}

		relativePath := "/" + strings.Trim(
			strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, router.Prefix), path.Ext(request.URL.Path)),
			"/",
		)

		handlers := router.routers[strings.ToUpper(request.Method)]
		for _, handler := range handlers {
			if params, ok := handler.try(relativePath); ok && handler.HasPermission(context.Context) {
				if len(params) > 0 {
					context.Request.URL.RawQuery = url.Values(params).Encode() + "&" + context.Request.URL.RawQuery
				}

				context.setResource(handler.Config.Resource)
				if context.Resource == nil {
					if matches := regexp.MustCompile(path.Join(router.Prefix, `([^/]+)`)).FindStringSubmatch(request.URL.Path); len(matches) > 1 {
						context.setResource(admin.GetResource(matches[1]))
					}
				}

				if ids, ok := context.Request.URL.Query()[":id"]; ok {
					context.ResourceID = ids[len(ids)-1]
				}

				handler.Handle(context)
				return
			}
		}

		http.NotFound(context.Writer, request)
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
