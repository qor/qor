package utils

import "strings"

func Humanize(str string) string {
	var human []rune
	for _, l := range str {
		if rune('A') <= l && l <= rune('Z') {
			human = append(human, rune(' '), rune(l))
		} else {
			human = append(human, rune(l))
		}
	}
	return strings.Title(string(human))
}
