Menu accept 4 parameters.

1. **Name**, the text displayed in menu link
2. **params**, parameters attached to menu link // TODO: Add example of this one's usage
3. **Link**, url the menu link linked to
4. **Items**, type is `[]*Menu`, used for build sub menus

Register menu like

  Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})
