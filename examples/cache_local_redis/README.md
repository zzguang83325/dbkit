# 本地缓存 vs Redis 缓存演示

本示例演示如何在 dbkit 中灵活使用本地缓存和 Redis 缓存。

## 功能特性

### 1. 显式指定缓存类型

```go
// 使用本地缓存（速度最快，单实例）
user, _ := dbkit.LocalCache("user_cache").QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// 使用 Redis 缓存（分布式共享）
order, _ := dbkit.RedisCache("order_cache").Query("SELECT * FROM orders WHERE user_id = ?", userId)
```

### 2. 使用默认缓存

```go
// 使用默认缓存（默认是本地缓存）
data, _ := dbkit.Cache("default_cache").QueryFirst("SELECT * FROM configs WHERE key = ?", key)
```

### 3. 切换默认缓存

```go
// 初始化 Redis 缓存
rc, _ := redis.NewRedisCache("localhost:6379", "", "", 0)
dbkit.InitRedisCache(rc)

// 切换默认缓存为 Redis
dbkit.SetDefaultCache(rc)

// 现在 Cache() 使用的是 Redis
data, _ := dbkit.Cache("cache_name").Query("SELECT * FROM table")
```

## API 说明

### LocalCache(cacheRepositoryName, ttl...)

显式使用本地缓存创建查询构建器。

**特点：**
- 速度最快（微秒级延迟）
- 内存存储，单实例独享
- 适合配置数据、字典数据等很少变化的数据

**示例：**
```go
// 基本用法
user, _ := dbkit.LocalCache("user_cache").QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// 自定义 TTL
configs, _ := dbkit.LocalCache("config_cache", 10*time.Minute).Query("SELECT * FROM configs")

// 分页查询
page, _ := dbkit.LocalCache("page_cache").Paginate(1, 10, "SELECT * FROM products ORDER BY id")
```

### RedisCache(cacheRepositoryName, ttl...)

显式使用 Redis 缓存创建查询构建器。

**特点：**
- 分布式共享，多实例可见
- 持久化存储，重启不丢失
- 适合业务数据、需要跨实例共享的数据

**示例：**
```go
// 基本用法
order, _ := dbkit.RedisCache("order_cache").QueryFirst("SELECT * FROM orders WHERE id = ?", orderId)

// 自定义 TTL
users, _ := dbkit.RedisCache("user_list", 5*time.Minute).Query("SELECT * FROM users WHERE age > ?", 18)

// 分页查询
page, _ := dbkit.RedisCache("order_page").Paginate(1, 20, "SELECT * FROM orders ORDER BY created_at DESC")
```

### Cache(cacheRepositoryName, ttl...)

使用默认缓存创建查询构建器。

**特点：**
- 默认使用本地缓存
- 可通过 `SetDefaultCache()` 切换为 Redis 或其他缓存
- 适合需要灵活切换缓存策略的场景

**示例：**
```go
// 使用默认缓存
data, _ := dbkit.Cache("cache_name").QueryFirst("SELECT * FROM table WHERE id = ?", id)

// 切换默认缓存后，所有 Cache() 调用都会使用新的缓存
dbkit.SetDefaultCache(redisCache)
data, _ := dbkit.Cache("cache_name").QueryFirst("SELECT * FROM table WHERE id = ?", id) // 现在用 Redis
```

## 使用场景

### 场景 1：配置管理（本地缓存）

```go
func GetSystemConfig(key string) (string, error) {
    record, err := dbkit.LocalCache("system_config", 10*time.Minute).
        QueryFirst("SELECT value FROM configs WHERE key = ?", key)
    if err != nil {
        return "", err
    }
    return record.GetString("value"), nil
}
```

### 场景 2：用户信息（Redis 缓存）

```go
func GetUserByID(userID int) (*User, error) {
    record, err := dbkit.RedisCache("user_info", 5*time.Minute).
        QueryFirst("SELECT * FROM users WHERE id = ?", userID)
    if err != nil {
        return nil, err
    }
    
    return &User{
        ID:    record.GetInt("id"),
        Name:  record.GetString("name"),
        Email: record.GetString("email"),
    }, nil
}
```

### 场景 3：混合使用

```go
func GetDashboardData(userID int) (*Dashboard, error) {
    // 配置数据用本地缓存（快速访问）
    configs, _ := dbkit.LocalCache("configs").Query("SELECT * FROM configs")
    
    // 用户数据用 Redis（多实例共享）
    user, _ := dbkit.RedisCache("users").QueryFirst("SELECT * FROM users WHERE id = ?", userID)
    
    // 统计数据用默认缓存
    stats, _ := dbkit.Cache("stats").Query("SELECT * FROM statistics WHERE user_id = ?", userID)
    
    return &Dashboard{
        Configs: configs,
        User:    user,
        Stats:   stats,
    }, nil
}
```

### 场景 4：环境切换

```go
// 开发环境：使用本地缓存
func InitDevelopment() {
    dbkit.OpenDatabase(dbkit.MySQL, devDSN, 10)
    // 默认就是本地缓存，无需额外配置
}

// 生产环境：使用 Redis 缓存
func InitProduction() {
    dbkit.OpenDatabase(dbkit.MySQL, prodDSN, 100)
    
    // 初始化 Redis
    rc, _ := redis.NewRedisCache("redis-cluster:6379", "", "password", 0)
    dbkit.InitRedisCache(rc)
    
    // 切换默认缓存为 Redis
    dbkit.SetDefaultCache(rc)
}

// 业务代码无需修改，自动使用对应环境的缓存
func GetUser(id int) (*User, error) {
    record, err := dbkit.Cache("user").
        QueryFirst("SELECT * FROM users WHERE id = ?", id)
    // ...
}
```

## 初始化步骤

### 1. 初始化数据库

```go
dbkit.OpenDatabase(dbkit.MySQL, "root:password@tcp(localhost:3306)/test", 10)
defer dbkit.Close()
```

### 2. 初始化 Redis 缓存（可选）

```go
rc, err := redis.NewRedisCache("localhost:6379", "", "", 0)
if err != nil {
    panic(err)
}
dbkit.InitRedisCache(rc)
```

### 3. 设置默认缓存（可选）

```go
// 如果需要将 Redis 设为默认缓存
dbkit.SetDefaultCache(rc)
```

## 性能对比

| 缓存类型 | 延迟 | 吞吐量 | 分布式 | 持久化 | 适用场景 |
|---------|------|--------|--------|--------|---------|
| 本地缓存 | ~1μs | 极高 | ❌ | ❌ | 配置、字典、单实例 |
| Redis 缓存 | ~1ms | 高 | ✅ | ✅ | 业务数据、多实例 |

## 注意事项

1. **本地缓存**：
   - 每个实例独立，不共享
   - 重启后数据丢失
   - 适合读多写少的配置数据

2. **Redis 缓存**：
   - 需要先调用 `InitRedisCache()` 初始化
   - 多实例共享，需要考虑缓存一致性
   - 适合需要分布式共享的业务数据

3. **默认缓存**：
   - 初始默认为本地缓存
   - 可通过 `SetDefaultCache()` 全局切换
   - 切换后所有 `Cache()` 调用都会使用新的缓存

## 运行示例

### 1. 完整演示（本地缓存 vs Redis 缓存）
```bash
go run .
```
或
```bash
go run main.go
```

### 2. Table() 方法缓存继承测试
```bash
go run . test
```

测试 `DB.LocalCache().Table()` 和 `Tx.LocalCache().Table()` 等链式调用的缓存继承功能。

### 3. SQL 模板分页测试
```bash
go run . sqltemplate
```

测试 SQL 模板的分页查询功能,包括:
- 无参数、单参数、多参数、命名参数的分页查询
- 本地缓存和 Redis 缓存的分页查询
- DB 实例和事务中的 SQL 模板分页
- 不同页码的缓存策略

## 前置条件

### MySQL 数据库
确保 MySQL 数据库正在运行,连接信息:
```
Host: localhost
Port: 3306
User: root
Password: 123456
Database: test
```

### Redis 服务器（可选）
如果要测试 Redis 缓存功能:
```
Host: 192.168.10.205
Port: 6379
Password: (空)
KeyPrefix: cpdv3
DB: 2
```

**注意**: 如果 Redis 未配置,程序会记录错误但不会中断,本地缓存功能仍然可用。

## 输出示例

```
=== 本地缓存 vs Redis 缓存演示 ===

[1] 连接 MySQL 数据库...
✓ 数据库连接成功

[2] 准备测试数据...
✓ 测试数据准备完成 (50 条记录)

[3] 初始化 Redis 缓存...
✓ Redis 缓存已初始化

【演示 1】显式使用本地缓存
  第 1 次查询 (从数据库): 耗时 2.5ms
  结果: id=1, name=User1
  第 2 次查询 (从本地缓存): 耗时 15μs
  结果: id=1, name=User1
  ⚡ 性能提升: 166.7x 倍

【演示 2】显式使用 Redis 缓存
  第 1 次查询 (从数据库): 耗时 2.3ms
  结果: id=2, name=User2
  第 2 次查询 (从 Redis 缓存): 耗时 850μs
  结果: id=2, name=User2
  ⚡ 性能提升: 2.7x 倍

...
```
