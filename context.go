package qor

import "net/http"

type CurrentUser interface {
	DisplayName() string
}

type Context struct {
	Writer       http.ResponseWriter
	Request      *http.Request
	CurrentUser  CurrentUser
	ResourceName string
	ResourceID   string
}
