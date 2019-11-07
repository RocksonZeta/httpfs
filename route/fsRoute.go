package route

import (
	"httpfs/base"
	"httpfs/service/fs"
	"httpfs/service/meta"
	"path/filepath"

	"github.com/kataras/iris/v12"
)

// /fs
func fsRoute(app iris.Party) {

	app.Post("/mkdir/{p:path}", func(ctx iris.Context) {
		c := ctx.(*Context)
		p := ctx.Params().Get("p")
		c.Ok(p)
	})
	app.Get("/stat/{p:path}", func(ctx iris.Context) {
		c := ctx.(*Context)
		p := filepath.Join("/", ctx.Params().Get("p"))
		info, err := fs.Stat(p)
		fm := meta.GetMeta().Get(p)
		if fm != nil {
			info.RawName = fm.FileName
		}
		c.Result(info, err)
	})
	app.Get("/ls/{p:path}", func(ctx iris.Context) {
		c := ctx.(*Context)
		p := filepath.Join("/", ctx.Params().Get("p"))
		c.Result(fs.Ls(p))
	})

	app.Post("/write/{collection:string}", func(ctx iris.Context) {
		c := ctx.(*Context)
		collection := ctx.Params().Get("collection")
		file, header, err := ctx.FormFile("file")
		if err != nil {
			c.Error(base.StateError, err)
			return
		}
		defer file.Close()
		rpath, _, err := fs.Write(file, collection, header.Filename, header.Size)
		c.Result(rpath, err)
	})
	app.Get("/read/{p:path}", func(ctx iris.Context) {
		p := filepath.Join("/", ctx.Params().Get("p"))
		fm := meta.GetMeta().Get(p)
		if fm != nil {
			ctx.Header("Content-Disposition", "attachment;filename="+fm.FileName)
		}
		_, err := fs.Read(p, ctx)
		if err != nil {
			ctx.NotFound()
		}
	})
	app.Get("/abspath/{p:path}", func(ctx iris.Context) {
		c := ctx.(*Context)
		p := filepath.Join("/", ctx.Params().Get("p"))
		c.Result(base.AbsPath(p))
	})
	app.Post("/rm/{p:path}", func(ctx iris.Context) {
		c := ctx.(*Context)
		p := filepath.Join("/", ctx.Params().Get("p"))
		c.Result(nil, fs.Remove(p))
	})
	// app.Post("/image/{p:path}", func(ctx iris.Context) {
	// 	c := ctx.(*Context)
	// 	cropStr := strings.TrimSpace(ctx.FormValue("crop"))
	// 	sizesStr := strings.TrimSpace(ctx.FormValue("sizes"))
	// 	file, _, err := ctx.FormFile("file")
	// 	if err != nil {
	// 		c.Error(base.StateError, err)
	// 		return
	// 	}
	// 	defer file.Close()
	// 	var files []string
	// 	filePath := ctx.Params().Get("p")
	// 	_, err = fs.Write(file, filePath)
	// 	if err != nil {
	// 		c.Error(base.StateError, err)
	// 		return
	// 	}
	// 	files = append(files, filePath)
	// 	var crop []int
	// 	if cropStr != "" {
	// 		coopStrs := strings.Split(cropStr, ",")
	// 		if len(coopStrs) != 4 {
	// 			c.Error(base.StateError, errors.New("crop format error"))
	// 			return
	// 		}
	// 		crop, _ = stringutil.ToInts(coopStrs)
	// 	}
	// 	var sizes [][]int
	// 	if sizesStr != "" {
	// 		sizesStrs := strings.Split(sizesStr, ",")
	// 		sizes = make([][]int, len(sizesStrs))
	// 		for i, s := range sizesStrs {
	// 			size, err := stringutil.ToInts(strings.Split(s, "x"))
	// 			if err != nil {
	// 				c.Error(base.StateError, errors.New("image sizes format error"))
	// 				return
	// 			}
	// 			sizes[i] = size
	// 		}
	// 	}
	// 	resizedFiles, err := fs.ResizeImage(filePath, crop, sizes)
	// 	if err != nil {
	// 		c.Error(base.StateError, err)
	// 		return
	// 	}
	// 	files = append(files, resizedFiles...)
	// 	c.Result(files, nil)
	// })
	// app.Get("/exec/{cmd:string}", func(ctx iris.Context) {
	// 	c := ctx.(*Context)
	// 	argsJson := ctx.URLParam("args")
	// 	timeout := ctx.URLParamIntDefault("timeout", 3)
	// 	var args []string
	// 	if argsJson != "" {
	// 		err := json.Unmarshal([]byte(argsJson), &args)
	// 		c.Result(nil, err)
	// 		return
	// 	}
	// 	state := fs.Exec(time.Duration(timeout)*time.Second, ctx.Params().Get("cmd"), args...)
	// 	c.Result(state, nil)
	// })
	app.Post("/zip/{p:path}", func(ctx iris.Context) {

	})
	app.Post("/unzip/{p:path}", func(ctx iris.Context) {

	})

}
