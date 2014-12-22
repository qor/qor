package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

type Scope struct {
	Name    string
	Handler func(*gorm.DB, *qor.Context) *gorm.DB
	Default bool
}
