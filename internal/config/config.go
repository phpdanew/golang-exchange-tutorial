package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	DataSource string
	Redis      redis.RedisConf
	RateLimit  struct {
		Seconds int
		Quota   int
	}
}
