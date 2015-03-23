package l10n

import (
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

type L10nController struct{}

func (L10nController) Index(context *admin.Context) {
}

func (L10nController) Show(context *admin.Context) {
}

func AvailableLocales(*qor.Context) []string {
	return []string{}
}

func EditableLocales(*qor.Context) []string {
	return []string{}
}

func ViewableLocales(*qor.Context) []string {
	return []string{}
}
