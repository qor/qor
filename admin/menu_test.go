package admin

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

func generateResourceMenu(resource *Resource) *Menu {
	return &Menu{rawPath: resource.ToParam(), Name: resource.Name}
}

func TestAddMenuAndGetMenus(t *testing.T) {
	admin := New(&qor.Config{})
	menu := &Menu{Name: "Dashboard", Link: "/admin"}
	admin.AddMenu(menu)

	if menu != admin.GetMenus()[0] {
		t.Error("menu not added")
	}
}

func TestAddMenuWithAncestorsAndGetSubMenus(t *testing.T) {
	admin := New(&qor.Config{})
	admin.AddMenu(&Menu{Name: "Dashboard", Link: "/admin"})
	admin.AddMenu(&Menu{Name: "dashboard-sub", Link: "/link", Ancestors: []string{"Dashboard"}})

	menu := admin.GetMenus()[0]
	subMenu := menu.GetSubMenus()[0]

	if subMenu.Name != "dashboard-sub" || subMenu.Link != "/link" {
		t.Error("sub menu not added")
	}
}

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

	menus = appendMenu(menus, res7.Config.Menu, generateResourceMenu(res7))
	menus = appendMenu(menus, res1.Config.Menu, generateResourceMenu(res1))
	menus = appendMenu(menus, res2.Config.Menu, generateResourceMenu(res2))
	menus = appendMenu(menus, res3.Config.Menu, generateResourceMenu(res3))
	menus = appendMenu(menus, res4.Config.Menu, generateResourceMenu(res4))
	menus = appendMenu(menus, res5.Config.Menu, generateResourceMenu(res5))
	menus = appendMenu(menus, res6.Config.Menu, generateResourceMenu(res6))
	prefixMenuLinks(menus, "/admin")

	expect := []*Menu{
		{Name: "menu1", subMenus: []*Menu{
			{Name: "menu1-1", subMenus: []*Menu{
				{Name: "menu1-1-1", subMenus: []*Menu{
					{Name: res7.Name, rawPath: "res7", Link: "/admin/res7"},
				}},
				{Name: res3.Name, rawPath: "res3", Link: "/admin/res3"},
			}},
			{Name: res1.Name, rawPath: "res1", Link: "/admin/res1"},
			{Name: res2.Name, rawPath: "res2", Link: "/admin/res2"},
			{Name: "menu1-2", subMenus: []*Menu{
				{Name: res6.Name, rawPath: "res6", Link: "/admin/res6"},
			}},
		}},
		{Name: "menu2", subMenus: []*Menu{
			{Name: res4.Name, rawPath: "res4", Link: "/admin/res4"},
		}},
		{Name: res5.Name, rawPath: "res5", Link: "/admin/res5"},
	}

	if !reflect.DeepEqual(expect, menus) {
		g, err := json.MarshalIndent(menus, "", "  ")
		if err != nil {
			t.Error(err)
		}
		w, err := json.MarshalIndent(expect, "", "  ")
		if err != nil {
			t.Error(err)
		}
		t.Errorf("add menu errors: got %s; expect %s", g, w)
	}

	menu := getMenu(menus, "res6")
	if menu == nil {
		t.Error("failed to get menu")
	} else if menu.Name != "res6" {
		t.Error("failed to get correct menu")
	}
}

func TestAddSubMenuViaParents(t *testing.T) {
	var menus []*Menu
	subMenuName := "Dashboard-subMenu"
	sub2MenuName := "Dashboard-subMenu-subMenu"

	pMenu1 := "pMenu1"
	pMenu2 := "pMenu2"
	pMenu1_1 := "pMenu1_1"

	menus = appendMenu(menus, []string{}, &Menu{Name: "Dashboard"})
	menus = appendMenu(menus, []string{"Dashboard"}, &Menu{Name: subMenuName})
	menus = appendMenu(menus, []string{"Dashboard", subMenuName}, &Menu{Name: sub2MenuName})

	menus = appendMenu(menus, []string{}, &Menu{Name: "Product"})
	menus = appendMenu(menus, []string{"Product"}, &Menu{Name: pMenu1})
	menus = appendMenu(menus, []string{"Product"}, &Menu{Name: pMenu2})
	menus = appendMenu(menus, []string{"Product", pMenu1}, &Menu{Name: pMenu1_1})

	expected := []*Menu{
		{Name: "Dashboard", subMenus: []*Menu{
			{Name: subMenuName, subMenus: []*Menu{
				{Name: sub2MenuName},
			}},
		}},

		{Name: "Product", subMenus: []*Menu{
			{Name: pMenu1, subMenus: []*Menu{
				{Name: pMenu1_1},
			}},
			{Name: pMenu2},
		}},
	}

	if !reflect.DeepEqual(expected, menus) {
		g, err := json.MarshalIndent(menus, "", "  ")
		if err != nil {
			t.Error(err)
		}
		w, err := json.MarshalIndent(expected, "", "  ")
		if err != nil {
			t.Error(err)
		}
		t.Errorf("add menu errors: got %s; expected %s", g, w)
	}
}
