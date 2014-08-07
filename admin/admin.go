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

func (admin *Admin) NewResource(name string, value interface{}) {
	resource := &Resource{Name: name}
	resource.Value = value
	admin.Resources[name] = resource
}

func (admin *Admin) SetAuth(auth auth.Auth) {
	admin.auth = auth
}
