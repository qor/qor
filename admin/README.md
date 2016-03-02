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
  // amount to /admin, so visit `/admin` to view the admin interface
  Admin.MountTo("/admin", mux)

  fmt.Println("Listening on: 9000")
  http.ListenAndServe(":9000", mux)
}
```

`go run main.go` and visit `localhost:9000/admin` to see the result !

## General Setting

### Site Name

Use `SetSiteName` to set admin's HTML title, it will not only set title, but also will auto load javascripts, stylesheets files according to your name, so it could be used to customize the admin interface.

For example, if you named it as `Qor Demo`, admin will look up `qor_demo.js`, `qor_demo.css` in [QOR view paths](#customize-views), and load them if found

```go
Admin.SetSiteName("Qor DEMO")
```

### Dashboard

Qor provide a default dashboard page with some dummary text, if you want to overwrite it, you could create a file named as `dashboard.tmpl` in [QOR view paths](#customize-views), Qor will load it as golang templates when render dashboard

If you want to disable the dashboard, you could redirect the dashboard page to some other page, for example:

```go
Admin.GetRouter().Get("/", func(c *admin.Context) {
  http.Redirect(c.Writer, c.Request, "/admin/clients", http.StatusSeeOther)
})
```

### Authentication

Qor provides pretty flexable authorization solution, with it, you could integrate admin with your current authorization method.

What you need to do is implement below `Auth` interface, and set it to the admin

```go
type Auth interface {
	GetCurrentUser(*Context) qor.CurrentUser // get current user, if don't have permission, then return nil
	LoginURL(*Context) string // get login url, if don't have permission, will redirect to this url
	LogoutURL(*Context) string // get logout url, if click logout link from admin interface, will visit this page
}

// Once struct `auth` has implemented above interface, use `SetAuth` set it to admin
Admin.SetAuth(auth{})
```

### Menu

Registered Resources will be shown in menu in order, use `admin.Config` to group them

```go
Admin.AddResource(&Product{}, &admin.Config{Menu: []string{"Product Management"}})
Admin.AddResource(&Color{}, &admin.Config{Menu: []string{"Product Management"}})
Admin.AddResource(&Size{}, &admin.Config{Menu: []string{"Product Management"}})

Admin.AddResource(&Order{}, &admin.Config{Menu: []string{"Order Management"}})
```

### Internationalization

To translate your admin interface to a new language, you could use the `i18n` [https://github.com/qor/i18n](https://github.com/qor/i18n)

## Working with Resource

### Customizing CURD pages

### Search

### Actions

### Customizing the Form

## JSON API

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
