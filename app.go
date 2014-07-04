package qor

import "net/http"

type App struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Params         interface{}
	Resource       interface{}
}
