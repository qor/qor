package media_library

import (
	"io"
	"mime/multipart"
	"os"
)

type FileSystem struct {
	Base
}

func (f FileSystem) Store(path string, fileHeader *multipart.FileHeader) error {
	if dst, err := os.Create(path); err == nil {
		if file, err := fileHeader.Open(); err == nil {
			_, err := io.Copy(dst, file)
			return err
		} else {
			return err
		}
	} else {
		return err
	}
}

func (f FileSystem) Receive(path string) (*os.File, error) {
	return os.Open(path)
}
