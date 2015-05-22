package i18n

import "github.com/qor/qor/admin"

type I18nController struct {
	*I18n
}

func (controller *I18nController) Index(context *admin.Context) {
	context.Execute("index", controller.I18n)
}
