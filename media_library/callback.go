package media_library

import (
	"bytes"
	"text/template"

	"github.com/jinzhu/gorm"
)

func getFuncMap(scope *gorm.Scope, field *gorm.Field, filename string) template.FuncMap {
	return template.FuncMap{
		"class":       scope.TableName,
		"primary_key": scope.PrimaryKeyValue,
		"column":      field.Name,
		"filename":    filename,
	}
}

func SaveAndCropImage(scope *gorm.Scope) {
	for _, field := range scope.Fields() {
		if media, ok := field.Field.Interface().(MediaLibrary); ok && media.GetFileHeader() != nil {
			if path := media.GetPathTemplate(field.Tag.Get("media_library")); path != "" {
				if tmpl, err := template.New("").Parse(path); err == nil {
					var result = bytes.NewBufferString("")
					tmpl = tmpl.Funcs(getFuncMap(scope, field, media.GetFileName()))
					if err := tmpl.Execute(result, scope.Value); err != nil {
						filePath := result.String()
						scope.NewDB().Model(scope.Value).UpdateColumn(field.Name, filePath)
						media.Store(filePath, media.GetFileHeader())
					}
				}
			}
		}
	}
}

func init() {
	gorm.DefaultCallback.Update().Before("gorm:after_update").
		Register("media_library:save_and_crop", SaveAndCropImage)
	gorm.DefaultCallback.Create().After("gorm:after_create").
		Register("media_library:save_and_crop", SaveAndCropImage)
}
