QOR admin support fully customizable & infinitely nested menu

### Register menu in QOR admin

    Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})

### Register nested menu in QOR admin
Pass ancestor menus name in `Ancestors` to create nested menu.

    Admin.AddMenu(&admin.Menu{Name: "menu", Link: "/link", Ancestors: []string{"Dashboard"}})

In this example, "menu" will appears under "Dashboard".

If an ancestor menu name doesn't included in existing menus, all menus after this one will be treated as not exists. For example

    Admin.AddMenu(&admin.Menu{Name: "menu1", Link: "/link1", Ancestors: []string{"Non-exists", "Dashboard"}})

"menu1" will appears under "Non-exists" > "Dashboard". "Non-exists" and "Dashboard" are have **no href**.

### Add resource to menu
By below code, A "User" menu will be registered in navigation.

    Admin.AddResource(&User{})

If you don't want resource to be displayed in navigation, pass `Invisible` option in like this

    Admin.AddResource(&User{}, &admin.Config{Invisible: true})

To add resource under nested menu, basically same as register nested menu, pass ancestor menus name like this

    Admin.AddResource(&User{}, &admin.Config{Menu: []string{"Dashboard", "User Management"}})

"User" menu will appears under "Dashboard" > "User Management"
