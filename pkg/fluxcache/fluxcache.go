package fluxcache

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/rs/zerolog/log"

	"github.com/souvik150/fluxcache/internal/cache"
	"github.com/souvik150/fluxcache/internal/protohelper"
	"github.com/souvik150/fluxcache/internal/redis"
	"github.com/souvik150/fluxcache/internal/syncer"
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

func encode(value any) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case int, int8, int16, int32, int64:
		return []byte(strconv.FormatInt(reflect.ValueOf(v).Int(), 10)), nil
	case uint, uint8, uint16, uint32, uint64:
		return []byte(strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)), nil
	case float32, float64:
		return []byte(strconv.FormatFloat(reflect.ValueOf(v).Float(), 'g', -1, 64)), nil
	case bool:
		if v {
			return []byte("1"), nil
		}
		return []byte("0"), nil
	default:
		return json.Marshal(v)
	}
}

// Set saves a key-value pair into memory and Redis
func (f *FluxCache) Set(key string, value any) error {
	f.mem.Set(key, value)
	log.Debug().Str("key", key).Msg("FluxCache: Set in memory")

	payload, err := encode(value)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("FluxCache: encode failed")
		return err
	}
	if err := f.rdb.Set(f.ctx, key, payload); err != nil {
		return err
	}
	log.Debug().Str("key", key).Msg("FluxCache: Set in Redis")
	return nil
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

func (f *FluxCache) IncrInt(key string, delta int64) (int64, error) {
	// durable op
	newVal, err := f.rdb.IncrBy(f.ctx, key, delta)
	if err != nil {
		return 0, err
	}
	// hot copy
	f.mem.Set(key, newVal)
	return newVal, nil
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

func (f *FluxCache) InjectRedisOutage(d time.Duration) {
	f.rdb.InjectRedisOutage(d)
}
