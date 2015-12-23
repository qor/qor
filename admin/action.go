package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/roles"
)

func (res *Resource) Action(action *Action) {
	res.actions = append(res.actions, action)
}

type ActionArgument struct {
	IDs      []string
	Argument interface{}
	Context  *qor.Context
}

type Action struct {
	Name       string
	Label      string
	Handle     func(scope *gorm.DB, context *qor.Context) error
	Handle1    func(arg *ActionArgument) error
	Resource   *Resource
	Permission *roles.Permission
	Visibles   []string
}

func (action *Action) NewStruct() interface{} {
	return action.Resource
}
