package admin

import (
	"encoding/json"
	"fmt"

	"github.com/qor/qor/media_library"
)

type AssetManager struct {
	ID   int
	File media_library.FileSystem
}

func (*AssetManager) InjectQorAdmin(res *Resource) {
	router := res.GetAdmin().GetRouter()
	router.Post(fmt.Sprintf("^/%v/upload", res.ToParam()), func(context *Context) {
		result := AssetManager{}
		if context.Request.MultipartForm != nil {
			result.File.Scan(context.Request.MultipartForm.File["file"])
		}
		context.GetDB().Save(&result)
		bytes, _ := json.Marshal(map[string]string{"filelink": result.File.URL(), "filename": result.File.GetFileName()})
		context.Writer.Write(bytes)
	})
}
