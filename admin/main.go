package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/resource"
)

type Admin struct {
	resources map[string]*resource.Resource
	DB        *gorm.DB
}

func New(db *gorm.DB) *Admin {
	admin := Admin{resources: make(map[string]*resource.Resource), DB: db}
	return &admin
}

func (admin *Admin) AddResource(resource *resource.Resource) {
	admin.resources[resource.RelativePath()] = resource
}
