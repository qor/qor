package media_library

import (
	"io"
	"os"
	"path/filepath"
)

type FileSystem struct {
	Base
}

func (f FileSystem) GetFullPath(url string, option *Option) (path string, err error) {
	if option != nil && option.Get("path") != "" {
		path = filepath.Join(option.Get("path"), url)
	} else {
		path = filepath.Join("./public", url)
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	}

	return
}

func (f FileSystem) Store(url string, option *Option, reader io.Reader) error {
	if fullpath, err := f.GetFullPath(url, option); err == nil {
		if dst, err := os.Create(fullpath); err == nil {
			_, err := io.Copy(dst, reader)
			return err
		} else {
			return err
		}
	} else {
		return err
	}
}

func (f FileSystem) Retrieve(url string) (*os.File, error) {
	if fullpath, err := f.GetFullPath(url, nil); err == nil {
		return os.Open(fullpath)
	} else {
		return nil, os.ErrNotExist
	}
}
