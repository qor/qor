package admin

import "path"

type Menu struct {
	Name     string
	rawPath  string
	Link     string
	SubMenus []*Menu
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

		if len(m.SubMenus) > 0 {
			if mc := getMenu(m.SubMenus, name); mc != nil {
				return mc
			}
		}
	}

	return nil
}

// Generate menu links by current route. e.g "/products" to "/admin/products"
func (admin *Admin) generateMenuLinks() {
	prefixMenuLinks(admin.menus, admin.router.Prefix)
}

func prefixMenuLinks(menus []*Menu, prefix string) {
	for _, m := range menus {
		if m.rawPath != "" {
			m.Link = path.Join(prefix, m.rawPath)
		}
		if len(m.SubMenus) > 0 {
			prefixMenuLinks(m.SubMenus, prefix)
		}
	}
}

func newMenu(menus []string, res *Resource) (menu *Menu) {
	menu = &Menu{rawPath: res.ToParam(), Name: res.Name}

	menuCount := len(menus)
	for index, _ := range menus {
		menu = &Menu{Name: menus[menuCount-index-1], SubMenus: []*Menu{menu}}
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
				m.SubMenus = appendMenu(m.SubMenus, resMenus[1:], res)
			} else {
				m.SubMenus = append(m.SubMenus, newMenu(nil, res))
			}
			return menus
		}
	}

	return append(menus, newMenu(resMenus, res))
}
