package qor

import (
	"github.com/jinzhu/gorm"

	"net/http"
)

type CurrentUser interface {
	DisplayName() string
}

type Context struct {
	Request    *http.Request
	ResourceID string
	Config     *Config
	Roles      []string
	DB         *gorm.DB
}

func (context *Context) New() *Context {
	return &Context{
		Request:    context.Request,
		ResourceID: context.ResourceID,
		Config:     context.Config,
		DB:         context.DB,
	}
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
