Resource:

    order = Resource.New(Order, "orders")

    order.Attrs().Index("order_id", "status", "amount")
    order.Attrs().New("order_id", "status", "amount")
    order.Attrs().Edit("order_id", "status", "amount")
    order.Attrs().Show("order_id", "status", "amount")

    order.Meta().Register(qor.Meta{Name: "username", Type: "select", Label: "hello", Value: "", Collection: ""})
    order.Meta().Register(qor.Meta{Name: "credit_card", Resource: creditcard})
    qor.Meta{Name: "credit_card", Resource: creditcard, Permission: rule.Permission}

    order.Search().Name("Name").Register(func() {} (Collection)).Suggestion(func() {})

    order.DefaultScope(func (d *gorm.DB, App) *gorm.DB {
      return d.Where("pay_mode_sign = ?", "C")
    })

    order.Filter().Group("Name").Scope("Cool", func (d *gorm.DB, App) *gorm.DB {
      return d.Where("pay_mode_sign = ?", "C")
    })
    order.Action().Register("name", func() {}).If(func() {})
    order.Download().Register("name", Downloader())

Publish:

    Find RelationShip, Publish

Rule:

    READ, WRITE, RDWR, CREATE, DELETE, ALL
    Allow(ALL, "admin", "dev").Deny(CREATE, "admin")
    type Permission struct {}
    HasPermission(CREATE, App)

    rule.Define("admin", function (App) bool {})

Worker:

    Worker.New("name", resource).Handle(func() {})

Exchanger:

    Exchange.New("products", resource)

Admin: (TBD)

    admin = Admin.New()
    admin.UseResource(order)
    admin.AddToMux("/admin", mux)

Api: (TBD)

    api = Api.New()
    api.UseResources(order)

Mail: (TBD)
