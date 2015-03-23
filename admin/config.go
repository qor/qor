package admin

import "github.com/qor/qor/roles"

type Config struct {
	Name       string
	Menu       []string
	Invisible  bool
	Permission *roles.Permission
	Theme      string
}
