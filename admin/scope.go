package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

func (res *Resource) Scope(scope *Scope) {
	res.scopes[scope.Name] = scope
}

type Scope struct {
	Name    string
	Label   string
	Handle  func(*gorm.DB, *qor.Context) *gorm.DB
	Default bool
}
