package imageutil

import (
	"testing"
)

func TestImageTransform(t *testing.T) {
	it, err := NewImageFileTransform("/Users/ququ/projects/go/src/good/static/img/demo/i1.jpg")
	if err != nil {
		t.Error(err)
	}
	it.Crop(10, 10, 100, 100).Resize(200, 200).Write("b.jpg", 80)
}
