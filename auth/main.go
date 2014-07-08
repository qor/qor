package auth

import "github.com/qor/qor"

type Auth interface {
	GetCurrentUser(*qor.App) qor.CurrentUser
	Logged(*qor.App) bool
	Login(*qor.App)
	Logout(*qor.App)
}
