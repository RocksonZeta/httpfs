package route

import (
	"errors"
	"httpfs/base"
	"httpfs/service/handle"
	"httpfs/service/handle/async"

	"github.com/kataras/iris"
)

// /call/module/method
func hanlderRoute(app iris.Party) {
	taskActor := async.NewTaskActor()
	app.Post("/async/{module:string}/{method:string}", func(ctx iris.Context) {
		c := ctx.(*Context)
		module := ctx.Params().Get("module")
		method := ctx.Params().Get("method")
		args := ctx.FormValue("args")
		taskActor.TaskChan <- async.Task{Module: module, Method: method, Args: args}
		c.Ok(nil)
	})

	app.Post("/{module:string}/{method:string}", func(ctx iris.Context) {
		c := ctx.(*Context)
		module := ctx.Params().Get("module")
		method := ctx.Params().Get("method")
		args := ctx.FormValue("args")
		fn, _ := handle.Get(module)
		if nil == fn {
			c.Error(base.StateError, errors.New("no such method:"+method))
		}
		c.Result(fn.Do(method, args))
	})
}
