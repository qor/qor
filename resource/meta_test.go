package resource_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	testutils "github.com/qor/qor/test/utils"
	"github.com/qor/qor/utils"
)

func format(value interface{}) string {
	return fmt.Sprint(utils.Indirect(reflect.ValueOf(value)).Interface())
}

func checkMeta(record interface{}, meta *resource.Meta, value interface{}, t *testing.T) {
	var (
		context   = &qor.Context{DB: testutils.TestDB()}
		metaValue = &resource.MetaValue{Name: meta.Name, Value: value}
	)

	meta.PreInitialize()
	meta.Initialize()

	meta.Setter(record, metaValue, context)
	if context.HasError() {
		t.Errorf("No error should happen, but got %v", context.Errors)
	}

	if result := meta.Valuer(record, context); format(result) != fmt.Sprint(value) {
		t.Errorf("Wrong value, should be %v, but got %v", format(value), result)
	}
}

func TestStringMetaValuerAndSetter(t *testing.T) {
	user := &struct {
		Name  string
		Name2 *string
	}{}

	res := resource.New(&user)

	meta := &resource.Meta{
		Name:         "Name",
		BaseResource: res,
	}

	checkMeta(&user, meta, "hello world", t)

	meta2 := &resource.Meta{
		Name:         "Name2",
		BaseResource: res,
	}

	checkMeta(&user, meta2, "hello world2", t)
}

func TestIntMetaValuerAndSetter(t *testing.T) {
	user := &struct {
		Age  int
		Age2 uint
		Age3 *int8
		Age4 *uint8
	}{}

	res := resource.New(&user)

	meta := &resource.Meta{
		Name:         "Age",
		BaseResource: res,
	}

	checkMeta(&user, meta, 18, t)

	meta2 := &resource.Meta{
		Name:         "Age2",
		BaseResource: res,
	}

	checkMeta(&user, meta2, "28", t)

	meta3 := &resource.Meta{
		Name:         "Age3",
		BaseResource: res,
	}

	checkMeta(&user, meta3, 38, t)

	meta4 := &resource.Meta{
		Name:         "Age4",
		BaseResource: res,
	}

	checkMeta(&user, meta4, "48", t)
}
