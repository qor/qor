package resource_test

import (
	"fmt"
	"testing"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/test/utils"
)

func checkMeta(record interface{}, meta *resource.Meta, value interface{}, t *testing.T) {
	var (
		context   = &qor.Context{DB: utils.TestDB()}
		metaValue = &resource.MetaValue{Name: meta.Name, Value: value}
	)

	meta.PreInitialize()
	meta.Initialize()

	meta.Setter(record, metaValue, context)
	if context.HasError() {
		t.Errorf("No error should happen, but got %v", context.Errors)
	}

	if result := meta.Valuer(record, context); fmt.Sprint(result) != fmt.Sprint(value) {
		t.Errorf("Wrong value, should be %v, but got %v", fmt.Sprint(value), result)
	}
}

func TestStringMetaValuerAndSetter(t *testing.T) {
	user := &struct {
		Name string
	}{}

	res := resource.New(&user)

	nameMeta := &resource.Meta{
		Name:         "Name",
		BaseResource: res,
	}

	checkMeta(&user, nameMeta, "hello world", t)
}

func TestIntMetaValuerAndSetter(t *testing.T) {
	user := &struct {
		Age int
	}{}

	res := resource.New(&user)

	meta := &resource.Meta{
		Name:         "Age",
		BaseResource: res,
	}

	checkMeta(&user, meta, 18, t)
}
