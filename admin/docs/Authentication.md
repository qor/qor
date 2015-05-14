Authentication

Auth must implement 4 functions

1. Login(*Context)
2. Logout(*Context)
3. GetCurrentUser(*Context) qor.CurrentUser
4. func (user User) DisplayName() string

Example:

    type Auth struct{}

    func (Auth) Login(c *admin.Context) {
      http.Redirect(c.Writer, c.Request, "/login", http.StatusSeeOther)
    }

    func (Auth) Logout(c *admin.Context) {
      http.Redirect(c.Writer, c.Request, "/logout", http.StatusSeeOther)
    }

    func (Auth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
      if userid, err := c.Request.Cookie("userid"); err == nil {
        var user User
        if !DB.First(&user, "id = ?", userid.Value).RecordNotFound() {
          return &user
        }
      }
      return nil
    }

    func (u User) DisplayName() string {
      return u.Name
    }

    // Register customized Auth in Qor admin
    Admin.SetAuth(&Auth{})
