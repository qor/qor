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
}

func (context *Context) DB() *gorm.DB {
	return context.Config.DB
}
