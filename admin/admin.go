package admin

import (
	"net/http"
	"text/template"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type Admin struct {
	Config    *qor.Config
	Resources map[string]*Resource
	auth      Auth
	router    *Router
	funcMaps  template.FuncMap
}

func New(config *qor.Config) *Admin {
	admin := Admin{
		Resources: map[string]*Resource{},
		funcMaps:  make(template.FuncMap),
		Config:    config,
		router:    newRouter(),
	}
	return &admin
}

func NewResource(value interface{}, names ...string) *Resource {
	res := resource.New(value, names...)
	return &Resource{
		Resource:    *res,
		cachedMetas: &map[string][]*resource.Meta{},
		scopes:      map[string]*Scope{},
		filters:     map[string]*Filter{},
		actions:     map[string]*Action{},
	}
}

func (admin *Admin) NewResource(value interface{}, names ...string) *Resource {
	res := NewResource(value, names...)
	admin.Resources[res.Name] = res
	return res
}

func (admin *Admin) GetResource(name string) *Resource {
	return admin.Resources[name]
}

func (admin *Admin) UseResource(res *Resource) {
	admin.Resources[res.Name] = res
}

func (admin *Admin) SetAuth(auth Auth) {
	admin.auth = auth
}

func (admin *Admin) GetRouter() *Router {
	return admin.router
}

func (admin *Admin) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	var currentUser *qor.CurrentUser
	context := Context{Context: &qor.Context{Config: admin.Config, Request: r}, Writer: w, Admin: admin}
	if admin.auth != nil {
		currentUser = admin.auth.GetCurrentUser(&context)
	}
	context.Roles = roles.MatchedRoles(r, currentUser)

	return &context
}

func (admin *Admin) RegisterFuncMap(name string, fc interface{}) {
	admin.funcMaps[name] = fc
}
