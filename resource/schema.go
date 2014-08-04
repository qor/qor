package resource

import (
	"github.com/qor/qor"
	"github.com/qor/qor/rules"

	"reflect"
	"regexp"
)

func Decode(result interface{}, metas []Meta, context *qor.Context, prefix string) bool {
	request := context.Request

	var formKeys = []string{}
	for key := range request.Form {
		formKeys = append(formKeys, key)
	}

	if values, ok := request.Form[prefix+"_id"]; ok {
		primaryKey := values[0]
		context.DB.First(result, primaryKey)
		if destroyValues, ok := request.Form[prefix+"_destroy"]; ok {
			if destroyValues[0] != "0" {
				context.DB.Delete(result, primaryKey)
				return false
			}
		}
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
			reg := regexp.MustCompile("(" + prefix + meta.Name + `\[\d+\]\.)([^.]+)`)
			for _, key := range formKeys {
				matches := reg.FindStringSubmatch(key)
				if len(matches) == 3 && !matchedFormKeys[matches[1]] {
					matchedFormKeys[matches[1]] = true
					result := reflect.New(field.Type().Elem())
					if Decode(result.Interface(), metas, context, matches[1]) {
						if !reflect.DeepEqual(reflect.Zero(result.Type().Elem()).Interface(), result.Elem().Interface()) {
							field.Set(reflect.Append(field, result.Elem()))
						}
					}
				}
			}
		} else {
			key := prefix + meta.Name
			if _, ok := request.Form[key]; ok {
				meta.Setter(result, key, context)
			} else if request.MultipartForm != nil {
				if _, ok := request.MultipartForm.File[key]; ok {
					meta.Setter(result, key, context)
				}
			}
		}
	}
	return true
}
