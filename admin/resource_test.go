package admin

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fatih/color"

	"github.com/qor/qor"
)

type Product struct {
	Name        string
	Code        string
	URL         string
	Description string
}

// Test Edit Attrs
type EditAttrsTestCase struct {
	Params []string
	Result []string
}

func TestEditAttrs(t *testing.T) {
	var testCases []EditAttrsTestCase
	testCases = append(testCases,
		EditAttrsTestCase{Params: []string{"Name", "Code"}, Result: []string{"Name", "Code"}},
		EditAttrsTestCase{Params: []string{"-Name"}, Result: []string{"Code", "URL", "Description"}},
		EditAttrsTestCase{Params: []string{"Name", "-Code"}, Result: []string{"Name"}},
		EditAttrsTestCase{Params: []string{"Name", "-Code", "-Name"}, Result: []string{}},
		EditAttrsTestCase{Params: []string{"Name", "Code", "-Name"}, Result: []string{"Code"}},
		EditAttrsTestCase{Params: []string{"-Name", "Code", "Name"}, Result: []string{"Code", "Name"}},
		EditAttrsTestCase{Params: []string{"Section:Name+Code+Description", "-Name"}, Result: []string{"Code+Description"}},
	)

	admin := New(&qor.Config{DB: db})
	product := admin.AddResource(&Product{})
	i := 1
	for _, testCase := range testCases {
		var attrs []interface{}
		for _, param := range testCase.Params {
			if strings.HasPrefix(param, "Section:") {
				var rows [][]string
				param = strings.Replace(param, "Section:", "", 1)
				rows = append(rows, strings.Split(param, "+"))
				attrs = append(attrs, &Section{Rows: rows})
			} else {
				attrs = append(attrs, param)
			}
		}

		editSections := product.EditAttrs(attrs...)
		var results []string
		for _, section := range editSections {
			columnStr := strings.Join(section.Rows[0], "+")
			results = append(results, columnStr)
		}
		if compareStringSlice(results, testCase.Result) {
			color.Green(fmt.Sprintf("Edit Attrs TestCase #%d: Success\n", i))
		} else {
			t.Errorf(color.RedString(fmt.Sprintf("\nEdit Attrs TestCase #%d: Failure Result:%v\n", i, results)))
		}
		i++
	}
}

func compareStringSlice(slice1 []string, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	i := 0
	for _, s := range slice1 {
		if s != slice2[i] {
			return false
		}
		i++
	}
	return true
}
