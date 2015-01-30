package admin

import (
	"encoding/json"
	"testing"

	"github.com/qor/qor/resource"
)

func TestMenu(t *testing.T) {
	var menus []*Menu
	res1 := &Resource{
		Resource: resource.Resource{Name: "res1"},
		Config:   &Config{Menu: []string{"menu1"}},
	}
	res2 := &Resource{
		Resource: resource.Resource{Name: "res2"},
		Config:   &Config{Menu: []string{"menu1"}},
	}
	res3 := &Resource{
		Resource: resource.Resource{Name: "res3"},
		Config:   &Config{Menu: []string{"menu1", "menu1-1"}},
	}
	res4 := &Resource{
		Resource: resource.Resource{Name: "res4"},
		Config:   &Config{Menu: []string{"menu2"}},
	}
	res5 := &Resource{
		Resource: resource.Resource{Name: "res5"},
		Config:   &Config{},
	}
	res6 := &Resource{
		Resource: resource.Resource{Name: "res6"},
		Config:   &Config{Menu: []string{"menu1", "menu1-2"}},
	}
	res7 := &Resource{
		Resource: resource.Resource{Name: "res7"},
		Config:   &Config{Menu: []string{"menu1", "menu1-1", "menu1-1-1"}},
	}

	menus = appendMenu(menus, res1.Config.Menu, res1)
	menus = appendMenu(menus, res2.Config.Menu, res2)
	menus = appendMenu(menus, res3.Config.Menu, res3)
	menus = appendMenu(menus, res4.Config.Menu, res4)
	menus = appendMenu(menus, res5.Config.Menu, res5)
	menus = appendMenu(menus, res6.Config.Menu, res6)
	menus = appendMenu(menus, res7.Config.Menu, res7)
	relinkMenus(menus, "/admin")

	expect := []*Menu{
		&Menu{Name: "menu1", Items: []*Menu{
			&Menu{Name: res1.Name, Link: "/admin/res1"},
			&Menu{Name: res2.Name, Link: "/admin/res2"},
			&Menu{Name: "menu1-1", Items: []*Menu{
				&Menu{Name: res3.Name, Link: "/admin/res3"},
				&Menu{Name: "menu1-1-1", Items: []*Menu{
					&Menu{Name: res7.Name, Link: "/admin/res7"},
				}},
			}},
			&Menu{Name: "menu1-2", Items: []*Menu{
				&Menu{Name: res6.Name, Link: "/admin/res6"},
			}},
		}},
		&Menu{Name: "menu2", Items: []*Menu{
			&Menu{Name: res4.Name, Link: "/admin/res4"},
		}},
		&Menu{Name: res5.Name, Link: "/admin/res5"},
	}

	isEqual := func(expect, got []*Menu) bool {
		e, err := json.Marshal(expect)
		if err != nil {
			t.Error("marshal expect error:", err)
		}
		g, err := json.Marshal(got)
		if err != nil {
			t.Error("marshal got error:", err)
		}

		return string(e) == string(g)
	}

	if !isEqual(expect, menus) {
		t.Errorf("add menu errors: expect %s got %s", expect, menus)
	}

	menu := getMenu(menus, "res6")
	if menu == nil {
		t.Error("failed to get menu")
	} else if menu.Name != "res6" {
		t.Error("failed to get correct menu")
	}
}
