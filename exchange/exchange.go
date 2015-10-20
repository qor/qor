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

func (res *Resource) Meta(meta Meta) {
}

func (res *Resource) Import(file interface{}, context qor.Context) {
	// file To MetaValues
	// decode to resource
	// save each value
}

func (res *Resource) Export(scope *gorm.DB, writer io.Writer, logger interface{}, context qor.Context) {
	// scope to values
	// write to file
	// write logger
}
