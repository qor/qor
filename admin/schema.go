package admin

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"regexp"
	"strings"
)

func ConvertMapToMetaValues(context *qor.Context, values map[string]interface{}, res *Resource) (metaValues resource.MetaValues) {
	metas := make(map[string]resource.Metaor)
	if res != nil {
		for _, attr := range res.AllAttrs() {
			metas[attr.Name] = attr
		}
	}

	for key, value := range values {
		meta := metas[key]
		if str, ok := value.(string); ok {
			metaValue := &resource.MetaValue{Name: key, Value: str, Meta: meta}
			metaValues.Values = append(metaValues.Values, metaValue)
		} else {
			var res *Resource
			if meta != nil && meta.GetMeta() != nil && meta.GetMeta().Resource != nil {
				res, _ = meta.GetMeta().Resource.(*Resource)
			}

			if vs, ok := value.(map[string]interface{}); ok {
				children := ConvertMapToMetaValues(context, vs, res)
				metaValue := &resource.MetaValue{Name: key, Meta: meta, MetaValues: children}
				metaValues.Values = append(metaValues.Values, metaValue)
			} else if vs, ok := value.([]interface{}); ok {
				for _, v := range vs {
					if mv, ok := v.(map[string]interface{}); ok {
						children := ConvertMapToMetaValues(context, mv, res)
						metaValue := &resource.MetaValue{Name: key, Meta: meta, MetaValues: children}
						metaValues.Values = append(metaValues.Values, metaValue)
					} else if meta != nil {
						metaValue := &resource.MetaValue{Name: key, Value: vs, Meta: meta}
						metaValues.Values = append(metaValues.Values, metaValue)
						break
					}
				}
			}
		}
	}
	return
}

func ConvertFormToMetaValues(context *qor.Context, prefix string, res *Resource) (metaValues resource.MetaValues) {
	request := context.Request
	convertedMap := make(map[string]bool)
	metas := make(map[string]resource.Metaor)
	if res != nil {
		for _, attr := range res.AllAttrs() {
			metas[attr.Name] = attr
		}
	}

	for key := range request.Form {
		if strings.HasPrefix(key, prefix) {
			key = strings.TrimPrefix(key, prefix)
			isCurrent := regexp.MustCompile("^[^.]+$")
			isNext := regexp.MustCompile(`^(([^.\[\]]+)(\[\d+\])?)(?:\.([^.]+)+)$`)

			if matches := isCurrent.FindStringSubmatch(key); len(matches) > 0 {
				meta := metas[matches[0]]
				metaValue := &resource.MetaValue{Name: matches[0], Value: request.Form[prefix+key], Meta: meta}
				metaValues.Values = append(metaValues.Values, metaValue)
			} else if matches := isNext.FindStringSubmatch(key); len(matches) > 0 {
				if _, ok := convertedMap[matches[1]]; !ok {
					convertedMap[matches[1]] = true
					meta := metas[matches[2]]
					var res *Resource
					if meta != nil && meta.GetMeta() != nil {
						res = meta.GetMeta().Resource.(*Resource)
					}
					children := ConvertFormToMetaValues(context, prefix+matches[1]+".", res)
					metaValue := &resource.MetaValue{Name: matches[2], Meta: meta, MetaValues: children}
					metaValues.Values = append(metaValues.Values, metaValue)
				}
			}
		}
	}

	if request.MultipartForm != nil {
		// for key, header := range request.MultipartForm.File {
		// xxxxx
		// }
	}
	return
}
