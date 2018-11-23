package videoutil

import (
	"fmt"
	"httpfs/util/stringutil"
	"testing"
)

func TestGetVideoInfo(t *testing.T) {
	r, _ := GetVideoInfo("/Users/ququ/Movies/1.MOV")
	fmt.Println(stringutil.JsonPretty(r))
	fmt.Println(r.Width(), r.Height(), r.BitRate(), r.Duration(), r.VideoBitRate(), r.VideoCodec(), r.AudioBitRate(), r.AudioCodec())
}
