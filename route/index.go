package route

import (
	"github.com/kataras/iris"
)

func Init(app *iris.Application) {

	fsParty := app.Party("/fs")
	handlerParty := app.Party("/call")
	zipParty := app.Party("/zip")

	fsRoute(fsParty)
	zipRoute(zipParty)
	hanlderRoute(handlerParty)
}
