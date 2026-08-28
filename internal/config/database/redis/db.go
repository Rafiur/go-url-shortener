package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/Rafiur/go-url-shortener/internal/config"
	"github.com/redis/go-redis/v9"
)

func SetupRedis(conf *config.Config) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:     conf.RedisAddress,
		Password: conf.RedisPassword,
		DB:       0, // use default DB
	}

	// Managed Redis (Upstash and friends) requires TLS; a local container does not.
	if conf.RedisTLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: strings.Split(conf.RedisAddress, ":")[0],
		}
	}

	client := redis.NewClient(opts)

	// Check if Redis is connected
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		//appLogger.Error("Failed to connect to Redis", "error", err)
		return nil, fmt.Errorf("could not connect to Redis: %v", err)
	}

	//appLogger.Info("Successfully connected to Redis")
	return client, nil
}
