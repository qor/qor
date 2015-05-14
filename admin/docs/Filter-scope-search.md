## Search

## Filter

## Scope

Define a scope that show all active users, `Name` will be the link text displayed in front-end.

    user.Scope(&admin.Scope{Name: "Active", Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
      return db.Where("active = ?", true)
    }})
