Authentication

Auth must implement 3 functions

1. Login(*Context)
2. Logout(*Context)
3. GetCurrentUser(*Context) qor.CurrentUser

Then use

  Admin.SetAuth(&Auth{})
