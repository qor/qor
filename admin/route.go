package admin

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type (
	handle func(c *Context)

	handler struct {
		Path   *regexp.Regexp
		Handle handle
	}
)

type Router struct {
	Prefix      string
	routers     map[string][]handler
	middlewares []*Middleware
}

func newRouter() *Router {
	return &Router{routers: map[string][]handler{
		"GET":    []handler{},
		"PUT":    []handler{},
		"POST":   []handler{},
		"DELETE": []handler{},
	}}
}

type Middleware struct {
	Handler func(*Context, *Middleware)
	next    *Middleware
}

func (middleware Middleware) Next(context *Context) {
	if next := middleware.next; next != nil {
		next.Handler(context, next)
	}
}

func (r *Router) Use(handler func(*Context, *Middleware)) {
	r.middlewares = append(r.middlewares, &Middleware{Handler: handler})
}

func (r *Router) Get(path string, handle handle) {
	r.routers["GET"] = append(r.routers["GET"], handler{Path: regexp.MustCompile(path), Handle: handle})
}

func (r *Router) Post(path string, handle handle) {
	r.routers["POST"] = append(r.routers["POST"], handler{Path: regexp.MustCompile(path), Handle: handle})
}

func (r *Router) Put(path string, handle handle) {
	r.routers["PUT"] = append(r.routers["PUT"], handler{Path: regexp.MustCompile(path), Handle: handle})
}

func (r *Router) Delete(path string, handle handle) {
	r.routers["DELETE"] = append(r.routers["DELETE"], handler{Path: regexp.MustCompile(path), Handle: handle})
}

func (admin *Admin) MountTo(prefix string, mux *http.ServeMux) {
	prefix = "/" + strings.Trim(prefix, "/")
	router := admin.router
	router.Prefix = prefix

	admin.compile()

	controller := &controller{admin}
	router.Get("^/?$", controller.Dashboard)
	router.Get("^/!search$", controller.SearchCenter)

	var registerToRouter = func(res *Resource) {
		// New
		router.Get(fmt.Sprintf("/%v/new", res.ToParam()), controller.New)

		// Create
		router.Post(fmt.Sprintf("/%v", res.ToParam()), controller.Create)

		// Show
		router.Get(fmt.Sprintf("/%v/:id", res.ToParam()), controller.Show)

		// Edit
		router.Get(fmt.Sprintf("/%v/:id/edit", res.ToParam()), controller.Edit)

		// Update
		router.Put(fmt.Sprintf("/%v/:id", res.ToParam()), controller.Update)
		router.Post(fmt.Sprintf("/%v/:id", res.ToParam()), controller.Update)

		// Delete
		router.Delete(fmt.Sprintf("/%v/:id", res.ToParam()), controller.Delete)

		// Index
		router.Get(fmt.Sprintf("/%v", res.ToParam()), controller.Index)

		// Action
		router.Post(fmt.Sprintf("/%v/action/[^/]+(\\?.*)?$", res.ToParam()), controller.Action)
	}

	for _, res := range admin.resources {
		if !res.Config.Invisible {
			registerToRouter(res)
		}
	}

	mux.Handle(prefix, admin)     // /:prefix
	mux.Handle(prefix+"/", admin) // /:prefix/:xxx
}

func (res *Resource) configure() {
	modelType := res.GetAdmin().Config.DB.NewScope(res.Value).GetModelStruct().ModelType
	for i := 0; i < modelType.NumField(); i++ {
		if fieldStruct := modelType.Field(i); fieldStruct.Anonymous {
			if injector, ok := reflect.New(fieldStruct.Type).Interface().(resource.ConfigureResourceInterface); ok {
				injector.ConfigureQorResource(res)
			}
		}
	}

	if injector, ok := res.Value.(resource.ConfigureResourceInterface); ok {
		injector.ConfigureQorResource(res)
	}
}

func (admin *Admin) compile() {
	admin.generateMenuLinks()

	router := admin.GetRouter()

	for _, res := range admin.resources {
		res.configure()
	}

	router.Use(func(context *Context, middleware *Middleware) {
		w := context.Writer
		req := context.Request

		// 128 MB
		req.ParseMultipartForm(32 << 22)
		if len(req.Form["_method"]) > 0 {
			req.Method = strings.ToUpper(req.Form["_method"][0])
		}

		var pathMatch = regexp.MustCompile(path.Join(router.Prefix, `(\w+)(?:/(\w+))?[^/]*`))
		var matches = pathMatch.FindStringSubmatch(req.URL.Path)
		if len(matches) > 1 {
			context.setResource(admin.GetResource(matches[1]))
			if len(matches) > 2 {
				context.ResourceID = matches[2]
			}
		}

		handlers := router.routers[strings.ToUpper(req.Method)]
		relativePath := strings.TrimPrefix(req.URL.Path, router.Prefix)
		for _, handler := range handlers {
			if handler.Path.MatchString(relativePath) {
				handler.Handle(context)
				return
			}
		}
		http.NotFound(w, req)
	})

	for index, middleware := range router.middlewares {
		var next *Middleware
		if len(router.middlewares) > index+1 {
			next = router.middlewares[index+1]
		}
		middleware.next = next
	}
}

func (admin *Admin) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	context := Context{Context: &qor.Context{Config: admin.Config, Request: r, Writer: w}, Admin: admin}

	return &context
}

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

	firstStack := admin.router.middlewares[0]
	firstStack.Handler(context, firstStack)
}
