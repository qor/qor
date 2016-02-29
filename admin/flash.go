package admin

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
)

// FlashNow add flash message to current request
func (context *Context) FlashNow(message, typ string) {
	context.Flashes = append(context.Flashes, Flash{Type: typ, Message: message})
}

// Flash add flash message to next request
func (context *Context) Flash(message, typ string) {
	flash := Flash{Type: typ, Message: message, Keep: true}
	context.Flashes = append(context.Flashes, flash)
	context.writeFlashes()
}

// GetFlashes get flash messages for this request
func (context *Context) GetFlashes() []Flash {
	flashes := context.readFlashFromCookie()
	flashes = append(flashes, context.Flashes...)
	context.writeFlashes()
	return flashes
}

// Flash flash message definiation
type Flash struct {
	Type    string
	Message string
	Keep    bool
}

func (context *Context) readFlashFromCookie() (flashes []Flash) {
	if cookie, err := context.Request.Cookie("qor-flashes"); err == nil {
		if bytes, err := base64.StdEncoding.DecodeString(cookie.Value); err == nil {
			json.Unmarshal(bytes, &flashes)
		}
	}
	return
}

func (context *Context) writeFlashes() {
	var flashes []Flash
	for _, flash := range context.Flashes {
		if flash.Keep {
			flashes = append(flashes, flash)
		}
	}

	if bytes, err := json.Marshal(flashes); err == nil {
		prefix := context.Admin.GetRouter().Prefix
		cookie := http.Cookie{Name: "qor-flashes", Value: base64.StdEncoding.EncodeToString(bytes), Path: prefix}
		http.SetCookie(context.Writer, &cookie)
	}
}
