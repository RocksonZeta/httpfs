package kv

import (
	"testing"
)

func TestMain(t *testing.T) {
	s := MainRedis.Get()
	defer s.Close()

	// ok, err := s.PingPong()
	// assert.Nil(t, err)
	// assert.True(t, ok)
	// s.SetJson("1", 2, 3)
	// var r1 int
	// s.GetJson("1", &r1)
	// fmt.Println("r1:", r1)
	// s.Set(1, 1, 3)
	// fmt.Println(redis.Int(s.Get(1)))
}
