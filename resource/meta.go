package resource

import (
	"github.com/qor/qor"
	"github.com/qor/qor/roles"
)

type Metaor interface {
	GetName() string
	GetFieldName() string
	GetMetas() []Metaor
	GetResource() Resourcer
	GetSetter() func(resource interface{}, metaValue *MetaValue, context *qor.Context)
	HasPermission(roles.PermissionMode, *qor.Context) bool
}
