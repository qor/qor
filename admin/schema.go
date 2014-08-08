package admin

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"regexp"
	"strings"
)

func ConvertFormToMetaDatas(context *qor.Context, prefix string, res *Resource) (metaDatas resource.MetaDatas) {
	request := context.Request
	convertedMap := make(map[string]bool)
	metas := make(map[string]resource.Metaor)
	for _, attr := range res.AllAttrs() {
		metas[attr.Name] = attr
	}

	for key := range request.Form {
		if strings.HasPrefix(key, prefix) {
			isCurrent := regexp.MustCompile(`(` + prefix + `([^.]+))$`)
			isNext := regexp.MustCompile(`(` + prefix + `([^.\[\]]+)(\[\d+\])?)(?:\.([^.]+))+$`)

			if matches := isCurrent.FindStringSubmatch(key); len(matches) > 0 {
				meta := metas[matches[1]]
				metaData := resource.MetaData{Name: matches[2], Value: request.Form[key], Meta: meta}
				metaDatas = append(metaDatas, metaData)
			} else if matches := isNext.FindStringSubmatch(key); len(matches) > 0 {
				if _, ok := convertedMap[matches[1]]; !ok {
					convertedMap[matches[1]] = true
					meta := metas[matches[2]]
					children := ConvertFormToMetaDatas(context, matches[1]+".", meta.GetMeta().Resource.(*Resource))
					metaData := resource.MetaData{Name: matches[2], Meta: meta, MetaDatas: children}
					metaData.MetaDatas = append(metaData.MetaDatas, metaData)
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
