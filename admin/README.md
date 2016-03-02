## Qor Service

Instantly create a beautiful, cross platform, configurable Admin Interface and API for managing your data in minutes.

[![GoDoc](https://godoc.org/github.com/qor/log?status.svg)](https://godoc.org/github.com/qor/log)

## Features

- Admin Interface for managing data
- JSON API
- Handle Associations
- Search and filtering
- Actions/Batch Actions
- Authentication
- Extendable

## Quick Start

```go
package main

import (
    "net/http"

    "github.com/jinzhu/gorm"
    _ "github.com/mattn/go-sqlite3"
    "github.com/qor/qor"
    "github.com/qor/service"
)

// Create a GORM-backend model
type User struct {
  gorm.Model
  Name string
}

// Create another GORM-backend model
type Product struct {
  gorm.Model
  Name        string
  Description string
}

func main() {
  DB, _ := gorm.Open("sqlite3", "demo.db")
  DB.AutoMigrate(&User{}, &Product{})

  Admin := service.New(&qor.Config{DB: &DB})

  // Create resources from GORM-backend model
  Admin.AddResource(&User{})
  Admin.AddResource(&Product{})

  mux := http.NewServeMux()
  Admin.MountTo("/admin", mux)

  fmt.Println("Listening on: 9000")
  http.ListenAndServe(":9000", mux)
}
```

`go run main.go` and visit `localhost:9000/admin` to see the result !

## General Setting

### Site Name

### Dashboard

### Authentication

### Menu

### Internationalization

## Working with Resource

### Customizing CURD pages

### Search

### Actions

### Customizing the Form

## Extendable

### Customize Views

// view paths

### Register route

### Plugins

```go
router := Admin.GetRouter()

router.Get("/path", func(context *admin.Context) {
    // do something here
})

router.Post("/path", func(context *admin.Context) {
    // do something here
})

router.Put("/path", func(context *admin.Context) {
    // do something here
})

router.Delete("/path", func(context *admin.Context) {
    // do something here
})

// naming route
router.Get("/path/:name", func(context *admin.Context) {
    context.Request.URL.Query().Get(":name")
})

// regexp support
router.Get("/path/:name[world]", func(context *admin.Context) { // "/hello/world"
    context.Request.URL.Query().Get(":name")
})

router.Get("/path/:name[\\d+]", func(context *admin.Context) { // "/hello/123"
    context.Request.URL.Query().Get(":name")
})
```

## Live DEMO

* Live Demo [http://demo.getqor.com/admin](http://demo.getqor.com/admin)
* Source Code of Live Demo [https://github.com/qor/qor-example](https://github.com/qor/qor-example)

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
