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
		if media, ok := field.Field.Addr().Interface().(MediaLibrary); ok {
			option := parseTagOption(field.Tag.Get("media_library"))

			// Store
			if media.GetFileHeader() != nil {
				if path := media.GetURLTemplate(option); path != "" {
					tmpl := template.New("").Funcs(getFuncMap(scope, field, media.GetFileName()))
					if tmpl, err := tmpl.Parse(path); err == nil {
						var result = bytes.NewBufferString("")
						if err := tmpl.Execute(result, scope.Value); err == nil {
							media.Scan(result.String())
							if file, err := media.GetFileHeader().Open(); err == nil {
								updateAttrs := map[string]interface{}{field.DBName: media.URL()}
								gorm.Update(scope.New(scope.Value).InstanceSet("gorm:update_attrs", updateAttrs))
								scope.Err(media.Store(media.URL("original"), option, file))
								scope.Err(media.Store(media.URL(), option, file))
							}
						} else {
							scope.Err(err)
						}
					}
				}
			}

			// Crop
			if !scope.HasError() && media.GetCropOption() != nil {
				media.Crop(media, option)
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
