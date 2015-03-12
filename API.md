Admin:

    Admin := admin.New(db *gorm.DB)

    order := Admin.AddResource(&order{}, admin.Config{Name: string, Menus: []string, Invisible: bool, Permission:})
    order.IndexAttrs("Id", "Amount", "Email")
    order.Finder()
    order.Deleter()
    order.Meta(&admin.Meta{Name: name, Valuer: func(), Setter: func()}) | Valuer (type, value, meta values), Setter
    order.Scope(&admin.Scope{Name: name, Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {}})
    order.Filter(&admin.Filter{Name: name, Handle: func(string, string, *gorm.DB, *qor.Context) *gorm.DB})
    order.Action(&admin.Action{Name: name, Handle: func(scope *gorm.DB, context *qor.Context) error {}, Inline, Metas})

    Admin.GetResource(name string) *Resource
    Admin.SetAuth(Auth{CurrentUser, Login, Logout})

    router := Admin.GetRouter()
    router.Get("/admin", func())

    Admin.MountTo(prefix string, mux *http.ServeMux)
    Admin.RegisterFuncMap(name string, fc interface{})

Publish:

    Publish := publish.New(db *gorm.DB)
    Publish.AddModel(&Order{}, publish.Config{Permission: permission, IgnoredAttrs: []string, Resource: admin.Resource}) // -> default scope, permission
    Publish.DraftDB()
    Publish.ProductionDB()
    Publish.Publish(records...)

    Admin.AddResource(Publish) -> router

Worker:

    Worker = worker.New()
    Worker.SetQueue()
    job := Worker.AddJob("name", qor.Config{Handle: , OnKill: , Queue: , Permission: })
    job.Meta(admin.Meta{})

    worker.Run(jobId) -> QorJob (file system render MetaValues) -> Worker -> job
    job.Run(QorJob)

    Admin.AddResource(Worker) -> router -> jobs -> metas -> POST -> QorJob (meta values -> file system)

Exchange:

    Exchange := exchange.New()
    Exchange.Meta{exchange.Meta{Name: , Value:, Setter: }}
    Exchange.Import(file, logger, context)
    Exchange.Export(records, writer, logger, context)

L10n:
    Admin.AddResource(l10n.Model(&Product{}), Permission: rules.Allow(roles.CRUD, "global_admin", "locale_admin"))

    current_user {
      current_locale -> locale_admin
    }

    type News struct {
      ID  int
      l10n.Locale
    }

    type Blog struct {
      ID int
      l10n.Locale
    }

    type Comment struct {
      ID int
      l10n.Locale
    }

    var blogs []Blog
    db.Find(&blogs)

    db.Set("l10n:locale", "zh_CN").Find(&blogs)
    db.Set("l10n:locale", "zh_CN").Find(&blogs)
    db.Set("l10n:locale", "zh_CN").Find(&blogs).Related(&comments)

    type Product struct {
      ID              int64
      LangaugeCode    string
      Name            string
      Brand           Brand
      Category        Category
      ColorVariations ColorVariation
    }

    type Brand struct {
      ID           int64
      LangaugeCode string
      Name         string
    }

    type Category struct {
      ID           int64
      Name         string
    }

    type ColorVariation struct {
      ID              int64
      LangaugeCode    string
      Color           string
    }

    localization.Localize(&ZHProduct{}, &GlobalProduct, "ZH")

    db.Model(&product).Related(&Brand()) // SELECT * FROM brands WHERE id = 10 AND language_code = 'ZH' OR language_code IS NULL

    localeDB := db.Set("localization:locale", "ZH" / global)
    localeDB.Set("localization", "localized").Model(&product).Related(&Brand())
    // SELECT * FROM brands WHERE id = 10 AND language_code = 'ZH'

    localeDB.Set("localization", "global").Model(&product).Related(&Brand())
    // SELECT * FROM brands WHERE id = 10 AND language_code IS NULL

    localeDB.Model(&product).Related(&Brand())
    // SELECT * FROM brands WHERE id = 10 AND (language_code = 'ZH' OR language_code IS NULL)

    // Query:  depends on mode (localized, global, mixed)
    // Update:
        Global: sync attrs that need sync
        Locale: can only update localized attrs
    // Delete: locale as condition
        Global: sync to all locales
        Locale: delete with locale
    // Create:
        Global: create global record
        Locale: create with locale

    // Has One
        Query:
          localize: localized mode
          sync: mixed mode
        Create:
          localize: change foreign key
          sync: not possible from locale
        Delete:
          localize: change foreign key
          sync: not possible from locale
    // Has Many
        Query:
          localize: localized mode
          sync: mixed mode
        Create:
          localize: add new records with locale
          sync: not possible from locale
        Delete:
          localize: delete the record with locale
          sync: not possible from locale
    // Belongs To:
        Query:
          localize: localized mode
          sync: mixed mode
        Create:
          localize: add new records with locale
          sync: not possible from locale
        Delete:
          localize: delete the record with locale
          sync: not possible from locale
    // Many To Many
        Query:
          localize: Scope(language_code: locale)
          sync: Same
        Delete:
          localize: Delete(language_code: locale)
          sync: not possible from locale
        Create:
          localize: Assign(language_code: locale)
          sync: not possible from locale

Role:

    roles.Register("admin", func (req *http.Request, currentUser qor.CurrentUser) bool) {
      return currentUser.(*User).Role == "admin"
    })

    roles.Register("admin", func (req *http.Request, currentUser qor.CurrentUser) bool) {
      return req.RemoteAddr == "127.0.0.1"
    })

    roles.Register("translator", func (req *http.Request, currentUser qor.CurrentUser) bool) {
      for _, role := range return currentUser.(*User).Roles {
        if role == "translator" {
          return true
        }
      }
      return false
    })

    order := Admin.AddResource(&order{}, admin.Config{Permission: roles.Allow(role.CRUD, "amdin").Allow(role.Read, "translator")})

    order.Permssion.HasPermission(role.Read, "admin", "translator"...)

i18n

    locale = i18n.Locale("zh-CN")
    locale.Scope("scope").T("key")
    locale.Params(map[string]string).T("missing", "default value", "another default")

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

    orderState.New("finish").Before().After().Do().Enter(Handle, "checkout", "paid").Enter(Handle, "paypal_paid").Exit(Handle, "hello")

    orderState.To("finish", &order)
      order.SetState("finish")
      order.NewStateLog("finish", tableName, Id, notes)
