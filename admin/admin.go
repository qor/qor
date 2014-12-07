package admin

import (
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"reflect"
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
	name := strings.ToLower(reflect.Indirect(reflect.ValueOf(value)).Type().Name())
	for _, n := range names {
		name = n
	}

	return &Resource{Name: name, Resource: resource.Resource{Value: value}, cachedMetas: &map[string][]*resource.Meta{}}
}

func (admin *Admin) NewResource(value interface{}, names ...string) *Resource {
	res := NewResource(value, names...)
	admin.Resources[res.Name] = res
	return res
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
