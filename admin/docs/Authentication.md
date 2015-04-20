Authentication

Auth must implement 3 functions

1. Login(*Context)
2. Logout(*Context)
3. GetCurrentUser(*Context) qor.CurrentUser

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

    // Register customized Auth in Qor admin
    Admin.SetAuth(&Auth{})
