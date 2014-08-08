package admin

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/auth"
	"github.com/qor/qor/resource"
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
	res.SetFinder(func(result interface{}, metaDatas resource.MetaDatas, context *qor.Context) error {
		if id := metaDatas.Get("_id"); id != nil {
			if destroy := metaDatas.Get("_destroy"); destroy != nil {
				if fmt.Sprintf("%v", destroy.Value) != "0" {
					context.DB.Delete(result, id.Value)
					return resource.ErrProcessorSkipLeft
				}
			}
			return context.DB.First(result, id.Value).Error
		}
		return nil
	})

	admin.Resources[name] = res
	return res
}

func (admin *Admin) SetAuth(auth auth.Auth) {
	admin.auth = auth
}
