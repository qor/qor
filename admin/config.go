package admin

import "github.com/qor/roles"

type Config struct {
	Name       string
	Menu       []string
	Invisible  bool
	Permission *roles.Permission
	Themes     []string
	PageCount  int
	Singleton  bool
}
