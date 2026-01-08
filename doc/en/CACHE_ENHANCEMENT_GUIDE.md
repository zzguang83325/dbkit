# DBKit Cache Usage Guide

## Overview

DBKit  supports three cache invocation methods, covering all database query scenarios:
- **Cache()** - Use default cache (switchable via `SetDefaultCache`)
- **LocalCache()** - Force use of local memory cache
- **RedisCache()** - Force use of Redis cache

## Supported Scenarios

### 1. DB Instance Queries

#### Basic Queries
```go
// Use default cache
records, err := dbkit.Cache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)

// Use local cache
records, err := dbkit.LocalCache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)

// Use Redis cache
records, err := dbkit.RedisCache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)

// Cache with TTL
records, err := dbkit.LocalCache("user_cache", 5*time.Minute).Query("SELECT * FROM users")
```

#### QueryFirst
```go
// Query single record with local cache
record, err := dbkit.LocalCache("user_cache").QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// Use Redis cache
record, err := dbkit.RedisCache("user_cache", 10*time.Minute).QueryFirst("SELECT * FROM users WHERE id = ?", 1)
```

#### QueryMap
```go
// Return results in map format
results, err := dbkit.LocalCache("user_cache").QueryMap("SELECT * FROM users")
```

#### Count
```go
// Count query
count, err := dbkit.LocalCache("user_cache").Count("users", "age > ?", 18)
```

#### Pagination
```go
// Pagination with local cache
page, err := dbkit.LocalCache("user_cache").Paginate(1, 20, "SELECT * FROM users WHERE age > ?", 18)

// Pagination with Redis cache
page, err := dbkit.RedisCache("user_cache", 5*time.Minute).Paginate(1, 20, "SELECT * FROM users")
```

### 2. QueryBuilder Chain Queries

#### Basic Chain Queries
```go
// Use default cache
records, err := dbkit.Table("users").
    Cache("user_cache").
    Where("age > ?", 18).
    OrderBy("created_at DESC").
    Find()

// Use local cache
records, err := dbkit.Table("users").
    LocalCache("user_cache").
    Where("age > ?", 18).
    Find()

// Use Redis cache
records, err := dbkit.Table("users").
    RedisCache("user_cache", 10*time.Minute).
    Where("status = ?", "active").
    Find()
```

#### QueryBuilder Pagination
```go
// Pagination with local cache
page, err := dbkit.Table("users").
    LocalCache("user_cache").
    Where("age > ?", 18).
    OrderBy("created_at DESC").
    Paginate(1, 20)

// Pagination with Redis cache
page, err := dbkit.Table("users").
    RedisCache("user_cache", 5*time.Minute).
    Where("status = ?", "active").
    Paginate(1, 20)
```

#### QueryBuilder Count
```go
// Count query
count, err := dbkit.Table("users").
    LocalCache("user_cache").
    Where("age > ?", 18).
    Count()
```

#### QueryBuilder FindFirst
```go
// Query single record
record, err := dbkit.Table("users").
    LocalCache("user_cache").
    Where("email = ?", "user@example.com").
    FindFirst()
```

### 3. Transaction Cache

#### Transaction Basic Queries
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // Use default cache
    records, err := tx.Cache("tx_cache").Query("SELECT * FROM users WHERE age > ?", 18)
    if err != nil {
        return err
    }
    
    // Use local cache
    records, err = tx.LocalCache("tx_cache").Query("SELECT * FROM orders WHERE user_id = ?", 1)
    if err != nil {
        return err
    }
    
    // Use Redis cache
    records, err = tx.RedisCache("tx_cache", 3*time.Minute).Query("SELECT * FROM products")
    if err != nil {
        return err
    }
    
    return nil
})
```

#### Transaction QueryFirst
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    record, err := tx.LocalCache("tx_cache").QueryFirst("SELECT * FROM users WHERE id = ?", 1)
    if err != nil {
        return err
    }
    // Process record...
    return nil
})
```

#### Transaction Count
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    count, err := tx.LocalCache("tx_cache").Count("users", "age > ?", 18)
    if err != nil {
        return err
    }
    // Process count...
    return nil
})
```

#### Transaction Pagination
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    page, err := tx.LocalCache("tx_cache").Paginate(1, 20, "SELECT * FROM users WHERE age > ?", 18)
    if err != nil {
        return err
    }
    // Process page...
    return nil
})
```

#### Transaction QueryBuilder
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // QueryBuilder with local cache
    records, err := tx.Table("users").
        LocalCache("tx_cache").
        Where("age > ?", 18).
        Find()
    if err != nil {
        return err
    }
    
    // QueryBuilder with Redis cache
    page, err := tx.Table("orders").
        RedisCache("tx_cache", 5*time.Minute).
        Where("status = ?", "pending").
        Paginate(1, 20)
    if err != nil {
        return err
    }
    
    return nil
})
```

### 4. SQL Template Cache

#### SQL Template Basic Queries
```go
// Use default cache
records, err := dbkit.SqlTemplate("getUsersByAge", map[string]interface{}{
    "age": 18,
}).Cache("template_cache").Query()

// Use local cache
records, err := dbkit.SqlTemplate("getUsersByAge", map[string]interface{}{
    "age": 18,
}).LocalCache("template_cache").Query()

// Use Redis cache
records, err := dbkit.SqlTemplate("getUsersByAge", map[string]interface{}{
    "age": 18,
}).RedisCache("template_cache", 10*time.Minute).Query()
```

#### SQL Template QueryFirst
```go
// Query single record
record, err := dbkit.SqlTemplate("getUserById", map[string]interface{}{
    "id": 1,
}).LocalCache("template_cache").QueryFirst()
```

#### SQL Template Pagination
```go
// Pagination query
page, err := dbkit.SqlTemplate("getUserList", map[string]interface{}{
    "status": "active",
}).LocalCache("template_cache", 5*time.Minute).Paginate(1, 20)
```

#### SQL Template in Transaction
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    records, err := tx.SqlTemplate("getUsersByAge", map[string]interface{}{
        "age": 18,
    }).LocalCache("tx_template_cache").Query()
    if err != nil {
        return err
    }
    // Process results...
    return nil
})
```

#### SQL Template on Specific Database
```go
// Execute SQL template on specific database with cache
records, err := dbkit.Use("secondary_db").SqlTemplate("getUsersByAge", map[string]interface{}{
    "age": 18,
}).LocalCache("template_cache").Query()
```

## Cache Configuration

### Initialize Cache

```go
// 1. Initialize local cache (already initialized by default)
dbkit.InitLocalCache(1 * time.Minute) // Cleanup interval

// 2. Initialize Redis cache
redisCache, err := redis.NewRedisCache("localhost:6379", "", "", 0)
if err != nil {
    log.Fatal(err)
}
dbkit.InitRedisCache(redisCache)

// 3. Set default cache provider
dbkit.SetDefaultCache(redisCache) // Set Redis as default cache
```

### Set Default TTL

```go
// Set global default TTL
dbkit.SetDefaultTtl(5 * time.Minute)

// Set TTL for specific cache repository
dbkit.CreateCache("user_cache", 10 * time.Minute)
dbkit.CreateCache("order_cache", 3 * time.Minute)
```

## Cache Management

### Clear Cache

```go
// Clear specific repository
dbkit.CacheClearRepository("user_cache")

// Clear specific repository in local cache
dbkit.LocalCacheClearRepository("user_cache")

// Clear specific repository in Redis cache
dbkit.RedisCacheClearRepository("order_cache")

// Clear all caches
dbkit.ClearAllCaches()

// Clear all repositories in local cache
dbkit.LocalCacheClearAll()

// Clear all repositories in Redis cache
dbkit.RedisCacheClearAll()
```

### View Cache Status

```go
// View default cache status
status := dbkit.CacheStatus()
fmt.Printf("Cache status: %+v\n", status)

// View local cache status
localStatus := dbkit.LocalCacheStatus()
fmt.Printf("Local cache status: %+v\n", localStatus)

// View Redis cache status
redisStatus, err := dbkit.RedisCacheStatus()
if err == nil {
    fmt.Printf("Redis cache status: %+v\n", redisStatus)
}
```

## Cache Key Generation Rules

DBKit uses MD5 hash to generate cache keys, ensuring the same query uses the same cache:

```go
// Cache key for normal queries
key = MD5(dbName + sql + args)

// Cache key for pagination queries
key = MD5(dbName + "PAGINATE:" + sql + args)

// Cache key for Count queries
key = MD5(dbName + "COUNT:" + table + ":" + whereSql + args)

// Cache key for QueryFirst
key = MD5(dbName + sql + args) + "_first"

// Cache key for SQL template pagination
key = MD5(dbName + "PAGINATE_TEMPLATE:" + sql + args)
```

## Best Practices

### 1. Choose Appropriate Cache Type

- **Local Cache (LocalCache)**: Suitable for single-server deployment, small data volume, frequent access
- **Redis Cache (RedisCache)**: Suitable for distributed deployment, shared cache scenarios
- **Default Cache (Cache)**: Flexible switching, convenient unified management

### 2. Set Reasonable TTL

```go
// Frequently changing data uses shorter TTL
dbkit.LocalCache("hot_data", 1*time.Minute).Query(sql)

// Relatively stable data uses longer TTL
dbkit.LocalCache("config_data", 30*time.Minute).Query(sql)

// Almost unchanging data can use even longer TTL
dbkit.LocalCache("static_data", 24*time.Hour).Query(sql)
```

### 3. Cache Repository Naming Convention

```go
// Name by business module
dbkit.LocalCache("user_cache").Query(sql)
dbkit.LocalCache("order_cache").Query(sql)
dbkit.LocalCache("product_cache").Query(sql)

// Name by data type
dbkit.LocalCache("user_list_cache").Query(sql)
dbkit.LocalCache("user_detail_cache").QueryFirst(sql)
dbkit.LocalCache("user_count_cache").Count(table, where)
```

### 4. Use Cache Carefully in Transactions

```go
// Queries in transactions generally should not use cache, as they may read stale data
// But if you're certain the data won't be modified in the transaction, cache can improve performance
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // Reading config data that won't be modified, cache is OK
    config, err := tx.LocalCache("config_cache").QueryFirst("SELECT * FROM config WHERE key = ?", "app_name")
    
    // Reading data that might be modified, cache not recommended
    user, err := tx.Query("SELECT * FROM users WHERE id = ?", userId)
    
    return nil
})
```

### 5. Cache Invalidation Strategy

```go
// Clear related cache after data update
_, err := dbkit.Exec("UPDATE users SET name = ? WHERE id = ?", "New Name", 1)
if err == nil {
    // Clear user-related caches
    dbkit.CacheClearRepository("user_cache")
    dbkit.CacheClearRepository("user_list_cache")
}
```

## Performance Comparison

### Local Cache vs Redis Cache

| Feature | Local Cache | Redis Cache |
|---------|-------------|-------------|
| Access Speed | Very Fast (direct memory access) | Fast (network overhead) |
| Memory Usage | Uses application memory | Independent Redis service |
| Distributed Support | No | Yes |
| Persistence | No | Yes |
| Use Case | Single-server deployment | Distributed deployment |

### Cache Performance Example

```go
// Query without cache
start := time.Now()
records, _ := dbkit.Query("SELECT * FROM users WHERE age > ?", 18)
fmt.Printf("Query without cache: %v\n", time.Since(start))
// Output: Query without cache: 50ms

// First cache query (needs to query database)
start = time.Now()
records, _ = dbkit.LocalCache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)
fmt.Printf("First cache query: %v\n", time.Since(start))
// Output: First cache query: 50ms

// Second cache query (read from cache)
start = time.Now()
records, _ = dbkit.LocalCache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)
fmt.Printf("Cache hit query: %v\n", time.Since(start))
// Output: Cache hit query: 0.1ms
```

## Notes

1. **Redis cache requires initialization**: Must call `InitRedisCache()` before using `RedisCache()`
2. **Cache key uniqueness**: Same SQL and parameters generate the same cache key
3. **Cache consistency**: Related caches need to be manually cleared after data updates
4. **Memory management**: Local cache uses application memory, need reasonable TTL and cleanup strategy
5. **Transaction isolation**: Cache queries in transactions may read stale data from outside the transaction

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/zzguang83325/dbkit"
    "github.com/zzguang83325/dbkit/redis"
)

func main() {
    // 1. Initialize database
    err := dbkit.InitDatabase(&dbkit.Config{
        Driver:   "mysql",
        Host:     "localhost",
        Port:     3306,
        Database: "myapp",
        Username: "root",
        Password: "password",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. Initialize Redis cache (optional)
    redisCache, err := redis.NewRedisCache("localhost:6379", "", "", 0)
    if err != nil {
        log.Printf("Redis initialization failed: %v", err)
    } else {
        dbkit.InitRedisCache(redisCache)
    }
    
    // 3. Set cache configuration
    dbkit.SetDefaultTtl(5 * time.Minute)
    dbkit.CreateCache("user_cache", 10 * time.Minute)
    
    // 4. Query with local cache
    users, err := dbkit.LocalCache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d users\n", len(users))
    
    // 5. Use QueryBuilder with local cache
    activeUsers, err := dbkit.Table("users").
        LocalCache("user_cache").
        Where("status = ?", "active").
        OrderBy("created_at DESC").
        Find()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Active users: %d\n", len(activeUsers))
    
    // 6. Pagination query with Redis cache
    page, err := dbkit.RedisCache("user_cache", 5*time.Minute).
        Paginate(1, 20, "SELECT * FROM users WHERE age > ?", 18)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Page result: page %d, total %d records\n", page.PageNumber, page.TotalRow)
    
    // 7. Use cache in transaction
    err = dbkit.Transaction(func(tx *dbkit.Tx) error {
        // Query config (can use cache)
        config, err := tx.LocalCache("config_cache").
            QueryFirst("SELECT * FROM config WHERE key = ?", "app_name")
        if err != nil {
            return err
        }
        
        // Update user data
        _, err = tx.Exec("UPDATE users SET last_login = NOW() WHERE id = ?", 1)
        return err
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 8. Clear cache
    dbkit.CacheClearRepository("user_cache")
    
    // 9. View cache status
    status := dbkit.LocalCacheStatus()
    fmt.Printf("Local cache status: %+v\n", status)
}
```

Flexible cache strategies can significantly improve application performance. Choose the appropriate cache method based on your actual business scenarios.
