package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var (
	Ctx         = context.Background()
	RedisClient *redis.Client
)

func InitRedisClient() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		DB:       0,  // use default DB
		Password: "", // no password set
	})
}
