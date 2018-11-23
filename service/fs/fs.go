package fs

import (
	"errors"
	"fmt"
	"httpfs/base"
	"httpfs/service/meta"
	"httpfs/service/pathmaker"
	"httpfs/util/imageutil"
	"httpfs/util/stringutil"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"zipx"

	"github.com/go-cmd/cmd"
)

// func base.AbsPath(relativePath string) (string, error) {
// 	return filepath.Abs(filepath.Join(base.Config.Fs.Root, relativePath))
// }
func MkDir(relativePath string) error {
	absPath, err := base.AbsPath(relativePath)
	fmt.Println("MkDir - ", absPath)
	if err != nil {
		return err
	}
	if stat, err := os.Stat(absPath); err == nil && stat.IsDir() {
		return nil
	}
	return os.MkdirAll(absPath, 0755)
}

func Read(srcPath string, dst io.Writer) (int64, error) {
	absPath, err := base.AbsPath(srcPath)
	if err != nil {
		return 0, err
	}
	src, err := os.OpenFile(absPath, os.O_RDONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer src.Close()
	return io.Copy(dst, src)
}

var pathMaker = pathmaker.RandomPathMaker(nil)

func Write(src io.Reader, collection, fileName string, size int64) (rpath string, written int64, err error) {
	rpath = filepath.Join("/"+collection, pathMaker(fileName, size))

	absPath, err := base.AbsPath(rpath)
	fmt.Println(absPath)
	fmt.Println(filepath.Dir(absPath))
	if err != nil {
		return "", 0, err
	}
	err = MkDir(filepath.Dir(rpath))
	if err != nil {
		return "", 0, err
	}
	dst, err := os.OpenFile(absPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return "", 0, err
	}
	defer dst.Close()
	written, err = io.Copy(dst, src)
	if err != nil {
		return "", written, err
	}
	err = meta.GetMeta().Register(rpath, meta.FileMeta{FileName: fileName, Size: written})

	return

}

type FileInfo struct {
	Name     string
	Size     int64
	Mode     os.FileMode
	ModeTime time.Time
	IsDir    bool
	RawName  string
}

func NewFileInfo(info os.FileInfo) *FileInfo {
	if nil == info {
		return nil
	}
	return &FileInfo{
		Name:     info.Name(),
		Size:     info.Size(),
		Mode:     info.Mode(),
		ModeTime: info.ModTime(),
		IsDir:    info.IsDir(),
	}
}
func Ls(relativePath string) ([]*FileInfo, error) {
	absPath, err := base.AbsPath(relativePath)
	if err != nil {
		return nil, err
	}
	infos, err := ioutil.ReadDir(absPath)
	if err != nil {
		return nil, err
	}
	r := make([]*FileInfo, len(infos))
	for i, info := range infos {
		r[i] = NewFileInfo(info)
	}
	return r, nil
}
func Stat(relativePath string) (*FileInfo, error) {
	absPath, err := base.AbsPath(relativePath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	return NewFileInfo(info), nil
}
func Remove(relativePath string) error {
	absPath, err := base.AbsPath(relativePath)
	if err != nil {
		return err
	}
	err = os.Remove(absPath)
	if err != nil {
		return err
	}
	meta.GetMeta().Remove(relativePath)
	return nil
}

// func RemoveAll(relativePath string) error {
// 	absPath, err := base.AbsPath(relativePath)
// 	if err != nil {
// 		return err
// 	}
// 	return os.RemoveAll(absPath)
// }

func ZipRead(zipFile, zipEntry string) ([]byte, error) {
	absPath, err := base.AbsPath(zipFile)
	if err != nil {
		return nil, err
	}
	bs, ok := zipx.GetCopy(absPath, zipEntry)
	if !ok {
		return nil, nil
	}
	return bs, nil
}

func ResizeImage(filePath string, crop []int, sizes [][]int) ([]string, error) {
	absPath, err := base.AbsPath(filePath)
	trans, err := imageutil.NewImageFileTransform(absPath)
	if err != nil {
		return nil, err
	}
	lens := len(sizes)
	if len(crop) > 0 {
		lens++
	}
	filenames := make([]string, lens)
	i := 1
	if len(crop) > 0 {
		if len(crop) != 4 {
			return nil, errors.New("image crop format error")
		}
		filename := stringutil.FileNameAppend(filePath, "_"+strconv.Itoa(i))
		absFilename, err := base.AbsPath(filename)
		if err != nil {
			return nil, err
		}
		trans.Crop(crop[0], crop[1], crop[2], crop[3]).Save(absFilename)
		filenames[i-1] = filename
		i++
	}
	if len(sizes) > 0 {
		for _, size := range sizes {
			if len(size) < 2 {
				return nil, errors.New("image resize format error")
			}
			filename := stringutil.FileNameAppend(filePath, "_"+strconv.Itoa(i))
			absFilename, err := base.AbsPath(filename)
			if err != nil {
				return nil, err
			}
			trans.Resize(size[0], size[1]).Save(absFilename)
			filenames[i-1] = filename
			i++
		}
	}

	return filenames, nil
}

func Exec(timeout time.Duration, cmdStr string, args ...string) cmd.Status {
	exec := cmd.NewCmd(cmdStr, args...)
	statusChan := exec.Start()
	go func() {
		<-time.After(timeout)
		exec.Stop()
	}()
	return <-statusChan
}
