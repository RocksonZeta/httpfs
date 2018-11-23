package cluster

import (
	"fmt"
	"testing"
)

func TestCluster(t *testing.T) {
	c := new(cluster)
	// fmt.Println(c.GetServers())
	// c.Report()
	fmt.Println(c.GetServers())

	// s := MainRedis.Get()
	// defer s.Close()

	// var cluster map[string]cluster.Server
	// s.HMGetAll("static", &cluster)
	// fmt.Println(cluster)
}
