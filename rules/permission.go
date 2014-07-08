package rules

import "github.com/qor/qor"

type PermissionMode uint32

const (
	Read PermissionMode = 1 << (32 - 1 - iota)
	Update
	Create
	Delete
	All
)

type Permission struct {
	rule       *Rule
	allowRoles map[PermissionMode][]string
	denyRoles  map[PermissionMode][]string
}

func (p *Permission) HasPermission(mode PermissionMode, context *qor.Context) bool {
	return false
}

func (p *Permission) Allow(mode PermissionMode, roles ...string) *Permission {
	return p
}

func (p *Permission) Deny(mode PermissionMode, roles ...string) *Permission {
	return p
}
