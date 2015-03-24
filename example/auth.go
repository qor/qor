package main

import (
	"net/http"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

type Auth struct{}

func (Auth) Login(c *admin.Context) {
	http.Redirect(c.Writer, c.Request, "/login", http.StatusSeeOther)
}

func (Auth) Logout(c *admin.Context) {
	http.Redirect(c.Writer, c.Request, "/logout", http.StatusSeeOther)
}

var loggedUserId int

func (Auth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	var user User
	if DB.First(&user, "id = ?", loggedUserId).RecordNotFound() {
		return nil
	} else {
		return &user
	}
}
