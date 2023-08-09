package rdsconn

import (
	"github.com/go-redis/redis/v8"
	"github.com/mangohow/imchat/cmd/chatserver/internal/conf"
	"github.com/mangohow/imchat/pkg/common/xredis"
)

var redisConn *redis.Client

func RedisConn() *redis.Client {
	return redisConn
}

func InitRedis() (err error) {
	redisConn, err = xredis.NewRedisInstance(conf.RedisConf)

	return
}

func CloseRedis() error {
	return redisConn.Close()
}

