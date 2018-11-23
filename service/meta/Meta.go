package meta

import (
	"encoding/json"
	"httpfs/base"
	"httpfs/base/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/boltdb/bolt"
)

const (
	bucketFileMetas  = "FileMetas"
	bucketHttpFsStat = "HttpFsStat"
	httpFsStatSize   = "Size"
	httpFsStatCount  = "Count"
)

func init() {
	db, err := bolt.Open(base.Config.Fs.Meta, 0644, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketFileMetas))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(bucketHttpFsStat))
		return err
	})
}

type Meta struct {
	db *bolt.DB
}

var meta *Meta

func GetMeta() *Meta {
	if meta != nil {
		return meta
	}
	db, err := bolt.Open(base.Config.Fs.Meta, 0644, nil)
	if err != nil {
		panic(err)
	}

	meta = &Meta{db: db}
	return meta
}

type FileMeta struct {
	FileName string
	Size     int64
	CheckSum string
	Backups  []string //server Id
}

type Stat struct {
	Count int
	Size  int64
}

func (m *Meta) Close() error {
	return m.db.Close()
}

func (m *Meta) Register(path string, fm FileMeta) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		return m.register(tx, path, fm)
	})

}

//register 无添加，有则更新
func (m *Meta) register(tx *bolt.Tx, path string, fm FileMeta) error {
	log.Log.Debug("Meta.register - path:"+path, ",file name:"+fm.FileName)
	metaBucket := tx.Bucket([]byte(bucketFileMetas))
	bs, _ := json.Marshal(fm)
	fmOldBs := metaBucket.Get([]byte(path))
	var oldFm FileMeta
	if nil != fmOldBs {
		err := json.Unmarshal(fmOldBs, &oldFm)
		if err == nil {
			if fm.FileName == "" {
				fm.FileName = oldFm.FileName
			}
			if fm.Size <= 0 {
				fm.Size = oldFm.Size
			}
		}
	}
	err := metaBucket.Put([]byte(path), bs)
	if err != nil {
		return nil
	}
	statBucket := tx.Bucket([]byte(bucketHttpFsStat))
	sizeDelta := fm.Size - oldFm.Size
	if sizeDelta != 0 {
		sizeStr := statBucket.Get([]byte(httpFsStatSize))
		if nil == sizeStr {
			statBucket.Put([]byte(httpFsStatSize), []byte(strconv.FormatInt(sizeDelta, 10)))
		} else {
			size, _ := strconv.ParseInt(string(sizeStr), 10, 64)
			statBucket.Put([]byte(httpFsStatSize), []byte(strconv.FormatInt(sizeDelta+size, 10)))
		}
	}
	if nil == fmOldBs {
		countStr := statBucket.Get([]byte(httpFsStatCount))
		if nil == countStr {
			statBucket.Put([]byte(httpFsStatCount), []byte("1"))
		} else {
			count, _ := strconv.Atoi(string(countStr))
			statBucket.Put([]byte(httpFsStatCount), []byte(strconv.Itoa(1+count)))
		}
	}
	return nil
}
func (m *Meta) Stat() Stat {
	var fm Stat
	m.db.View(func(tx *bolt.Tx) error {
		sizeStr := tx.Bucket([]byte(bucketHttpFsStat)).Get([]byte(httpFsStatSize))
		countStr := tx.Bucket([]byte(bucketHttpFsStat)).Get([]byte(httpFsStatCount))
		fm.Count, _ = strconv.Atoi(string(countStr))
		fm.Size, _ = strconv.ParseInt(string(sizeStr), 10, 64)
		return nil
	})
	return fm
}
func (m *Meta) Get(path string) *FileMeta {
	var fm *FileMeta
	m.db.View(func(tx *bolt.Tx) error {
		bs := tx.Bucket([]byte(bucketFileMetas)).Get([]byte(path))
		if nil == bs {
			return nil
		}
		fm = &FileMeta{}
		return json.Unmarshal(bs, fm)
	})
	return fm
}
func (m *Meta) Query(filter func(path string, fm FileMeta) bool) {
	m.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketFileMetas)).Cursor()
		if nil == c {
			return nil
		}
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fm := FileMeta{}
			json.Unmarshal(v, &fm)
			if !filter(string(k), fm) {
				return nil
			}
		}
		return nil
	})
}
func (m *Meta) Remove(path string) {
	m.db.Update(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket([]byte(bucketFileMetas))
		if metaBucket == nil {
			return nil
		}
		fmBs := metaBucket.Get([]byte(path))
		if nil == fmBs {
			return nil
		}
		err := metaBucket.Delete([]byte(path))
		if err != nil {
			return err
		}
		var oldFm FileMeta
		json.Unmarshal(fmBs, &oldFm)
		statBucket := tx.Bucket([]byte(bucketHttpFsStat))
		if statBucket != nil {
			sizeStr := statBucket.Get([]byte(httpFsStatSize))
			if nil != sizeStr {
				size, _ := strconv.ParseInt(string(sizeStr), 10, 64)
				statBucket.Put([]byte(httpFsStatSize), []byte(strconv.FormatInt(size-oldFm.Size, 10)))
			}
			countStr := statBucket.Get([]byte(httpFsStatCount))
			if nil != countStr {
				count, _ := strconv.Atoi(string(countStr))
				statBucket.Put([]byte(httpFsStatCount), []byte(strconv.Itoa(count-1)))
			}
		}
		return nil
	})
}

func (m *Meta) registerFile(tx *bolt.Tx, path, fileName string) error {
	log.Log.Debug("Meta.registerFile - path:", path, ",fileName:", fileName)
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	fm := FileMeta{}
	if !stat.IsDir() {
		if fileName == "" {
			fm.FileName = filepath.Base(path)
		} else {
			fm.FileName = fileName
		}
		fm.Size = stat.Size()
	}
	return m.register(tx, base.RelPath(path), fm)
}
func (m *Meta) RegisterFile(path, fileName string) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		return m.registerFile(tx, base.AbsPathMust(path), fileName)
	})
}
func (m *Meta) RegisterDir(dirName, fileNamePattern string) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		return m.registerDir(tx, base.AbsPathMust(dirName), fileNamePattern)
	})
}

func (m *Meta) registerDir(tx *bolt.Tx, dirpath, fileNamePattern string) error {
	infos, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return err
	}
	for _, info := range infos {
		if info.IsDir() {
			m.registerDir(tx, filepath.Join(dirpath, info.Name()), fileNamePattern)
		} else {
			name := info.Name()
			if fileNamePattern != "" {
				if ok, err := regexp.MatchString(fileNamePattern, name); !ok || err != nil {
					return err
				}
			}
			m.registerFile(tx, filepath.Join(dirpath, name), "")
		}
	}
	return nil
}
