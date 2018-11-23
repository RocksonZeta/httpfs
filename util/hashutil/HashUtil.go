package hashutil

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
func Md5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

func Sha1(str string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(str)))
}

func Random6() string {
	return strconv.Itoa(rand.Intn(99999) + 100000)
}
func Md5File(file string) (string, error) {
	f, err := os.Open(file)
	if nil != err {
		return "", err
	}
	defer f.Close()
	return Md5Reader(f), nil
}

func Md5Reader(reader io.Reader) string {
	m := md5.New()
	io.Copy(m, reader)
	return fmt.Sprintf("%x", m.Sum([]byte("")))
}
func RandomStr(n int, ignoreCase bool) string {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	if true == ignoreCase {
		letters = []byte("abcdefghijklmnopqrstuvwxyz0123456789")
	}
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandomStr32() string {
	return RandomStr(32, true)
}
