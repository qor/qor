package resource

import (
	"github.com/qor/qor"
	"github.com/qor/qor/rules"

	"reflect"
	"regexp"
)

func Decode(result interface{}, metas []Meta, context *qor.Context, prefix string) {
	request := context.Request
	request.ParseMultipartForm(32 << 22)
	// request.MultipartForm

	var formKeys = []string{}
	for key := range request.Form {
		formKeys = append(formKeys, key)
	}

	if values, ok := request.Form[prefix+"_id"]; ok {
		context.DB.First(result, values[0])
	}

	for _, meta := range metas {
		if meta.Type == "single_edit" {
			metas := meta.Resource.AllowedMetas(meta.Resource.AllAttrs(), context, rules.Update)
			field := reflect.Indirect(reflect.ValueOf(result)).FieldByName(meta.Name)
			Decode(field.Addr().Interface(), metas, context, prefix+meta.Name+".")
		} else if meta.Type == "collection_edit" {
			metas := meta.Resource.AllowedMetas(meta.Resource.AllAttrs(), context, rules.Update)
			field := reflect.Indirect(reflect.ValueOf(result)).FieldByName(meta.Name)

			matchedFormKeys := map[string]bool{}
			reg := regexp.MustCompile(prefix + meta.Name + `\[\d+\]\.`)
			for _, key := range formKeys {
				matches := reg.FindStringSubmatch(key)
				if _, ok := matchedFormKeys[key]; !ok && len(matches) > 0 {
					matchedFormKeys[key] = true
					result := reflect.New(field.Type().Elem())
					Decode(result.Interface(), metas, context, matches[0])
					field.Set(reflect.Append(field, result.Elem()))
				}
			}
		} else {
			if values, ok := request.Form[prefix+meta.Name]; ok {
				meta.Setter(result, values[0], context)
			}
		}
	}
}
