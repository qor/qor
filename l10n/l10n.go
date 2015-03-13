package l10n

import "github.com/jinzhu/gorm"

type Interface interface {
	IsGlobal() bool
	SetLocale(locale string)
}

type Locale struct {
	LanguageCode *string `sql:"size:6" gorm:"primary_key"`
}

func (l Locale) IsGlobal() bool {
	return l.LanguageCode == nil
}

func (l *Locale) SetLocale(locale string) {
	l.LanguageCode = &locale
}

func Localize(scope *gorm.Scope, global Interface, locale string) {
	// find deleted locale -> reset deleted at
	// sync attrs from global to locale
}
