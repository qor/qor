package responder

import (
	"net/http"
	"path/filepath"
	"strings"
)

type Responder struct {
	responds map[string]func()
}

func With(format string, fc func()) *Responder {
	rep := &Responder{responds: map[string]func(){}}
	return rep.With(format, fc)
}

func (rep *Responder) With(format string, fc func()) *Responder {
	rep.responds[format] = fc
	return rep
}

func (rep *Responder) Respond(writer http.ResponseWriter, request *http.Request) {
	format := "html"
	if ext := filepath.Ext(request.URL.Path); ext != "" {
		format = strings.TrimPrefix(ext, ".")
	}

	if request.Header.Get("Content-Type") == "application/json" {
		format = "json"
	}

	for _, str := range strings.Split(request.Header.Get("Accept"), ",") {
		if str == "application/json" {
			format = "json"
			break
		}
	}

	if fc, ok := rep.responds[format]; ok {
		fc()
		return
	}

	for _, respond := range rep.responds {
		respond()
		break
	}
	return
}
