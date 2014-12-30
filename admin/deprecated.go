package admin

import (
	"bytes"
	"runtime/debug"
)

func (admin *Admin) RenderError(err error, code int, c *Context) {
	stacks := append([]byte(err.Error()+"\n"), debug.Stack()...)
	data := struct {
		Code int
		Body string
	}{
		Code: code,
		Body: string(bytes.Replace(stacks, []byte("\n"), []byte("<br>"), -1)),
	}
	c.Writer.WriteHeader(data.Code)
	content := Content{Context: c, Result: data}
	content.Execute("error")
}
