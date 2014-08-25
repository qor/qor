package qor

import "github.com/jinzhu/gorm"

type CurrentUser interface {
	DisplayName() string
}

type Context struct {
	ResourceName string
	ResourceID   string
	Config       *Config
	Roles        []string
}

func (context *Context) DB() *gorm.DB {
	return context.Config.DB
}
