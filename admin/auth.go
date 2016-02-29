package admin

import "github.com/qor/qor"

// Auth is an auth interface that used to qor admin
// If you want to implement an authorization gateway for admin interface, you could implement this interface, and set it to the admin with `admin.SetAuth(auth)`
type Auth interface {
	GetCurrentUser(*Context) qor.CurrentUser
	LoginURL(*Context) string
	LogoutURL(*Context) string
}
