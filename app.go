package qor

import "net/http"

type CurrentUser struct {
	Name  string
	Roles []string
}

func (user *CurrentUser) Logout(App) {
}

func (user *CurrentUser) Login(App) {
}

type App struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	CurrentUser    CurrentUser
}
