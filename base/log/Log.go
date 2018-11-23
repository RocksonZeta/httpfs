package log

import (
	"fmt"
	"os"

	"github.com/gogap/logrus_mate"
	_ "github.com/gogap/logrus_mate/hooks/expander"
	_ "github.com/gogap/logrus_mate/hooks/file"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger = logrus.New()

// var Admin *logrus.Logger
func init() {
	confPath := "log.conf"
	var mate *logrus_mate.LogrusMate
	var err error
	for i := 0; i < 5; i++ {
		_, err = os.Stat(confPath)
		if err == nil {
			mate, err = logrus_mate.NewLogrusMate(logrus_mate.ConfigFile(confPath))
			break
		} else {
			confPath = "../" + confPath
			continue
		}
	}
	if err != nil {
		panic(err)
	}
	fmt.Println("find log.conf at:" + confPath)
	if err := mate.Hijack(Log, "main"); err != nil {
		fmt.Println(err)
		return
	}
}
