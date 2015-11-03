## Introduction

Qor admin provide easy-to-use interface for data management.

## Quick example

Use 35 lines of code to setup & run Qor admin.

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

You can view [qor example](https://github.com/qor/qor/tree/master/example) for a more detailed configuration example.

## Features

- CRUD of any resource
- Search and filtering
- Authentication
- Authorization(detail)
- Custom actions
- Customizable view
- Rich editor
- Image crop
- Integrate-able with [Publish](https://github.com/qor/qor/tree/master/publish)
- Integrate-able with [l10n](https://github.com/qor/qor/tree/master/l10n)
- JSON API supported
- Extendable

## Search Center

It provide feature set that let End User to be able to quickly find any resources Developer defnited, like Product, Order, User, etc. Developer define what field of information user can search on.

### Developer side usage

### End User side usage


## Documentation

https://github.com/qor/qor/wiki
