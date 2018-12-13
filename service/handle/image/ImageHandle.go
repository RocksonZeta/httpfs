package image

import (
	"encoding/json"
	"httpfs/base"
	"httpfs/base/log"
	"httpfs/service/meta"
	"httpfs/service/pathmaker"
	"httpfs/util/imageutil"
	"httpfs/util/stringutil"
	"os"
	"path/filepath"
	"strconv"
)

type ImageHandler struct {
}

func (h *ImageHandler) Do(method, params string) (interface{}, error) {
	if method == "cropresize" {
		var param ImageTransformParam
		err := json.Unmarshal([]byte(params), &param)
		if err != nil {
			return nil, err
		}
		return h.CropResize(param)
	}
	return nil, nil
}

type ImageTransformParam struct {
	FilePath string
	Crop     []int
	Resize   [][]int
}

//CropResize : Crop->Resizes
func (h *ImageHandler) CropResize(param ImageTransformParam) ([]string, error) {
	log.Log.Debug("ImageHandler.CropResize - param:", param)
	absPath, err := base.AbsPath(param.FilePath)
	if err != nil {
		return nil, err
	}
	trans, err := imageutil.NewImageFileTransform(absPath)
	if err != nil {
		return nil, err
	}
	fcount := len(param.Resize)
	hasCrop := false
	if len(param.Crop) > 0 {
		hasCrop = true
		fcount++
	}
	files := pathmaker.SubFiles(param.FilePath, fcount)
	// resizeBuf, err := trans.Buffer()
	// if err != nil {
	// 	return nil, err
	// }
	// resizeBytes := resizeBuf.Bytes()
	oldFileMeta := meta.GetMeta().Get(param.FilePath)
	fileName := filepath.Base(param.FilePath)
	if oldFileMeta != nil && oldFileMeta.FileName != "" {
		fileName = oldFileMeta.FileName
	}
	for i, file := range files {
		absFile, _ := base.AbsPath(file)
		if i == 0 && hasCrop {
			crop := param.Crop
			err := trans.Crop(crop[0], crop[1], crop[2], crop[3]).Save(absFile)
			if err != nil {
				return nil, err
			}
			stat, err := os.Stat(absFile)
			if err != nil {
				return nil, err
			}
			err = meta.GetMeta().Register(file, meta.FileMeta{FileName: stringutil.FileNameAppend(fileName, "_"+strconv.Itoa(i+1)), Size: stat.Size()})
			log.Log.Error(err)
			continue
		}
		j := i
		if hasCrop { //crop减一
			j = j - 1
		}
		// b := bytes.NewBuffer(resizeBytes)
		// t, err := imageutil.NewImageTransform(b, trans.Format)
		if err != nil {
			return nil, err
		}
		sizes := param.Resize
		err = trans.ResizeKeepRatio(sizes[j][0], sizes[j][1]).Save(absFile)
		if err != nil {
			return nil, err
		}
		stat, err := os.Stat(absFile)
		if err != nil {
			return nil, err
		}
		err = meta.GetMeta().Register(file, meta.FileMeta{FileName: stringutil.FileNameAppend(fileName, "_"+strconv.Itoa(i+1)), Size: stat.Size()})
		log.Log.Error(err)
	}
	return files, nil
}
