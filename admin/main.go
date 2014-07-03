package admin

import "github.com/qor/qor/resource"

type Admin struct {
}

func New() *Admin {
	admin := Admin{}
	return &admin
}

func (admin *Admin) AddResource(resource *resource.Resource) {
}
