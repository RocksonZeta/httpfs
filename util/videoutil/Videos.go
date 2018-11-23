package videoutil

type Videos struct {
	mp4 string
}

/*
考虑：
vcodec,acodec,vbr,abr,width,height
1.h264,aac OK
vbr,abr在合理范围内，直接resize,音频copy
2.abr>标准
直接压缩音频
直接使用ffmpeg
=====
3.分离音视频，启用x264,压缩视频，压缩音频if
合并

*/
