package rules

import (
	"fmt"

	"github.com/qor/qor"
)

type Rule struct {
	definitions map[string]func(*qor.Context) bool
}

func New() *Rule {
	return &Rule{}
}

func (rule *Rule) Register(role string, fc func(*qor.Context) bool) {
	if rule.definitions == nil {
		rule.definitions = map[string]func(*qor.Context) bool{}
	}

	definition := rule.definitions[role]
	if definition != nil {
		fmt.Println("%v already defined, overwrited it!", role)
	}
	rule.definitions[role] = fc
}

func (rule *Rule) newPermission() *Permission {
	return &Permission{
		rule:       rule,
		allowRoles: map[PermissionMode][]string{},
		denyRoles:  map[PermissionMode][]string{},
	}
}

func (rule *Rule) Allow(mode PermissionMode, roles ...string) *Permission {
	return rule.newPermission().Allow(mode, roles...)
}

func (rule *Rule) Deny(mode PermissionMode, roles ...string) *Permission {
	return rule.newPermission().Deny(mode, roles...)

}
