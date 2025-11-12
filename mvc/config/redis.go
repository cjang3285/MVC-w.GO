package config 

import (
    "github.com/redis/go-redis/v9"
    "context"
    "fmt"
)

var Ctx = context.Background()
var RedisClient *redis.Client

func ConnectRedis() error {
    RedisClient = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        Password: "",
        DB: 0,
    })


    pong, err := RedisClient.Ping(Ctx).Result()
    if err != nil {
	return fmt.Errorf("Redis connection failed: %w", err)
    }
    fmt.Println("Redis connection successful:", pong)
    return nil
}
