package admin

import "github.com/qor/qor"

type Auth interface {
	GetCurrentUser(*Context) *qor.CurrentUser
	Logged(*Context) bool
	Login(*Context)
	Logout(*Context)
}
