# DBKit 示例使用指南

本文档提供了 DBKit 各个示例的详细说明和使用方法。

## 目录

1. [Timestamp 示例](#timestamp-示例)
2. [Soft Delete 示例](#soft-delete-示例)
3. [Optimistic Lock 示例](#optimistic-lock-示例)
4. [MySQL 示例](#mysql-示例)
5. [PostgreSQL 示例](#postgresql-示例)

---

## Timestamp 示例

**位置**: `examples/timestamp/main.go`

**功能**: 演示自动时间戳功能

### 核心概念

- **created_at**: 记录创建时间，插入时自动填充，不会被修改
- **updated_at**: 记录最后修改时间，每次更新时自动填充
- **自定义字段名**: 支持使用不同的字段名称

### 主要 API

```go
// 启用时间戳功能（全局）
dbkit.EnableTimestamps()

// 为表配置时间戳
dbkit.ConfigTimestamps("users")

// 使用自定义字段名
dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")

// 只配置 created_at
dbkit.ConfigCreatedAt("logs", "log_time")

// 禁用时间戳更新
dbkit.Table("users").Where("id = ?", id).WithoutTimestamps().Update(record)
```

### 使用场景

- 记录数据创建和修改时间
- 审计日志
- 数据版本控制
- 数据恢复

### 示例代码

```go
// 1. 插入时自动填充 created_at
record := dbkit.NewRecord()
record.Set("name", "John Doe")
record.Set("email", "john@example.com")
id, _ := dbkit.Insert("users", record)
// created_at 会自动设置为当前时间

// 2. 更新时自动填充 updated_at
updateRecord := dbkit.NewRecord()
updateRecord.Set("name", "John Updated")
dbkit.Update("users", updateRecord, "id = ?", id)
// updated_at 会自动设置为当前时间

// 3. 禁用时间戳更新
dbkit.Table("users").Where("id = ?", id).WithoutTimestamps().Update(record)
// updated_at 不会被修改
```

---

## Soft Delete 示例

**位置**: `examples/soft_delete/main.go`

**功能**: 演示软删除功能

### 核心概念

- **软删除**: 标记记录为已删除，但不从数据库中移除
- **deleted_at**: 删除标记字段，NULL 表示活跃，有值表示已删除
- **恢复**: 将软删除的记录恢复为活跃状态
- **强制删除**: 永久删除记录

### 主要 API

```go
// 启用软删除功能（全局）
dbkit.EnableSoftDelete()

// 为表配置软删除
dbkit.ConfigSoftDelete("users", "deleted_at")

// 软删除记录
dbkit.Delete("users", "id = ?", id)

// 恢复软删除的记录
dbkit.Restore("users", "id = ?", id)

// 永久删除记录
dbkit.ForceDelete("users", "id = ?", id)

// 查询包含已删除的记录
dbkit.Table("users").WithTrashed().Find()

// 只查询已删除的记录
dbkit.Table("users").OnlyTrashed().Find()

// 查询活跃记录（默认行为）
dbkit.Table("users").Find()
```

### 使用场景

- 用户注销但保留历史数据
- 订单取消但保留审计日志
- 数据恢复和撤销操作
- 合规性要求（保留删除记录）

### 示例代码

```go
// 1. 软删除记录
dbkit.Delete("users", "id = ?", 2)
// 记录的 deleted_at 被设置为当前时间

// 2. 查询活跃用户（不包括已删除）
records, _ := dbkit.Table("users").Find()
// 只返回 deleted_at 为 NULL 的记录

// 3. 查询所有用户（包括已删除）
records, _ := dbkit.Table("users").WithTrashed().Find()
// 返回所有记录

// 4. 恢复已删除的记录
dbkit.Restore("users", "id = ?", 2)
// 将 deleted_at 设置为 NULL

// 5. 永久删除记录
dbkit.ForceDelete("users", "id = ?", 3)
// 从数据库中完全删除记录
```

---

## Optimistic Lock 示例

**位置**: `examples/optimistic_lock/main.go`

**功能**: 演示乐观锁功能

### 核心概念

- **乐观锁**: 通过版本号检测并发修改
- **版本号**: 每条记录有一个版本号，更新时版本号自动加 1
- **冲突检测**: 版本号不匹配时更新失败
- **ErrVersionMismatch**: 版本冲突错误

### 主要 API

```go
// 启用乐观锁功能（全局）
dbkit.EnableOptimisticLock()

// 为表配置乐观锁
dbkit.ConfigOptimisticLock("products")

// 使用自定义版本字段名
dbkit.ConfigOptimisticLockWithField("orders", "revision")

// 检查版本冲突错误
if errors.Is(err, dbkit.ErrVersionMismatch) {
    // 处理版本冲突
}
```

### 使用场景

- 防止并发更新冲突
- 库存管理（防止超卖）
- 订单状态更新
- 数据一致性保证

### 示例代码

```go
// 1. 插入记录时自动初始化版本号为 1
record := dbkit.NewRecord()
record.Set("name", "Laptop")
record.Set("price", 999.99)
id, _ := dbkit.Insert("products", record)
// version 自动设置为 1

// 2. 使用正确的版本号更新
updateRecord := dbkit.NewRecord()
updateRecord.Set("version", int64(1))  // 当前版本号
updateRecord.Set("price", 899.99)
rows, err := dbkit.Update("products", updateRecord, "id = ?", id)
// 更新成功，version 自动变为 2

// 3. 使用过期的版本号更新（会失败）
staleRecord := dbkit.NewRecord()
staleRecord.Set("version", int64(1))  // 过期的版本号
staleRecord.Set("price", 799.99)
rows, err := dbkit.Update("products", staleRecord, "id = ?", id)
// 返回 ErrVersionMismatch 错误

// 4. 正确处理并发更新
latestRecord, _ := dbkit.Table("products").Where("id = ?", id).FindFirst()
currentVersion := latestRecord.GetInt("version")
updateRecord := dbkit.NewRecord()
updateRecord.Set("version", currentVersion)
updateRecord.Set("price", 799.99)
rows, err := dbkit.Update("products", updateRecord, "id = ?", id)
// 使用最新版本号，更新成功
```

---

## MySQL 示例

**位置**: `examples/mysql/main.go`

**功能**: 演示 MySQL 数据库的各种操作

### 前置条件

- MySQL 数据库已启动
- 数据库连接信息正确
- 有相应的数据库和表权限

### 连接字符串格式

```
user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
```

参数说明:
- `user`: 数据库用户名
- `password`: 数据库密码
- `host`: 数据库主机地址
- `port`: 数据库端口（默认 3306）
- `dbname`: 数据库名称
- `charset`: 字符集（推荐 utf8mb4）
- `parseTime`: 是否解析时间字段
- `loc`: 时区设置

### 主要 API

#### 连接管理

```go
// 打开数据库连接
dbkit.OpenDatabase(dbkit.MySQL, dsn, 25)

// 使用指定的数据库名称打开连接
dbkit.OpenDatabaseWithDBName("mysql", dbkit.MySQL, dsn, 25)

// 测试连接
dbkit.PingDB("mysql")

// 启用调试模式
dbkit.SetDebugMode(true)
```

#### Record CRUD 操作

```go
// 插入记录
record := dbkit.NewRecord().Set("name", "John").Set("age", 30)
id, err := dbkit.Use("mysql").Insert("users", record)

// 查询多条记录
records, err := dbkit.Use("mysql").Query("SELECT * FROM users WHERE age > ?", 25)

// 查询单条记录
record, err := dbkit.Use("mysql").QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// 更新记录
updateRec := dbkit.NewRecord().Set("age", 35)
affected, err := dbkit.Use("mysql").Update("users", updateRec, "id = ?", 1)

// 保存记录（插入或更新）
affected, err := dbkit.Use("mysql").Save("users", record)

// 统计记录数
count, err := dbkit.Use("mysql").Count("users", "age > ?", 25)

// 检查记录是否存在
exists, err := dbkit.Use("mysql").Exists("users", "id = ?", 1)

// 删除记录
affected, err := dbkit.Use("mysql").Delete("users", "id = ?", 1)
```

#### 链式查询

```go
// 基本链式查询
records, err := dbkit.Table("users").
    Where("age > ?", 25).
    OrderBy("age DESC").
    Limit(10).
    Find()

// 分页查询
page, err := dbkit.Table("users").
    Where("age > ?", 25).
    Paginate(1, 10)

// 查询单条记录
record, err := dbkit.Table("users").
    Where("id = ?", 1).
    FindFirst()

// 统计记录数
count, err := dbkit.Table("users").
    Where("age > ?", 25).
    Count()
```

#### 批量操作

```go
// 批量插入
records := make([]*dbkit.Record, 0, 100)
for i := 1; i <= 100; i++ {
    record := dbkit.NewRecord().Set("name", fmt.Sprintf("User_%d", i))
    records = append(records, record)
}
affected, err := dbkit.Use("mysql").BatchInsert("users", records, 50)
```

#### 事务处理

```go
// 事务操作
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 在事务中执行操作
    _, err := tx.Insert("users", record)
    if err != nil {
        return err
    }
    
    _, err = tx.Update("users", updateRec, "id = ?", 1)
    return err
})
```

#### 缓存操作

```go
// 查询并缓存结果
records, err := dbkit.Cache("user_cache").Query("SELECT * FROM users WHERE age > ?", 25)

// 分页查询并缓存
page, err := dbkit.Cache("user_page_cache").Paginate(1, 10, "SELECT * FROM users", "users", "", "")

// 统计并缓存
count, err := dbkit.Cache("user_count_cache").Count("users", "age > ?", 25)
```

### 使用场景

- 基本的 CRUD 操作
- 复杂的查询和过滤
- 数据分页
- 事务处理
- 性能优化（缓存）

---

## PostgreSQL 示例

**位置**: `examples/postgresql/main.go`

**功能**: 演示 PostgreSQL 数据库的各种操作

### 前置条件

- PostgreSQL 数据库已启动
- 数据库连接信息正确
- 有相应的数据库和表权限

### 连接字符串格式

```
user=postgres password=123456 host=127.0.0.1 port=5432 dbname=postgres sslmode=disable
```

参数说明:
- `user`: 数据库用户名
- `password`: 数据库密码
- `host`: 数据库主机地址
- `port`: 数据库端口（默认 5432）
- `dbname`: 数据库名称
- `sslmode`: SSL 模式（disable/require/prefer）

### 主要 API

API 与 MySQL 基本相同，主要区别在于：

1. **连接字符串格式不同**
2. **参数占位符**: PostgreSQL 使用 `$1, $2` 而不是 `?`
3. **JSONB 支持**: PostgreSQL 原生支持 JSONB 类型
4. **数组类型**: PostgreSQL 支持数组类型

### 示例代码

```go
// 连接 PostgreSQL
dsn := "user=postgres password=123456 host=127.0.0.1 port=5432 dbname=postgres sslmode=disable"
dbkit.OpenDatabaseWithDBName("postgresql", dbkit.PostgreSQL, dsn, 25)

// 其他操作与 MySQL 相同
records, err := dbkit.Table("users").Where("age > ?", 25).Find()
```

---

## 快速开始

### 1. 时间戳功能

```bash
cd examples/timestamp
go run main.go
```

### 2. 软删除功能

```bash
cd examples/soft_delete
go run main.go
```

### 3. 乐观锁功能

```bash
cd examples/optimistic_lock
go run main.go
```

### 4. MySQL 综合测试

```bash
cd examples/mysql
go run main.go
```

### 5. PostgreSQL 综合测试

```bash
cd examples/postgresql
go run main.go
```

---

## 常见问题

### Q: 如何选择 Record 还是 DbModel？

**A**: 
- **Record**: 灵活、轻量级，适合动态数据和快速开发
- **DbModel**: 类型安全、结构化，适合大型项目和复杂业务逻辑

### Q: 如何处理并发更新冲突？

**A**: 使用乐观锁：
1. 读取最新版本号
2. 在更新时指定版本号
3. 如果版本号不匹配，重试操作

### Q: 如何提高查询性能？

**A**: 
1. 使用缓存机制
2. 使用分页查询
3. 添加适当的数据库索引
4. 使用连接池

### Q: 如何处理事务中的错误？

**A**: 在事务回调函数中返回错误，DBKit 会自动回滚：

```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    if err := doSomething(tx); err != nil {
        return err  // 自动回滚
    }
    return nil  // 自动提交
})
```

---

## 更多资源

- [DBKit 官方文档](https://github.com/zzguang83325/dbkit)
- [API 参考](../../api.md)
- [最佳实践](./BEST_PRACTICES.md)
