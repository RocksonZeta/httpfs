package async

type AsyncHandler interface {
	Do(method, params string, taskId int, finishChan chan TaskReslt) error
}

var handlers map[string]AsyncHandler

func init() {
	handlers = make(map[string]AsyncHandler)
	handlers["video"] = VideoHandler{}
}
