package media_library

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"

	"github.com/jinzhu/gorm"

	"mime/multipart"
)

var ErrNotImplemented = errors.New("not implemented")

type Base struct {
	Path       string
	Valid      bool
	Option     Option
	CropOption CropOption
	File       multipart.File
}

func (b Base) Scan(value interface{}) error {
	if v, ok := value.(string); ok {
		b.Path, b.Valid = v, true
		return nil
	}
	return nil
}

func (b Base) Value() (driver.Value, error) {
	if b.Valid {
		return b.Path, nil
	}
	return nil, nil
}

func (b Base) GetOption() Option {
	return Option{}
}

func (b Base) ParseOption(option string) {
}

func (b Base) GetPath(value interface{}, column string, header *multipart.FileHeader) string {
	scope := gorm.Scope{Value: value}
	// ":model_name/:column_name/:primary_key/:filename"
	kind := reflect.Indirect(reflect.ValueOf(value)).Type().Name()
	primaryKey := fmt.Sprintf("%v", scope.PrimaryKeyValue())
	filename := header.Filename
	return fmt.Sprintf(path.Join("/tmp", kind, column, primaryKey, filename))
}

func (b Base) Store(path string, src io.Reader) error {
	b.Path, b.Valid = path, true
	return ErrNotImplemented
}

func (Base) Receive(filename string) (*os.File, error) {
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
