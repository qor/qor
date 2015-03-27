package admin

import (
	"fmt"
	"text/template"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
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

func (admin *Admin) NewResource(value interface{}, config *Config) *Resource {
	if config == nil {
		config = &Config{}
	}

	res := &Resource{
		Resource:    *resource.New(value),
		Config:      config,
		cachedMetas: &map[string][]*Meta{},
		scopes:      map[string]*Scope{},
		filters:     map[string]*Filter{},
		admin:       admin,
	}
	res.Finder = res.finder

	if config.Name != "" {
		res.Name = config.Name
	} else if namer, ok := value.(ResourceNamer); ok {
		res.Name = namer.ResourceName()
	}

	if injector, ok := value.(Injector); ok {
		injector.InjectQorAdmin(res)
	}
	return res
}

func (admin *Admin) AddResource(value interface{}, config *Config) *Resource {
	res := admin.NewResource(value, config)

	if !res.Config.Invisible {
		admin.menus = appendMenu(admin.menus, res.Config.Menu, res)
	}

	admin.resources = append(admin.resources, res)
	return res
}

func (admin *Admin) GetResource(name string) *Resource {
	for _, res := range admin.resources {
		if res.ToParam() == name || res.Name == name {
			return res
		}
	}
	return nil
}
