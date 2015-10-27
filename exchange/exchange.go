package exchange

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type Resource struct {
	resource.Resource
	Config *Config
	metas  []*Meta
}

type Config struct {
	PrimaryField  string
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
	res.Permission = res.Config.Permission

	if res.Config.PrimaryField != "" {
		res.FindOneHandler = func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
			scope := context.GetDB().NewScope(res.Value)
			if field, ok := scope.FieldByName(res.Config.PrimaryField); ok {
				return context.GetDB().First(result, fmt.Sprintf("%v = ?", scope.Quote(field.DBName)), metaValues.Get(res.Config.PrimaryField).Value).Error
			} else {
				return errors.New("failed to find primary field")
			}
		}
	}
	return &res
}

func (res *Resource) Meta(meta Meta) *Meta {
	meta.base = res
	meta.updateMeta()
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
	rows, err := container.NewReader(res, context)
	if err == nil {
		for rows.Next() {
			var metaValues *resource.MetaValues
			if metaValues, err = rows.ReadRow(); err == nil {
				result := res.NewStruct()
				res.FindOneHandler(result, metaValues, context)
				if err = resource.DecodeToResource(res, result, metaValues, context).Start(); err == nil {
					if err = res.CallSave(result, context); err != nil {
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
		reflectValue := reflect.Indirect(reflect.ValueOf(results))
		if writer, err := container.NewWriter(res, context); err == nil {
			writer.WriteHeader()
			for i := 0; i < reflectValue.Len(); i++ {
				var result = reflectValue.Index(i).Interface()
				if err := writer.WriteRow(result); err != nil {
					return err
				}
			}
			writer.Flush()
		} else {
			return err
		}
	} else {
		return err
	}
	return nil
}
