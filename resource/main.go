package resource

import (
	"reflect"
	"strings"
)

type Resource struct {
	Model interface{}
	Name  string
	attrs *attrs
	meta  *meta
}

func New(value interface{}) *Resource {
	data := reflect.Indirect(reflect.ValueOf(value))
	resourceName := data.Type().Name()
	resource := Resource{Name: resourceName, Model: value}
	return &resource
}

func (r *Resource) Attrs() *attrs {
	return r.attrs
}

func (r *Resource) Meta() *meta {
	return r.meta
}

func (r *Resource) RelativePath() string {
	return strings.ToLower(r.Name)
}
