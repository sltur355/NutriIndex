package redis

import (
	"LAB1/internal/app/config"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "nutriscan_service." // наш префикс сервиса

type Client struct {
	cfg    config.RedisConfig
	client *redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := &Client{}
	client.cfg = cfg

	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Password:     cfg.Password,
		Username:     cfg.User,
		DB:           0, // используем базу по умолчанию
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.ReadTimeout,
		PoolTimeout:  30 * time.Second,
		MinIdleConns: 5,
	})

	client.client = redisClient

	// Проверяем подключение
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return client, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// GetClient возвращает внутренний redis клиент
func (c *Client) GetClient() *redis.Client {
	return c.client
}
