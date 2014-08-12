package admin

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"regexp"
	"strings"
)

func ConvertFormToMetaValues(context *qor.Context, prefix string, res *Resource) (metaValues resource.MetaValues) {
	request := context.Request
	convertedMap := make(map[string]bool)
	metas := make(map[string]resource.Metaor)
	for _, attr := range res.AllAttrs() {
		metas[attr.Name] = attr
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
					children := ConvertFormToMetaValues(context, prefix+matches[1]+".", meta.GetMeta().Resource.(*Resource))
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
