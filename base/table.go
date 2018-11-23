package base

import "github.com/guregu/null"

type File struct {
	Id         int
	ServerId   int
	Path       string
	Bytes      int64
	State      int
	CreateTime int64
	Mime       null.String
	RawName    null.String
	Backup1    null.Int
	Backup2    null.Int
}
type Server struct {
	Id         int
	Server     string
	Proxy      string
	Root       string
	Ready      bool
	RatedSpace int
	UsedSpace  int64
	Backup1    null.Int
	Backup2    null.Int
}
