package osutil_test

import (
	"fmt"
	"httpfs/util/osutil"
	"testing"
	"time"

	"github.com/go-cmd/cmd"
)

func TestParseArgs(t *testing.T) {
	fmt.Println(osutil.ParseCmd(`ls -a a 
	34 fddfd \
	hello`))
}
func TestExec(t *testing.T) {
	_, stdout, _, _ := osutil.Exec("find ./ -type d -print -exec ls {} \\;")
	fmt.Println(stdout)
}
func TestExecCB(t *testing.T) {
	state := osutil.ExecCb("sleep 4", time.NewTicker(1*time.Second), 3*time.Second, func(s cmd.Status) bool {
		fmt.Println("stdout:", s.Stdout)
		return true
	})
	fmt.Println("final:", state)
}
