package media_library

import (
	"database/sql/driver"
	"io"
	"mime/multipart"
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
	GetPath(interface{}, string, *multipart.FileHeader) string

	Store(string, io.Reader) error
	Receive(filename string) (*os.File, error)
	Crop(CropOption) error

	Url(style ...string) string
	String() string
}
