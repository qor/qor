package media_library

import (
	"database/sql/driver"

	"github.com/jinzhu/gorm"

	"io"
	"os"
)

type Option map[string]string

type CropOption struct {
	X      int
	Y      int
	Width  int
	Height int
}

type MediaLibrary interface {
	Scan(value interface{}) error
	Value() (driver.Value, error)

	GetOption() Option
	ParseOption(string)
	GetPath(resource interface{}, column string, filename string) string

	SetFile(filename string, reader io.Reader)
	GetFile() io.Reader
	GetFileName() string
	SetCropOption(CropOption)

	Store(string, io.Reader) error
	Retrieve(filename string) (*os.File, error)

	Url(style ...string) string
	String() string
}

func SaveAndCropImage(scope *gorm.Scope) {
	for _, field := range scope.Fields() {
		if media, ok := field.Field.Interface().(MediaLibrary); ok {
			media.ParseOption(field.Tag.Get("media_library"))
			path := media.GetPath(scope.Value, field.Name, media.GetFileName())
			media.Store(path, media.GetFile())
		}
	}
}

func init() {
	// gorm.DefaultCallback.Update().After("gorm:save_after_associations").
	// 	Register("media_library:save_and_crop", SaveAndCropImage)
	// gorm.DefaultCallback.Create().After("gorm:save_after_associations").
	// 	Register("media_library:save_and_crop", SaveAndCropImage)
}
