package vips

import (
	"bytes"
	"io"
	"mime/multipart"

	"github.com/qor/qor/media_library"

	"gopkg.in/h2non/bimg.v0"
)

type bimgImageHandler struct{}

func (bimgImageHandler) CouldHandle(media media_library.MediaLibrary) bool {
	return media.IsImage()
}

func (bimgImageHandler) Handle(media media_library.MediaLibrary, file multipart.File, option *media_library.Option) error {
	// Save Original Image
	if err := media.Store(media.URL("original"), option, file); err == nil {
		file.Seek(0, 0)

		// Crop & Resize
		var buffer bytes.Buffer
		if _, err := io.Copy(&buffer, file); err != nil {
			return err
		}

		img := bimg.NewImage(buffer.Bytes())

		// Handle original image
		{
			bimgOption := bimg.Options{Interlace: true}

			// Crop original image if specified
			if cropOption := media.GetCropOption("original"); cropOption != nil {
				bimgOption.Top = cropOption.Min.Y
				bimgOption.Left = cropOption.Min.X
				bimgOption.AreaWidth = cropOption.Max.X - cropOption.Min.X
				bimgOption.AreaHeight = cropOption.Max.Y - cropOption.Min.Y
			}

			// Process & Save original image
			if buf, err := img.Process(bimgOption); err == nil {
				media.Store(media.URL(), option, bytes.NewReader(buf))
			} else {
				return err
			}
		}

		// Handle size images
		for key, size := range media.GetSizes() {
			img := bimg.NewImage(buffer.Bytes())

			bimgOption := bimg.Options{
				Interlace: true,
			}

			if cropOption := media.GetCropOption(key); cropOption != nil {
				bimgOption.Top = cropOption.Min.Y
				bimgOption.Left = cropOption.Min.X
				bimgOption.AreaWidth = cropOption.Max.X - cropOption.Min.X
				bimgOption.AreaHeight = cropOption.Max.Y - cropOption.Min.Y
				bimgOption.Crop = true
				bimgOption.Force = true
			}

			// Process & Save size image
			if _, err := img.Process(bimgOption); err == nil {
				if buf, err := img.Process(bimg.Options{
					Width:   size.Width,
					Height:  size.Height,
					Crop:    true,
					Enlarge: true,
					Force:   true,
				}); err == nil {
					media.Store(media.URL(key), option, bytes.NewReader(buf))
				} else {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	} else {
		return err
	}
}

func init() {
	media_library.RegisterMediaLibraryHandler("image_handler", bimgImageHandler{})
}
