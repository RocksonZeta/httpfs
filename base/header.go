package base

const (
	StateOk    = 0
	StateError = 1
)

type JsonResult struct {
	State int
	Data  interface{}
	Err   error
}

func (r *JsonResult) Error() string {
	return r.Err.Error()
}
