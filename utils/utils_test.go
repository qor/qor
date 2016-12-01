package utils

import "testing"

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
	cases := [][2]string{
		{"OrderItem", "order_item"},
		{"order item", "order_item"},
		{"Order Item", "order_item"},
		{"FAQ", "faq"},
		{"FAQPage", "faq_page"},
	}
	for _, c := range cases {
		if got := ToParamString(c[0]); got != c[1] {
			t.Errorf("ToParamString(%q) = %q; want %q", c[0], got, c[1])
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
