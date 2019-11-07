package main

import (
	"fmt"
	"httpfs/base"
	"httpfs/base/log"
	"httpfs/route"
	"httpfs/service/cluster"
	"httpfs/util/stringutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	parseCmdOptions()
	if base.Config.Http.Static > 0 {
		go startStaticServer()
	}
	startHttpFsServer()

}
func startHttpFsServer() {
	fmt.Println(stringutil.JsonPretty(base.Config))
	app := iris.Default()
	app.ContextPool.Attach(func() context.Context {
		return &route.Context{
			Context: context.NewContext(app),
		}
	})
	app.RegisterView(iris.HTML("./views", ".html"))
	app.HandleDir("/static", "./static")

	conf := iris.YAML("./conf.yml")

	route.Init(app)
	cluster.InitCluster()
	err := app.Run(iris.Addr(":"+strconv.Itoa(base.Config.Http.LocalUrl.Port)), iris.WithConfiguration(conf))
	if err != nil {
		log.Log.Error(err)
	}
}

func startStaticServer() {
	app := iris.Default()
	app.HandleDir("/", base.Config.Fs.Root)
	err := app.Run(iris.Addr(":"+strconv.Itoa(base.Config.Http.Static)), iris.WithConfiguration(iris.YAML("./conf.yml")))
	if err != nil {
		log.Log.Error(err)
	}
}

func parseCmdOptions() {
	local := kingpin.Flag("local", "app run local mode").Short('l').Bool()
	debug := kingpin.Flag("debug", "app work development mode").Short('d').Bool()
	kingpin.Parse()
	if debug != nil {
		base.Config.Debug = *debug
	}
	if local != nil {
		base.Config.RunLocal = *local
	}
}

func onSignal(server *fuse.Server) {
	c := make(chan os.Signal)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		fmt.Println("unmount ddfs fuse")
		// umount -f test/mount
		server.Unmount()
		os.Exit(0)
	}()

}
