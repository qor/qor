package media_library

import (
	"io"
	"os"
)

type FileSystem struct {
	Base
}

func (f FileSystem) fullpath(path string) string {
	return path
}

func (f FileSystem) Store(path string, src io.Reader) error {
	path = f.fullpath(path)

	if dst, err := os.Create(path); err == nil {
		f.FileName, f.Valid = path, true
		io.Copy(dst, src)
		return nil
	} else {
		return err
	}
}

func (f FileSystem) Receive(path string) (*os.File, error) {
	return os.Open(f.fullpath(path))
}

func (f FileSystem) Crop(option CropOption) error {
	return ErrNotImplemented
}

func (f FileSystem) Url(...string) string {
	return ""
}
