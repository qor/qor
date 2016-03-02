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

// Admin is a struct that used to generate admin/api interface
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

// ResourceNamer is an interface for models that defined method `ResourceName`
type ResourceNamer interface {
	ResourceName() string
}

// New new admin with configuration
func New(config *qor.Config) *Admin {
	admin := Admin{
		funcMaps:         make(template.FuncMap),
		Config:           config,
		router:           newRouter(),
		metaConfigorMaps: metaConfigorMaps,
	}
	return &admin
}

// SetSiteName set site's name, the name will be used as admin HTML title and admin interface will auto load javascripts, stylesheets files based on its value
// For example, if you named it as `Qor Demo`, admin will look up `qor_demo.js`, `qor_demo.css` in QOR view paths, and load them if found
func (admin *Admin) SetSiteName(siteName string) {
	admin.SiteName = siteName
}

// SetAuth set admin's authorization gateway
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

// GetRouter get router from admin
func (admin *Admin) GetRouter() *Router {
	return admin.router
}

// NewResource initialize a new qor resource, won't add it to admin, just initialize it
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
		configuration.PageCount = 20
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

// AddResource make a model manageable from admin interface
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

// AddSearchResource make a resource searchable from search center
func (admin *Admin) AddSearchResource(resources ...*Resource) {
	admin.searchResources = append(admin.searchResources, resources...)
}

func (admin *Admin) EnabledSearchCenter() bool {
	return len(admin.searchResources) > 0
}

// GetResource get resource with name
func (admin *Admin) GetResource(name string) *Resource {
	for _, res := range admin.resources {
		var typeName = reflect.Indirect(reflect.ValueOf(res.Value)).Type().String()
		if res.ToParam() == name || res.Name == name || typeName == name {
			return res
		}
	}
	return nil
}

// GetResources get defined resources from admin
func (admin *Admin) GetResources() []*Resource {
	return admin.resources
}

// I18n define admin's i18n interface
type I18n interface {
	Scope(scope string) I18n
	Default(value string) I18n
	T(locale string, key string, args ...interface{}) template.HTML
}

// T call i18n backend to translate
func (admin *Admin) T(context *qor.Context, key string, value string, values ...interface{}) template.HTML {
	locale := utils.GetLocale(context)

	if admin.I18n == nil {
		if result, err := cldr.Parse(locale, value, values...); err == nil {
			return template.HTML(result)
		}
		return template.HTML(key)
	}

	return admin.I18n.Default(value).T(locale, key, values...)
}
