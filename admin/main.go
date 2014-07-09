package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/auth"
	"github.com/qor/qor/resource"
)

type Admin struct {
	DB        *gorm.DB
	resources map[string]*resource.Resource
	auth      auth.Auth
}

func New(db *gorm.DB) *Admin {
	admin := Admin{resources: map[string]*resource.Resource{}, DB: db}
	return &admin
}

func (admin *Admin) AddResource(resource *resource.Resource) {
	admin.resources[resource.RelativePath()] = resource
}

func (admin *Admin) SetAuth(auth auth.Auth) {
	admin.auth = auth
}
