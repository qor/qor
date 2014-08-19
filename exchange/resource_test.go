package exchange

import (
	"testing"

	"github.com/qor/qor/resource"
)

func TestGetMetaValues(t *testing.T) {
	phone := NewResource(Phone{})
	phone.HasSequentialColumns = true
	phone.RegisterMeta(&resource.Meta{Name: "Num", Label: "Phone"})

	address := NewResource(Address{})
	address.HasSequentialColumns = true
	address.RegisterMeta(&resource.Meta{Name: "Name", Label: "Address"})
	address.RegisterMeta(&resource.Meta{Name: "Phones", Resource: phone})

	user := NewResource(User{})
	user.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	user.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})
	user.RegisterMeta(&resource.Meta{Name: "Addresses", Resource: address})

	mvs, _ := user.getMetaValues(map[string]string{
		"Name":       "Van",
		"Address 01": "China",
		"Address 02": "USA",
		"Phone 1":    "xxx-xxx-xxx-1",
		"Phone 2":    "xxx-xxx-xxx-2",
	}, 0)

	expect := 3
	if len(mvs.Values) != expect {
		t.Errorf("expecting to retrieve %d MetaValues instead of %d", expect, len(mvs.Values))
	}

	walk(mvs)

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

var hasChina, hasUSA, hasPhone1, hasPhone2 bool

func walk(mvs *resource.MetaValues) {
	for _, v := range mvs.Values {
		if v.MetaValues != nil {
			for _, vs := range v.MetaValues.Values {
				if vs.Value != nil {
					switch vs.Value.(string) {
					case "China":
						hasChina = true
					case "USA":
						hasUSA = true
					}
				} else if mvs := vs.MetaValues; mvs != nil {
					walk(mvs)
				}
			}
		} else if v.Value != nil {
			switch v.Value.(string) {
			case "xxx-xxx-xxx-1":
				hasPhone1 = true
			case "xxx-xxx-xxx-2":
				hasPhone2 = true
			}
		}
	}
}
