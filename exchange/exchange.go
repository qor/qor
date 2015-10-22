package exchange

import (
	"reflect"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type Exchange struct {
	Config    *qor.Config
	resources []*Resource
}

func New(config qor.Config) *Exchange {
	return &Exchange{Config: &config}
}

type Resource struct {
	resource.Resource
	Config *Config
	metas  []*Meta
}

type Config struct {
	Permission    *roles.Permission
	WithoutHeader bool
}

func NewResource(value interface{}, config ...Config) *Resource {
	res := Resource{Resource: *resource.New(value)}
	if len(config) > 0 {
		res.Config = &config[0]
	} else {
		res.Config = &Config{}
	}
	return &res
}

func (res *Resource) Meta(meta Meta) *Meta {
	res.metas = append(res.metas, &meta)
	return &meta
}

func (res *Resource) GetMeta(name string) *Meta {
	for _, meta := range res.metas {
		if meta.Name == name {
			return meta
		}
	}
	return nil
}

func (res *Resource) GetMetas([]string) []resource.Metaor {
	metas := []resource.Metaor{}
	for _, meta := range res.metas {
		metas = append(metas, meta)
	}
	return metas
}

func (res *Resource) Import(container Container, context *qor.Context) error {
	rows, err := container.Rows(res)
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
	results := res.NewSlice()
	if err := context.GetDB().Find(results).Error; err == nil {

		reflectValue := reflect.ValueOf(results)
		for i := 0; i < reflectValue.Len(); i++ {
			var metaValues *resource.MetaValues
			container.WriteRow(metaValues)
		}
	} else {
		return err
	}
	return nil
}
