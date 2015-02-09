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

	GetURLTemplate(tag string) string

	GetFileHeader() *multipart.FileHeader
	GetFileName() string
	SetCropOption(*CropOption)
	Crop() error

	Store(url string, option *Option, fileHeader *multipart.FileHeader) error
	Retrieve(url string) (*os.File, error)

	URL(style ...string) string
	String() string
}

type Option map[string]string

func (option Option) Get(key string) string {
	return option[key]
}

func parseTagOption(str string) *Option {
	tags := strings.Split(str, ";")
	setting := Option{}
	for _, value := range tags {
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) == 2 {
			setting[k] = v[1]
		} else {
			setting[k] = k
		}
	}
	return &setting
}
