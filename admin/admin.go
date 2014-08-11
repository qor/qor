package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/auth"
	"github.com/qor/qor/resource"
	"strings"

	"reflect"
)

type Admin struct {
	Prefix    string
	DB        *gorm.DB
	Resources map[string]*Resource
	auth      auth.Auth
}

func New(db *gorm.DB) *Admin {
	admin := Admin{Resources: map[string]*Resource{}, DB: db}
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

func (admin *Admin) SetAuth(auth auth.Auth) {
	admin.auth = auth
}
