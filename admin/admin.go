package admin

import (
	"fmt"
	"text/template"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

const (
	DEFAULT_PAGE_COUNT = 10
)

type Admin struct {
	Config    *qor.Config
	menus     []*Menu
	resources []*Resource
	auth      Auth
	router    *Router
	funcMaps  template.FuncMap
}

type Injector interface {
	InjectQorAdmin(*Resource)
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

func (admin *Admin) SetAuth(auth Auth) {
	admin.auth = auth
}

func (admin *Admin) RegisterFuncMap(name string, fc interface{}) {
	admin.funcMaps[name] = fc
}

func (admin *Admin) GetRouter() *Router {
	return admin.router
}

func (res *Resource) finder(result interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
	var primaryKey string
	if metaValues == nil {
		primaryKey = context.ResourceID
	} else if id := metaValues.Get(res.PrimaryFieldName()); id != nil {
		primaryKey = utils.ToString(id.Value)
	}

	if primaryKey != "" {
		if metaValues != nil {
			if destroy := metaValues.Get("_destroy"); destroy != nil {
				if fmt.Sprintf("%v", destroy.Value) != "0" {
					context.GetDB().Delete(result, primaryKey)
					return resource.ErrProcessorSkipLeft
				}
			}
		}
		return context.GetDB().First(result, primaryKey).Error
	}
	return nil
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
		scopes:      map[string]*Scope{},
		filters:     map[string]*Filter{},
		admin:       admin,
	}
	res.FindOneHandler = res.finder

	if configuration.PageCount == 0 {
		configuration.PageCount = DEFAULT_PAGE_COUNT
	}

	if configuration.Name != "" {
		res.Name = configuration.Name
	} else if namer, ok := value.(ResourceNamer); ok {
		res.Name = namer.ResourceName()
	}

	if injector, ok := value.(Injector); ok {
		injector.InjectQorAdmin(res)
	}
	return res
}

func (admin *Admin) AddResource(value interface{}, config ...*Config) *Resource {
	res := admin.NewResource(value, config...)

	if !res.Config.Invisible {
		// TODO: move Menu out of res.Config, make the API looks better
		menu := &Menu{rawPath: res.ToParam(), Name: res.Name}
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
