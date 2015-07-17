package i18n

import (
	"log"

	"github.com/qor/qor/admin"
)

type i18nController struct {
	*I18n
}

func (controller *i18nController) Index(context *admin.Context) {
	context.Execute("index", controller.I18n)
}

func (controller *i18nController) Update(context *admin.Context) {
	form := context.Request.Form
	translation := Translation{Key: form.Get("Key"), Locale: form.Get("Locale"), Value: form.Get("Value")}

	if results := controller.I18n.Translations[translation.Locale]; results != nil {
		if result := results[translation.Key]; result != nil {
			translation.Backend = result.Backend
		}
	}

	err := controller.I18n.SaveTranslation(&translation)
	if err != nil {
		context.Flash(context.T("Could not save translation: {{.Key}}: {{.Error}}", translation, err), "failure")
		log.Println(err.Error, translation.Key)
	}
	context.Writer.Write([]byte("OK"))
}
