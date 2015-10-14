package media_library

import (
	"bytes"
	"mime/multipart"

	"github.com/disintegration/imaging"
)

var mediaLibraryHandlers = make(map[string]MediaLibraryHandler)

type MediaLibraryHandler interface {
	CouldHandle(media MediaLibrary) bool
	Handle(media MediaLibrary, file multipart.File, option *Option) error
}

func RegisterMediaLibraryHandler(name string, handler MediaLibraryHandler) {
	mediaLibraryHandlers[name] = handler
}

// Register default image handler
type imageHandler struct{}

func (imageHandler) CouldHandle(media MediaLibrary) bool {
	return media.IsImage()
}

func (imageHandler) Handle(media MediaLibrary, file multipart.File, option *Option) error {
	if err := media.Store(media.URL("original"), option, file); err == nil {
		file.Seek(0, 0)

		if img, err := imaging.Decode(file); err == nil {
			if format, err := getImageFormat(media.URL()); err == nil {
				if cropOption := media.GetCropOption("original"); cropOption != nil {
					img = imaging.Crop(img, *cropOption)
				}

				// Save default image
				var buffer bytes.Buffer
				imaging.Encode(&buffer, img, *format)
				media.Store(media.URL(), option, &buffer)

				for key, size := range media.GetSizes() {
					newImage := img
					if cropOption := media.GetCropOption(key); cropOption != nil {
						newImage = imaging.Crop(newImage, *cropOption)
					}

					dst := imaging.Thumbnail(newImage, size.Width, size.Height, imaging.Lanczos)
					var buffer bytes.Buffer
					imaging.Encode(&buffer, dst, *format)
					media.Store(media.URL(key), option, &buffer)
				}
				return nil
			} else {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func init() {
	RegisterMediaLibraryHandler("image_handler", imageHandler{})
}
