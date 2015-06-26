package admin

import (
	"encoding/json"
	"net/http"
)

type Flash struct {
	Type    string
	Message string
}

func (context *Context) readFlashFromCookie() (flashes []Flash) {
	if cookie, err := context.Request.Cookie("qor-flashes"); err == nil {
		json.Unmarshal([]byte(cookie.Value), &flashes)
	}
	return
}

func (context *Context) FlashNow(message, typ string) {
	context.Flashs = append(context.Flashs, Flash{Type: typ, Message: message})
}

func (context *Context) Flash(message, typ string) {
	flashes := context.readFlashFromCookie()
	flashes = append(flashes, Flash{Type: typ, Message: message})
	context.Flashs = append(context.Flashs, Flash{Type: typ, Message: message})

	if bytes, err := json.Marshal(context.Flashs); err == nil {
		http.SetCookie(context.Writer, &http.Cookie{Name: "qor-flashes", Value: string(bytes)})
	}
}

func (context *Context) GetFlashes() []Flash {
	flashes := context.readFlashFromCookie()
	flashes = append(flashes, context.Flashs...)
	return flashes
}
