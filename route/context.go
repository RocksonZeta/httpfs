package route

import (
	"httpfs/base"

	"github.com/kataras/iris/context"
)

type Context struct {
	context.Context
}

func (ctx *Context) Do(handlers context.Handlers) {
	context.Do(ctx, handlers)
}
func (ctx *Context) Next() {
	context.Next(ctx)
}

func (c *Context) Result(result interface{}, err error) {
	r := base.JsonResult{State: base.StateOk}
	if err != nil {
		r.Err = err
		r.State = base.StateError
	} else {
		r.State = base.StateOk
		r.Data = result
	}
	c.JSON(r)
}
func (c *Context) Ok(data interface{}) {
	c.JSON(base.JsonResult{State: base.StateOk, Data: data})
}
func (c *Context) Error(state int, err error, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	c.JSON(base.JsonResult{State: state, Err: err, Data: d})
}
