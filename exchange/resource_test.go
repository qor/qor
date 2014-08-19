package exchange

import (
	"testing"

	"github.com/qor/qor/resource"
)

func TestGetMetaValues(t *testing.T) {
	phone := NewResource(Phone{})
	phone.RegisterMeta(&resource.Meta{Name: "Num", Label: "Phone"})

	address := NewResource(Address{})
	address.HasSequentialColumns = true
	address.RegisterMeta(&resource.Meta{Name: "Name", Label: "Address"})
	address.RegisterMeta(&resource.Meta{Name: "Phone", Resource: phone})

	user := NewResource(User{})
	user.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	user.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})
	user.RegisterMeta(&resource.Meta{Name: "Addresses", Resource: address})

	mvs, _ := user.getMetaValues(map[string]string{
		"Name":       "Van",
		"Address 01": "China",
		"Address 02": "USA",
		"Phone 01":   "xxx-xxx-xxx-1",
		"Phone 02":   "xxx-xxx-xxx-2",

		// Should not be included in returned mvs
		"Phone 03": "xxx-xxx-xxx-2",
	}, 0)

	expect := 3
	if len(mvs.Values) != expect {
		t.Errorf("expecting to retrieve %d MetaValues instead of %d", expect, len(mvs.Values))
	}

	var hasChina, hasUSA, hasPhone1, hasPhone2 bool
	for _, mv := range mvs.Values {
		switch mv.Name {
		case "Addresses":
			for _, mv := range mv.MetaValues.Values {
				switch mv.Name {
				case "Name":
					switch mv.Value.(string) {
					case "China":
						hasChina = true
					case "USA":
						hasUSA = true
					}
				case "Phone":
					if len(mv.MetaValues.Values) != 1 {
						t.Errorf("Expect 1 phone value per address instead of %d", len(mv.MetaValues.Values))
					}
					for _, mv := range mv.MetaValues.Values {
						switch mv.Name {
						case "Num":
							switch mv.Value.(string) {
							case "xxx-xxx-xxx-1":
								hasPhone1 = true
							case "xxx-xxx-xxx-2":
								hasPhone2 = true
							}
						}
					}
				}
			}
		case "Name":
			if name := mv.Value.(string); name != "Van" {
				t.Errorf(`Expect name "Van" but got %s`, name)
			}
		}
	}

	if !hasChina {
		t.Error("Should contains China in mvs.Values")
	}
	if !hasUSA {
		t.Error("Should contains USA in mvs.Values")
	}
	if !hasPhone1 {
		t.Error("Should contains xxx-xxx-xxx-1 in mvs.Values")
	}
	if !hasPhone2 {
		t.Error("Should contains xxx-xxx-xxx-2 in mvs.Values")
	}
}
