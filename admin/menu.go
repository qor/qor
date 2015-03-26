package admin

import "path"

type Menu struct {
	Name   string
	params string
	Link   string
	Items  []*Menu
}

func (admin Admin) GetMenus() []*Menu {
	return admin.menus
}

func (admin *Admin) AddMenu(menu *Menu) {
	admin.menus = append(admin.menus, menu)
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

func newMenu(menus []string, res *Resource) (menu *Menu) {
	menu = &Menu{params: res.ToParam(), Name: res.Name}

	menuCount := len(menus)
	for index, _ := range menus {
		menu = &Menu{Name: menus[menuCount-index-1], Items: []*Menu{menu}}
	}

	return
}

func appendMenu(menus []*Menu, resMenus []string, res *Resource) []*Menu {
	if len(resMenus) > 0 {
		for _, m := range menus {
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
	}

	return append(menus, newMenu(resMenus, res))
}
