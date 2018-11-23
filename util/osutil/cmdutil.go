package osutil

import (
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/go-cmd/cmd"
)

func ExecCb(cmdStr string, ticker *time.Ticker, timeout time.Duration, cb func(c cmd.Status) bool) cmd.Status {
	// args := ParseCmd(cmdStr)
	// name := args[0]
	// c := cmd.NewCmd(name, args[1:]...)
	c := cmd.NewCmd("sh", "-c", cmdStr)
	statusChan := c.Start()
	if timeout > 0 {
		go func() {
			<-time.After(timeout)
			c.Stop()
		}()
	}
	go func() {
		for range ticker.C {
			if !cb(c.Status()) {
				c.Stop()
			}
		}
	}()
	return <-statusChan
}
func Exec(cmdStr string) (*os.ProcessState, string, string, error) {
	// args := ParseCmd(cmdStr)
	// name := args[0]
	// cmd := exec.Command(name, args[1:]...)
	cmd := exec.Command("sh", "-c", cmdStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", "", err
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, "", "", err
	}
	defer stderr.Close()
	err = cmd.Start()
	if err != nil {
		return nil, "", "", err
	}
	stdoutBs, err := ioutil.ReadAll(stdout)
	stderrBs, err := ioutil.ReadAll(stderr)
	return cmd.ProcessState, string(stdoutBs), string(stderrBs), nil
}

func isEmptyChar(c byte) bool {
	return ' ' == c || '\t' == c || '\r' == c || '\n' == c
}

func ParseCmd(cmd string) []string {
	token := ""
	i := 0
	state := "start"
	var tokens []string
	var c byte
	for i < len(cmd) {
		c = cmd[i]
		i++
		switch state {
		case "start":
			if isEmptyChar(c) {
				continue
			}
			if '\'' == c {
				state = "squot"
				continue
			}
			if '"' == c {
				state = "dquot"
				continue
			}
			i--
			state = "token"
			continue
		case "token":
			if isEmptyChar(c) {
				tokens = append(tokens, token)
				token = ""
				state = "sep"
				continue
			}
			token += string(c)
		case "sep":
			if isEmptyChar(c) {
				continue
			}
			if '\'' == c {
				state = "squot"
				continue
			}
			if '"' == c {
				state = "dquot"
				continue
			}
			if '\\' == c {
				state = "sep"
				continue
			}
			i--
			state = "token"
		case "squot":
			if '\'' == c && cmd[i-1] != '\\' {
				tokens = append(tokens, token)
				token = ""
				state = "sep"
				continue
			}
			token += string(c)
		case "dquot":
			if '"' == c && cmd[i-1] != '\\' {
				tokens = append(tokens, token)
				token = ""
				state = "sep"
				continue
			}
			token += string(c)
		}
	}
	if token != "" {
		tokens = append(tokens, token)
	}
	return tokens
}
