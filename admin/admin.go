package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/auth"
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

func (admin *Admin) NewResource(name string, value interface{}) *Resource {
	res := &Resource{Name: name}
	res.Value = value
	admin.Resources[name] = res
	return res
}

func (admin *Admin) SetAuth(auth auth.Auth) {
	admin.auth = auth
}
