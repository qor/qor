package admin

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/qor/qor/media_library"
)

type AssetManager struct {
	ID   int
	File media_library.FileSystem `media_library:"URL:/system/asset_manager/{{primary_key}}/{{filename_with_hash}}"`
}

func (*AssetManager) InjectQorAdmin(res *Resource) {
	router := res.GetAdmin().GetRouter()
	router.Post(fmt.Sprintf("^/%v/upload", res.ToParam()), func(context *Context) {
		result := AssetManager{}
		result.File.Scan(context.Request.MultipartForm.File["file"])
		context.GetDB().Save(&result)
		bytes, _ := json.Marshal(map[string]string{"filelink": result.File.URL(), "filename": result.File.GetFileName()})
		context.Writer.Write(bytes)
	})

	assetURL := regexp.MustCompile(`^/system/asset_manager/(\d+)/`)
	router.Post(fmt.Sprintf("^/%v/crop", res.ToParam()), func(context *Context) {
		var err error
		var cropOption struct{ url, option string }
		defer context.Request.Body.Close()
		if err = json.NewDecoder(context.Request.Body).Decode(&cropOption); err == nil {
			if matches := assetURL.FindStringSubmatch(cropOption.url); len(matches) > 1 {
				result := AssetManager{}
				if err = context.GetDB().Find(&result, matches[1]).Error; err == nil {
					if err = result.File.Scan(cropOption.option); err == nil {
						if err = context.GetDB().Save(result).Error; err == nil {
							bytes, _ := json.Marshal(map[string]string{"url": result.File.URL(), "filename": result.File.GetFileName()})
							context.Writer.Write(bytes)
						}
					}
				}
			}
		}

		bytes, _ := json.Marshal(map[string]string{"err": err.Error()})
		context.Writer.Write(bytes)
	})
}
