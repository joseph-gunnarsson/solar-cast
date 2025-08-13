package red

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func GetRedisConnection() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	return rdb
}
