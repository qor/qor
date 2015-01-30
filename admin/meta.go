package admin

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type Meta struct {
	*resource.Meta
	Base       resource.Resourcer
	Name       string
	Label      string
	Type       string
	Valuer     func(interface{}, *qor.Context) interface{}
	Setter     func(resource interface{}, metaValues *resource.MetaValues, context *qor.Context)
	Collection interface{}
	Permission *roles.Permission
}
