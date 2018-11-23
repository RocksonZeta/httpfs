package pathmaker

import (
	"fmt"
	"testing"
)

func TestGenDir(t *testing.T) {
	fmt.Println(GenDir(0, 100, 3))
	fmt.Println(GenDir(1010000, 100, 3))
	fmt.Println(GenDir(199, 100, 2))
	fmt.Println(GenDir(200, 100, 2))
	fmt.Println(GenDir(200, 200, 2))
	fmt.Println(GenDir(201, 200, 2))
	fmt.Println(GenDir(400, 200, 3))
}
