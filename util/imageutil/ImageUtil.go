package imageutil

import (
	"bytes"
	"image"
	"io"
	"math"
	"os"

	"github.com/disintegration/imaging"
)

type ImageTransform struct {
	Im     image.Image
	Format imaging.Format
}

func NewImageFileTransform(filename string) (*ImageTransform, error) {
	format, err := imaging.FormatFromFilename(filename)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewImageTransform(f, format)
}
func NewImageTransform(file io.Reader, format imaging.Format) (*ImageTransform, error) {
	t := &ImageTransform{Format: format}
	im, err := imaging.Decode(file)
	if nil != err {
		return nil, err
	}
	t.Im = im
	return t, nil
}
func (t *ImageTransform) CropRect(rect image.Rectangle) *ImageTransform {
	t.Im = imaging.Crop(t.Im, rect)
	return t
}
func (t *ImageTransform) Crop(x, y, w, h int) *ImageTransform {
	t.Im = imaging.Crop(t.Im, image.Rect(x, y, x+w, y+h))
	return t
}
func (t *ImageTransform) Resize(w, h int) *ImageTransform {
	t.Im = imaging.Resize(t.Im, w, h, imaging.Linear)
	return t
}
func (t *ImageTransform) ResizeKeepRatio(w, h int) *ImageTransform {
	ow, oh := t.Size()
	rw := float64(w) / float64(ow)
	rh := float64(h) / float64(oh)
	x, y := 0, 0
	nw, nh := w, h
	if math.Abs(rw-1) < math.Abs(rh-1) { //选择压缩比最小的 ，
		nh = int(float64(oh) * rw)
		y = (h - nh) / 2
		if y < 0 {
			y = -y
		}
	} else {
		nw = int(float64(ow) * rh)
		x = (w - nw) / 2
		if x < 0 {
			x = -x
		}
	}
	t.Im = imaging.Resize(t.Im, nw, nh, imaging.Linear)
	t.Crop(x, y, w, h)
	return t
}
func (t *ImageTransform) Size() (int, int) {
	return t.Im.Bounds().Dx(), t.Im.Bounds().Dy()
}
func (t *ImageTransform) Buffer() (*bytes.Buffer, error) {
	buff := new(bytes.Buffer)
	err := imaging.Encode(buff, t.Im, t.Format)
	if nil != err {
		return nil, err
	}
	return buff, nil
}
func (t *ImageTransform) Save(filePath string) error {
	return imaging.Save(t.Im, filePath)
}
func (t *ImageTransform) Write(filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if nil != err {
		return err
	}
	defer f.Close()
	src, err := t.Buffer()
	if nil != err {
		return err
	}
	io.Copy(f, src)
	return nil
}

// func GetImageFormat(filename string) (imaging.Format, error) {
// 	formats := map[string]imaging.Format{
// 		".jpg":  imaging.JPEG,
// 		".jpeg": imaging.JPEG,
// 		".png":  imaging.PNG,
// 		".tif":  imaging.TIFF,
// 		".tiff": imaging.TIFF,
// 		".bmp":  imaging.BMP,
// 		".gif":  imaging.GIF,
// 	}

// 	ext := strings.ToLower(filepath.Ext(filename))
// 	f, ok := formats[ext]
// 	if !ok {
// 		return 0, imaging.ErrUnsupportedFormat
// 	}
// 	return f, nil
// }
