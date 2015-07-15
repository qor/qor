package utils

import (
	"regexp"
	"strings"
)

var pluralInflections = [][]string{
	[]string{"([a-z])$", "${1}s"},
	[]string{"s$", "s"},
	[]string{"^(ax|test)is$", "${1}es"},
	[]string{"(octop|vir)us$", "${1}i"},
	[]string{"(octop|vir)i$", "${1}i"},
	[]string{"(alias|status)$", "${1}es"},
	[]string{"(bu)s$", "${1}ses"},
	[]string{"(buffal|tomat)o$", "${1}oes"},
	[]string{"([ti])um$", "${1}a"},
	[]string{"([ti])a$", "${1}a"},
	[]string{"sis$", "ses"},
	[]string{"(?:([^f])fe|([lr])f)$", "${1}${2}ves"},
	[]string{"(hive)$", "${1}s"},
	[]string{"([^aeiouy]|qu)y$", "${1}ies"},
	[]string{"(x|ch|ss|sh)$", "${1}es"},
	[]string{"(matr|vert|ind)(?:ix|ex)$", "${1}ices"},
	[]string{"^(m|l)ouse$", "${1}ice"},
	[]string{"^(m|l)ice$", "${1}ice"},
	[]string{"^(ox)$", "${1}en"},
	[]string{"^(oxen)$", "${1}"},
	[]string{"(quiz)$", "${1}zes"},
}

var singularInflections = [][]string{
	[]string{"s$", ""},
	[]string{"(ss)$", "${1}"},
	[]string{"(n)ews$", "${1}ews"},
	[]string{"([ti])a$", "${1}um"},
	[]string{"((a)naly|(b)a|(d)iagno|(p)arenthe|(p)rogno|(s)ynop|(t)he)(sis|ses)$", "${1}sis"},
	[]string{"(^analy)(sis|ses)$", "${1}sis"},
	[]string{"([^f])ves$", "${1}fe"},
	[]string{"(hive)s$", "${1}"},
	[]string{"(tive)s$", "${1}"},
	[]string{"([lr])ves$", "${1}f"},
	[]string{"([^aeiouy]|qu)ies$", "${1}y"},
	[]string{"(s)eries$", "${1}eries"},
	[]string{"(m)ovies$", "${1}ovie"},
	[]string{"(x|ch|ss|sh)es$", "${1}"},
	[]string{"^(m|l)ice$", "${1}ouse"},
	[]string{"(bus)(es)?$", "${1}"},
	[]string{"(o)es$", "${1}"},
	[]string{"(shoe)s$", "${1}"},
	[]string{"(cris|test)(is|es)$", "${1}is"},
	[]string{"^(a)x[ie]s$", "${1}xis"},
	[]string{"(octop|vir)(us|i)$", "${1}us"},
	[]string{"(alias|status)(es)?$", "${1}"},
	[]string{"^(ox)en", "${1}"},
	[]string{"(vert|ind)ices$", "${1}ex"},
	[]string{"(matr)ices$", "${1}ix"},
	[]string{"(quiz)zes$", "${1}"},
	[]string{"(database)s$", "${1}"},
}

var irregularInflections = [][]string{
	[]string{"person", "people"},
	[]string{"man", "men"},
	[]string{"child", "children"},
	[]string{"sex", "sexes"},
	[]string{"move", "moves"},
	[]string{"mombie", "mombies"},
}

var uncountableInflections = []string{"equipment", "information", "rice", "money", "species", "series", "fish", "sheep", "jeans", "police"}

type inflection struct {
	regexp  *regexp.Regexp
	replace string
}

var compiledPluralMaps []inflection
var compiledSingularMaps []inflection

func compile() {
	compiledPluralMaps = []inflection{}
	compiledSingularMaps = []inflection{}
	for _, uncountable := range uncountableInflections {
		inf := inflection{
			regexp:  regexp.MustCompile("^(?i)(" + uncountable + ")$"),
			replace: "${1}",
		}
		compiledPluralMaps = append(compiledPluralMaps, inf)
		compiledSingularMaps = append(compiledSingularMaps, inf)
	}

	for _, value := range irregularInflections {
		infs := []inflection{
			inflection{regexp: regexp.MustCompile(strings.ToUpper(value[0]) + "$"), replace: strings.ToUpper(value[1])},
			inflection{regexp: regexp.MustCompile(strings.Title(value[0]) + "$"), replace: strings.Title(value[1])},
			inflection{regexp: regexp.MustCompile(value[0] + "$"), replace: value[1]},
		}
		compiledPluralMaps = append(compiledPluralMaps, infs...)
	}

	for _, value := range irregularInflections {
		infs := []inflection{
			inflection{regexp: regexp.MustCompile(strings.ToUpper(value[1]) + "$"), replace: strings.ToUpper(value[0])},
			inflection{regexp: regexp.MustCompile(strings.Title(value[1]) + "$"), replace: strings.Title(value[0])},
			inflection{regexp: regexp.MustCompile(value[1] + "$"), replace: value[0]},
		}
		compiledSingularMaps = append(compiledSingularMaps, infs...)
	}

	for i := len(pluralInflections) - 1; i >= 0; i-- {
		value := pluralInflections[i]
		infs := []inflection{
			inflection{regexp: regexp.MustCompile(strings.ToUpper(value[0])), replace: strings.ToUpper(value[1])},
			inflection{regexp: regexp.MustCompile(value[0]), replace: value[1]},
			inflection{regexp: regexp.MustCompile("(?i)" + value[0]), replace: value[1]},
		}
		compiledPluralMaps = append(compiledPluralMaps, infs...)
	}

	for i := len(singularInflections) - 1; i >= 0; i-- {
		value := singularInflections[i]
		infs := []inflection{
			inflection{regexp: regexp.MustCompile(strings.ToUpper(value[0])), replace: strings.ToUpper(value[1])},
			inflection{regexp: regexp.MustCompile(value[0]), replace: value[1]},
			inflection{regexp: regexp.MustCompile("(?i)" + value[0]), replace: value[1]},
		}
		compiledSingularMaps = append(compiledSingularMaps, infs...)
	}
}

func init() {
	compile()
}

func AddPlural(key, value string) {
	pluralInflections = append(pluralInflections, []string{key, value})
	compile()
}

func AddSingular(key, value string) {
	singularInflections = append(singularInflections, []string{key, value})
	compile()
}

func AddIrregular(key, value string) {
	irregularInflections = append(irregularInflections, []string{key, value})
	compile()
}

func AddUncountable(value string) {
	uncountableInflections = append(uncountableInflections, value)
	compile()
}

func Plural(str string) string {
	for _, inflection := range compiledPluralMaps {
		if inflection.regexp.MatchString(str) {
			return inflection.regexp.ReplaceAllString(str, inflection.replace)
		}
	}
	return str
}

func Singular(str string) string {
	for _, inflection := range compiledSingularMaps {
		if inflection.regexp.MatchString(str) {
			return inflection.regexp.ReplaceAllString(str, inflection.replace)
		}
	}
	return str
}
