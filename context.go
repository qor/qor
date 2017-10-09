package qor

import (
	"net/http"

	"github.com/jinzhu/gorm"
)

// CurrentUser is an interface, which is used for qor admin to get current logged user
type CurrentUser interface {
	DisplayName() string
}

// Context qor context, which is used for many qor components, used to share information between them
type Context struct {
	CurrentUser CurrentUser
	Request     *http.Request
	Writer      http.ResponseWriter
	Roles       []string
	ResourceID  string
	Errors

	DB     *gorm.DB
	Config *Config
}

// Clone clone current context
func (context *Context) Clone() *Context {
	var clone = *context
	return &clone
}

// GetDB get db from current context
func (context *Context) GetDB() *gorm.DB {
	if context.DB != nil {
		return context.DB
	}
	return context.Config.DB
}

// SetDB set db into current context
func (context *Context) SetDB(db *gorm.DB) {
	context.DB = db
}
