package exchange

import (
	"io"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

type Exchange struct {
	Config    *qor.Config
	resources []*Resource
}

type Resource struct {
	resource.Resource
}

type Meta struct {
	Name string
}

func (exchange *Exchange) AddResource(value interface{}) {
	res := Resource{Resource: *resource.New(value)}
	exchange.resources = append(exchange.resources, &res)
}

func (exchange *Exchange) Import(value interface{}, context qor.Context) {
}

func (exchange *Exchange) Export(scope *gorm.DB, writer io.Writer, logger interface{}, context qor.Context) {
}

func (res *Resource) Meta(meta Meta) {
}
