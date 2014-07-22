package admin

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/rules"

	"reflect"
)

func Decode(result interface{}, metas []resource.Meta, context *qor.Context, prefix string) {
	request := context.Request
	request.ParseMultipartForm(32 << 22)
	// request.MultipartForm

	for _, meta := range metas {
		if meta.Type == "single_edit" {
			metas := meta.Resource.AllowedMetas(meta.Resource.AllAttrs(), context, rules.Update)
			field := reflect.Indirect(reflect.ValueOf(result)).FieldByName(meta.Name)
			Decode(field.Addr().Interface(), metas, context, prefix+meta.Name+".")
		} else {
			if values, ok := request.Form[prefix+meta.Name]; ok {
				meta.Setter(result, values[0], context)
			}
		}
	}
}
