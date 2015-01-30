package admin

import "github.com/qor/qor/roles"

type Config struct {
	Name       string
	Menus      []string
	Invisible  bool
	Permission *roles.Permission
}
