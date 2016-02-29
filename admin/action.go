package admin

import (
	"fmt"
	"reflect"

	"github.com/qor/qor"
	"github.com/qor/qor/utils"
	"github.com/qor/roles"
)

// Action register action for qor resource
func (res *Resource) Action(action *Action) {
	if action.Label == "" {
		action.Label = utils.HumanizeString(action.Name)
	}
	res.Actions = append(res.Actions, action)
}

// ActionArgument action argument that used in handle
type ActionArgument struct {
	PrimaryValues []string
	Context       *Context
	Argument      interface{}
}

// Action action definiation
type Action struct {
	Name       string
	Label      string
	Handle     func(arg *ActionArgument) error
	Resource   *Resource
	Permission *roles.Permission
	Visibles   []string
}

// ToParam used to register routes for actions
func (action Action) ToParam() string {
	return utils.ToParamString(action.Name)
}

// HasPermission check if current user has permission for the action
func (action Action) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if action.Permission == nil {
		return true
	}
	return action.Permission.HasPermission(mode, context.Roles...)
}

// FindSelectedRecords find selected records when run bulk actions
func (actionArgument *ActionArgument) FindSelectedRecords() []interface{} {
	var (
		context  = actionArgument.Context
		resource = context.Resource
		records  = []interface{}{}
	)

	clone := context.clone()
	clone.SetDB(clone.GetDB().Where(fmt.Sprintf("%v IN (?)", resource.PrimaryDBName()), actionArgument.PrimaryValues))
	results, _ := clone.FindMany()

	resultValues := reflect.Indirect(reflect.ValueOf(results))
	for i := 0; i < resultValues.Len(); i++ {
		records = append(records, resultValues.Index(i).Interface())
	}
	return records
}
