package media_library

import (
	"database/sql/driver"
	"errors"
	"io"
	"os"

	"mime/multipart"
)

var ErrNotImplemented = errors.New("not implemented")

type Base struct {
	Url        string
	Valid      bool
	FileName   string
	FileHeader *multipart.FileHeader
	CropOption *CropOption
	Reader     io.Reader
}

func (b Base) Scan(value interface{}) error {
	switch v := value.(type) {
	case *multipart.FileHeader:
		b.FileHeader, b.FileName, b.Valid = v, v.Filename, true
	case string:
		b.Url, b.Valid = v, true
	}
	return nil
}

func (b Base) Value() (driver.Value, error) {
	if b.Valid {
		return b.FileName, nil
	}
	return nil, nil
}

func (b Base) URL(...string) string {
	return b.Url
}

func (b Base) String() string {
	return b.URL()
}

func (b Base) GetFileName() string {
	return b.FileName
}

func (b Base) GetFileHeader() *multipart.FileHeader {
	return b.FileHeader
}

func (b Base) GetPathTemplate(tag string) (path string) {
	if path = parseTagSetting(tag)["url"]; path == "" {
		path = "/system/{class}/{primary_key}/{column}/{basename}.{nanotime}.{extension}"
	}
	return
}

func (b Base) SetCropOption(option *CropOption) {
	b.CropOption = option
}

func (b Base) Store(url string, file *multipart.FileHeader) error {
	return ErrNotImplemented
}

func (b Base) Retrieve(url string) (*os.File, error) {
	return nil, ErrNotImplemented
}

func (b Base) Crop() error {
	return ErrNotImplemented
}
