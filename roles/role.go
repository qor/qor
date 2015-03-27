package roles

import (
	"fmt"
	"net/http"

	"github.com/qor/qor"
)

type Role struct {
	definitions map[string]func(req *http.Request, currentUser qor.CurrentUser) bool
}

func New() *Role {
	return &Role{}
}

var role = &Role{}

func Register(name string, fc func(req *http.Request, currentUser qor.CurrentUser) bool) {
	role.Register(name, fc)
}

func NewPermission() *Permission {
	return role.newPermission()
}

func (role *Role) newPermission() *Permission {
	return &Permission{
		Role:       role,
		allowRoles: map[PermissionMode][]string{},
		denyRoles:  map[PermissionMode][]string{},
	}
}

func Allow(mode PermissionMode, roles ...string) *Permission {
	return role.Allow(mode, roles...)
}

func Deny(mode PermissionMode, roles ...string) *Permission {
	return role.Deny(mode, roles...)
}

func MatchedRoles(req *http.Request, currentUser qor.CurrentUser) []string {
	return role.MatchedRoles(req, currentUser)
}

func (role *Role) MatchedRoles(req *http.Request, currentUser qor.CurrentUser) (roles []string) {
	if definitions := role.definitions; definitions != nil {
		for name, definition := range definitions {
			if definition(req, currentUser) {
				roles = append(roles, name)
			}
		}
	}
	return
}

func (role *Role) Get(name string) (func(req *http.Request, currentUser qor.CurrentUser) bool, bool) {
	fc, ok := role.definitions[name]
	return fc, ok
}

func (role *Role) Register(name string, fc func(req *http.Request, currentUser qor.CurrentUser) bool) {
	if role.definitions == nil {
		role.definitions = map[string]func(req *http.Request, currentUser qor.CurrentUser) bool{}
	}

	definition := role.definitions[name]
	if definition != nil {
		fmt.Printf("%v already defined, overwrited it!\n", name)
	}
	role.definitions[name] = fc
}

func (role *Role) Allow(mode PermissionMode, roles ...string) *Permission {
	return role.newPermission().Allow(mode, roles...)
}

func (role *Role) Deny(mode PermissionMode, roles ...string) *Permission {
	return role.newPermission().Deny(mode, roles...)
}
