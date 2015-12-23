## Introduction

Qor admin provide easy-to-use interface for data management.

## Features

- CRUD of any resource
- JSON API supported
- Authentication
- Search and filtering
- Custom actions
- Customizable view
- Extendable

## Quick Example

```go
package main

import (
    "net/http"

    "github.com/jinzhu/gorm"
    _ "github.com/mattn/go-sqlite3"
    "github.com/qor/qor"
    "github.com/qor/qor/admin"
)

type User struct {
  gorm.Model
    Name string
}

type Product struct {
  gorm.Model
    Name        string
    Description string
}

func main() {
  DB, _ := gorm.Open("sqlite3", "demo.db")
  DB.AutoMigrate(&User{}, &Product{})

  Admin := admin.New(&qor.Config{DB: &DB})
  Admin.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})
  Admin.AddResource(&Product{}, &admin.Config{Menu: []string{"Product Management"}})

  mux := http.NewServeMux()
  Admin.MountTo("/admin", mux)
  http.ListenAndServe(":9000", mux)
}

// TODO: add screenshot after QOR admin UI finished
`go run main.go` and visit `localhost:9000/admin` to see the result !
```

## Usage

### Register route

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

You can view [qor example](https://github.com/qor/qor-example) for a more detailed configuration example.
