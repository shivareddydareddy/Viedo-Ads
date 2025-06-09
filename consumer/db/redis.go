// package db

// import (
//     "github.com/redis/go-redis/v9"
// )

// type RedisClient struct {
//     Client *redis.Client
// }

// func NewRedisClient(addr string) RedisClient {
//     rdb := redis.NewClient(&redis.Options{
//         Addr: addr,
//         DB:   0,
//     })

//     return RedisClient{Client: rdb}
// }


package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(addr string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		// RetryDelay:   time.Second,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{Client: rdb}, nil
}

func (r *RedisClient) Close() {
	if r.Client != nil {
		r.Client.Close()
	}
}