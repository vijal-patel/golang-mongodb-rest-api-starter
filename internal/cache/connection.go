package cache

import (
	"context"
	"golang-mongodb-rest-api-starter/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func Init(cfg *config.Config, log *zap.Logger) *redis.Client {
	log.Info("Connecting to Redis...")
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Cache.Host,
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.DB,
	})
	log.Info("Successfully connected to Redis!")

	return client
}

func Disconnect(client *redis.Client) error {
	return client.Close()
}

func Ping(client *redis.Client) error {
	status := client.Ping(context.TODO())
	if status.Err() != nil {
		return status.Err()
	}
	return nil
}
