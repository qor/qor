package resource

import (
	"fmt"
	"os"
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
	if reflect.ValueOf(value).Kind() != reflect.Ptr {
		fmt.Println("resource.New only accept Pointer as argument, but got", reflect.ValueOf(value).Type())
		os.Exit(1)
	}

	data := reflect.Indirect(reflect.ValueOf(value))
	resourceName := data.Type().Name()
	resource := Resource{Name: resourceName, Model: value, attrs: &attrs{}}
	resource.meta = &meta{resource: &resource}
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
