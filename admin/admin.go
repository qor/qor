package admin

import (
	"reflect"
	"text/template"

	"github.com/qor/inflection"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

const (
	DEFAULT_PAGE_COUNT = 20
)

type I18n interface {
	Scope(scope string) I18n
	Default(value string) I18n
	T(locale string, key string, args ...interface{}) string
}

type Admin struct {
	Config    *qor.Config
	SiteName  string
	I18n      I18n
	menus     []*Menu
	resources []*Resource
	auth      Auth
	router    *Router
	funcMaps  template.FuncMap
}

type ResourceNamer interface {
	ResourceName() string
}

func New(config *qor.Config) *Admin {
	admin := Admin{
		funcMaps: make(template.FuncMap),
		Config:   config,
		router:   newRouter(),
	}
	return &admin
}

func (admin *Admin) SetSiteName(siteName string) {
	admin.SiteName = siteName
}

func (admin *Admin) SetAuth(auth Auth) {
	admin.auth = auth
}

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

	if configuration.PageCount == 0 {
		configuration.PageCount = DEFAULT_PAGE_COUNT
	}

	if configuration.Name != "" {
		res.Name = configuration.Name
	} else if namer, ok := value.(ResourceNamer); ok {
		res.Name = namer.ResourceName()
	}

	scope := admin.Config.DB.NewScope(res.Value)
	modelType := scope.GetModelStruct().ModelType
	for i := 0; i < modelType.NumField(); i++ {
		fieldStruct := modelType.Field(i)
		if field, ok := scope.FieldByName(fieldStruct.Name); !ok || field.Relationship == nil {
			if injector, ok := reflect.New(fieldStruct.Type).Interface().(configureAfterNewInjector); ok {
				injector.ConfigureQorResourceAfterNew(res)
			}
		}
	}

	if injector, ok := res.Value.(configureAfterNewInjector); ok {
		injector.ConfigureQorResourceAfterNew(res)
	}
	return res
}

type configureAfterNewInjector interface {
	ConfigureQorResourceAfterNew(*Resource)
}

func (admin *Admin) AddResource(value interface{}, config ...*Config) *Resource {
	res := admin.NewResource(value, config...)

	if !res.Config.Invisible {
		var menuName string
		if res.Config.Singleton {
			menuName = inflection.Singular(res.Name)
		} else {
			menuName = inflection.Plural(res.Name)
		}

		menu := &Menu{rawPath: res.ToParam(), Name: menuName}
		admin.menus = appendMenu(admin.menus, res.Config.Menu, menu)
	}

	admin.resources = append(admin.resources, res)
	return res
}

func (admin *Admin) GetResource(name string) *Resource {
	for _, res := range admin.resources {
		if res.ToParam() == name || res.Name == name || res.StructType == name {
			return res
		}
	}
	return nil
}

func (admin *Admin) GetResources() []*Resource {
	return admin.resources
}
