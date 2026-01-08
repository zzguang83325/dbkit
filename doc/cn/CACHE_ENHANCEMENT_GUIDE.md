# DBKit 缓存使用指南

## 概述

DBKit 支持三种缓存调用方式，覆盖所有查询数据库的场景：
- **Cache()** - 使用默认缓存（可通过 `SetDefaultCache` 切换）
- **LocalCache()** - 强制使用本地内存缓存
- **RedisCache()** - 强制使用 Redis 缓存

## 支持的场景

### 1. DB 实例查询

#### 基础查询
```go
// 使用默认缓存
records, err := dbkit.Cache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)

// 使用本地缓存
records, err := dbkit.LocalCache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)

// 使用 Redis 缓存
records, err := dbkit.RedisCache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)

// 带 TTL 的缓存
records, err := dbkit.LocalCache("user_cache", 5*time.Minute).Query("SELECT * FROM users")
```

#### QueryFirst 查询
```go
// 使用本地缓存查询单条记录
record, err := dbkit.LocalCache("user_cache").QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// 使用 Redis 缓存
record, err := dbkit.RedisCache("user_cache", 10*time.Minute).QueryFirst("SELECT * FROM users WHERE id = ?", 1)
```

#### QueryMap 查询
```go
// 返回 map 格式的结果
results, err := dbkit.LocalCache("user_cache").QueryMap("SELECT * FROM users")
```

#### Count 查询
```go
// 统计查询
count, err := dbkit.LocalCache("user_cache").Count("users", "age > ?", 18)
```

#### 分页查询
```go
// 使用本地缓存的分页查询
page, err := dbkit.LocalCache("user_cache").Paginate(1, 20, "SELECT * FROM users WHERE age > ?", 18)

// 使用 Redis 缓存的分页查询
page, err := dbkit.RedisCache("user_cache", 5*time.Minute).Paginate(1, 20, "SELECT * FROM users")
```

### 2. QueryBuilder 链式查询

#### 基础链式查询
```go
// 使用默认缓存
records, err := dbkit.Table("users").
    Cache("user_cache").
    Where("age > ?", 18).
    OrderBy("created_at DESC").
    Find()

// 使用本地缓存
records, err := dbkit.Table("users").
    LocalCache("user_cache").
    Where("age > ?", 18).
    Find()

// 使用 Redis 缓存
records, err := dbkit.Table("users").
    RedisCache("user_cache", 10*time.Minute).
    Where("status = ?", "active").
    Find()
```

#### QueryBuilder 分页
```go
// 使用本地缓存的分页
page, err := dbkit.Table("users").
    LocalCache("user_cache").
    Where("age > ?", 18).
    OrderBy("created_at DESC").
    Paginate(1, 20)

// 使用 Redis 缓存的分页
page, err := dbkit.Table("users").
    RedisCache("user_cache", 5*time.Minute).
    Where("status = ?", "active").
    Paginate(1, 20)
```

#### QueryBuilder Count
```go
// 统计查询
count, err := dbkit.Table("users").
    LocalCache("user_cache").
    Where("age > ?", 18).
    Count()
```

#### QueryBuilder FindFirst
```go
// 查询单条记录
record, err := dbkit.Table("users").
    LocalCache("user_cache").
    Where("email = ?", "user@example.com").
    FindFirst()
```

### 3. 事务中的缓存

#### 事务基础查询
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 使用默认缓存
    records, err := tx.Cache("tx_cache").Query("SELECT * FROM users WHERE age > ?", 18)
    if err != nil {
        return err
    }
    
    // 使用本地缓存
    records, err = tx.LocalCache("tx_cache").Query("SELECT * FROM orders WHERE user_id = ?", 1)
    if err != nil {
        return err
    }
    
    // 使用 Redis 缓存
    records, err = tx.RedisCache("tx_cache", 3*time.Minute).Query("SELECT * FROM products")
    if err != nil {
        return err
    }
    
    return nil
})
```

#### 事务 QueryFirst
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    record, err := tx.LocalCache("tx_cache").QueryFirst("SELECT * FROM users WHERE id = ?", 1)
    if err != nil {
        return err
    }
    // 处理记录...
    return nil
})
```

#### 事务 Count
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    count, err := tx.LocalCache("tx_cache").Count("users", "age > ?", 18)
    if err != nil {
        return err
    }
    // 处理统计结果...
    return nil
})
```

#### 事务分页
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    page, err := tx.LocalCache("tx_cache").Paginate(1, 20, "SELECT * FROM users WHERE age > ?", 18)
    if err != nil {
        return err
    }
    // 处理分页结果...
    return nil
})
```

#### 事务 QueryBuilder
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 使用本地缓存的 QueryBuilder
    records, err := tx.Table("users").
        LocalCache("tx_cache").
        Where("age > ?", 18).
        Find()
    if err != nil {
        return err
    }
    
    // 使用 Redis 缓存的 QueryBuilder
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

### 4. SQL 模板缓存

#### SQL 模板基础查询
```go
// 使用默认缓存
records, err := dbkit.SqlTemplate("getUsersByAge", map[string]interface{}{
    "age": 18,
}).Cache("template_cache").Query()

// 使用本地缓存
records, err := dbkit.SqlTemplate("getUsersByAge", map[string]interface{}{
    "age": 18,
}).LocalCache("template_cache").Query()

// 使用 Redis 缓存
records, err := dbkit.SqlTemplate("getUsersByAge", map[string]interface{}{
    "age": 18,
}).RedisCache("template_cache", 10*time.Minute).Query()
```

#### SQL 模板 QueryFirst
```go
// 查询单条记录
record, err := dbkit.SqlTemplate("getUserById", map[string]interface{}{
    "id": 1,
}).LocalCache("template_cache").QueryFirst()
```

#### SQL 模板分页
```go
// 分页查询
page, err := dbkit.SqlTemplate("getUserList", map[string]interface{}{
    "status": "active",
}).LocalCache("template_cache", 5*time.Minute).Paginate(1, 20)
```

#### SQL 模板在事务中使用
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    records, err := tx.SqlTemplate("getUsersByAge", map[string]interface{}{
        "age": 18,
    }).LocalCache("tx_template_cache").Query()
    if err != nil {
        return err
    }
    // 处理结果...
    return nil
})
```

#### SQL 模板在指定数据库上使用
```go
// 在指定数据库上执行 SQL 模板并使用缓存
records, err := dbkit.Use("secondary_db").SqlTemplate("getUsersByAge", map[string]interface{}{
    "age": 18,
}).LocalCache("template_cache").Query()
```

## 缓存配置

### 初始化缓存

```go
// 1. 初始化本地缓存（默认已初始化）
dbkit.InitLocalCache(1 * time.Minute) // 清理间隔

// 2. 初始化 Redis 缓存
redisCache, err := redis.NewRedisCache("localhost:6379", "", "", 0)
if err != nil {
    log.Fatal(err)
}
dbkit.InitRedisCache(redisCache)

// 3. 设置默认缓存提供者
dbkit.SetDefaultCache(redisCache) // 将 Redis 设为默认缓存
```

### 设置默认 TTL

```go
// 设置全局默认 TTL
dbkit.SetDefaultTtl(5 * time.Minute)

// 为特定缓存仓库设置 TTL
dbkit.CreateCache("user_cache", 10 * time.Minute)
dbkit.CreateCache("order_cache", 3 * time.Minute)
```

## 缓存管理

### 清除缓存

```go
// 清除指定仓库的缓存
dbkit.CacheClearRepository("user_cache")

// 清除本地缓存中的指定仓库
dbkit.LocalCacheClearRepository("user_cache")

// 清除 Redis 缓存中的指定仓库
dbkit.RedisCacheClearRepository("order_cache")

// 清除所有缓存
dbkit.ClearAllCaches()

// 清除本地缓存的所有仓库
dbkit.LocalCacheClearAll()

// 清除 Redis 缓存的所有仓库
dbkit.RedisCacheClearAll()
```

### 查看缓存状态

```go
// 查看默认缓存状态
status := dbkit.CacheStatus()
fmt.Printf("缓存状态: %+v\n", status)

// 查看本地缓存状态
localStatus := dbkit.LocalCacheStatus()
fmt.Printf("本地缓存状态: %+v\n", localStatus)

// 查看 Redis 缓存状态
redisStatus, err := dbkit.RedisCacheStatus()
if err == nil {
    fmt.Printf("Redis 缓存状态: %+v\n", redisStatus)
}
```

## 缓存键生成规则

DBKit 使用 MD5 哈希生成缓存键，确保相同的查询使用相同的缓存：

```go
// 普通查询的缓存键
key = MD5(dbName + sql + args)

// 分页查询的缓存键
key = MD5(dbName + "PAGINATE:" + sql + args)

// Count 查询的缓存键
key = MD5(dbName + "COUNT:" + table + ":" + whereSql + args)

// QueryFirst 的缓存键
key = MD5(dbName + sql + args) + "_first"

// SQL 模板分页的缓存键
key = MD5(dbName + "PAGINATE_TEMPLATE:" + sql + args)
```

## 最佳实践

### 1. 选择合适的缓存类型

- **本地缓存（LocalCache）**：适用于单机部署、数据量小、访问频繁的场景
- **Redis 缓存（RedisCache）**：适用于分布式部署、需要共享缓存的场景
- **默认缓存（Cache）**：灵活切换，便于统一管理

### 2. 设置合理的 TTL

```go
// 频繁变化的数据使用较短的 TTL
dbkit.LocalCache("hot_data", 1*time.Minute).Query(sql)

// 相对稳定的数据使用较长的 TTL
dbkit.LocalCache("config_data", 30*time.Minute).Query(sql)

// 几乎不变的数据可以使用更长的 TTL
dbkit.LocalCache("static_data", 24*time.Hour).Query(sql)
```

### 3. 缓存仓库命名规范

```go
// 按业务模块命名
dbkit.LocalCache("user_cache").Query(sql)
dbkit.LocalCache("order_cache").Query(sql)
dbkit.LocalCache("product_cache").Query(sql)

// 按数据类型命名
dbkit.LocalCache("user_list_cache").Query(sql)
dbkit.LocalCache("user_detail_cache").QueryFirst(sql)
dbkit.LocalCache("user_count_cache").Count(table, where)
```

### 4. 事务中谨慎使用缓存

```go
// 事务中的查询通常不建议使用缓存，因为可能读取到过期数据
// 但如果确定数据不会在事务中被修改，可以使用缓存提升性能
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 读取不会被修改的配置数据，可以使用缓存
    config, err := tx.LocalCache("config_cache").QueryFirst("SELECT * FROM config WHERE key = ?", "app_name")
    
    // 读取可能被修改的数据，不建议使用缓存
    user, err := tx.Query("SELECT * FROM users WHERE id = ?", userId)
    
    return nil
})
```

### 5. 缓存失效策略

```go
// 数据更新后清除相关缓存
_, err := dbkit.Exec("UPDATE users SET name = ? WHERE id = ?", "新名字", 1)
if err == nil {
    // 清除用户相关的缓存
    dbkit.CacheClearRepository("user_cache")
    dbkit.CacheClearRepository("user_list_cache")
}
```

## 性能对比

### 本地缓存 vs Redis 缓存

| 特性 | 本地缓存 | Redis 缓存 |
|------|---------|-----------|
| 访问速度 | 极快（内存直接访问） | 快（网络开销） |
| 内存占用 | 占用应用内存 | 独立 Redis 服务 |
| 分布式支持 | 不支持 | 支持 |
| 持久化 | 不支持 | 支持 |
| 适用场景 | 单机部署 | 分布式部署 |

### 缓存效果示例

```go
// 无缓存查询
start := time.Now()
records, _ := dbkit.Query("SELECT * FROM users WHERE age > ?", 18)
fmt.Printf("无缓存查询耗时: %v\n", time.Since(start))
// 输出: 无缓存查询耗时: 50ms

// 首次缓存查询（需要查询数据库）
start = time.Now()
records, _ = dbkit.LocalCache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)
fmt.Printf("首次缓存查询耗时: %v\n", time.Since(start))
// 输出: 首次缓存查询耗时: 50ms

// 第二次缓存查询（从缓存读取）
start = time.Now()
records, _ = dbkit.LocalCache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)
fmt.Printf("缓存命中查询耗时: %v\n", time.Since(start))
// 输出: 缓存命中查询耗时: 0.1ms
```

## 注意事项

1. **Redis 缓存需要先初始化**：使用 `RedisCache()` 前必须先调用 `InitRedisCache()`
2. **缓存键的唯一性**：相同的 SQL 和参数会生成相同的缓存键
3. **缓存一致性**：数据更新后需要手动清除相关缓存
4. **内存管理**：本地缓存会占用应用内存，需要合理设置 TTL 和清理策略
5. **事务隔离**：事务中的缓存查询可能读取到事务外的旧数据

## 完整示例

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
    // 1. 初始化数据库
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
    
    // 2. 初始化 Redis 缓存（可选）
    redisCache, err := redis.NewRedisCache("localhost:6379", "", "", 0)
    if err != nil {
        log.Printf("Redis 初始化失败: %v", err)
    } else {
        dbkit.InitRedisCache(redisCache)
    }
    
    // 3. 设置缓存配置
    dbkit.SetDefaultTtl(5 * time.Minute)
    dbkit.CreateCache("user_cache", 10 * time.Minute)
    
    // 4. 使用本地缓存查询
    users, err := dbkit.LocalCache("user_cache").Query("SELECT * FROM users WHERE age > ?", 18)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("查询到 %d 个用户\n", len(users))
    
    // 5. 使用 QueryBuilder 和本地缓存
    activeUsers, err := dbkit.Table("users").
        LocalCache("user_cache").
        Where("status = ?", "active").
        OrderBy("created_at DESC").
        Find()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("活跃用户: %d 个\n", len(activeUsers))
    
    // 6. 使用 Redis 缓存分页查询
    page, err := dbkit.RedisCache("user_cache", 5*time.Minute).
        Paginate(1, 20, "SELECT * FROM users WHERE age > ?", 18)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("分页结果: 第 %d 页，共 %d 条记录\n", page.PageNumber, page.TotalRow)
    
    // 7. 在事务中使用缓存
    err = dbkit.Transaction(func(tx *dbkit.Tx) error {
        // 查询配置（可以使用缓存）
        config, err := tx.LocalCache("config_cache").
            QueryFirst("SELECT * FROM config WHERE key = ?", "app_name")
        if err != nil {
            return err
        }
        
        // 更新用户数据
        _, err = tx.Exec("UPDATE users SET last_login = NOW() WHERE id = ?", 1)
        return err
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 8. 清除缓存
    dbkit.CacheClearRepository("user_cache")
    
    // 9. 查看缓存状态
    status := dbkit.LocalCacheStatus()
    fmt.Printf("本地缓存状态: %+v\n", status)
}
```



灵活的缓存策略可以显著提升应用性能，建议根据实际业务场景选择合适的缓存方式。
