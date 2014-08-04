package media_library

import (
	"database/sql/driver"
	"errors"

	"github.com/jinzhu/gorm"

	"mime/multipart"
	"os"
)

var ErrNotImplemented = errors.New("not implemented")

type Base struct {
	Path       string
	Valid      bool
	Option     Option
	CropOption CropOption
	File       multipart.File
}

func (b *Base) Scan(value interface{}) error {
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

func (b Base) GetPath(value interface{}, column string, header *multipart.FileHeader) string {
	scope := gorm.Scope{Value: value}
	primaryKey := scope.PrimaryKeyValue()
	return column
}

func (b Base) Store(path string, file *os.File) error {
	b.Path, b.Valid = path, true
	b.File = file
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
