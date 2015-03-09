package localization

type Interface interface {
	IsGlobal() bool
	SetLocale(language string)
}

type Localization struct {
	LangaugeCode *string
}

func (localization Localization) IsGlobal() bool {
	return localization.LangaugeCode == nil
}

func (localization Localization) SetLocale(language string) {
	localization.LangaugeCode = &language
}
