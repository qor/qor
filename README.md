Resource:

    order = Resource.New(Order, "orders")

    order.Attrs().Index("order_id", "status", "amount")
    order.Attrs().New("order_id", "status", "amount")
    order.Attrs().Edit("order_id", "status", "amount")
    order.Attrs().Show("order_id", "status", "amount")

    order.Meta().Role("admin").Register(qor.Meta{Name: "username", Type: "select", Label: "hello", Value: "", Collection: ""})
    order.Meta().Register(qor.Meta{Name: "credit_card", Resource: creditcard})

    order.Search().Name("Name").Register(function() {} (Collection)).Suggestion(function() {})
    order.Filter().Group("Name").Register("Cool", function() {})
    order.Role("admin").DefaultScope(function() {})
    order.Action().Register("name", function() {}).If(function() {})
    order.Download().Register("name", Downloader())

Publish:

    Find RelationShip, Publish

Rule:

    Rule.New("admin").Allow(function() {}).Deny(function() {})

Worker:

    Worker.New("name", resource).Handle(function() {})

Exchanger:

    Exchange.New("products", resource)

Admin: (TBD)

    admin = Admin.New()
    admin.Route().Get("/", function(r, w) {})
    admin.Route().Post("/", function(r, w) {})
    admin.UseResource(order)
    admin.Menu().New().UseResource(order)
    admin.Menu().Edit().UseResource(order)
    admin.AddToMux("/admin", mux)

Api: (TBD)

    api = Api.New()
    api.UseResources(order)

Mail: (TBD)
