package admin

import "path"

type Menu struct {
	Name      string
	rawPath   string
	Link      string
	SubMenus  []*Menu
	Ancestors []string
}

func (admin Admin) GetMenus() []*Menu {
	return admin.menus
}

func (admin *Admin) AddMenu(menu *Menu) {
	admin.menus = appendMenu(admin.menus, menu.Ancestors, menu)
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

func newMenu(menus []string, neoMenu *Menu) *Menu {
	menuCount := len(menus)
	for index, _ := range menus {
		neoMenu = &Menu{Name: menus[menuCount-index-1], SubMenus: []*Menu{neoMenu}}
	}

	return neoMenu
}

func appendMenu(menus []*Menu, ancestors []string, neoMenu *Menu) []*Menu {
	if len(ancestors) > 0 {
		for _, m := range menus {
			if m.Name != ancestors[0] {
				continue
			}

			if len(ancestors) > 1 {
				m.SubMenus = appendMenu(m.SubMenus, ancestors[1:], neoMenu)
			} else {
				m.SubMenus = append(m.SubMenus, newMenu(nil, neoMenu))
			}
			return menus
		}
	}

	return append(menus, newMenu(ancestors, neoMenu))
}
