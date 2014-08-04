package media_library

import (
	"database/sql/driver"
	"errors"
	"mime/multipart"
)

var ErrNotImplemented = errors.New("not implemented")

type Base struct {
	Option     Option
	Path       string
	CropOption CropOption
	Valid      bool
	File       multipart.File
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

func (b Base) Store(path string, header *multipart.FileHeader) error {
	if header.Filename != "" {
		b.Path, b.Valid = path, true
		if src, err := header.Open(); err == nil {
			b.File = src
			return nil
		} else {
			return err
		}
	}
	return ErrNotImplemented
}

func (Base) Receive(filename string) (multipart.File, error) {
	return nil, ErrNotImplemented
}

func (Base) Crop(option CropOption) error {
	return ErrNotImplemented
}

func (Base) Url(...string) string {
	return ""
}

func (b Base) String() string {
	return b.Url()
}

func (b Base) ParseOption(option string) {
}

func (b Base) GetOption() Option {
	return Option{}
}
