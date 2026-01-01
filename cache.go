package dbkit

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// CacheProvider interface defines the behavior of a cache provider
type CacheProvider interface {
	CacheGet(cacheName, key string) (interface{}, bool)
	CacheSet(cacheName, key string, value interface{}, ttl time.Duration)
	CacheDelete(cacheName, key string)
	CacheClear(cacheName string)
	Status() map[string]interface{}
}

// cacheEntry represents a single item in the local cache
type cacheEntry struct {
	value      interface{}
	expiration time.Time
	createdAt  time.Time
}

func (e cacheEntry) isExpired() bool {
	if e.expiration.IsZero() {
		return false
	}
	return time.Now().After(e.expiration)
}

// localCache implements CacheProvider using in-memory storage
type localCache struct {
	stores          sync.Map // map[string]*sync.Map (cacheName -> map[key]cacheEntry)
	cleanupInterval time.Duration
}

// newLocalCache creates a new in-memory cache provider
func newLocalCache(cleanupInterval time.Duration) *localCache {
	lc := &localCache{
		stores:          sync.Map{},
		cleanupInterval: cleanupInterval,
	}
	// 启动定期清理过期缓存的任务
	cleanupOnce.Do(func() {
		go lc.startCleanupTimer()
	})
	return lc
}

func (lc *localCache) startCleanupTimer() {
	ticker := time.NewTicker(lc.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		lc.cleanupExpired()
	}
}

func (lc *localCache) cleanupExpired() {
	lc.stores.Range(func(name, store interface{}) bool {
		s := store.(*sync.Map)
		s.Range(func(key, value interface{}) bool {
			entry := value.(cacheEntry)
			if entry.isExpired() {
				s.Delete(key)
			}
			return true
		})
		return true
	})
}

func (lc *localCache) CacheGet(cacheName, key string) (interface{}, bool) {
	if store, ok := lc.stores.Load(cacheName); ok {
		if entry, ok := store.(*sync.Map).Load(key); ok {
			e := entry.(cacheEntry)
			if !e.isExpired() {
				return e.value, true
			}
			// 过期了，顺手删掉
			store.(*sync.Map).Delete(key)
		}
	}
	return nil, false
}

func (lc *localCache) CacheSet(cacheName, key string, value interface{}, ttl time.Duration) {
	store, _ := lc.stores.LoadOrStore(cacheName, &sync.Map{})
	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}
	store.(*sync.Map).Store(key, cacheEntry{
		value:      value,
		expiration: expiration,
		createdAt:  time.Now(),
	})
}

func (lc *localCache) CacheDelete(cacheName, key string) {
	if store, ok := lc.stores.Load(cacheName); ok {
		store.(*sync.Map).Delete(key)
	}
}

func (lc *localCache) CacheClear(cacheName string) {
	lc.stores.Delete(cacheName)
}

func (lc *localCache) Status() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["type"] = "LocalCache"
	stats["cleanup_interval"] = lc.cleanupInterval.String()

	var totalItems int64
	var storeCount int64
	var totalMemory int64

	lc.stores.Range(func(name, store interface{}) bool {
		storeCount++
		s := store.(*sync.Map)
		s.Range(func(key, value interface{}) bool {
			totalItems++
			entry := value.(cacheEntry)
			totalMemory += estimateSize(key)
			totalMemory += estimateSize(entry.value)
			return true
		})
		return true
	})

	stats["total_items"] = totalItems
	stats["store_count"] = storeCount
	stats["estimated_memory_bytes"] = totalMemory
	stats["estimated_memory_human"] = formatBytes(totalMemory)

	return stats
}

func estimateSize(v interface{}) int64 {
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case string:
		return int64(len(val))
	case []byte:
		return int64(len(val))
	case int, int32, uint, uint32, float32:
		return 4
	case int64, uint64, float64:
		return 8
	case bool:
		return 1
	case *Record:
		if val == nil {
			return 0
		}
		var size int64
		val.mu.RLock()
		for k, v := range val.columns {
			size += int64(len(k))
			size += estimateSize(v)
		}
		val.mu.RUnlock()
		return size
	case []Record:
		var size int64
		for _, r := range val {
			size += estimateSize(&r)
		}
		return size
	case []*Record:
		var size int64
		for _, r := range val {
			size += estimateSize(r)
		}
		return size
	case map[string]interface{}:
		var size int64
		for k, v := range val {
			size += int64(len(k))
			size += estimateSize(v)
		}
		return size
	case []interface{}:
		var size int64
		for _, item := range val {
			size += estimateSize(item)
		}
		return size
	default:
		// Fallback for other types
		return 16 // Assume a pointer or small struct size
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// Global cache state
var (
	defaultCache CacheProvider = newLocalCache(1 * time.Minute)
	defaultTTL   time.Duration = time.Minute
	cacheConfigs sync.Map      // map[cacheName]time.Duration
	cacheMu      sync.RWMutex
	cleanupOnce  sync.Once
)

// GetCache returns the current global cache provider
func GetCache() CacheProvider {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return defaultCache
}

// SetCache sets the global cache provider
func SetCache(c CacheProvider) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	defaultCache = c
}

// SetLocalCacheConfig sets the global cache provider to a new localCache with custom interval
func SetLocalCacheConfig(cleanupInterval time.Duration) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	defaultCache = newLocalCache(cleanupInterval)
}

// SetDefaultTtl sets the global default TTL for caching
func SetDefaultTtl(ttl time.Duration) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	defaultTTL = ttl
}

// CreateCache pre-configures a cache store with a specific TTL
func CreateCache(cacheName string, ttl time.Duration) {
	cacheConfigs.Store(cacheName, ttl)
}

// CacheSet stores a value in a specific cache store
func CacheSet(cacheName, key string, value interface{}, ttl ...time.Duration) {
	expiration := defaultTTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	} else if configTTL, ok := cacheConfigs.Load(cacheName); ok {
		expiration = configTTL.(time.Duration)
	}

	defaultCache.CacheSet(cacheName, key, value, expiration)
}

// CacheGet retrieves a value from a specific cache store
func CacheGet(cacheName, key string) (interface{}, bool) {
	return defaultCache.CacheGet(cacheName, key)
}

// CacheDelete removes a specific key from a cache store
func CacheDelete(cacheName, key string) {
	defaultCache.CacheDelete(cacheName, key)
}

// CacheClear clears all keys from a cache store
func CacheClear(cacheName string) {
	defaultCache.CacheClear(cacheName)
}

// CacheStatus returns the current cache provider's status
func CacheStatus() map[string]interface{} {
	return defaultCache.Status()
}

// Cache sets the cache name and TTL for the default database query
func Cache(name string, ttl ...time.Duration) *DB {
	db, err := defaultDB()
	if err != nil {
		return &DB{lastErr: err}
	}
	return db.Cache(name, ttl...)
}

// GenerateCacheKey creates a unique key for a query
func GenerateCacheKey(dbName, sql string, args ...interface{}) string {
	hash := md5.New()
	hash.Write([]byte(dbName))
	hash.Write([]byte(sql))
	if len(args) > 0 {
		hash.Write([]byte(fmt.Sprintf("%v", args)))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func getEffectiveTTL(cacheName string, customTTL time.Duration) time.Duration {
	if customTTL >= 0 {
		return customTTL
	}
	if configTTL, ok := cacheConfigs.Load(cacheName); ok {
		return configTTL.(time.Duration)
	}
	return defaultTTL
}
