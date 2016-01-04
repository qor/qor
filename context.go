package qor

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

type CurrentUser interface {
	DisplayName() string
}

type Context struct {
	Request     *http.Request
	Writer      http.ResponseWriter
	ResourceID  string
	Config      *Config
	Roles       []string
	DB          *gorm.DB
	CurrentUser CurrentUser
	Errors
}

func (context *Context) Clone() *Context {
	var clone = *context
	return &clone
}

func (context *Context) GetDB() *gorm.DB {
	if context.DB != nil {
		return context.DB
	} else {
		return context.Config.DB
	}
}

func (context *Context) SetDB(db *gorm.DB) {
	context.DB = db
}
