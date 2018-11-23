package kv

import "httpfs/base"

var MainRedis *ServiceFactory

func init() {

	conf := RedisConfig{
		Addr:     base.Config.Redis.Addr,
		Password: base.Config.Redis.Password,
		Database: base.Config.Redis.Db,
	}
	MainRedis = NewFactory(&conf)
}
