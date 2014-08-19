package qor

import (
	"github.com/jinzhu/gorm"

	"net/http"
)

type CurrentUser interface {
	DisplayName() string
}

type Context struct {
	Writer       http.ResponseWriter
	Request      *http.Request
	CurrentUser  CurrentUser
	ResourceName string
	ResourceID   string
	Config       *Config
}

func (context *Context) DB() *gorm.DB {
	return context.Config.DB
}
