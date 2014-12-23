package admin

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
)

type Filter struct {
	Name       string
	Operations []string
	Handler    func(name string, value string, scope *gorm.DB, context *qor.Context) *gorm.DB
}
