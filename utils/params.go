package utils

import (
	"net/url"
	"regexp"
	"strings"
)

func isAlpha(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '!'
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

// ParamsMatch match string by param
func ParamsMatch(source string, path string) (url.Values, string, bool) {
	var i, j int
	var p = make(url.Values)

	for i < len(path) {
		switch {
		case j >= len(source):

			if source != "/" && len(source) > 0 && source[len(source)-1] == '/' {
				return p, path, true
			}

			if source == "" && path == "/" {
				return p, path, true
			}
			return p, path[:i], false
		case source[j] == ':':
			var name, val string
			var nextc byte

			name, nextc, j = match(source, isAlnum, j+1)
			val, _, i = match(path, matchPart(nextc), i)

			if (j < len(source)) && source[j] == '[' {
				var index int
				if idx := strings.Index(source[j:], "]/"); idx > 0 {
					index = idx
				} else if source[len(source)-1] == ']' {
					index = len(source) - j - 1
				}

				if index > 0 {
					match := strings.TrimSuffix(strings.TrimPrefix(source[j:j+index+1], "["), "]")
					if reg, err := regexp.Compile("^" + match + "$"); err == nil && reg.MatchString(val) {
						j = j + index + 1
					} else {
						return nil, "", false
					}
				}
			}

			p.Add(":"+name, val)
		case path[i] == source[j]:
			i++
			j++
		default:
			return nil, "", false
		}
	}

	if j != len(source) {
		if (len(source) == j+1) && source[j] == '/' {
			return p, path, true
		}

		return nil, "", false
	}
	return p, path, true
}
