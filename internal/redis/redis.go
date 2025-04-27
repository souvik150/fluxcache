package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisClient{client: rdb}
}

func (r *RedisClient) Set(ctx context.Context, key string, value []byte) error {
	return withRetry(func() error {
		return r.client.Set(ctx, key, value, 0).Err()
	})
}

func (r *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	var result []byte
	err := withRetry(func() error {
		val, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			return err
		}
		result = val
		return nil
	})
	return result, err
}

func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return withRetry(func() error {
		return r.client.Del(ctx, key).Err()
	})
}

func (r *RedisClient) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	var cursor uint64
	var keys []string
	for {
		var batch []string
		var err error
		batch, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

func withRetry(fn func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * 100) // simple backoff
	}
	return err
}
