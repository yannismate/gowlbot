package cache

import (
	"context"
	"github.com/go-redis/redis/v9"
	"github.com/yannismate/gowlbot/internal/config"
	"time"
)

func ProvideRedisClient(cfg *config.OwlBotConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Cache.URL,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return client, nil
}
