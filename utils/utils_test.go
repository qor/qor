package utils

import "testing"

func TestToParamString(t *testing.T) {
	cases := [][2]string{
		{"OrderItem", "order_item"},
		{"order item", "order_item"},
		{"Order Item", "order_item"},
	}
	for _, c := range cases {
		if got := ToParamString(c[0]); got != c[1] {
			t.Errorf("ToParamString(%q) = %q; want %q", c[0], got, c[1])
		}
	}
}
