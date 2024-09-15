package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Mysql struct {
		Datasource string
	}

	Cache  cache.CacheConf
	RedisX redis.RedisConf
	Jwt    struct {
		AccessSecret string
		AccessExpire int64
	}
}
