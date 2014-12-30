package admin

import (
	"bytes"
	"runtime/debug"
)

func (admin *Admin) RenderError(err error, code int, context *Context) {
	stacks := append([]byte(err.Error()+"\n"), debug.Stack()...)
	data := struct {
		Code int
		Body string
	}{
		Code: code,
		Body: string(bytes.Replace(stacks, []byte("\n"), []byte("<br>"), -1)),
	}
	context.Writer.WriteHeader(data.Code)
	context.Execute("error", data)
}
