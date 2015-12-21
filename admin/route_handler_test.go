package admin

import (
	"fmt"
	"testing"
)

func TestRouteHandler(t *testing.T) {
	var urls = map[string]string{
		"/hello": "/hello",
	}

	for key, value := range urls {
		fmt.Println(routeHandler{Path: key}.try(value))
	}
}
