package admin

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

type Admin struct {
	Config    *qor.Config
	Resources map[string]*Resource
	auth      Auth
	router    *Router
}

func New(config *qor.Config) *Admin {
	admin := Admin{
		Resources: map[string]*Resource{},
		Config:    config,
		router:    newRouter(),
	}
	return &admin
}

func NewResource(value interface{}, names ...string) *Resource {
	res := resource.New(value, names...)
	return &Resource{Resource: *res, cachedMetas: &map[string][]*resource.Meta{}}
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
