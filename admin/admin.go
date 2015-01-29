package admin

import (
	"text/template"

	"github.com/qor/qor"
)

type Admin struct {
	Config    *qor.Config
	resources []*Resource
	auth      Auth
	router    *Router
	funcMaps  template.FuncMap
}

type Injector interface {
	InjectQorAdmin(*Admin)
}

func New(config *qor.Config) *Admin {
	admin := Admin{
		funcMaps: make(template.FuncMap),
		Config:   config,
		router:   newRouter(),
	}
	return &admin
}

func (admin *Admin) NewResource(value interface{}, names ...string) *Resource {
	res := &Resource{Value: value}
	admin.resources = append(admin.resources, res)

	if injector, ok := value.(Injector); ok {
		injector.InjectQorAdmin(admin)
	}
	return res
}

func (admin *Admin) GetResource(name string) *Resource {
	for _, res := range admin.resources {
		if res.ToParam() == name {
			return res
		}
	}
	return nil
}

func (admin *Admin) SetAuth(auth Auth) {
	admin.auth = auth
}

func (admin *Admin) GetRouter() *Router {
	return admin.router
}

func (admin *Admin) RegisterFuncMap(name string, fc interface{}) {
	admin.funcMaps[name] = fc
}
