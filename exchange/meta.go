package exchange

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

type Meta struct {
	Name       string
	Valuer     func(interface{}, *qor.Context) interface{}
	Setter     func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context)
	Permission *roles.Permission
}

func (meta *Meta) GetName() string {
	return meta.Name
}

func (meta *Meta) GetFieldName() string {
	return meta.Name
}

func (meta *Meta) GetMetas() []resource.Metaor {
	// FIXME
	return []resource.Metaor{}
}

func (meta *Meta) GetResource() resource.Resourcer {
	return nil
}

func (meta *Meta) GetValuer() func(interface{}, *qor.Context) interface{} {
	return meta.Valuer
}

func (meta *Meta) GetSetter() func(resource interface{}, metaValue *resource.MetaValue, context *qor.Context) {
	return meta.Setter
}

func (meta *Meta) HasPermission(mode roles.PermissionMode, context *qor.Context) bool {
	if meta.Permission == nil {
		return true
	}
	return meta.Permission.HasPermission(mode, context.Roles...)
}
