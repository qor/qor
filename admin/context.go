package admin

import (
	"net/http"

	"github.com/qor/qor"
)

type Context struct {
	*qor.Context
	Admin        *Admin
	ResourceName string
	Writer       http.ResponseWriter
}
