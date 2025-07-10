package database

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"

	"matchmaker/internal/logging"
)

var Redis *redis.Client

// InitRedis connects to Redis using REDIS_URL.
func InitRedis() (*redis.Client, error) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		logging.Log.Warn("REDIS_URL not set")
	}
	opts, err := redis.ParseURL(url)
	if err != nil {
		logging.Log.WithError(err).Error("invalid redis url")
		return nil, err
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		logging.Log.WithError(err).Error("failed to connect to redis")
		return nil, err
	}
	Redis = client
	return client, nil
}
