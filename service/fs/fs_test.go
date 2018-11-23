package fs_test

import (
	"bytes"
	"fmt"
	"httpfs/service/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLs(t *testing.T) {
	files, err := fs.Ls(".")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(files))
}
func TestStat(t *testing.T) {
	stat, err := fs.Stat(".")
	assert.Nil(t, err)
	assert.True(t, stat.IsDir)
}
func TestAbsPath(t *testing.T) {
	_, err := fs.AbsPath(".")
	assert.Nil(t, err)
}
func TestExecPath(t *testing.T) {
	state := fs.Exec(1*time.Second, "ls")
	fmt.Println(state)
}
func TestMain(t *testing.T) {
	collection := "txt"
	err := fs.MkDir(collection)
	assert.Nil(t, err)
	_, err = fs.AbsPath(collection)
	assert.Nil(t, err)
	hello := "hello"

	//write read
	p, written, err := fs.Write(bytes.NewBufferString(hello), collection, "1", int64(len(hello)))
	fmt.Println("filepath:", p)
	assert.Nil(t, err)
	assert.Equal(t, len(hello), int(written))
	buf := &bytes.Buffer{}
	readed, err := fs.Read(p, buf)
	assert.Nil(t, err)
	assert.Equal(t, len(hello), int(readed))
	assert.Equal(t, hello, string(buf.Bytes()))

	//info
	infos, err := fs.Ls(collection)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(infos))
	info, err := fs.Stat(p)
	assert.Nil(t, err)
	assert.Equal(t, len(hello), int(info.Size))

	bs, err := fs.ZipRead("test.zip", "test")
	assert.Nil(t, err)
	//zip
	fmt.Println(string(bs))

	//delete
	err = fs.Remove(p)
	assert.Nil(t, err)
	// err = fs.RemoveAll(collection)
	// assert.Nil(t, err)
}
