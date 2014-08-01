package media_library

import (
	"database/sql/driver"
	"errors"
	"mime/multipart"
)

var NotImplementError = errors.New("not implemented")

type Base struct {
	Path       string
	CropOption CropOption
	Valid      bool
	file       multipart.File
}

func (b Base) Scan(value interface{}) error {
	if v, ok := value.(string); ok {
		b.Path, b.Valid = v, true
		return nil
	}
	return errors.New("scan value is not string")
}

func (b Base) Value() (driver.Value, error) {
	if b.Valid {
		return b.Path, nil
	}
	return nil, errors.New("file is invalid")
}

func (b Base) Store(path string, file multipart.File, header *multipart.FileHeader) error {
	if header.Filename != "" {
		b.Path, b.Valid = path, true
		b.file = file
		// save
	}
	return NotImplementError
}

func (Base) Receive(filename string) (multipart.File, error) {
	return nil, NotImplementError
}

func (Base) Crop(option CropOption) error {
	return NotImplementError
}

func (Base) Url(...string) string {
	return ""
}

func (b Base) String() string {
	return b.Url()
}
