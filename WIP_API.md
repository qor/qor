Resource:

    order.Search().Name("Name").Register(func (d *gorm.DB, App) *gorm.DB {
      return d.Where("pay_mode_sign = ?", "C")
    }) //.Suggestion(func() {})
    order.Filter()

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

Worker:

    Worker.New("name", resource).Handle(func() {})

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
