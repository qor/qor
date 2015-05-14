Setup usable Admin step by step.

### 1. Database
Use [gorm](https://github.com/jinzhu/gorm) as ORM for example.

      DB, _ := gorm.Open("sqlite3", "demo.db")
      DB.AutoMigrate(&User{}, &Product{})

### 2. Initialize admin and register resources to it.

      Admin := admin.New(&qor.Config{DB: &DB})
      Admin.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})
      Admin.AddResource(&Product{}, &admin.Config{Menu: []string{"Product Management"}})

### 3. Authentication
[Authentication]()

### 4. Register route and start server

      mux := http.NewServeMux()
      Admin.MountTo("/admin", mux)
      http.ListenAndServe(":9000", mux)

Now a run-able admin has been created. You can start adding [Navigation](), [Roles](), Configure [Field](), [Filter](), [Search]() and [Localization]().
