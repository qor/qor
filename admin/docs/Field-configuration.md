### Field configuration
---

Use resource "user" as example to introduce supported field type. [gorm](https://github.com/jinzhu/gorm) as ORM.

    type User struct {
      gorm.Model
      Name      string
      Gender    string
      Description string
      Languages []Language
    }

    type Language struct {
      gorm.Model
      Name string
    }

#### text field

  User name will be displayed as text input in form with label "Full Name"

    user.Meta(&admin.Meta{Name: "Name", Label: "Full Name"})

#### select one

  User gender will be displayed as select in form with options "M", "F", "U"

    user.Meta(&admin.Meta{Name: "Gender", Type: "select_one", Collection: []string{"M", "F", "U"}})

#### select many

  User languages will be displayed as select(multiple selectable) in form with all languages in database. `Collection` can be a Array or a Function

    user.Meta(&admin.Meta{Name: "Languages", Type: "select_many",
      Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
        if languages := []Language{}; !context.GetDB().Find(&languages).RecordNotFound() {
          for _, language := range languages {
            results = append(results, []string{fmt.Sprintf("%v", language.ID), language.Name})
          }
        }
        return
      },
    })

#### rich editor

  User description will be displayed as rich editor in form.

    user.Meta(&admin.Meta{Name: "Description", Type: "rich_editor", Resource: Admin.NewResource(&admin.AssetManager{})}})

  `Resource: Admin.NewResource(&admin.AssetManager{})` here means enable image upload feature for this rich editor.

#### media upload

  resource image will have a file uploader in form by this, Please visit [media library]() for more detail.

    type Image struct {
      File media_library.FileSystem
    }

    image.Meta(&admin.Meta{Name: "File"})

### Additional features
---

#### permission control

  "translator" can fully control language's name but "user" could only read language's name. For more usage please visit [Roles]()

    language.Meta(&admin.Meta{Name: "name", Permission: roles.Allow(role.CRUD, "translator").Allow(role.Read, "user")})

#### nested resources

  A languages section will appears in user form as nested resource

    language := Admin.AddResource(&Language{})
    language.Meta(&admin.Meta{Name: "Name"})
    user.Meta(&admin.Meta{Name: "Languages", Resource: language})

#### alias

// TODO: this option need a appropriate name
Point to db field name when it is different with the name used in API. For example, if db field name is "name" but wants to use "fullname" in API. you need define {Name: "fullname", Alias: "name"}, **Alias point to db field name**.
