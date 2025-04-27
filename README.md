# FluxCache

> ⚡ Ultra-fast, production-grade, in-memory + Redis synchronized cache for Golang.  
> Designed for highly concurrent systems with crash recovery, compression, and multi-instance support.

---

## ✨ Features

- Memory + Redis hybrid caching
- LRU eviction in memory
- Proto (gRPC) native support (compressed via GZIP)
- Automatic background Redis sync (optional)
- Resilient to service restarts (re-hydrate from Redis)
- Safe for multi-instance deployments
- Batched SCAN fetching from Redis
- Optimized retry handling on Redis failures
- Highly concurrent (thread-safe Get/Set/Delete)

---

## 📦 Installation

```bash
go get github.com/your-username/fluxcache
```

---

## 🚀 Quick Example

```go
package main

import (
	"context"
	"time"

	"github.com/souvik150/fluxcache/pkg/fluxcache"
)

func main() {
	ctx := context.Background()

	cache, _ := fluxcache.NewFluxCache(ctx, fluxcache.Options{
		RedisAddr:    "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		MemoryCap:     1_000_000,
		SyncAtStart:   true,
		SyncInterval:  5 * time.Minute,
		KeyPattern:    "user:*",
	})

	// Set a key
	_ = cache.Set("user:123", []byte("example"))

	// Get a key
	val, _ := cache.Get("user:123")
	fmt.Println(string(val.([]byte)))
}
```

---

## 📊 Benchmark Results

| Benchmark | Result |
|:---|:---|
| **SET Benchmark** | 1,000,000 users in 3m41.76s (4509 ops/sec) |
| **GET Benchmark** | 1,000,000 users in 1m23.23s (12014 ops/sec) |
| **MIXED Benchmark** | 500,000 ops in 1m13.63s (6790 ops/sec) — 0 errors |
| **Concurrent Correctness** | 994,277 reads, 59,089 writes, 0 mismatches ✅ |

### Notes:
- ✅ No data races detected under full `-race` mode.
- ✅ Memory + Redis hybrid caching worked flawlessly.
- ✅ Benchmark conducted under concurrency of 200.

---

## ⚙️ Options

| Field | Description |
|:---|:---|
| RedisAddr | Redis address (`localhost:6379`) |
| RedisPassword | Redis password (optional) |
| RedisDB | Redis DB index (0 by default) |
| MemoryCap | Max number of keys to keep in memory (LRU) |
| SyncAtStart | Whether to preload from Redis on start |
| SyncInterval | How often to re-sync from Redis (optional) |
| KeyPattern | Redis key pattern to scan |

---

## 🛡️ License

FluxCache is released under the [MIT License](LICENSE).

---

## 🙏 Contributing

Pull Requests welcome! 🚀  
If you find bugs, optimizations, or want to add more features (e.g., TTL support, clustering) — feel free to open an issue or PR.

---

## ✉️ Contact

If you like this project or want to collaborate on high-performance caching systems,  
feel free to connect with me on GitHub: [souvik150](https://github.com/souvik150).

---
