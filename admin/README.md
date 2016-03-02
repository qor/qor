## Qor Service

Instantly create a beautiful, cross platform, configurable Admin Interface and API for managing your data in minutes.

[![GoDoc](https://godoc.org/github.com/qor/log?status.svg)](https://godoc.org/github.com/qor/log)

## Features

- Admin Interface for managing data
- JSON API
- Handle Associations
- Search and filtering
- Actions/Batch Actions
- Authentication & Permission
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

  // Initalize
  Admin := service.New(&qor.Config{DB: &DB})

  // Create resources from GORM-backend model
  Admin.AddResource(&User{})
  Admin.AddResource(&Product{})

  // Register route
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
```

Here is an example

```go
type Auth struct{}

func (Auth) LoginURL(c *admin.Context) string {
  return "/login"
}

func (Auth) LogoutURL(*Context) string
  return "/logout"
}

func (Auth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
  if userid, err := c.Request.Cookie("userid"); err == nil {
    var user User
    if !DB.First(&user, "id = ?", userid.Value).RecordNotFound() {
      return &user
    }
  }
  return nil
}

func (u User) DisplayName() string {
  return u.Name
}

// Register Auth for Qor admin
Admin.SetAuth(&Auth{})
```

### Menu

#### Register Menu

```go
Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})

// Register nested menu
Admin.AddMenu(&admin.Menu{Name: "menu", Link: "/link", Ancestors: []string{"Dashboard"}})
```

#### Add resource to menu

```go
Admin.AddResource(&User{})

Admin.AddResource(&Product{}, &service.Config{Menu: []string{"Product Management"}})
Admin.AddResource(&Color{}, &service.Config{Menu: []string{"Product Management"}})
Admin.AddResource(&Size{}, &service.Config{Menu: []string{"Product Management"}})

Admin.AddResource(&Order{}, &service.Config{Menu: []string{"Order Management"}})
```

If you don't want resource to be displayed in navigation, pass Invisible option in like this

```go
Admin.AddResource(&User{}, &admin.Config{Invisible: true})
```

### Internationalization

To translate your admin interface to a new language, you could use the `i18n` [https://github.com/qor/i18n](https://github.com/qor/i18n)

## Working with Resource

Every Qor Resource need a [GORM-backend](https://github.com/jinzhu/gorm) model, so you need to define the model first, after that you could create qor resource with `Admin.AddResource(&Product{})`

After add resource to admin, qor service will generate the admin interface to manage the resource, and also it will generate a JSON based RESTFul API.

So for above example, you could visit `localhost:9000/admin/products` to manage `Product` in HTML interface, or use the RESTFul JSON api `localhost:9000/admin/products.json` to any CRUD work

### Customizing CURD pages

```go
// Set attributes will be shown in the index page

// show given attributes in the index page
order.IndexAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
// show all attributes except `State` in the index page
order.IndexAttrs("-State")

// Set attributes will be shown in the new page
order.NewAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
// show all attributes except `State` in the new page
order.NewAttrs("-State")
// Structure the new form to make it tidy and clean with `Section`
product.NewAttrs(
  &admin.Section{
		Title: "Basic Information",
		Rows: [][]string{
			{"Name"},
			{"Code", "Price"},
		}
  },
  &admin.Section{
		Title: "Organization",
		Rows: [][]string{
			{"Category", "Collections", "MadeCountry"},
    }
  },
  "Description",
  "ColorVariations",
}

// Set attributes will be shown for the edit page, similiar with new page
order.EditAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")

// Set attributes will be shown for the show page, similiar with new page
// If ShowAttrs haven't been configured, there will be no show page generated, by will show the edit from instead
order.ShowAttrs("User", "PaymentAmount", "ShippedAt", "CancelledAt", "State", "ShippingAddress")
```

### Search

Use `SearchAttrs` to set search attributes, when search resources, will use those columns to search, it aslo support search nested relations

```go
// Search products with its name, code, category's name, brand's name
product.SearchAttrs("Name", "Code", "Category.Name", "Brand.Name")
```

### Scopes

Define a scope that show all active users

```go
user.Scope(&admin.Scope{Name: "Active", Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
  return db.Where("active = ?", true)
}})
```

Group Scopes

```go
order.Scope(&admin.Scope{Name: "Paid", Group: "State", Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
  return db.Where("state = ?", "paid")
}})

order.Scope(&admin.Scope{Name: "Shipped", Group: "State", Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
  return db.Where("state = ?", "shipped")
}})
```

### Actions

Qor Service provide four kinds of actions:

* Bulk actions
* Edit form action
* Show page action
* Menu item action

They all using qor reosurce's `Action` method to register themselves, and using `Visible` to contol where to show them

```go
product.Action(&admin.Action{
	Name: "enable",
	Handle: func(actionArgument *admin.ActionArgument) error {
    // `FindSelectedRecords` => return selected record in bulk action mode, return current record in other mode
		for _, record := range actionArgument.FindSelectedRecords() {
			actionArgument.Context.DB.Model(record.(*models.Product)).Update("disabled", false)
		}
		return nil
	},
	Visibles: []string{"index", "edit", "show", "menu_item"},
})

// Register Actions need user's input
order.Action(&admin.Action{
  Name: "Ship",
  Handle: func(argument *admin.ActionArgument) error {
    trackingNumberArgument := argument.Argument.(*trackingNumberArgument)
    for _, record := range argument.FindSelectedRecords() {
      argument.Context.GetDB().Model(record).UpdateColumn("tracking_number", trackingNumberArgument.TrackingNumber)
    }
    return nil
  },
  Resource: Admin.NewResource(&trackingNumberArgument{}),
  Visibles: []string{"show", "menu_item"},
})

// the ship action's argument
type trackingNumberArgument struct {
  TrackingNumber string
}
```

### Customizing the Form

When Qor Service generating the admin interface, if will get your resource's data type and relations, based on those information, create `Meta` for your registered resource.

Then when render the index/show/new/edit pages, will generate it based on the resource's Meta definition.

If you want to change those defaults, you need to change the resource's `Meta` definition.

```go
// Qor has defined many meta's types by default, including `string`, `password`, `date`, `rich_editor`, `select_one` and so on
user.Meta(&service.Meta{Name: "Password", Type: "password"}) // change resource user's Password field's type from `string` to `password`

// Change resource user's Gender field's to be a select, options are `M`, `F`
user.Meta(&service.Meta{Name: "Gender", Type: "select_one", Collection: []string{"M", "F"}})
```

### Permission

Qor Service is using [https://github.com/qor/roles](https://github.com/qor/roles) for permission management, refer it for how to define roles, permissions

```go
// CURD permission for admin users, deny create permission for manager
user := Admin.AddResource(&User{}, &service.Config{Permission: roles.Allow(roles.CRUD, "admin").Deny(roles.Create, "manager")})

// For user's Email field, allow CURD for admin users, deny update for manager
user.Meta(&service.Meta{Name: "Email", Permission: roles.Allow(roles.CRUD, "admin").Deny(roles.Create, "manager")})
```

### RESTFul JSON API

The RESTFul JSON shared same configuration with your admin interface, including actions, permission, so after you configured your admin interface, you will get an API for free!

## Extendable

#### Configure Qor Resources Automatically

If your model has defined below two methods, it will be call when registing

```go
func ConfigureQorResourceBeforeInitialize(resource) {
  // resource.(*service.Resource)
}

func ConfigureQorResource(resource) {
  // resource.(*service.Resource)
}
```

#### Configure Qor Meta Automatically

If your field's type has defined below two methods, it will be call when registing

```go
func ConfigureQorMetaBeforeInitialize(meta) {
  // resource.(*service.Meta)
}

func ConfigureMetaInterface(meta) {
  // resource.(*service.Meta)
}
```

#### Use Theme

Use theme `fancy` for products, when visting product's CRUD pages, will load `assets/javascripts/fancy.js` and `assets/stylesheets/fancy.css` from [QOR view paths](#customize-views)

```go
product.UseTheme("fancy")
```

#### Customize Views

When rendering pages, qor will look up templates from qor view paths, and use them to render the page, qor has registered `{current path}/app/views/qor` for your to allow you extend your application from there. If you want to customize your views from other places, you could register new path with `service.RegisterViewPath`

Customize Views Rules:

* To overwrite a template, create a file under `{current path}/app/views/qor` with same name
* To overwrite templates for one resource, put templates with same name to `{qor view paths}/{resource param}`, for example `{current path}/app/views/qor/products/index.tmpl`
* To overwrite templates for resources using theme, put templates with same name to `{qor view paths}/themes/{theme name}`

#### Register routes

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

#### Plugins

There are couple of plugins created for qor already, you could find some of them from here [https://github.com/qor](https://github.com/qor), visit them to learn more how to extend qor

## Live DEMO

* Live Demo [http://demo.getqor.com/admin](http://demo.getqor.com/admin)
* Source Code of Live Demo [https://github.com/qor/qor-example](https://github.com/qor/qor-example)

## Q & A

* How to integrate with beego

```go
mux := http.NewServeMux()
Admin.MountTo("/admin", mux)

beego.Handler("/admin/*", mux)
beego.Run()
```

* How to integrate with Gin

```go
mux := http.NewServeMux()
Admin.MountTo("/admin", mux)

r := gin.Default()
r.Any("/admin/*w", gin.WrapH(mux))
r.Run()
```

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
