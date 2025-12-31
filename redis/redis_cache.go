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

func (r *redisCache) CacheGet(cacheName, key string) (interface{}, bool) {
	fullKey := fmt.Sprintf("dbkit:%s:%s", cacheName, key)
	val, err := r.client.Get(r.ctx, fullKey).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		return nil, false
	}

	return val, true
}

func (r *redisCache) CacheSet(cacheName, key string, value interface{}, ttl time.Duration) {
	fullKey := fmt.Sprintf("dbkit:%s:%s", cacheName, key)

	var data interface{} = value
	switch value.(type) {
	case string, []byte:
		data = value
	default:
		if jsonData, err := json.Marshal(value); err == nil {
			data = jsonData
		}
	}

	r.client.Set(r.ctx, fullKey, data, ttl)
}

func (r *redisCache) CacheDelete(cacheName, key string) {
	fullKey := fmt.Sprintf("dbkit:%s:%s", cacheName, key)
	r.client.Del(r.ctx, fullKey)
}

func (r *redisCache) CacheClear(cacheName string) {
	pattern := fmt.Sprintf("dbkit:%s:*", cacheName)
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
