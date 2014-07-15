package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/auth"
	"github.com/qor/qor/resource"
)

type Admin struct {
	Prefix    string
	DB        *gorm.DB
	Resources map[string]*resource.Resource
	auth      auth.Auth
}

func New(db *gorm.DB) *Admin {
	admin := Admin{Resources: map[string]*resource.Resource{}, DB: db}
	return &admin
}

func (admin *Admin) AddResource(resource *resource.Resource) {
	admin.Resources[resource.RelativePath()] = resource
}

func (admin *Admin) SetAuth(auth auth.Auth) {
	admin.auth = auth
}
