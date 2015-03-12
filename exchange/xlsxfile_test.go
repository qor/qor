package exchange

import (
	"reflect"

	"testing"
)

func TestNewXLSXFile(t *testing.T) {
	got, err := NewXLSXFile("fixture/simple.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	expect := &XLSXFile{lines: [][]string{
		[]string{"Index", "Name", "Age", "", "", "", "", "", ""},
		[]string{"1", "Van", "24", "", "", "", "", "", ""},
		[]string{"2", "Dave", "32", "", "", "", "", "", ""},
		[]string{"3", "Kate", "25", "", "", "", "", "", ""},
	}}

	if !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v\ngot %#v", expect, got)
	}
}
