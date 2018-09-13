package utils

import (
	"fmt"
	"testing"

	"github.com/theplant/testingutils"
)

func TestHumanizeString(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"API", "API"},
		{"OrderID", "Order ID"},
		{"OrderItem", "Order Item"},
		{"orderItem", "Order Item"},
		{"OrderIDItem", "Order ID Item"},
		{"OrderItemID", "Order Item ID"},
		{"VIEW SITE", "VIEW SITE"},
		{"Order Item", "Order Item"},
		{"Order ITEM", "Order ITEM"},
		{"ORDER Item", "ORDER Item"},
	}
	for _, c := range cases {
		if got := HumanizeString(c.input); got != c.want {
			t.Errorf("HumanizeString(%q) = %q; want %q", c.input, got, c.want)
		}
	}
}

func TestToParamString(t *testing.T) {
	results := map[string]string{
		"OrderItem":  "order_item",
		"order item": "order_item",
		"Order Item": "order_item",
		"FAQ":        "faq",
		"FAQPage":    "faq_page",
		"!help_id":   "!help_id",
		"help id":    "help_id",
		"语言":         "yu-yan",
	}

	for key, value := range results {
		if ToParamString(key) != value {
			t.Errorf("%v to params should be %v, but got %v", key, value, ToParamString(key))
		}
	}
}

func TestPatchURL(t *testing.T) {
	var cases = []struct {
		original string
		input    []interface{}
		want     string
		err      error
	}{
		{
			original: "http://qor.com/admin/orders?locale=global&q=dotnet&test=1#test",
			input:    []interface{}{"locale", "cn"},
			want:     "http://qor.com/admin/orders?locale=cn&q=dotnet&test=1#test",
		},
		{
			original: "http://qor.com/admin/orders?locale=global&q=dotnet&test=1#test",
			input:    []interface{}{"locale", ""},
			want:     "http://qor.com/admin/orders?q=dotnet&test=1#test",
		},
	}
	for _, c := range cases {
		// u, _ := url.Parse(c.original)
		// context := Context{Context: &qor.Context{Request: &http.Request{URL: u}}}
		got, err := PatchURL(c.original, c.input...)
		if c.err != nil {
			if err == nil || err.Error() != c.err.Error() {
				t.Errorf("got error %s; want %s", err, c.err)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
			if got != c.want {
				t.Errorf("context.PatchURL = %s; c.want %s", got, c.want)
			}
		}
	}
}

func TestJoinURL(t *testing.T) {
	var cases = []struct {
		original string
		input    []interface{}
		want     string
		err      error
	}{
		{
			original: "http://qor.com",
			input:    []interface{}{"admin"},
			want:     "http://qor.com/admin",
		},
		{
			original: "http://qor.com",
			input:    []interface{}{"/admin"},
			want:     "http://qor.com/admin",
		},
		{
			original: "http://qor.com/",
			input:    []interface{}{"/admin"},
			want:     "http://qor.com/admin",
		},
		{
			original: "http://qor.com?q=keyword",
			input:    []interface{}{"admin"},
			want:     "http://qor.com/admin?q=keyword",
		},
		{
			original: "http://qor.com/?q=keyword",
			input:    []interface{}{"admin"},
			want:     "http://qor.com/admin?q=keyword",
		},
		{
			original: "http://qor.com/?q=keyword",
			input:    []interface{}{"admin/"},
			want:     "http://qor.com/admin/?q=keyword",
		},
	}
	for _, c := range cases {
		// u, _ := url.Parse(c.original)
		// context := Context{Context: &qor.Context{Request: &http.Request{URL: u}}}
		got, err := JoinURL(c.original, c.input...)
		if c.err != nil {
			if err == nil || err.Error() != c.err.Error() {
				t.Errorf("got error %s; want %s", err, c.err)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
			if got != c.want {
				t.Errorf("context.JoinURL = %s; c.want %s", got, c.want)
			}
		}
	}
}

func TestSortFormKeys(t *testing.T) {
	keys := []string{"QorResource.Category", "QorResource.Addresses[2].Address1", "QorResource.Addresses[1].Address1", "QorResource.Addresses[11].Address1", "QorResource.Addresses[0].Address1", "QorResource.Code", "QorResource.ColorVariations[0].Color", "QorResource.ColorVariations[0].ID", "QorResource.ColorVariations[0].SizeVariations[2].AvailableQuantity", "QorResource.ColorVariations[0].SizeVariations[11].AvailableQuantity", "QorResource.ColorVariations[0].SizeVariations[22].AvailableQuantity", "QorResource.ColorVariations[0].SizeVariations[3].AvailableQuantity", "QorResource.ColorVariations[0].SizeVariations[4].AvailableQuantity", "QorResource.ColorVariations[1].SizeVariations[0].AvailableQuantity", "QorResource.ColorVariations[1].SizeVariations[1].AvailableQuantity", "QorResource.ColorVariations[1].ID", "QorResource.ColorVariations[0].SizeVariations[1].ID", "QorResource.ColorVariations[0].SizeVariations[4].ID", "QorResource.ColorVariations[0].SizeVariations[3].ID", "QorResource.Z[0]"}

	SortFormKeys(keys)

	orderedKeys := []string{"QorResource.Addresses[0].Address1", "QorResource.Addresses[1].Address1", "QorResource.Addresses[2].Address1", "QorResource.Addresses[11].Address1", "QorResource.Category", "QorResource.Code", "QorResource.ColorVariations[0].Color", "QorResource.ColorVariations[0].ID", "QorResource.ColorVariations[0].SizeVariations[1].ID", "QorResource.ColorVariations[0].SizeVariations[2].AvailableQuantity", "QorResource.ColorVariations[0].SizeVariations[3].AvailableQuantity", "QorResource.ColorVariations[0].SizeVariations[3].ID", "QorResource.ColorVariations[0].SizeVariations[4].AvailableQuantity", "QorResource.ColorVariations[0].SizeVariations[4].ID", "QorResource.ColorVariations[0].SizeVariations[11].AvailableQuantity", "QorResource.ColorVariations[0].SizeVariations[22].AvailableQuantity", "QorResource.ColorVariations[1].ID", "QorResource.ColorVariations[1].SizeVariations[0].AvailableQuantity", "QorResource.ColorVariations[1].SizeVariations[1].AvailableQuantity", "QorResource.Z[0]"}

	if fmt.Sprint(keys) != fmt.Sprint(orderedKeys) {
		t.Errorf("ordered form keys should be \n%v\n, but got\n%v", orderedKeys, keys)
	}
}

func TestSafeJoin(t *testing.T) {
	pth1, err := SafeJoin("hello", "world")
	if err != nil || pth1 != "hello/world" {
		t.Errorf("no error should happen")
	}

	// test possible vulnerability https://snyk.io/research/zip-slip-vulnerability#go
	pth2, err := SafeJoin("hello", "../world")
	if err == nil || pth2 != "" {
		t.Errorf("no error should happen")
	}
}

func TestFormattedDecimal(t *testing.T) {
	testCases := []struct {
		Float    float64
		Expected string
	}{
		{Float: 0.0, Expected: "0"},
		{Float: 0.12345678, Expected: "0.12345678"},
		{Float: 0.123456789, Expected: "0.12345678"},
		{Float: 0.1234, Expected: "0.1234"},
		{Float: 0.00001234, Expected: "0.00001234"},
		{Float: 0.000012345, Expected: "0.00001234"},
		{Float: 1.1234, Expected: "1.1234"},
		{Float: 12345.1234, Expected: "12345.1234"},
		{Float: 123456789.1234, Expected: "123456789.1234"},
		{Float: 1234567890.1234, Expected: "1234567890"},
		{Float: 123, Expected: "123"},
		{Float: 1230000000, Expected: "1230000000"},
		{Float: 12300000009, Expected: "12300000009"},
		{Float: 123000000091, Expected: "123000000091"},
		{Float: 0.000000005, Expected: "0.000000005"},
	}

	for _, testCase := range testCases {
		if diff := testingutils.PrettyJsonDiff(testCase.Expected, FormattedDecimal(testCase.Float)); len(diff) > 0 {
			t.Errorf(diff)
		}
	}
}
