package xredis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/mangohow/imchat/pkg/common/xconfig"
	"time"
)

func NewRedisInstance(conf *xconfig.RedisConfig) (*redis.Client, error) {
	redisConn := redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           int(conf.DB),
		PoolSize:     int(conf.PoolSize),     // 连接池最大socket连接数
		MinIdleConns: int(conf.MinIdleConns), // 最少连接维持数
	})

	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err := redisConn.Ping(timeoutCtx).Result()
	if err != nil {
		return nil, err
	}

	return redisConn, nil
}

