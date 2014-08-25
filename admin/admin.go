package admin

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"strings"

	"reflect"
)

type Admin struct {
	Prefix    string
	Config    *qor.Config
	Resources map[string]*Resource
	auth      Auth
}

func New(config *qor.Config) *Admin {
	admin := Admin{Resources: map[string]*Resource{}, Config: config}
	return &admin
}

func NewResource(value interface{}, names ...string) *Resource {
	name := strings.ToLower(reflect.Indirect(reflect.ValueOf(value)).Type().Name())
	for _, n := range names {
		name = n
	}

	return &Resource{Name: name, Resource: resource.Resource{Value: value}}
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
