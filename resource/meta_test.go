package resource_test

import (
	"fmt"
	"testing"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/test/utils"
)

func TestValuerAndSetter(t *testing.T) {
	checker := func(record interface{}, meta *resource.Meta, value interface{}) {
		var (
			context   = &qor.Context{DB: utils.TestDB()}
			metaValue = &resource.MetaValue{Name: meta.Name, Value: value}
		)

		meta.PreInitialize()
		meta.Initialize()

		meta.Setter(record, metaValue, context)

		if result := meta.Valuer(record, context); fmt.Sprint(result) != fmt.Sprint(value) {
			t.Errorf("Wrong value, should be %v, but got %v", fmt.Sprint(value), result)
		}
	}

	user := &struct {
		Name string
	}{}

	res := resource.New(&user)

	nameMeta := &resource.Meta{
		Name:         "Name",
		BaseResource: res,
	}

	checker(&user, nameMeta, "hello world")
}
