package admin

import "github.com/qor/qor"

type Auth interface {
	GetCurrentUser(*Context) qor.CurrentUser
	Login(*Context)
	Logout(*Context)
}
