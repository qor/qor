package admin

import (
	"text/template"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

type Admin struct {
	Config    *qor.Config
	resources []*Resource
	auth      Auth
	router    *Router
	funcMaps  template.FuncMap

	menus []*Menu
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

func (admin *Admin) AddResource(value interface{}, config *Config) *Resource {
	res := &Resource{
		Resource:    *resource.New(value),
		Config:      config,
		cachedMetas: &map[string][]*Meta{},
		scopes:      map[string]*Scope{},
		filters:     map[string]*Filter{},
		admin:       admin,
	}

	var noName, noMenu bool
	if config != nil {
		if config.Name != "" {
			noName = true
			res.Name = config.Name
		}

		if config.Invisible {
			noMenu = true
		} else {
			if len(config.Menu) > 0 {
				noMenu = true
				admin.menus = appendMenu(admin.menus, config.Menu, res)
			}
		}
	}
	if !noName {
		if namer, ok := value.(ResourceNamer); ok {
			res.Name = namer.ResourceName()
		}
	}
	if !noMenu {
		admin.AddMenu(&Menu{Name: res.Name, params: res.ToParam()})
	}

	admin.resources = append(admin.resources, res)

	if injector, ok := value.(Injector); ok {
		injector.InjectQorAdmin(res)
	}
	return res
}

func (admin *Admin) GetResource(name string) *Resource {
	for _, res := range admin.resources {
		if res.ToParam() == name {
			return res
		}
	}
	return nil
}

func (admin *Admin) SetAuth(auth Auth) {
	admin.auth = auth
}

func (admin *Admin) GetRouter() *Router {
	return admin.router
}

func (admin *Admin) RegisterFuncMap(name string, fc interface{}) {
	admin.funcMaps[name] = fc
}
