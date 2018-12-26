package util

import (
	"github.com/dejavuzhou/dejavuzhou.github.io/config"
	"github.com/go-redis/redis"
)

var RedisClient *redis.Client

func init() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.REDIS_ADDR,
		Password: config.REDIS_PASSWORD, // no password set
		DB:       config.REDIS_DB_IDX,   // use default DB
	})
	
	// pong, err := RedisClient.Ping().Result()
	// fmt.Println(pong, err)
}
