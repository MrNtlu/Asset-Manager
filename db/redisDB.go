package db

import (
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	RedisDB       *redis.Client
	RedisSExpire  = 5 * time.Minute
	RedisMExpire  = 10 * time.Minute
	RedisLExpire  = 30 * time.Minute
	RedisXLExpire = 1 * time.Hour
)

func SetupRedis() {
	if os.Getenv("REDIS_TYPE") != "Remote" {
		RedisDB = redis.NewClient(&redis.Options{
			Addr: "redis:6379",
			DB:   0,
		})
	} else {
		RedisDB = redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_ENDPOINT"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		})
	}
}
