package admin

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/qor/qor"
)

func TestPatchUrl(t *testing.T) {
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
		{
			original: "http://qor.com/admin/orders?locale=global&q=dotnet&test=1#test",
			input:    []interface{}{"locale", 1},
			err:      errors.New("1 type is int, want string"),
		},
	}
	for _, c := range cases {
		u, _ := url.Parse(c.original)
		context := Context{Context: &qor.Context{Request: &http.Request{URL: u}}}
		got, err := context.PatchURL(c.input...)
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
