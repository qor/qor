package resource

import "reflect"

type Resource struct {
	Name  string
	attrs *attrs
	meta  *meta
}

func New(value interface{}) *Resource {
	data := reflect.Indirect(reflect.ValueOf(value))
	resourceName := data.Type().Name()
	resource := Resource{Name: resourceName}
	return &resource
}

func (r *Resource) Attrs() *attrs {
	return r.attrs
}

func (r *Resource) Meta() *meta {
	return r.meta
}

func (r *Resource) RelativePath() string {
	return r.Name
}
