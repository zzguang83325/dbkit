package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// redisCache implements CacheProvider using Redis
type redisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new Redis-backed cache provider
func NewRedisCache(addr, username, password string, db int) (*redisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       db,
	})

	rc := &redisCache{
		client: client,
		ctx:    context.Background(),
	}

	// 测试连接
	if err := rc.client.Ping(rc.ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %v", err)
	}

	return rc, nil
}

// CacheGet 从 Redis 获取缓存值
// 优化：直接返回 JSON 字节数组，避免字符串转换开销
func (r *redisCache) CacheGet(cacheRepositoryName, key string) (interface{}, bool) {
	fullKey := fmt.Sprintf("dbkit:%s:%s", cacheRepositoryName, key)

	// 使用 Bytes() 而不是 Result()，避免字符串转换
	val, err := r.client.Get(r.ctx, fullKey).Bytes()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		return nil, false
	}

	// 直接返回字节数组，让 convertCacheValue 处理反序列化
	// 这样可以避免双重序列化：string → []byte → object
	return val, true
}

func (r *redisCache) CacheSet(cacheRepositoryName, key string, value interface{}, ttl time.Duration) {
	fullKey := fmt.Sprintf("dbkit:%s:%s", cacheRepositoryName, key)

	var data interface{} = value
	switch value.(type) {
	case string, []byte:
		data = value
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			// 序列化失败，记录日志并跳过存储
			fmt.Printf("dbkit: redis cache marshal failed, key=%s, error=%v\n", fullKey, err)
			return
		}
		data = jsonData
	}

	r.client.Set(r.ctx, fullKey, data, ttl)
}

func (r *redisCache) CacheDelete(cacheRepositoryName, key string) {
	fullKey := fmt.Sprintf("dbkit:%s:%s", cacheRepositoryName, key)
	r.client.Del(r.ctx, fullKey)
}

func (r *redisCache) CacheClear(cacheRepositoryName string) {
	pattern := fmt.Sprintf("dbkit:%s:*", cacheRepositoryName)
	iter := r.client.Scan(r.ctx, 0, pattern, 0).Iterator()
	for iter.Next(r.ctx) {
		r.client.Del(r.ctx, iter.Val())
	}
}

func (r *redisCache) Status() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["type"] = "RedisCache"
	stats["address"] = r.client.Options().Addr

	info, err := r.client.Info(r.ctx, "memory").Result()
	if err == nil {
		stats["redis_info_memory"] = info
	}

	dbSize, err := r.client.DBSize(r.ctx).Result()
	if err == nil {
		stats["db_size"] = dbSize
	}

	return stats
}
