Resource:

    order = Resource.New(Order, "orders")

    order.Attrs().Index("order_id", "status", "amount")
    order.Attrs().New("order_id", "status", "amount")
    order.Attrs().Edit("order_id", "status", "amount")
    order.Attrs().Show("order_id", "status", "amount")

    order.Meta().Role("admin").Register(qor.Meta{Name: "username", Type: "select", Label: "hello", Value: "", Collection: ""})
    order.Meta().Register(qor.Meta{Name: "credit_card", Resource: creditcard})

    order.Search().Name("Name").Register(func() {} (Collection)).Suggestion(func() {})
    order.Filter().Group("Name").Register("Cool", func() {})
    order.Role("admin").DefaultScope(func() {})
    order.Action().Register("name", func() {}).If(func() {})
    order.Download().Register("name", Downloader())

Publish:

    Find RelationShip, Publish

Rule:

    Rule.New("admin").Allow(func() {}).Deny(func() {})

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
