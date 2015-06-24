package admin

import "github.com/qor/qor"

type Auth interface {
	GetCurrentUser(*Context) qor.CurrentUser
	LoginURL(*Context) string
	LogoutURL(*Context) string
}
