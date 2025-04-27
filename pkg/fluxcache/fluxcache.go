package fluxcache

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	"fluxcache/internal/cache"
	"fluxcache/internal/protohelper"
	"fluxcache/internal/redis"
	"fluxcache/internal/syncer"
)

type FluxCache struct {
	mem *cache.MemoryCache
	rdb *redis.RedisClient
	ctx context.Context
}

// NewFluxCache initializes a new FluxCache instance
func NewFluxCache(ctx context.Context, opts Options) (*FluxCache, error) {
	rdb := redis.NewRedisClient(opts.RedisAddr, opts.RedisPassword, opts.RedisDB)
	mem := cache.NewMemoryCache(opts.MemoryCap)

	fc := &FluxCache{
		mem: mem,
		rdb: rdb,
		ctx: ctx,
	}

	// Sync all existing keys at startup
	if opts.SyncAtStart {
		log.Info().Msg("FluxCache: Performing initial sync from Redis")
		syncer.SyncFromRedis(ctx, rdb, mem, opts.KeyPattern)
	}

	// Background periodic sync (optional)
	if opts.SyncInterval > 0 {
		log.Info().Msgf("FluxCache: Starting periodic sync every %v", opts.SyncInterval)
		go syncer.PeriodicSync(ctx, rdb, mem, opts.SyncInterval, opts.KeyPattern)
	}

	return fc, nil
}

// Set saves a key-value pair into memory and Redis
func (f *FluxCache) Set(key string, value any) error {
	f.mem.Set(key, value)
	log.Debug().Str("key", key).Msg("FluxCache: Set in memory")

	switch v := value.(type) {
	case []byte:
		err := f.rdb.Set(f.ctx, key, v)
		if err == nil {
			log.Debug().Str("key", key).Msg("FluxCache: Set in Redis (bytes)")
		}
		return err
	case string:
		err := f.rdb.Set(f.ctx, key, []byte(v))
		if err == nil {
			log.Debug().Str("key", key).Msg("FluxCache: Set in Redis (string)")
		}
		return err
	default:
		log.Error().Str("key", key).Msg("FluxCache: Unsupported type for Set")
		return errors.New("unsupported value type for Set")
	}
}

// Get fetches a value from memory first, fallback to Redis
func (f *FluxCache) Get(key string) (any, error) {
	val, ok := f.mem.Get(key)
	if ok {
		log.Debug().Str("key", key).Msg("FluxCache: Cache hit (memory)")
		return val, nil
	}

	log.Debug().Str("key", key).Msg("FluxCache: Cache miss, trying Redis")

	b, err := f.rdb.Get(f.ctx, key)
	if err != nil {
		log.Error().Str("key", key).Err(err).Msg("FluxCache: Redis Get failed")
		return nil, err
	}

	// Save back into memory
	f.mem.Set(key, b)
	log.Debug().Str("key", key).Msg("FluxCache: Loaded from Redis into memory")

	return b, nil
}

// SetProto saves a proto.Message compressed into memory and Redis
func (f *FluxCache) SetProto(key string, msg proto.Message) error {
	data, err := protohelper.Marshal(msg)
	if err != nil {
		log.Error().Str("key", key).Err(err).Msg("FluxCache: Proto marshal failed")
		return err
	}
	return f.Set(key, data)
}

// GetProto fetches a compressed proto.Message from memory or Redis
func (f *FluxCache) GetProto(key string, out proto.Message) error {
	val, err := f.Get(key)
	if err != nil {
		return err
	}

	data, ok := val.([]byte)
	if !ok {
		log.Error().Str("key", key).Msg("FluxCache: Expected []byte from Get for proto")
		return errors.New("value is not []byte")
	}

	return protohelper.Unmarshal(data, out)
}

// Delete removes a key from both memory and Redis
func (f *FluxCache) Delete(key string) error {
	f.mem.Delete(key)
	log.Debug().Str("key", key).Msg("FluxCache: Deleted from memory")

	err := f.rdb.Delete(f.ctx, key)
	if err == nil {
		log.Debug().Str("key", key).Msg("FluxCache: Deleted from Redis")
	} else {
		log.Error().Str("key", key).Err(err).Msg("FluxCache: Redis Delete failed")
	}
	return err
}
