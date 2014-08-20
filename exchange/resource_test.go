package exchange

import (
	"log"
	"testing"

	"github.com/qor/qor"

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
	user.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"}).Set("AliasHeaders", []string{"Aeon"})
	user.RegisterMeta(&resource.Meta{Name: "Addresses", Resource: address})

	mvs, _ := user.getMetaValues(map[string]string{
		"Name":       "Van",
		"Aeon":       "24",
		"Address 01": "China",
		"Address 02": "USA",
		"Phone 01":   "xxx-xxx-xxx-1",
		"Phone 02":   "xxx-xxx-xxx-2",

		// Should not be included in returned mvs
		"Phone 03": "xxx-xxx-xxx-2",
	}, 0)

	expect := 4
	if len(mvs.Values) != expect {
		t.Errorf("expecting to retrieve %d MetaValues instead of %d", expect, len(mvs.Values))
	}

	if testing.Verbose() {
		for _, mv := range mvs.Values {
			log.Printf("--> %+v\n", mv.Value)
		}
	}

	var hasChina, hasUSA, hasPhone1, hasPhone2, hasName, hasAeon bool
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
			hasName = true
			if name := mv.Value.(string); name != "Van" {
				t.Errorf(`Expect name "Van" but got %s`, name)
			}
		case "Age":
			hasAeon = true
			if aeon := mv.Value.(string); aeon != "24" {
				t.Errorf(`Expect Aeon "24" but got "%s"`, aeon)
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
	if !hasName {
		t.Error("Should contains Van in mvs.Values")
	}
	if !hasAeon {
		t.Error("Should contains 24 in mvs.Values")
	}
}

func TestMetaOptional(t *testing.T) {
	cleanup()

	address := NewResource(Address{})
	name := address.RegisterMeta(&resource.Meta{Name: "Name", Label: "Address"})
	ex := New(address)
	for i := 0; i < 2; i++ {
		if i == 1 {
			name.Set("Optional", true)
		}
		fi, _, err := ex.Import(&XLSXFile{
			lines: [][]string{
				[]string{"Country"},
				[]string{"USA"},
			},
		}, &qor.Context{
			DB: testdb,
		})
		if err != nil {
			t.Fatal(err)
		}
		var hasError bool
		select {
		case <-fi.Done:
		case <-fi.Error:
			hasError = true
		}

		switch i {
		case 0:
			if !hasError {
				t.Error("Should receive error when Name is not optional")
			}
		case 1:
			if hasError {
				t.Error("Should not receive error when Name is optional")
			}
		}
	}
}
