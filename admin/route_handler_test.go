package admin

import (
	"net/url"
	"reflect"
	"testing"
)

type result struct {
	Route    string
	URL      string
	NotMatch bool
	Values   url.Values
}

func TestRouteHandler(t *testing.T) {
	var results = []result{
		{
			Route:  "/hello",
			URL:    "/hello",
			Values: url.Values{},
		}, {
			Route:  "/hello/:name",
			URL:    "/hello/!world",
			Values: url.Values{":name": []string{"!world"}},
		}, {
			Route:  "/hello/!world",
			URL:    "/hello/!world",
			Values: url.Values{},
		}, {
			Route:  "/hello/:id",
			URL:    "/hello/12",
			Values: url.Values{":id": []string{"12"}},
		}, {
			Route:  "/hello/:name[world]",
			URL:    "/hello/world",
			Values: url.Values{":name": []string{"world"}},
		}, {
			Route:    "/hello/:name[world]",
			URL:      "/hello/world/123",
			NotMatch: true,
		}, {
			Route:  "/hello/:name[world]/123",
			URL:    "/hello/world/123",
			Values: url.Values{":name": []string{"world"}},
		}, {
			Route:    "/hello/:name[world]",
			URL:      "/hello/jinzhu",
			NotMatch: true,
		}, {
			Route:  "/hello/:id[\\d+]",
			URL:    "/hello/12",
			Values: url.Values{":id": []string{"12"}},
		}, {
			Route:    "/hello/:id[\\d+]",
			URL:      "/hello/world",
			NotMatch: true,
		}, {
			Route:  "/hello/:id[[\\d]]",
			URL:    "/hello/1",
			Values: url.Values{":id": []string{"1"}},
		}, {
			Route:    "/hello/:id[[\\d]]",
			URL:      "/hello/12",
			NotMatch: true,
		}, {
			Route:  "/hello/:name[world]/:id[\\d+]",
			URL:    "/hello/world/123",
			Values: url.Values{":name": []string{"world"}, ":id": []string{"123"}},
		},
	}

	for _, result := range results {
		values, ok := routeHandler{Path: result.Route}.try(result.URL)

		if ok == result.NotMatch {
			t.Errorf("%v & %v expect match: %v, but got: %v", result.Route, result.URL, !result.NotMatch, ok)
		}

		if ok && !reflect.DeepEqual(values, result.Values) {
			t.Errorf("%v & %v matched url values should be %v, but got %v", result.Route, result.URL, result.Values, values)
		}
	}
}
