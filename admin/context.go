package admin

import (
	"github.com/qor/qor"

	"net/http"
)

type Context struct {
	*qor.Context
	ResourceName string
	Writer       http.ResponseWriter
}
