package media_library

import (
	"encoding/json"
	"errors"
	"mime/multipart"

	"github.com/jinzhu/gorm"
)

func SaveAndCropImage(isCreate bool) func(scope *gorm.Scope) {
	return func(scope *gorm.Scope) {
		if !scope.HasError() {
			var updateColumns = map[string]interface{}{}
			for _, field := range scope.Fields() {
				if media, ok := field.Field.Addr().Interface().(MediaLibrary); ok && !media.Cropped() {

					option := parseTagOption(field.Tag.Get("media_library"))
					if media.GetFileHeader() != nil || media.NeedCrop() {
						var file multipart.File
						var err error
						if fileHeader := media.GetFileHeader(); fileHeader != nil {
							file, err = media.GetFileHeader().Open()
						} else {
							file, err = media.Retrieve(media.URL("original"))
						}

						if scope.Err(err) != nil {
							return
						}
						media.Cropped(true)

						if url := media.GetURL(option, scope, field, media); url == "" {
							scope.Err(errors.New("invalid URL"))
						} else {
							result, _ := json.Marshal(map[string]string{"Url": url})
							media.Scan(string(result))
						}

						if isCreate {
							if value, err := media.Value(); err == nil {
								updateColumns[field.DBName] = value
							}
						}

						if file != nil {
							defer file.Close()
							var handled = false
							for _, handler := range mediaLibraryHandlers {
								if handler.CouldHandle(media) {
									if scope.Err(handler.Handle(media, file, option)) == nil {
										handled = true
									}
								}
							}

							// Save File
							if !handled {
								scope.Err(media.Store(media.URL(), option, file))
							}
						}
					}
				}
			}

			if isCreate && !scope.HasError() && len(updateColumns) != 0 {
				scope.NewDB().Model(scope.Value).UpdateColumns(updateColumns)
			}
		}
	}
}

func init() {
	gorm.DefaultCallback.Update().Before("gorm:before_update").Register("media_library:save_and_crop", SaveAndCropImage(false))
	gorm.DefaultCallback.Create().After("gorm:after_create").Register("media_library:save_and_crop", SaveAndCropImage(true))
}
