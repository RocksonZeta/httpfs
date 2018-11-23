package pathmaker

import (
	"fmt"
	"httpfs/service/meta"
	"httpfs/util/hashutil"
	"path/filepath"
	"strconv"
)

func RandomPathMaker(params map[string]interface{}) func(fileName string, size int64) string {
	dirLimit := 200
	dirDepth := 2
	if nil != params {
		if p, ok := params["dirLimit"]; ok {
			dirLimit = p.(int)
		}
		if p, ok := params["dirDepth"]; ok {
			dirDepth = p.(int)
		}
	}
	return func(fileName string, size int64) string {
		ext := filepath.Ext(fileName)
		count := meta.GetMeta().Stat().Count
		return fmt.Sprintf("%s/%s/%s%s", GenDir(count, dirLimit, dirDepth), hashutil.RandomStr(10, true), hashutil.RandomStr(10, true), ext)
	}
}

func GenDir(fileTotalCount, dirLimit, dirDepth int) string {
	dirCounts := fileTotalCount
	p := ""
	for i := 0; i < dirDepth; i++ {
		dirCounts = dirCounts / dirLimit
		p += "/" + strconv.Itoa(dirCounts%dirLimit)
	}
	return p
}
