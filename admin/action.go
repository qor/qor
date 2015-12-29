package admin

import (
	"fmt"
	"github.com/qor/qor/roles"
	"reflect"
)

func (res *Resource) Action(action *Action) {
	res.actions = append(res.actions, action)
}

type ActionArgument struct {
	IDs     []string
	Context *Context
}

type Action struct {
	Name       string
	Label      string
	Handle     func(arg *ActionArgument) error
	Resource   *Resource
	Permission *roles.Permission
	Visibles   []string
}

func (arg *ActionArgument) AllRecords() []interface{} {
	var records = []interface{}{}
	results := arg.Context.Resource.NewSlice()
	arg.Context.GetDB().Where(fmt.Sprintf("%v IN (?)", arg.Context.Resource.PrimaryDBName()), arg.IDs).Find(results)
	resultValues := reflect.Indirect(reflect.ValueOf(results))
	for i := 0; i < resultValues.Len(); i++ {
		records = append(records, resultValues.Index(i).Interface())
	}
	return records
}
