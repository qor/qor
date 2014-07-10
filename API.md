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

    Layout:

        views/themes/tis/resources/user/edit.tmpl
        views/themes/tis/edit.tmpl
        views/resources/user/edit.tmpl
        views/themes/default/edit.tmpl

        views/qor/themes/default/layout.tmpl
        views/qor/themes/default/header.tmpl
        views/qor/themes/default/footer.tmpl
        views/qor/dashboard.tmpl {{define content}}
        views/qor/index.tmpl
        views/qor/new.tmpl
        views/qor/edit.tmpl

Api: (TBD)

    api = Api.New()
    api.UseResources(order)

Mail: (TBD)

Form:
    qor.Meta{Name: "username", Type: "text", Label: "hello",
             Value: " ",
             DefaultValue: string/func() string,
             Collection: []string / map[string]string / func() []string / func() map[string]string / []Meta / func() []Meta
             Setter: func (value interface{}) error,
             InputHtml: map[string]string{"alt": "hello"}}

    qor.DefineMeta("name", template)

    <li>
      <label>{{.Label}}</label>
      <input type="{{.Type}}" name="{{.Name}}" value="{{.Value}}"/>
    </li>

    <label for="post_tag_ids">{{.Label}}</label>
    <select id="post_tag_ids" name="post[tag_ids]" multiple="true">
      <option value="1">Ruby</option>
      <option value="6">Rails</option>
      <option value="3">Forms</option>
      <option value="4">Awesome</option>
    </select>

    <li>
      <label for="post_author_id_1">
        <input type="radio" id="post_author_id_1" value="1"> Justin
      </label>
    </li>
    <li>
      <label for="post_author_id_3">
        <input type="radio" id="post_author_id_3" value="3"> Kate
      </label>
    </li>

L10n

    locale = l18n.Locale("zh-CN")
    locale.Scope("scope").T("key")
    locale.Params(map[string]string).T("missing", "default value", "another default")

Authentication:
   Auth := auth.New()

Localization:

    type Product struct {
      Id int64
      Name string
    }

    type LocalizedProduct struct {
      Product
      ProductId int64
      LanguageCode string
    }
