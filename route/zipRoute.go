package route

import (
	"errors"
	"fmt"
	"httpfs/base"
	"path/filepath"
	"strings"

	"github.com/RocksonZeta/zipx"
	"github.com/kataras/iris"
)

// /zip
func zipRoute(app iris.Party) {

	app.Get("/read/{p:path}", func(ctx iris.Context) {
		c := ctx.(*Context)
		p := filepath.Join("/", ctx.Params().Get("p"))

		indexDot := strings.Index(p, ".")
		i := indexDot
		for ; i < len(p); i++ {
			if '/' == p[i] {
				break
			}
		}
		filePath := p[0:i]

		abs, err := base.AbsPath(filePath)
		if err != nil {
			c.Error(base.StateError, err)
			return
		}

		bs, ok := zipx.GetCopy(abs, p[i+1:])
		if !ok {
			c.Error(base.StateError, errors.New("read failed"))
			return
		}
		ctx.Write(bs)
	})
	app.Get("/get/{p:path}", func(ctx iris.Context) {
		c := ctx.(*Context)
		p := filepath.Join("/", ctx.Params().Get("p"))
		abs, err := base.AbsPath(p)
		if err != nil {
			c.Error(base.StateError, err)
			return
		}
		entry := ctx.URLParam("entry")
		bs, ok := zipx.GetCopy(abs, entry)
		fmt.Println(entry+" zip bs:", string(bs))
		if !ok {
			c.Error(base.StateError, errors.New("read failed"))
			return
		}
		ctx.Write(bs)
	})
	// app.Get("/ls/{p:path}", func(ctx iris.Context) {
	// })
	// app.Get("/stat/{p:path}", func(ctx iris.Context) {
	// })
}
