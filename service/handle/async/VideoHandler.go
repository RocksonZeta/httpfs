package async

import (
	"encoding/json"
	"errors"
	"httpfs/base"
	"httpfs/base/log"
	"httpfs/service/meta"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
)

type VideoHandler struct{}

type VideoCompressParam struct {
	File             string
	ProgressRedisKey string
	VideoId          int
}

func (v VideoHandler) Do(method, params string, taskId int, finishChan chan TaskReslt) error {
	log.Log.Debug("VideoHandler.Do - method:", method, ",params:", params, ",taskId:", taskId)
	if method == "CompressDash" {
		return v.CompressDash(params, taskId, finishChan)
	}

	return errors.New("VideoHandler no such method:" + method)
}

// func (v VideoHandler) CompressDash1(params string, taskId int, finishChan chan TaskReslt) error {
// 	log.Log.Debug("CompressDash1:", params)
// 	var param VideoCompressParam
// 	err := json.Unmarshal([]byte(params), &param)
// 	if err != nil {
// 		return err
// 	}
// 	c := exec.Command("python3", "bin/mp4.py", base.AbsPathMust(param.File))
// 	stdout, err := c.StdoutPipe()
// 	if err != nil {
// 		return err
// 	}
// 	defer stdout.Close()
// 	err = c.Start()
// 	if err != nil {
// 		return err
// 	}
// 	ticker := time.NewTicker(1 * time.Second)
// 	go func() {
// 		for {
// 			select {
// 			case _, ok := <-ticker.C:
// 				fmt.Println("ticker:", ok)
// 				if !ok {
// 					return
// 				}
// 				reader := bufio.NewReader(stdout)
// 				bs, _, err := reader.ReadLine()
// 				if err != nil {
// 					log.Log.Error(err)
// 					// return
// 				}
// 				fmt.Println("stdout:", string(bs))
// 			}
// 		}
// 	}()
// 	err = c.Wait()
// 	ticker.Stop()
// 	return err

// }
func (v VideoHandler) CompressDash(params string, taskId int, finishChan chan TaskReslt) error {
	var param VideoCompressParam
	err := json.Unmarshal([]byte(params), &param)
	if err != nil {
		return err
	}
	cs := base.Config.ClusterId + ":" + base.Config.ServerId
	c := cmd.NewCmd("python3", "bin/mp4.py", "-f", param.File, "-rk", param.ProgressRedisKey, "-vid", strconv.Itoa(param.VideoId), "-root", base.AbsRoot(), "-cs", cs)
	statusChan := c.Start()
	// ticker := time.NewTicker(3 * time.Second)

	go func() {
		<-time.After(24 * time.Hour)
		log.Log.Warn("Task run timeout. >24h")
		c.Stop()
	}()

	finished := <-statusChan
	for _, line := range finished.Stdout {
		log.Log.Info(line)
	}
	for _, line := range finished.Stderr {
		log.Log.Error(line)
	}
	d, f := filepath.Split(param.File)
	i := strings.Index(f, ".")
	meta.GetMeta().RegisterDir(d, "^"+f[0:i])
	finishChan <- TaskReslt{TaskId: taskId, State: finished.Exit}
	return nil
}
