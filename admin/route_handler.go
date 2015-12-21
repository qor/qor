package admin

import "net/url"

type requestHandler func(c *Context)

type routeHandler struct {
	Path   string
	Handle requestHandler
}

func isAlpha(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlnum(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

func matchPart(b byte) func(byte) bool {
	return func(c byte) bool {
		return c != b && c != '/'
	}
}

func match(s string, f func(byte) bool, i int) (matched string, next byte, j int) {
	j = i
	for j < len(s) && f(s[j]) {
		j++
	}
	if j < len(s) {
		next = s[j]
	}
	return s[i:j], next, j
}

// copied from pat https://github.com/bmizerany/pat
func (h routeHandler) try(path string) (url.Values, bool) {
	p := make(url.Values)
	var i, j int
	for i < len(path) {
		switch {
		case j >= len(h.Path):
			if h.Path != "/" && len(h.Path) > 0 && h.Path[len(h.Path)-1] == '/' {
				return p, true
			}
			return nil, false
		case h.Path[j] == ':':
			var name, val string
			var nextc byte
			name, nextc, j = match(h.Path, isAlnum, j+1)
			val, _, i = match(path, matchPart(nextc), i)
			p.Add(":"+name, val)
		case path[i] == h.Path[j]:
			i++
			j++
		default:
			return nil, false
		}
	}

	if j != len(h.Path) {
		return nil, false
	}
	return p, true
}
