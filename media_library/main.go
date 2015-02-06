package media_library

import (
	"database/sql/driver"
	"mime/multipart"
	"strings"

	"os"
)

type CropOption struct {
	X      int
	Y      int
	Width  int
	Height int
}

type MediaLibrary interface {
	Scan(value interface{}) error
	Value() (driver.Value, error)

	GetPathTemplate(tag string) string

	GetFileHeader() *multipart.FileHeader
	GetFileName() string
	SetCropOption(*CropOption)

	Store(url string, fileHeader *multipart.FileHeader) error
	Retrieve(url string) (*os.File, error)

	URL(style ...string) string
	String() string
}

func parseTagSetting(str string) map[string]string {
	tags := strings.Split(str, ";")
	setting := map[string]string{}
	for _, value := range tags {
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) == 2 {
			setting[k] = v[1]
		} else {
			setting[k] = k
		}
	}
	return setting
}
