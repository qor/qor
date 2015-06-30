package main

import (
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

type Auth struct{}

func (Auth) LoginURL(c *admin.Context) string {
	return "/login"
}

func (Auth) LogoutURL(c *admin.Context) string {
	return "/logout"
}

func (Auth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	if userid, err := c.Request.Cookie("userid"); err == nil {
		var user User
		if !db.First(&user, "id = ?", userid.Value).RecordNotFound() {
			return &user
		}
	}
	return nil
}
