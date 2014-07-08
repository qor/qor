package qor

import "net/http"

type CurrentUser interface {
	DisplayName() string
}

type App struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	CurrentUser    CurrentUser
}
