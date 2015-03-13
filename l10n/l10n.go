package l10n

type Interface interface {
	IsGlobal() bool
	SetLocale(locale string)
}

type Locale struct {
	LanguageCode string `sql:"size:6" gorm:"primary_key"`
}

func (l Locale) IsGlobal() bool {
	return l.LanguageCode == ""
}

func (l *Locale) SetLocale(locale string) {
	l.LanguageCode = locale
}
