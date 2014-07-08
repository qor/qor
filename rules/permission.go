package rules

import "github.com/qor/qor"

type PermissionMode uint32

const (
	Read PermissionMode = 1 << (32 - 1 - iota)
	Update
	Create
	Delete
	CURD
)

type Permission struct {
	rule       *Rule
	allowRoles map[PermissionMode][]string
	denyRoles  map[PermissionMode][]string
}

func (permission *Permission) HasPermission(mode PermissionMode, context *qor.Context) bool {
	if len(permission.denyRoles) != 0 {
		if roles := permission.denyRoles[mode]; roles != nil {
			if definitions := permission.rule.definitions; definitions != nil {
				for _, r := range roles {
					if definitions[r](context) {
						return false
					}
				}
			}
		}
	}

	if len(permission.allowRoles) != 0 {
		if roles := permission.allowRoles[mode]; roles != nil {
			if definitions := permission.rule.definitions; definitions != nil {
				for _, r := range roles {
					if definitions[r](context) {
						return true
					}
				}
			}
		}
	} else if len(permission.denyRoles) != 0 {
		return true
	}

	return false
}

func (permission *Permission) Allow(mode PermissionMode, roles ...string) *Permission {
	if mode == CURD {
		return permission.Allow(Create, roles...).Allow(Update, roles...).Allow(Read, roles...).Allow(Delete, roles...)
	}

	if permission.allowRoles[mode] == nil {
		permission.allowRoles[mode] = []string{}
	}
	permission.allowRoles[mode] = append(permission.allowRoles[mode], roles...)
	return permission
}

func (permission *Permission) Deny(mode PermissionMode, roles ...string) *Permission {
	if mode == CURD {
		return permission.Deny(Create, roles...).Deny(Update, roles...).Deny(Read, roles...).Deny(Delete, roles...)
	}

	if permission.denyRoles[mode] == nil {
		permission.denyRoles[mode] = []string{}
	}
	permission.denyRoles[mode] = append(permission.denyRoles[mode], roles...)
	return permission
}
