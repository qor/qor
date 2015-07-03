package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

func (res *Resource) Scope(scope *Scope) {
	res.scopes = append(res.scopes, scope)
}

type Scope struct {
	Name    string
	Label   string
	Group   string
	Handle  func(*gorm.DB, *qor.Context) *gorm.DB
	Default bool
}
