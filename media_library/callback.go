package media_library

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/jinzhu/gorm"
)

func getFuncMap(scope *gorm.Scope, field *gorm.Field, filename string) template.FuncMap {
	return template.FuncMap{
		"class":       scope.TableName,
		"primary_key": func() string { return fmt.Sprintf("%v", scope.PrimaryKeyValue()) },
		"column":      func() string { return field.Name },
		"filename":    func() string { return filename },
		"basename":    func() string { return strings.TrimSuffix(path.Base(filename), path.Ext(filename)) },
		"nanotime":    func() string { return strings.Replace(time.Now().Format("20060102150506.000000000"), ".", "", -1) },
		"extension":   func() string { return strings.TrimPrefix(path.Ext(filename), ".") },
	}
}

func SaveAndCropImage(scope *gorm.Scope) {
	for _, field := range scope.Fields() {
		if media, ok := field.Field.Addr().Interface().(MediaLibrary); ok && media.GetFileHeader() != nil {
			tag := field.Tag.Get("media_library")
			if path := media.GetURLTemplate(tag); path != "" {
				tmpl := template.New("").Funcs(getFuncMap(scope, field, media.GetFileName()))
				if tmpl, err := tmpl.Parse(path); err == nil {
					var result = bytes.NewBufferString("")
					if err := tmpl.Execute(result, scope.Value); err == nil {
						url := result.String()
						updateAttrs := map[string]interface{}{field.Name: url}
						gorm.Update(scope.New(scope.Value).InstanceSet("gorm:update_attrs", updateAttrs))
						if scope.Err(media.Store(url, parseTagOption(tag), media.GetFileHeader())) == nil {
							media.Crop(media)
						}
					} else {
						scope.Err(err)
					}
				} else {
					scope.Err(err)
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
