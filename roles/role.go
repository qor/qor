package roles

import (
	"fmt"

	"github.com/qor/qor"
)

type Role struct {
	definitions map[string]func(*qor.Context) bool
}

func New() *Role {
	return &Role{}
}

func (role *Role) Register(name string, fc func(*qor.Context) bool) {
	if role.definitions == nil {
		role.definitions = map[string]func(*qor.Context) bool{}
	}

	definition := role.definitions[name]
	if definition != nil {
		fmt.Println("%v already defined, overwrited it!", name)
	}
	role.definitions[name] = fc
}

func (role *Role) newPermission() *Permission {
	return &Permission{
		role:       role,
		allowRoles: map[PermissionMode][]string{},
		denyRoles:  map[PermissionMode][]string{},
	}
}

func (role *Role) Allow(mode PermissionMode, roles ...string) *Permission {
	return role.newPermission().Allow(mode, roles...)
}

func (role *Role) Deny(mode PermissionMode, roles ...string) *Permission {
	return role.newPermission().Deny(mode, roles...)

}
