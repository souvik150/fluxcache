package syncer

import (
	"context"
	"fluxcache/internal/cache"
	"fluxcache/internal/redis"
	"time"
)

func SyncFromRedis(ctx context.Context, rdb *redis.RedisClient, mem *cache.MemoryCache, pattern string) {
	keys, err := rdb.ScanKeys(ctx, pattern)
	if err != nil {
		return
	}
	for _, key := range keys {
		data, err := rdb.Get(ctx, key)
		if err != nil {
			continue
		}
		mem.Set(key, data)
	}
}

func PeriodicSync(ctx context.Context, rdb *redis.RedisClient, mem *cache.MemoryCache, interval time.Duration, pattern string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			SyncFromRedis(ctx, rdb, mem, pattern)
		case <-ctx.Done():
			return
		}
	}
}
