Resource:

    order.Search("Name", func (d *gorm.DB, App) *gorm.DB {
      return d.Where("pay_mode_sign = ?", "C")
    }) //.Suggestion(func() {})

    order.Filter("name")
    order.Filter("amount")
    order.Filter(qor.Meta{Name: "name", Label: "Name", Collection: func() []string{ return ["12", "23"]}})

    order.Scope().Group("Name").Register("Cool", func (db *gorm.DB, context *qor.Context) *gorm.DB {
      return d.Where("pay_mode_sign = ?", "C")
    })

    order.BulkEdit().UpdateAttrs("name", "md_week", "gender", "categories")
    order.BulkEdit().Register("name", func(gorm.DB, App) {
      "xxx"
    })

    order.Action().Register("name", func() {db, App} {}).If(func(interface{}, App) {} bool)
    order.Download().Register("name", Downloader())

Publish:

    create data in production -> save in draft db first. (always use the id from draft db)
    Review before publish, diff in popup

    type Product struct {
       Title     string
       ColorCode string
       Price     float64
       Ext       string
       PublishAt time.Time
       Image     MediaLibrary `media_library:"path:/system/:table_name/:id/:filename;"`
    }

    Product{Title: "product A", Image: os.Open("xxxx")}
    db.Save(&product)

    db, err := publish.Open("sqlite", "/tmp/qor.db")
    user := db.NewResource(&Product{})
    user.InstantPublishAttrs("title", "color_code", "price", "colorA", "colorB")
    user.IgnoredAttrs("ext")

    /system_draft/products/xxx.png
    /system/products/xxx.png

    publish.GetDependencies(objects...)
    for _, object := range objects {
      get relations
      get many to many relations
    }
    publish.Publish(struct...) // insert into products_draft (name, color_id) select name, color_id from products;

Mail: (TBD)

Form:

    qor.RegisterMetaType("name", qor.Meta{Setter: xxx, Template: xxx, Value: xxx, Collection: xxx, InputHTML: xxx})

L10n

    locale = l18n.Locale("zh-CN")
    locale.Scope("scope").T("key")
    locale.Params(map[string]string).T("missing", "default value", "another default")

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

Layout:

    render_layout "xxx"

    type Action struct {
      Name string
    }
    layout.RegisterAction(name string, fc func)

    type Widget struct {
      Name          string
      Attrs         string
      RequiredAttrs string
      Template      string
    }

    button := Resource{Name: "Button"}

    slide := Resource{Name: "Slide", Master: false}
    slide.RegisterMeta(qor.Meta{Name: "link", Type: "string"})
    slide.RegisterMeta(qor.Meta{Name: "image", Type: "media"})
    slide.RegisterMeta(qor.Meta{Name: "button", Resource:  button})

    slides := Resource{Name: "Slides"}
    slides.RegisterMeta(qor.Meta{Name: "slides", Resource: slide})
    slides.RegisterMeta(qor.Meta{Name: "slide_menu", Value: func{}, Resource: slide_menu})

    type Layout struct {
      Name       string
      WidgetName string
      Style      string
      Value      string
    }

    layout.Render(name)

StateMachine
  type StateChangeLog struct {
    Id         uint64
    ReferTable string
    ReferId    string
    State      string
    Note       string
    CreatedAt  time.Time
    UpdatedAt  time.Time
    DeletedAt  time.Time
  }

  type StateMachine struct {
    State string
    StateChangeLogs []StateChangeLog
  }

  type Order struct {
    StateMachine // SetState
  }

  orderState := state.New(&Order{})
  orderState.New("finish").Before().After().Do().From("ready").To("paid")

  orderState.To("finish", &order)
    order.SetState("finish")
    order.NewStateLog("finish", tableName, Id, notes)

Action

    type Action struct {
      Name string "update_name"
      Metas []string

      Handle func(scope gorm.DB, context qor.context) error
      Single bool
    }

    order := admin.NewResource(Order{})
    order.Action(action)

    /admin/order/action
    /admin/order/action/confirm?ids=[1,2,3]
    /admin/order/action/confirm?ids=[1]
