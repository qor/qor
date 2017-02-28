package utils

import (
	"net/url"
	"reflect"
	"testing"
)

func TestParamsMatch(t *testing.T) {
	type paramMatchChecker struct {
		Source  string
		Path    string
		Results url.Values
	}

	checkers := []paramMatchChecker{
		{Source: "/hello/:name", Path: "/hello/world", Results: url.Values{":name": []string{"world"}}},
		{Source: "/hello/:name/:id", Path: "/hello/world/444", Results: url.Values{":name": []string{"world"}, ":id": []string{"444"}}},
		{Source: "/hello/:name/:id", Path: "/bye/world/444", Results: nil},
	}

	for _, checker := range checkers {
		results, ok := ParamsMatch(checker.Source, checker.Path)

		if (checker.Results != nil) != ok {
			t.Errorf("%+v should matched correctly, matched should be %v, but got %v", checker, checker.Results != nil, ok)
		}

		if !reflect.DeepEqual(results, checker.Results) {
			t.Errorf("%+v's match results should be same, should got %v, but got %v", checker, checker.Results, results)
		}
	}
}
