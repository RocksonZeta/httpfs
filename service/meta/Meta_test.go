package meta

import (
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {
	m := GetMeta()
	m.Register("/1", FileMeta{FileName: "1.jpg", Size: 100})
	fmt.Println(m.Get("/1"))
	fmt.Println(m.Stat())
	// m.Remove("/1")
	// fmt.Println(m.Get("/1"))
	fmt.Println(m.Stat())
}
func TestRegister(t *testing.T) {
	m := GetMeta()
	m.RegisterDir("1", "")
	fmt.Println(m.Get("/1/1.txt"))
	fmt.Println(m.Stat())
	m.Query(func(path string, fm FileMeta) bool {
		fmt.Println("->", path, "fm:", fm)
		return true
	})
}
func TestRegisterDir(t *testing.T) {
	m := GetMeta()
	m.RegisterDir("1", "^2")
	fmt.Println(m.Get("/1/1.txt"))
	fmt.Println(m.Stat())
	m.Query(func(path string, fm FileMeta) bool {
		fmt.Println("->", path, "fm:", fm)
		return true
	})
}
