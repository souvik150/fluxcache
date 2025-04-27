# FluxCache

> ⚡ Ultra-fast, production-grade, in-memory + Redis synchronized cache for Golang.
>  
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
go get github.com/souvik150/fluxcache
```

