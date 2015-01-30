package admin

import (
	"path"

	"github.com/bom-d-van/goutils/printutils"
)

type Menu struct {
	Name   string
	params string
	Link   string
	Items  []*Menu
}

func (admin Admin) GetMenus() []*Menu {
	printutils.PrettyPrint(admin.menus)
	return admin.menus
}

func (admin Admin) GetMenu(name string) (m *Menu) {
	return getMenu(admin.menus, name)
}

func getMenu(menus []*Menu, name string) *Menu {
	for _, m := range menus {
		if m.Name == name {
			return m
		}

		if len(m.Items) > 0 {
			if mc := getMenu(m.Items, name); mc != nil {
				return mc
			}
		}
	}

	return nil
}

func (admin *Admin) AddMenu(menu *Menu) {
	admin.menus = append(admin.menus, menu)
}

func (admin *Admin) linkMenus() {
	relinkMenus(admin.menus, admin.router.Prefix)
}

func relinkMenus(menus []*Menu, prefix string) {
	for _, m := range menus {
		if m.params != "" {
			m.Link = path.Join(prefix, m.params)
		}
		if len(m.Items) > 0 {
			relinkMenus(m.Items, prefix)
		}
	}
}

func (m *Menu) AddChild(menu *Menu) {
	m.Items = append(m.Items, menu)
}

func newMenu(menus []string, res *Resource) (m *Menu) {
	if mlen := len(menus); mlen == 0 {
		m = &Menu{params: res.ToParam(), Name: res.Name}
	} else {
		m = &Menu{
			Name:  menus[0],
			Items: []*Menu{&Menu{params: res.ToParam(), Name: res.Name}},
		}

		for i := mlen - 2; i >= 0; i-- {
			m = &Menu{
				Name:  menus[i],
				Items: []*Menu{m},
			}
		}
	}

	return
}

func appendMenu(menus []*Menu, resMenus []string, res *Resource) []*Menu {
	if len(resMenus) == 0 {
		return append(menus, newMenu(resMenus, res))
	}

	for _, m := range menus {
		if m.Link != "" {
			continue
		}

		if m.Name != resMenus[0] {
			continue
		}

		if len(resMenus) > 1 {
			m.Items = appendMenu(m.Items, resMenus[1:], res)
		} else {
			m.Items = append(m.Items, newMenu(nil, res))
		}
		return menus
	}

	return append(menus, newMenu(resMenus, res))
}
