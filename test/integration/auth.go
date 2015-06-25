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
	var currentUser User

	if !DB.Where("name = ?", "currentUser").First(&currentUser).RecordNotFound() {
		return &currentUser
	}

	return nil
}
