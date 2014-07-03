package admin

import "github.com/qor/qor/resource"

type Admin struct {
	resources map[string]*resource.Resource
}

func New() *Admin {
	admin := Admin{resources: make(map[string]*resource.Resource)}
	return &admin
}

func (admin *Admin) AddResource(resource *resource.Resource) {
	admin.resources[resource.RelativePath()] = resource
}
