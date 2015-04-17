### Database
---

  Only accept https://github.com/jinzhu/gorm. "mysql", "sqlite", "postgresql" are supported by gorm

  Initialize db by

      DB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPwd, dbName))

### Initialize Admin
---

      config := qor.Config{DB: &DB}
      Admin := admin.New(&config)

### Add resource
---

  AddResource function accept 2 parameters

  1. resource
  2. menu that link to this resource

      user := Admin.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})

### Route
---

  Initialize a mux and mount Admin to it.

    mux := http.NewServeMux()
    Admin.MountTo("/admin", mux)

### Server
---

  Listen and serve and you're done !

    http.ListenAndServe(":9000", mux)
