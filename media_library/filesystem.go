package media_library

import (
	"io"
	"mime/multipart"
	"os"
)

type FileSystem struct {
	Base
}

func (f FileSystem) Store(path string, header *multipart.FileHeader) error {
	if header.Filename != "" {
		f.Path, f.Valid = path, true
	}

	if dst, err := os.Create(path); err == nil {
		if src, err := header.Open(); err == nil {
			f.File = src
			io.Copy(dst, src)
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func (f FileSystem) Receive(filename string) (multipart.File, error) {
	return nil, ErrNotImplemented
}

func (f FileSystem) Crop(option CropOption) error {
	return ErrNotImplemented
}

func (f FileSystem) Url(...string) string {
	return ""
}
