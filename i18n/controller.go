package i18n

import "github.com/qor/qor/admin"

type I18nController struct {
	*I18n
}

func (controller *I18nController) Index(context *admin.Context) {
	context.Execute("index", controller.I18n)
}

func (controller *I18nController) Update(context *admin.Context) {
	var translation Translation
	context.Resource.Decode(context, &translation)
	if results := controller.I18n.Translations[translation.Locale]; results != nil {
		if result := results[translation.Key]; result != nil {
			translation.Backend = result.Backend
		}
	}

	controller.I18n.SaveTransaltion(&translation)
	context.Writer.Write([]byte("OK"))
}
