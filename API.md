Resource:

    order = Resource.New(Order, "orders")

    order.Attrs().Index("order_id", "status", "amount")
    order.Attrs().New("order_id", "status", "amount")
    order.Attrs().Edit("order_id", "status", "amount")
    order.Attrs().Show("order_id", "status", "amount")

    order.Meta().Register(qor.Meta{Name: "username", Type: "select", Label: "hello", Value: "", Collection: "", Setter: ""})
    order.Meta().Register(qor.Meta{Name: "credit_card", Resource: creditcard})
    qor.Meta{Name: "credit_card", Resource: creditcard, Permission: rule.Permission}

    order.Search().Name("Name").Register(func (d *gorm.DB, App) *gorm.DB {
      return d.Where("pay_mode_sign = ?", "C")
    }) //.Suggestion(func() {})
    order.Filter()

    order.DefaultScope(func (d *gorm.DB, App) *gorm.DB {
      return d.Where("pay_mode_sign = ?", "C")
    })
    order.Scope().Group("Name").Register("Cool", func (d *gorm.DB, App) *gorm.DB {
      return d.Where("pay_mode_sign = ?", "C")
    })

    order.BulkEdit().UpdateAttrs("name", "md_week", "gender", "categories")
    order.BulkEdit().Register("name", func(gorm.DB, App) {
      "xxx"
    })

    order.Action().Register("name", func() {db, App} {}).If(func(interface{}, App) {} bool)
    order.Download().Register("name", Downloader())

Rule:

    READ, WRITE, RDWR, CREATE, DELETE, ALL
    Allow(ALL, "admin", "dev").Deny(CREATE, "admin")
    type Permission struct {}
    HasPermission(CREATE, App)

    rule.Register("admin", function (App) bool {})

Exchanger:

    Exchange.New("products", resource)

Publish:

    create data in production -> save in draft db first. (always use the id from draft db)
    Review before publish, diff in popup

Worker:

    Worker.New("name", resource).Handle(func() {})

Admin: (TBD)

    admin = Admin.New()
    admin.UseResource(order)
    admin.AddToMux("/admin", mux)

Api: (TBD)

    api = Api.New()
    api.UseResources(order)

Mail: (TBD)
