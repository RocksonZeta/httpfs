package cluster

import (
	"httpfs/base"
	"httpfs/base/log"
	"httpfs/service/cluster/kv"
	"httpfs/service/meta"
	"runtime"
	"time"
)

type Server struct {
	ClusterId, ServerId string
	Local, Proxy        string
	Ut                  int64 //update time
	RatedSpace          int   //MB
	FreeSpace           int   //MB
	Cpu                 int   //
	Mem                 int   //MB
	MemFree             int
	LoadAverage         int
}

func InitCluster() {
	new(cluster).HeartBeat()
}

type cluster struct{}

func (c *cluster) CollectSysInfo() Server {
	s := Server{}
	s.ClusterId = base.Config.ClusterId
	s.ServerId = base.Config.ServerId
	s.Local = base.Config.Http.Local
	s.Proxy = base.Config.Http.Proxy
	s.Ut = time.Now().Unix()
	s.RatedSpace = base.Config.Fs.RatedSpace
	s.FreeSpace = base.Config.Fs.RatedSpace*1024 - int(meta.GetMeta().Stat().Size/1024/1024)
	s.Cpu = runtime.NumCPU()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	s.Mem = int(mem.Sys / 1024 / 1024)
	s.MemFree = int(mem.Frees / 1024 / 1024)
	return s
}

func (c *cluster) Register() error {
	redis := kv.MainRedis.Get()
	defer redis.Close()
	s := c.CollectSysInfo()
	ttl := int64(base.Config.ClusterTimer + 60)
	return redis.HSet(s.ClusterId, s.ServerId, s, ttl)
}
func (c *cluster) GetServers() (map[string]Server, error) {
	redis := kv.MainRedis.Get()
	defer redis.Close()
	var servers map[string]Server
	err := redis.HMGetAll(base.Config.ClusterId, &servers)
	return servers, err
}

func (c *cluster) HeartBeat() {
	err := c.Register()
	if err != nil {
		log.Log.Error("cluster - HeartBeat error.", err)
		return
	}
	go func() {
		ticker := time.NewTicker(time.Duration(base.Config.ClusterTimer) * time.Second)
		for {
			select {
			case <-ticker.C:
				c.Register()
			}
		}
	}()
}
