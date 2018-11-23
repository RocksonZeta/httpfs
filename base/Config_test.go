package base

import (
	"fmt"
	"httpfs/util/stringutil"
	"testing"
)

func TestAppConfig(t *testing.T) {
	fmt.Println(stringutil.JsonPretty(Config))

}
func TestPath(t *testing.T) {
	fmt.Println(AbsPath("/0"))
	fmt.Println(RelPath(AbsPathMust("/0")))

}
