package videoutil

import (
	"encoding/json"
	"httpfs/util/osutil"
	"strconv"
	"strings"
)

type VideoInfo struct {
	Video  map[string]interface{}
	Audio  map[string]interface{}
	Format map[string]interface{}
}

func GetVideoInfo(media string) (VideoInfo, error) {

	cmd := "ffprobe -v quiet -print_format json -show_format -show_streams " + media
	// state, stdout, stderr = run(c)
	// video_info = json.loads(stdout)
	// streams = {s['codec_type'] : s for s in video_info['streams']}
	// format = video_info['format']
	info := VideoInfo{}
	_, stdout, _, err := osutil.Exec(cmd)
	if err != nil {
		return info, err
	}
	// fmt.Println(stdout, err)
	r := make(map[string]interface{})
	json.Unmarshal([]byte(stdout), &r)
	info.Format = r["format"].(map[string]interface{})
	streams := r["streams"].([]interface{})
	for i, s := range streams {
		v := s.(map[string]interface{})
		if "video" == v["codec_type"] {
			info.Video = streams[i].(map[string]interface{})
		}
		if "audio" == v["codec_type"] {
			info.Audio = streams[i].(map[string]interface{})
		}
	}
	// return streams, format
	return info, nil
}

func (v VideoInfo) Size() int64 {
	if v.Format != nil {
		r, _ := strconv.ParseInt(v.Format["size"].(string), 10, 64)
		return r
	}
	return 0
}
func (v VideoInfo) Duration() float64 {
	if v.Format != nil {
		r, _ := strconv.ParseFloat(v.Format["duration"].(string), 64)
		return r
	}
	return 0
}
func (v VideoInfo) BitRate() int {
	if v.Format != nil {
		r, _ := strconv.Atoi(v.Format["bit_rate"].(string))
		return r
	}
	return 0
}
func (v VideoInfo) Width() int {
	if v.Video != nil {
		return int(v.Video["width"].(float64))
	}
	return 0
}
func (v VideoInfo) Height() int {
	if v.Video != nil {
		return int(v.Video["height"].(float64))
	}
	return 0
}
func (v VideoInfo) VideoBitRate() int {
	if v.Video != nil {
		r, _ := strconv.Atoi(v.Video["bit_rate"].(string))
		return r
	}
	return 0
}
func (v VideoInfo) VideoCodec() string {
	if v.Video != nil {
		return strings.ToLower(v.Video["codec_name"].(string))
	}
	return ""
}
func (v VideoInfo) AudioBitRate() int {
	if v.Audio != nil {
		r, _ := strconv.Atoi(v.Audio["bit_rate"].(string))
		return r
	}
	return 0
}
func (v VideoInfo) AudioCodec() string {
	if v.Audio != nil {
		return strings.ToLower(v.Audio["codec_name"].(string))
	}
	return ""
}
func (v VideoInfo) IsH264() bool {
	return "h264" == v.VideoCodec()
}
func (v VideoInfo) IsAac() bool {
	return "aac" == v.AudioCodec()
}
