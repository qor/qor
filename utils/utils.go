package utils

import "strings"

// Humanize separates string based on capitalizd letters
// e.g. "OrderItem" -> "Order Item"
func HumanizeString(str string) string {
	var human []rune
	for _, l := range str {
		if rune('A') <= l && l <= rune('Z') {
			human = append(human, rune(' '), l)
		} else {
			human = append(human, l)
		}
	}
	return strings.Title(string(human))
}

// ToParamString replaces spaces and separates words (by uppercase letters) with
// underscores in a string, also downcase it
// e.g. ToParamString -> to_param_string, To ParamString -> to_param_string
func ToParamString(str string) string {
	if len(str) <= 1 {
		return strings.ToLower(str)
	}

	str = strings.Replace(str, " ", "_", -1)
	result := []rune{rune(str[0])}
	for _, l := range str[1:] {
		if rune('A') <= l && l <= rune('Z') {
			if lr := len(result); lr == 0 || result[lr-1] != '_' {
				result = append(result, rune('_'), l)
				continue
			}
		}

		result = append(result, l)
	}

	return strings.ToLower(string(result))
}
