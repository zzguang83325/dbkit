package dbkit

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// CacheProvider interface defines the behavior of a cache provider
type CacheProvider interface {
	CacheGet(cacheRepositoryName, key string) (interface{}, bool)
	CacheSet(cacheRepositoryName, key string, value interface{}, ttl time.Duration)
	CacheDelete(cacheRepositoryName, key string)
	CacheClearRepository(cacheRepositoryName string)
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
	stores          sync.Map // map[string]*sync.Map (cacheRepositoryName -> map[key]cacheEntry)
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
		// 如果是预编译语句缓存，需要在删除前关闭语句
		isStmtCache := name == "__dbkit_stmt_cache__"

		s.Range(func(key, value interface{}) bool {
			entry := value.(cacheEntry)
			if entry.isExpired() {
				// 如果是预编译语句缓存，关闭语句
				if isStmtCache {
					if stmt, ok := entry.value.(*sql.Stmt); ok {
						stmt.Close()
					}
				}
				s.Delete(key)
			}
			return true
		})
		return true
	})
}

func (lc *localCache) CacheGet(cacheRepositoryName, key string) (interface{}, bool) {
	if store, ok := lc.stores.Load(cacheRepositoryName); ok {
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

func (lc *localCache) CacheSet(cacheRepositoryName, key string, value interface{}, ttl time.Duration) {
	store, _ := lc.stores.LoadOrStore(cacheRepositoryName, &sync.Map{})
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

func (lc *localCache) CacheDelete(cacheRepositoryName, key string) {
	if cacheRepositoryName == "" || key == "" {
		return // 忽略空参数
	}
	if store, ok := lc.stores.Load(cacheRepositoryName); ok {
		store.(*sync.Map).Delete(key)
	}
}

func (lc *localCache) CacheClearRepository(cacheRepositoryName string) {
	if cacheRepositoryName == "" {
		return // 忽略空字符串，避免误操作
	}
	lc.stores.Delete(cacheRepositoryName)
}

// ClearAll 清空所有缓存存储库
func (lc *localCache) ClearAll() {
	lc.stores.Range(func(key, value interface{}) bool {
		lc.stores.Delete(key)
		return true
	})
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
	localCacheInstance CacheProvider               // 本地缓存实例
	redisCacheInstance CacheProvider               // Redis 缓存实例
	defaultCache       CacheProvider               // 默认缓存提供者
	defaultTTL         time.Duration = time.Minute // 默认 TTL
	cacheConfigs       sync.Map                    // map[cacheRepositoryName]time.Duration
	cacheMu            sync.RWMutex                // 缓存锁
	cleanupOnce        sync.Once                   // 清理任务只执行一次
)

// init 初始化默认使用本地缓存
func init() {
	localCacheInstance = newLocalCache(1 * time.Minute)
	defaultCache = localCacheInstance
}

// GetCache 获取当前默认缓存提供者
func GetCache() CacheProvider {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return defaultCache
}

// SetDefaultCache 设置默认缓存提供者
func SetDefaultCache(c CacheProvider) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	defaultCache = c
}

// InitLocalCache 初始化本地缓存实例
func InitLocalCache(cleanupInterval time.Duration) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	localCacheInstance = newLocalCache(cleanupInterval)
}

// InitRedisCache 初始化 Redis 缓存实例
func InitRedisCache(provider CacheProvider) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	redisCacheInstance = provider
}

// GetLocalCacheInstance 获取本地缓存实例
func GetLocalCacheInstance() CacheProvider {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	if localCacheInstance == nil {
		localCacheInstance = newLocalCache(1 * time.Minute)
	}
	return localCacheInstance
}

// GetRedisCacheInstance 获取 Redis 缓存实例
func GetRedisCacheInstance() CacheProvider {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return redisCacheInstance
}

// SetLocalCacheConfig 设置本地缓存配置并将其设为默认缓存
// Deprecated: 使用 InitLocalCache 和 SetDefaultCache 代替
func SetLocalCacheConfig(cleanupInterval time.Duration) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	localCacheInstance = newLocalCache(cleanupInterval)
	defaultCache = localCacheInstance
}

// SetDefaultTtl sets the global default TTL for caching
func SetDefaultTtl(ttl time.Duration) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	defaultTTL = ttl
}

// CreateCache pre-configures a cache store with a specific TTL
func CreateCache(cacheRepositoryName string, ttl time.Duration) {
	cacheConfigs.Store(cacheRepositoryName, ttl)
}

// CacheSet stores a value in a specific cache store
func CacheSet(cacheRepositoryName, key string, value interface{}, ttl ...time.Duration) {
	expiration := defaultTTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	} else if configTTL, ok := cacheConfigs.Load(cacheRepositoryName); ok {
		expiration = configTTL.(time.Duration)
	}

	defaultCache.CacheSet(cacheRepositoryName, key, value, expiration)
}

// CacheGet retrieves a value from a specific cache store
func CacheGet(cacheRepositoryName, key string) (interface{}, bool) {
	return defaultCache.CacheGet(cacheRepositoryName, key)
}

// CacheDelete removes a specific key from a cache store
func CacheDelete(cacheRepositoryName, key string) {
	defaultCache.CacheDelete(cacheRepositoryName, key)
}

// CacheClearRepository clears all keys from a cache store
func CacheClearRepository(cacheRepositoryName string) {
	defaultCache.CacheClearRepository(cacheRepositoryName)
}

// CacheStatus returns the current cache provider's status
func CacheStatus() map[string]interface{} {
	return defaultCache.Status()
}

// LocalCacheSet 在本地缓存中存储值
func LocalCacheSet(cacheRepositoryName, key string, value interface{}, ttl ...time.Duration) {
	expiration := defaultTTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	} else if configTTL, ok := cacheConfigs.Load(cacheRepositoryName); ok {
		expiration = configTTL.(time.Duration)
	}

	GetLocalCacheInstance().CacheSet(cacheRepositoryName, key, value, expiration)
}

// LocalCacheGet 从本地缓存中获取值
func LocalCacheGet(cacheRepositoryName, key string) (interface{}, bool) {
	return GetLocalCacheInstance().CacheGet(cacheRepositoryName, key)
}

// LocalCacheDelete 从本地缓存中删除指定键
func LocalCacheDelete(cacheRepositoryName, key string) {
	GetLocalCacheInstance().CacheDelete(cacheRepositoryName, key)
}

// LocalCacheClearRepository 清空本地缓存中的指定存储库
func LocalCacheClearRepository(cacheRepositoryName string) {
	GetLocalCacheInstance().CacheClearRepository(cacheRepositoryName)
}

// LocalCacheStatus 获取本地缓存状态
func LocalCacheStatus() map[string]interface{} {
	return GetLocalCacheInstance().Status()
}

// RedisCacheSet 在 Redis 缓存中存储值
func RedisCacheSet(cacheRepositoryName, key string, value interface{}, ttl ...time.Duration) error {
	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		return fmt.Errorf("redis cache not initialized, call InitRedisCache first")
	}

	expiration := defaultTTL
	if len(ttl) > 0 {
		expiration = ttl[0]
	} else if configTTL, ok := cacheConfigs.Load(cacheRepositoryName); ok {
		expiration = configTTL.(time.Duration)
	}

	redisCache.CacheSet(cacheRepositoryName, key, value, expiration)
	return nil
}

// RedisCacheGet 从 Redis 缓存中获取值
func RedisCacheGet(cacheRepositoryName, key string) (interface{}, bool, error) {
	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		return nil, false, fmt.Errorf("redis cache not initialized, call InitRedisCache first")
	}

	val, ok := redisCache.CacheGet(cacheRepositoryName, key)
	return val, ok, nil
}

// RedisCacheDelete 从 Redis 缓存中删除指定键
func RedisCacheDelete(cacheRepositoryName, key string) error {
	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		return fmt.Errorf("redis cache not initialized, call InitRedisCache first")
	}

	redisCache.CacheDelete(cacheRepositoryName, key)
	return nil
}

// RedisCacheClearRepository 清空 Redis 缓存中的指定存储库
func RedisCacheClearRepository(cacheRepositoryName string) error {
	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		return fmt.Errorf("redis cache not initialized, call InitRedisCache first")
	}

	redisCache.CacheClearRepository(cacheRepositoryName)
	return nil
}

// RedisCacheStatus 获取 Redis 缓存状态
func RedisCacheStatus() (map[string]interface{}, error) {
	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		return nil, fmt.Errorf("redis cache not initialized, call InitRedisCache first")
	}

	return redisCache.Status(), nil
}

// ClearAllCaches 清空默认缓存中的所有存储库
func ClearAllCaches() {
	if clearer, ok := defaultCache.(interface{ ClearAll() }); ok {
		clearer.ClearAll()
	}
}

// LocalCacheClearAll 清空本地缓存中的所有存储库
func LocalCacheClearAll() {
	if clearer, ok := GetLocalCacheInstance().(interface{ ClearAll() }); ok {
		clearer.ClearAll()
	}
}

// RedisCacheClearAll 清空 Redis 缓存中的所有 dbkit 相关缓存
func RedisCacheClearAll() error {
	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		return fmt.Errorf("redis cache not initialized, call InitRedisCache first")
	}

	if clearer, ok := redisCache.(interface{ ClearAll() }); ok {
		clearer.ClearAll()
	}
	return nil
}

// Cache 使用默认缓存创建查询构建器（可通过 SetDefaultCache 切换默认缓存）
// 示例: dbkit.Cache("user_cache").QueryFirst("SELECT * FROM users WHERE id = ?", 1)
func Cache(name string, ttl ...time.Duration) *DB {
	db, err := defaultDB()
	if err != nil {
		return &DB{lastErr: err}
	}
	return db.Cache(name, ttl...)
}

// LocalCache 创建一个使用本地缓存的查询构建器
// 示例: dbkit.LocalCache("user_cache").QueryFirst("SELECT * FROM users WHERE id = ?", 1)
func LocalCache(cacheRepositoryName string, ttl ...time.Duration) *DB {
	db, err := defaultDB()
	if err != nil {
		return &DB{lastErr: err}
	}

	newDB := &DB{
		dbMgr:               db.dbMgr,
		cacheProvider:       GetLocalCacheInstance(),
		cacheRepositoryName: cacheRepositoryName,
		cacheTTL:            -1,
	}

	if len(ttl) > 0 {
		newDB.cacheTTL = ttl[0]
	} else if configTTL, ok := cacheConfigs.Load(cacheRepositoryName); ok {
		newDB.cacheTTL = configTTL.(time.Duration)
	}

	return newDB
}

// RedisCache 创建一个使用 Redis 缓存的查询构建器
// 示例: dbkit.RedisCache("order_cache").Query("SELECT * FROM orders WHERE user_id = ?", userId)
func RedisCache(cacheRepositoryName string, ttl ...time.Duration) *DB {
	db, err := defaultDB()
	if err != nil {
		return &DB{lastErr: err}
	}

	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		return &DB{lastErr: fmt.Errorf("redis cache not initialized, call InitRedisCache first")}
	}

	newDB := &DB{
		dbMgr:               db.dbMgr,
		cacheProvider:       redisCache,
		cacheRepositoryName: cacheRepositoryName,
		cacheTTL:            -1,
	}

	if len(ttl) > 0 {
		newDB.cacheTTL = ttl[0]
	} else if configTTL, ok := cacheConfigs.Load(cacheRepositoryName); ok {
		newDB.cacheTTL = configTTL.(time.Duration)
	}

	return newDB
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

// GeneratePaginationCacheKey 为分页查询生成专门的缓存键
// 基于解析后的SQL结构生成更精确的缓存键，确保缓存键的唯一性和一致性
func GeneratePaginationCacheKey(dbName string, parsedSQL *ParsedSQL, page, pageSize int, args ...interface{}) string {
	hash := md5.New()

	// 数据库名称
	hash.Write([]byte(dbName))
	hash.Write([]byte(":"))

	// 分页标识
	hash.Write([]byte("PAGINATE"))
	hash.Write([]byte(":"))

	// 页码和页面大小
	hash.Write([]byte(fmt.Sprintf("p%d_s%d", page, pageSize)))
	hash.Write([]byte(":"))

	// SQL各个部分的哈希，确保结构化的一致性
	if parsedSQL.SelectClause != "" {
		hash.Write([]byte("SELECT:"))
		hash.Write([]byte(parsedSQL.SelectClause))
		hash.Write([]byte(":"))
	}

	if parsedSQL.FromClause != "" {
		hash.Write([]byte("FROM:"))
		hash.Write([]byte(parsedSQL.FromClause))
		hash.Write([]byte(":"))
	}

	if parsedSQL.WhereClause != "" {
		hash.Write([]byte("WHERE:"))
		hash.Write([]byte(parsedSQL.WhereClause))
		hash.Write([]byte(":"))
	}

	if parsedSQL.GroupByClause != "" {
		hash.Write([]byte("GROUP:"))
		hash.Write([]byte(parsedSQL.GroupByClause))
		hash.Write([]byte(":"))
	}

	if parsedSQL.HavingClause != "" {
		hash.Write([]byte("HAVING:"))
		hash.Write([]byte(parsedSQL.HavingClause))
		hash.Write([]byte(":"))
	}

	if parsedSQL.OrderByClause != "" {
		hash.Write([]byte("ORDER:"))
		hash.Write([]byte(parsedSQL.OrderByClause))
		hash.Write([]byte(":"))
	}

	// 复杂查询标识
	if parsedSQL.IsComplex {
		hash.Write([]byte("COMPLEX:"))
	}
	if parsedSQL.HasSubquery {
		hash.Write([]byte("SUBQUERY:"))
	}
	if parsedSQL.HasJoin {
		hash.Write([]byte("JOIN:"))
	}

	// 参数
	if len(args) > 0 {
		hash.Write([]byte("ARGS:"))
		hash.Write([]byte(fmt.Sprintf("%v", args)))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// GenerateCountCacheKey 为计数查询生成专门的缓存键
// 基于解析后的SQL结构生成计数查询的缓存键
func GenerateCountCacheKey(dbName string, parsedSQL *ParsedSQL, args ...interface{}) string {
	hash := md5.New()

	// 数据库名称
	hash.Write([]byte(dbName))
	hash.Write([]byte(":"))

	// 计数标识
	hash.Write([]byte("COUNT"))
	hash.Write([]byte(":"))

	// 只包含影响计数的SQL部分
	if parsedSQL.FromClause != "" {
		hash.Write([]byte("FROM:"))
		hash.Write([]byte(parsedSQL.FromClause))
		hash.Write([]byte(":"))
	}

	if parsedSQL.WhereClause != "" {
		hash.Write([]byte("WHERE:"))
		hash.Write([]byte(parsedSQL.WhereClause))
		hash.Write([]byte(":"))
	}

	if parsedSQL.GroupByClause != "" {
		hash.Write([]byte("GROUP:"))
		hash.Write([]byte(parsedSQL.GroupByClause))
		hash.Write([]byte(":"))
	}

	if parsedSQL.HavingClause != "" {
		hash.Write([]byte("HAVING:"))
		hash.Write([]byte(parsedSQL.HavingClause))
		hash.Write([]byte(":"))
	}

	// 复杂查询标识（影响计数逻辑）
	if parsedSQL.IsComplex {
		hash.Write([]byte("COMPLEX:"))
	}
	if parsedSQL.HasSubquery {
		hash.Write([]byte("SUBQUERY:"))
	}
	if parsedSQL.HasJoin {
		hash.Write([]byte("JOIN:"))
	}

	// 参数
	if len(args) > 0 {
		hash.Write([]byte("ARGS:"))
		hash.Write([]byte(fmt.Sprintf("%v", args)))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func getEffectiveTTL(cacheRepositoryName string, customTTL time.Duration) time.Duration {
	if customTTL >= 0 {
		return customTTL
	}
	if configTTL, ok := cacheConfigs.Load(cacheRepositoryName); ok {
		return configTTL.(time.Duration)
	}
	return defaultTTL
}
