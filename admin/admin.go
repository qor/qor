package admin

import (
	"html/template"
	"reflect"

	"github.com/jinzhu/inflection"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/theplant/cldr"
)

const (
	DEFAULT_PAGE_COUNT = 20
)

type Admin struct {
	Config           *qor.Config
	SiteName         string
	I18n             I18n
	menus            []*Menu
	resources        []*Resource
	searchResources  []*Resource
	auth             Auth
	router           *Router
	funcMaps         template.FuncMap
	metaConfigorMaps map[string]func(*Meta)
}

type ResourceNamer interface {
	ResourceName() string
}

func New(config *qor.Config) *Admin {
	admin := Admin{
		funcMaps:         make(template.FuncMap),
		Config:           config,
		router:           newRouter(),
		metaConfigorMaps: metaConfigorMaps,
	}
	return &admin
}

func (admin *Admin) SetSiteName(siteName string) {
	admin.SiteName = siteName
}

func (admin *Admin) SetAuth(auth Auth) {
	admin.auth = auth
}

// RegisterMetaConfigor register configor for a kind, it will be called when register those kind of metas
func (admin *Admin) RegisterMetaConfigor(kind string, fc func(*Meta)) {
	admin.metaConfigorMaps[kind] = fc
}

// RegisterFuncMap register view funcs, it could be used in view templates
func (admin *Admin) RegisterFuncMap(name string, fc interface{}) {
	admin.funcMaps[name] = fc
}

func (admin *Admin) GetRouter() *Router {
	return admin.router
}

func (admin *Admin) NewResource(value interface{}, config ...*Config) *Resource {
	var configuration *Config
	if len(config) > 0 {
		configuration = config[0]
	}

	if configuration == nil {
		configuration = &Config{}
	}

	res := &Resource{
		Resource:    *resource.New(value),
		Config:      configuration,
		cachedMetas: &map[string][]*Meta{},
		filters:     map[string]*Filter{},
		admin:       admin,
	}

	res.Permission = configuration.Permission

	if configuration.PageCount == 0 {
		configuration.PageCount = DEFAULT_PAGE_COUNT
	}

	if configuration.Name != "" {
		res.Name = configuration.Name
	} else if namer, ok := value.(ResourceNamer); ok {
		res.Name = namer.ResourceName()
	}

	// Configure resource when initializing
	modelType := admin.Config.DB.NewScope(res.Value).GetModelStruct().ModelType
	for i := 0; i < modelType.NumField(); i++ {
		if fieldStruct := modelType.Field(i); fieldStruct.Anonymous {
			if injector, ok := reflect.New(fieldStruct.Type).Interface().(resource.ConfigureResourceBeforeInitializeInterface); ok {
				injector.ConfigureQorResourceBeforeInitialize(res)
			}
		}
	}

	if injector, ok := res.Value.(resource.ConfigureResourceBeforeInitializeInterface); ok {
		injector.ConfigureQorResourceBeforeInitialize(res)
	}

	findOneHandler := res.FindOneHandler
	res.FindOneHandler = func(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
		if context.ResourceID == "" {
			context.ResourceID = res.GetPrimaryValue(context.Request)
		}
		return findOneHandler(result, metaValues, context)
	}

	return res
}

func (admin *Admin) AddResource(value interface{}, config ...*Config) *Resource {
	res := admin.NewResource(value, config...)

	if !res.Config.Invisible {
		var menuName string
		if res.Config.Singleton {
			menuName = res.Name
		} else {
			menuName = inflection.Plural(res.Name)
		}

		menu := &Menu{rawPath: res.ToParam(), Name: menuName}
		admin.menus = appendMenu(admin.menus, res.Config.Menu, menu)
	}

	admin.resources = append(admin.resources, res)
	return res
}

func (admin *Admin) AddSearchResource(resources ...*Resource) {
	admin.searchResources = append(admin.searchResources, resources...)
}

func (admin *Admin) EnabledSearchCenter() bool {
	return len(admin.searchResources) > 0
}

func (admin *Admin) GetResource(name string) *Resource {
	for _, res := range admin.resources {
		var typeName = reflect.Indirect(reflect.ValueOf(res.Value)).Type().String()
		if res.ToParam() == name || res.Name == name || typeName == name {
			return res
		}
	}
	return nil
}

func (admin *Admin) GetResources() []*Resource {
	return admin.resources
}

// I18n define admin's i18n interface
type I18n interface {
	Scope(scope string) I18n
	Default(value string) I18n
	T(locale string, key string, args ...interface{}) template.HTML
}

func (admin *Admin) T(context *qor.Context, key string, value string, values ...interface{}) template.HTML {
	locale := utils.GetLocale(context)

	if admin.I18n == nil {
		if result, err := cldr.Parse(locale, value, values...); err == nil {
			return template.HTML(result)
		}
		return template.HTML(key)
	} else {
		return admin.I18n.Default(value).T(locale, key, values...)
	}
}
