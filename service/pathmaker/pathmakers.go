package pathmaker

import (
	"httpfs/util/stringutil"
	"strconv"
)

var PathMakers map[string]func(map[string]interface{}) func(fileName string, size int64) string = make(map[string]func(map[string]interface{}) func(fileName string, size int64) string)

func init() {
	PathMakers["randomMaker"] = RandomPathMaker
}

func SubFiles(fp string, n int) []string {
	r := make([]string, n)
	for i := 1; i <= n; i++ {
		r[i-1] = stringutil.FileNameAppend(fp, "_"+strconv.Itoa(i))
	}
	return r
}
