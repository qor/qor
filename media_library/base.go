package media_library

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/disintegration/imaging"
	"github.com/jinzhu/gorm"
)

var ErrNotImplemented = errors.New("not implemented")

type Base struct {
	Url        string
	Valid      bool
	FileName   string
	FileHeader *multipart.FileHeader
	CropOption *image.Rectangle
	Reader     io.Reader
}

func (b *Base) Scan(value interface{}) error {
	switch v := value.(type) {
	case []*multipart.FileHeader:
		if len(v) > 0 {
			file := v[0]
			b.FileHeader, b.FileName, b.Valid = file, file.Filename, true
		}
	case []uint8:
		b.Url, b.Valid = string(v), true
	case string:
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			var cropOption struct{ X, Y, Width, Height int }
			if err := json.Unmarshal([]byte(v), &cropOption); err == nil {
				b.SetCropOption(image.Rectangle{
					Min: image.Point{X: cropOption.X, Y: cropOption.Y},
					Max: image.Point{X: cropOption.X + cropOption.Width, Y: cropOption.Y + cropOption.Height},
				})
			} else {
				return err
			}
		} else {
			b.Url, b.Valid = v, true
		}
	default:
		fmt.Errorf("unsupported driver -> Scan pair for MediaLibrary")
	}
	return nil
}

func (b Base) Value() (driver.Value, error) {
	if b.Valid {
		return b.Url, nil
	}
	return nil, nil
}

func (b Base) URL(styles ...string) string {
	if len(styles) > 0 {
		ext := path.Ext(b.Url)
		return fmt.Sprintf("%v.%v%v", strings.TrimSuffix(b.Url, ext), styles[0], ext)
	}
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

func (b Base) GetURLTemplate(option *Option) (path string) {
	if path = option.Get("URL"); path == "" {
		path = "/system/{{class}}/{{primary_key}}/{{column}}/{{filename_with_hash}}"
	}
	return
}

func getFuncMap(scope *gorm.Scope, field *gorm.Field, filename string) template.FuncMap {
	hash := func() string { return strings.Replace(time.Now().Format("20060102150506.000000000"), ".", "", -1) }
	return template.FuncMap{
		"class":       scope.TableName,
		"primary_key": func() string { return fmt.Sprintf("%v", scope.PrimaryKeyValue()) },
		"column":      func() string { return field.Name },
		"filename":    func() string { return filename },
		"basename":    func() string { return strings.TrimSuffix(path.Base(filename), path.Ext(filename)) },
		"hash":        hash,
		"filename_with_hash": func() string {
			return fmt.Sprintf("%v.%v%v", strings.TrimSuffix(filename, path.Ext(filename)), hash(), path.Ext(filename))
		},
		"extension": func() string { return strings.TrimPrefix(path.Ext(filename), ".") },
	}
}

func (b Base) GetURL(option *Option, scope *gorm.Scope, field *gorm.Field) string {
	if path := b.GetURLTemplate(option); path != "" {
		tmpl := template.New("").Funcs(getFuncMap(scope, field, b.GetFileName()))
		if tmpl, err := tmpl.Parse(path); err == nil {
			var result = bytes.NewBufferString("")
			if err := tmpl.Execute(result, scope.Value); err == nil {
				return result.String()
			}
		}
	}
	return ""
}

func (b *Base) SetCropOption(option image.Rectangle) {
	b.CropOption = &option
}

func (b *Base) GetCropOption() *image.Rectangle {
	return b.CropOption
}

func (b Base) Retrieve(url string) (*os.File, error) {
	return nil, ErrNotImplemented
}

func (b Base) GetSizes() map[string]Size {
	return map[string]Size{}
}

func (b Base) IsImage() bool {
	_, err := getImageFormat(b.URL())
	return err == nil
}

func getImageFormat(url string) (*imaging.Format, error) {
	formats := map[string]imaging.Format{
		".jpg":  imaging.JPEG,
		".jpeg": imaging.JPEG,
		".png":  imaging.PNG,
		".tif":  imaging.TIFF,
		".tiff": imaging.TIFF,
		".bmp":  imaging.BMP,
		".gif":  imaging.GIF,
	}

	ext := strings.ToLower(filepath.Ext(url))
	if f, ok := formats[ext]; ok {
		return &f, nil
	} else {
		return nil, imaging.ErrUnsupportedFormat
	}
}
