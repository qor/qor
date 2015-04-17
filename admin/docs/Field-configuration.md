### Meta config
---

  Meta accepts below parameters

  - **Name**, name of the attribute
  - **Alias**(TODO: this should switch name with Name), point to db field name when field name is different with the name used outside(API client). use Alias to point to field name. For example, if field name is "name" but wants to use "fullname" in Admin. you need define {Name: "fullname", Alias: "name"}, Alias point to field name.
  - **Label**, field label text
  - **Type**, define how to display this meta, see below list for detail
  - **Resource**, set nested resources, the nested resources's meta will be displayed in parent resource form. You can nest infinity resources you want.
  - **Collection**, data set of select one and select many meta.
  - **Permission**, control user's permission on current meta.

#### text field

  Set the name of the field and label(optional), "Type" is default as text input.

    user.Meta(&admin.Meta{Name: "name", Label: "Full Name"})

#### select one

  Set "Type" as "select_one" then set data source by parameter "Collection"

    user.Meta(&admin.Meta{Name: "gender", Type: "select_one", Collection: []string{"M", "F", "U"}})

#### select many

  Set "Type" as "select_many", "Collection" also support function

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

  Set "Name" and "Type" as "rich_editor"

    user.Meta(&admin.Meta{Name: "description", Type: "rich_editor"})

#### media upload

  // placeholder