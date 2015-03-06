package media_library

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/disintegration/imaging"
	"github.com/jinzhu/gorm"
)

func SaveAndCropImage(isCreate bool) func(scope *gorm.Scope) {
	return func(scope *gorm.Scope) {
		for _, field := range scope.Fields() {
			if media, ok := field.Field.Addr().Interface().(MediaLibrary); ok {
				option := parseTagOption(field.Tag.Get("media_library"))
				if media.GetFileHeader() != nil || media.GetCropOption() != nil {
					var file multipart.File
					fmt.Println("---")
					if fileHeader := media.GetFileHeader(); fileHeader != nil {
						file, _ = media.GetFileHeader().Open()
					} else {
						file, _ = media.Retrieve(media.URL("original"))
						fmt.Println("---2")
					}
					fmt.Println("---3")

					if url := media.GetURL(option, scope, field); url == "" {
						scope.Err(errors.New("invalid URL"))
					} else {
						result, _ := json.Marshal(map[string]string{"Url": url})
						media.Scan(string(result))
					}

					if isCreate && !scope.HasError() {
						if value, err := media.Value(); err == nil {
							gorm.Update(scope.New(scope.Value).InstanceSet("gorm:update_attrs", map[string]interface{}{field.DBName: value}))
						}
					}

					if file != nil {
						defer file.Close()

						if media.IsImage() {
							// Save Original Image
							fmt.Println(media.URL("original"))
							if scope.Err(media.Store(media.URL("original"), option, file)) == nil {
								file.Seek(0, 0)

								// Crop & Resize
								if img, err := imaging.Decode(file); scope.Err(err) == nil {
									if format, err := getImageFormat(media.URL()); scope.Err(err) == nil {
										if cropOption := media.GetCropOption(); cropOption != nil {
											img = imaging.Crop(img, *cropOption)
										}

										// Save default image
										var buffer bytes.Buffer
										imaging.Encode(&buffer, img, *format)
										media.Store(media.URL(), option, &buffer)

										for key, size := range media.GetSizes() {
											dst := imaging.Resize(img, size.Width, size.Height, imaging.Lanczos)
											var buffer bytes.Buffer
											imaging.Encode(&buffer, dst, *format)
											media.Store(media.URL(key), option, &buffer)
										}
									}
								}
							}
						} else {
							// Save File
							scope.Err(media.Store(media.URL(), option, file))
						}
					}
				}
			}
		}
	}
}

func init() {
	gorm.DefaultCallback.Update().Before("gorm:before_update").Register("media_library:save_and_crop", SaveAndCropImage(false))
	gorm.DefaultCallback.Create().After("gorm:after_create").Register("media_library:save_and_crop", SaveAndCropImage(true))
}
