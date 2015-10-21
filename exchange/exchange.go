package exchange

import (
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

func (res *Resource) Import(container Container, context *qor.Context) error {
	rows, err := container.Rows()
	if err == nil {
		for rows.Next() {
			if metaValues, err := rows.CurrentColumn(); err == nil {
				result := res.NewStruct()
				res.FindOneHandler(result, metaValues, context)
				if err = resource.DecodeToResource(res, result, metaValues, context).Start(); err == nil {
					if err = res.CallSaver(result, context); err != nil {
						return err
					}
				}
			}
		}
	}
	return err
}

func (res *Resource) Export(container Container, context *qor.Context) error {
	// scope to values
	// write to file
	// write logger
	return nil
}
