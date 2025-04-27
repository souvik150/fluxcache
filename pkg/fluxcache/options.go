package fluxcache

import "time"

type Options struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	MemoryCap     int           // Max memory entries for LRU
	SyncAtStart   bool          // Whether to load from Redis on startup
	SyncInterval  time.Duration // Background sync every X time
	KeyPattern    string        // Pattern to match keys in Redis (e.g., "*")
}
