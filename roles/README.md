### Roles

#### Permissions

Roles provide 5 permissions:

    `Read`, `Update`, `Create`, `Delete`, `CRUD`

#### Register role

    roles.Register("admin", func(req *http.Request, currentUser qor.CurrentUser) bool {
      return currentUser.(*User).Role == "admin"
    })

#### Get matched roles of user

    roles.MatchedRoles(httpRequest, user)

#### Allow

Allow permissions on role, white list pattern

    permission := roles.Allow(roles.Read, "admin")
    permission.HasPermission(roles.Read, "admin") // true
    permission.HasPermission(roles.Create, "admin") // false

You can allow permission in chain

    permission := roles.Allow(roles.Read, "admin").Allow(roles.CRUD, "manager")
    permission.HasPermission(roles.Read, "admin") // true
    permission.HasPermission(roles.Read, "manager") // true
    permission.HasPermission(roles.Create, "manager") // true
    permission.HasPermission(roles.Update, "manager") // true
    permission.HasPermission(roles.Delete, "manager") // true

#### Deny

Deny permissions on role, black list pattern

    permission := roles.Deny(roles.Read, "admin")
    permission.HasPermission(roles.Read, "admin") // false
    permission.HasPermission(roles.Create, "admin") // true

To deny permission in chain

    permission := roles.Deny(roles.Read, "admin").Deny(roles.Delete, "user")
    permission.HasPermission(roles.Read, "admin") // false
    permission.HasPermission(roles.Delete, "user") // false
