package handle

import (
	"httpfs/service/handle/image"
	"sync"
)

//Handler module.method(params) (return,error)
type Handler interface {
	Do(method, params string) (interface{}, error)
}

//not size
var handlers sync.Map

// var handlers map[string]Handler = make(map[string]Handler)

func init() {
	Register("image", new(image.ImageHandler))
}

func Register(module string, h Handler) {
	handlers.Store(module, h)
}
func Get(module string) (Handler, bool) {
	h, ok := handlers.Load(module)
	if ok {
		return h.(Handler), ok
	}
	return nil, false
}
