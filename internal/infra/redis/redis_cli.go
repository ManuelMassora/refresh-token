package redis

import (
    "context"
    "github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:		"localhost:6379",
		Password: 	"",
		DB:			0,
	})
}

func PingRedis(ctx context.Context) error {
    return RedisClient.Ping(ctx).Err()
}