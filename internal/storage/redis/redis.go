package redis

import (
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
)

type RedisClient struct {
	Client *redis.Client
	Logger *log.Logger
}

func NewRedisClient() *RedisClient {
	db, _ := convertor.ToInt(os.Getenv("RD_DB"))
	return &RedisClient{
		Logger: log.New(log.Writer(), "RedisClient", log.LstdFlags),
		Client: redis.NewClient(
			&redis.Options{
				Addr:     fmt.Sprintf("%s:%s", os.Getenv("RD_HOST"), os.Getenv("RD_PORT")),
				Username: os.Getenv("RD_USER"),
				Password: os.Getenv("RD_PASSWORD"),
				DB:       int(db),
			},
		),
	}
}
