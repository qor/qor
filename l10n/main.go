package l10n

type locale struct {
	Name   string
	scope  string
	params map[string]string
}

func Locale(name string) *locale {
	return &locale{Name: name}
}

func (l *locale) clone() *locale {
	return &locale{Name: l.Name, scope: l.scope, params: l.params}
}

func (l *locale) Scope(name string) *locale {
	lc := l.clone()
	lc.scope = name
	return lc
}

func (l *locale) Params(params map[string]string) *locale {
	lc := l.clone()
	lc.params = params
	return lc
}

func (l *locale) T(name string, values ...string) string {
	return name
}
