package admin

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"regexp"
	"strings"
)

func ConvertFormToMetaDatas(context *qor.Context, prefix string, res *Resource) (metaDatas resource.MetaDatas) {
	var convertedMap map[string]bool
	request := context.Request
	metas := res.Metas

	for key := range request.Form {
		if strings.HasSuffix(key, prefix) {
			isCurrent := regexp.MustCompile("(" + prefix + `([^\.]+))$`)
			isNext := regexp.MustCompile("(" + prefix + `([^\.\[\]]+)(\[\d+\])?)(\.[^\.]+)+$`)

			if matches := isCurrent.FindStringSubmatch(key); len(matches) > 0 {
				metaData := resource.MetaData{Name: matches[2], Value: request.Form[key], Meta: metas[matches[1]]}
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
