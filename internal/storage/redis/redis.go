package redis

import (
	"context"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
)

type RedisStore struct {
	Client *redis.Client
	Logger *log.Logger
}

func NewRedisStore() *RedisStore {
	db, _ := convertor.ToInt(os.Getenv("RD_DB"))
	client := &RedisStore{
		Logger: log.New(log.Writer(), "RedisStore", log.LstdFlags),
		Client: redis.NewClient(
			&redis.Options{
				Addr:     fmt.Sprintf("%s:%s", os.Getenv("RD_HOST"), os.Getenv("RD_PORT")),
				Username: os.Getenv("RD_USER"),
				Password: os.Getenv("RD_PASSWORD"),
				DB:       int(db),
			},
		),
	}
	_, err := client.Client.Ping(context.Background()).Result()
	if err != nil {
		client.Logger.Fatal("failed to connect redis", err)
		return nil
	}
	client.Logger.Println("connected to redis")
	return client
}
