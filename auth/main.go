package auth

import "github.com/qor/qor"

type Auth interface {
	GetCurrentUser(*qor.Context) qor.CurrentUser
	Logged(*qor.Context) bool
	Login(*qor.Context)
	Logout(*qor.Context)
}
