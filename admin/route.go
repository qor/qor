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
	"github.com/qor/roles"
)

// Middleware is a way to filter a request and response coming into your application
// Register new middleware with `admin.GetRouter().Use(Middleware{
//   Name: "middleware name", // use middleware with same name will overwrite old one
//   Handler: func(*Context, *Middleware) {
//     // do something
//     // run next middleware
//     middleware.Next(context)
//   },
// })`
// It will be called in order, it need to be registered before `admin.MountTo`
type Middleware struct {
	Name    string
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
func (r *Router) Use(middleware *Middleware) {
	// compile middleware
	for index, m := range r.middlewares {
		// replace middleware have same name
		if m.Name == middleware.Name {
			middleware.next = m.next
			r.middlewares[index] = middleware
			if index > 1 {
				r.middlewares[index-1].next = middleware
			}
			return
		} else if len(r.middlewares) > index+1 {
			m.next = r.middlewares[index+1]
		} else if len(r.middlewares) == index+1 {
			m.next = middleware
		}
	}

	r.middlewares = append(r.middlewares, middleware)
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

	admin.generateMenuLinks()

	adminController := &controller{Admin: admin}
	router.Get("", adminController.Dashboard)
	router.Get("/!search", adminController.SearchCenter)

	var registerResourceToRouter func(*Resource, ...string)
	registerResourceToRouter = func(res *Resource, modes ...string) {
		var prefix string
		var param = res.ToParam()
		var primaryKey = res.ParamIDName()
		if prefix = func(r *Resource) string {
			p := param

			for r.base != nil {
				bp := r.base.ToParam()
				if bp == param {
					return ""
				}
				p = path.Join(bp, r.base.ParamIDName(), p)
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
					router.Get(path.Join(prefix, "new"), adminController.New, RouteConfig{
						PermissionMode: roles.Create,
						Resource:       res,
					})

					// Create
					router.Post(prefix, adminController.Create, RouteConfig{
						PermissionMode: roles.Create,
						Resource:       res,
					})
				}
			}

			if mode == "read" {
				if res.Config.Singleton {
					// Index
					router.Get(prefix, adminController.Show, RouteConfig{
						PermissionMode: roles.Read,
						Resource:       res,
					})
				} else {
					// Index
					router.Get(prefix, adminController.Index, RouteConfig{
						PermissionMode: roles.Read,
						Resource:       res,
					})

					// Show
					router.Get(path.Join(prefix, primaryKey), adminController.Show, RouteConfig{
						PermissionMode: roles.Read,
						Resource:       res,
					})
				}
			}

			if mode == "update" {
				if res.Config.Singleton {
					// Update
					router.Put(prefix, adminController.Update, RouteConfig{
						PermissionMode: roles.Update,
						Resource:       res,
					})
				} else {
					// Action
					for _, action := range res.Actions {
						actionController := &controller{Admin: admin, action: action}
						router.Get(path.Join(prefix, action.ToParam()), actionController.Action, RouteConfig{
							PermissionMode: roles.Update,
							Resource:       res,
						})
						router.Put(path.Join(prefix, action.ToParam()), actionController.Action, RouteConfig{
							PermissionMode: roles.Update,
							Resource:       res,
						})
					}

					// Edit
					router.Get(path.Join(prefix, primaryKey, "edit"), adminController.Edit, RouteConfig{
						PermissionMode: roles.Update,
						Resource:       res,
					})

					// Update
					router.Post(path.Join(prefix, primaryKey), adminController.Update, RouteConfig{
						PermissionMode: roles.Update,
						Resource:       res,
					})
					router.Put(path.Join(prefix, primaryKey), adminController.Update, RouteConfig{
						PermissionMode: roles.Update,
						Resource:       res,
					})

					// Action
					for _, action := range res.Actions {
						actionController := &controller{Admin: admin, action: action}
						router.Get(path.Join(prefix, primaryKey, action.ToParam()), actionController.Action, RouteConfig{
							PermissionMode: roles.Update,
							Resource:       res,
						})
						router.Put(path.Join(prefix, primaryKey, action.ToParam()), actionController.Action, RouteConfig{
							PermissionMode: roles.Update,
							Resource:       res,
						})
					}
				}
			}

			if mode == "delete" {
				if !res.Config.Singleton {
					// Delete
					router.Delete(path.Join(prefix, primaryKey), adminController.Delete, RouteConfig{
						PermissionMode: roles.Delete,
						Resource:       res,
					})
				}
			}
		}

		// Sub Resources
		for _, meta := range res.ConvertSectionToMetas(res.NewAttrs()) {
			if meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil && meta.Resource.base != nil {
				if len(meta.Resource.newSections) > 0 {
					registerResourceToRouter(meta.Resource, "create")
				}
			}
		}

		for _, meta := range res.ConvertSectionToMetas(res.ShowAttrs()) {
			if meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil && meta.Resource.base != nil {
				if len(meta.Resource.showSections) > 0 {
					registerResourceToRouter(meta.Resource, "read")
				}
			}
		}

		for _, meta := range res.ConvertSectionToMetas(res.EditAttrs()) {
			if meta.FieldStruct != nil && meta.FieldStruct.Relationship != nil && meta.Resource.base != nil {
				if len(meta.Resource.editSections) > 0 {
					registerResourceToRouter(meta.Resource, "update", "delete")
				}
			}
		}
	}

	for _, res := range admin.resources {
		res.configure()
		if !res.Config.Invisible {
			registerResourceToRouter(res, "create", "read", "update", "delete")
		}
	}

	mux.Handle(prefix, admin)     // /:prefix
	mux.Handle(prefix+"/", admin) // /:prefix/:xxx

	admin.compile()
}

func (admin *Admin) compile() {
	router := admin.GetRouter()
	router.Use(&Middleware{
		Name: "qor_handler",
		Handler: func(context *Context, middleware *Middleware) {
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

					handler.Handle(context)
					return
				}
			}

			http.NotFound(context.Writer, request)
		},
	})
}

// ServeHTTP dispatches the handler registered in the matched route
func (admin *Admin) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var relativePath = strings.TrimPrefix(req.URL.Path, admin.router.Prefix)
	var context = admin.NewContext(w, req)

	if regexp.MustCompile("^/assets/.*$").MatchString(relativePath) {
		(&controller{Admin: admin}).Asset(context)
		return
	}

	defer func() func() {
		begin := time.Now()
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
