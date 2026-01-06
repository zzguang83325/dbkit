# DBKit API 手册

[English Version](api_en.md) | [README](README.md) | [English README](README_EN.md)

## 目录

- [数据库初始化](#数据库初始化)
- [查询操作](#查询操作)
- [查询超时控制](#查询超时控制)
- [插入与更新](#插入与更新)
- [删除操作](#删除操作)
- [软删除](#软删除)
- [自动时间戳](#自动时间戳)
- [乐观锁](#乐观锁)
- [事务处理](#事务处理)
- [Record 对象](#record-对象)
- [链式查询](#链式查询)
- [DbModel 操作](#dbmodel-操作)
- [缓存操作](#缓存操作)
- [SQL 模板](#sql-模板)
- [日志配置](#日志配置)
- [工具函数](#工具函数)

---

## 数据库初始化

### OpenDatabase
```go
func OpenDatabase(driver DriverType, dsn string, maxOpen int) error
```
使用默认配置打开数据库连接。

**参数:**
- `driver`: 数据库驱动类型 (MySQL, PostgreSQL, SQLite3, Oracle, SQLServer)
- `dsn`: 数据源名称（连接字符串）
- `maxOpen`: 最大打开连接数

**示例:**
```go
err := dbkit.OpenDatabase(dbkit.MySQL, "root:password@tcp(localhost:3306)/test", 10)
```

### OpenDatabaseWithConfig
```go
func OpenDatabaseWithConfig(config *Config) error
```
使用自定义配置打开数据库连接。

**Config 结构体:**
```go
type Config struct {
    Driver          DriverType    // 数据库驱动类型
    DSN             string        // 数据源名称
    MaxOpen         int           // 最大打开连接数
    MaxIdle         int           // 最大空闲连接数
    ConnMaxLifetime time.Duration // 连接最大生命周期
    QueryTimeout    time.Duration // 默认查询超时时间（0表示不限制）
}
```

### OpenDatabaseWithDBName
```go
func OpenDatabaseWithDBName(dbname string, driver DriverType, dsn string, maxOpen int) error
```
以指定名称打开数据库连接（多数据库模式）。

### Register
```go
func Register(dbname string, config *Config) error
```
使用自定义配置注册命名数据库。

### Use
```go
func Use(dbname string) *DB
```
切换到指定名称的数据库，返回 DB 对象用于链式调用。

**示例:**
```go
db := dbkit.Use("main")
records, err := db.Query("SELECT * FROM users")
```

### Close
```go
func Close() error
func CloseDB(dbname string) error
```
关闭数据库连接。

### Ping
```go
func Ping() error
func PingDB(dbname string) error
```
测试数据库连接。

---

## 查询操作

### Query
```go
func Query(querySQL string, args ...interface{}) ([]Record, error)
func (db *DB) Query(querySQL string, args ...interface{}) ([]Record, error)
func (tx *Tx) Query(querySQL string, args ...interface{}) ([]Record, error)
```
执行查询并返回多条记录。

**示例:**
```go
records, err := dbkit.Query("SELECT * FROM users WHERE age > ?", 18)
```

### QueryFirst
```go
func QueryFirst(querySQL string, args ...interface{}) (*Record, error)
func (db *DB) QueryFirst(querySQL string, args ...interface{}) (*Record, error)
func (tx *Tx) QueryFirst(querySQL string, args ...interface{}) (*Record, error)
```
执行查询并返回第一条记录，无记录时返回 nil。

### QueryMap
```go
func QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error)
func (db *DB) QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error)
```
执行查询并返回 map 切片。

### QueryToDbModel
```go
func QueryToDbModel(dest interface{}, querySQL string, args ...interface{}) error
func (db *DB) QueryToDbModel(dest interface{}, querySQL string, args ...interface{}) error
```
执行查询并将结果映射到结构体切片。

### QueryFirstToDbModel
```go
func QueryFirstToDbModel(dest interface{}, querySQL string, args ...interface{}) error
func (db *DB) QueryFirstToDbModel(dest interface{}, querySQL string, args ...interface{}) error
```
执行查询并将第一条结果映射到结构体。

### Count
```go
func Count(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Count(table string, whereSql string, whereArgs ...interface{}) (int64, error)
```
统计符合条件的记录数。

**示例:**
```go
count, err := dbkit.Count("users", "age > ?", 18)
```

### Exists
```go
func Exists(table string, whereSql string, whereArgs ...interface{}) (bool, error)
func (db *DB) Exists(table string, whereSql string, whereArgs ...interface{}) (bool, error)
```
检查是否存在符合条件的记录。

### FindAll
```go
func FindAll(table string) ([]Record, error)
func (db *DB) FindAll(table string) ([]Record, error)
```
查询表中所有记录。

### Paginate
```go
func Paginate(page, pageSize int, querySQL string, args ...interface{}) (*Page[Record], error)
func (db *DB) Paginate(page, pageSize int, querySQL string, args ...interface{}) (*Page[Record], error)
```
分页查询（推荐使用）。使用完整SQL语句进行分页查询，自动解析SQL并根据数据库类型生成相应的分页语句。

### PaginateBuilder
```go
func PaginateBuilder(page, pageSize int, selectSql, table, whereSql, orderBySql string, args ...interface{}) (*Page[Record], error)
func (db *DB) PaginateBuilder(page, pageSize int, selectSql, table, whereSql, orderBySql string, args ...interface{}) (*Page[Record], error)
```
传统构建式分页查询。通过分别指定SELECT、表名、WHERE和ORDER BY子句进行分页查询。

**参数:**
- `page`: 页码（从1开始）
- `pageSize`: 每页记录数
- `selectSql`: SELECT 部分
- `table`: 表名
- `whereSql`: WHERE 条件
- `orderBySql`: ORDER BY 部分
- `args`: 查询参数

**返回 Page 结构体:**
```go
type Page[T any] struct {
    List       []T   // 数据列表
    PageNumber int   // 当前页码
    PageSize   int   // 每页大小
    TotalPage  int   // 总页数
    TotalRow   int64 // 总记录数
}
```

---

## 查询超时控制

DBKit 支持全局和单次查询超时设置，使用 Go 标准库的 `context.Context` 实现。

### 全局超时配置
在 Config 中设置 `QueryTimeout` 字段：
```go
config := &dbkit.Config{
    Driver:       dbkit.MySQL,
    DSN:          "root:password@tcp(localhost:3306)/test",
    MaxOpen:      10,
    QueryTimeout: 30 * time.Second,  // 所有查询默认30秒超时
}
dbkit.OpenDatabaseWithConfig(config)
```

### Timeout (全局函数)
```go
func Timeout(d time.Duration) *DB
```
返回带有指定超时时间的 DB 实例。

**示例:**
```go
users, err := dbkit.Timeout(5 * time.Second).Query("SELECT * FROM users")
```

### DB.Timeout
```go
func (db *DB) Timeout(d time.Duration) *DB
```
为 DB 实例设置查询超时时间。

**示例:**
```go
users, err := dbkit.Use("default").Timeout(5 * time.Second).Query("SELECT * FROM users")
```

### Tx.Timeout
```go
func (tx *Tx) Timeout(d time.Duration) *Tx
```
为事务设置查询超时时间。

**示例:**
```go
dbkit.Transaction(func(tx *dbkit.Tx) error {
    _, err := tx.Timeout(5 * time.Second).Query("SELECT * FROM orders")
    return err
})
```

### QueryBuilder.Timeout
```go
func (qb *QueryBuilder) Timeout(d time.Duration) *QueryBuilder
```
为链式查询设置超时时间。

**示例:**
```go
users, err := dbkit.Table("users").
    Where("age > ?", 18).
    Timeout(10 * time.Second).
    Find()
```

### 超时错误处理
超时后返回 `context.DeadlineExceeded` 错误：
```go
import "context"
import "errors"

users, err := dbkit.Timeout(1 * time.Second).Query("SELECT SLEEP(5)")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("查询超时")
    }
}
```

---

## 插入与更新

### Exec
```go
func Exec(querySQL string, args ...interface{}) (sql.Result, error)
func (db *DB) Exec(querySQL string, args ...interface{}) (sql.Result, error)
func (tx *Tx) Exec(querySQL string, args ...interface{}) (sql.Result, error)
```
执行 SQL 语句（INSERT, UPDATE, DELETE 等）。

### Save
```go
func Save(table string, record *Record) (int64, error)
func (db *DB) Save(table string, record *Record) (int64, error)
func (tx *Tx) Save(table string, record *Record) (int64, error)
```
智能保存记录。如果主键存在且记录已存在则更新，否则插入。

**返回值:** 插入时返回新ID，更新时返回影响行数。

### Insert
```go
func Insert(table string, record *Record) (int64, error)
func (db *DB) Insert(table string, record *Record) (int64, error)
func (tx *Tx) Insert(table string, record *Record) (int64, error)
```
强制插入新记录。

**返回值:** 新插入记录的ID。

### Update
```go
func Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
func (tx *Tx) Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
```
根据条件更新记录。

**返回值:** 影响的行数。

**注意:** DBKit 默认关闭了时间戳自动更新、乐观锁和软删除功能，以获得最佳性能。如需启用这些功能，请分别使用 `EnableTimestamps()`、`EnableOptimisticLock()` 和 `EnableSoftDelete()`。

### UpdateFast
```go
func UpdateFast(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) UpdateFast(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
```
轻量级更新，始终跳过时间戳和乐观锁检查，提供最佳性能。

**返回值:** 影响的行数。

**使用场景:**

1. **高频更新场景**: 需要极致性能的高并发更新操作
   ```go
   // 游戏服务器更新玩家积分
   record := dbkit.NewRecord().Set("score", newScore)
   dbkit.UpdateFast("players", record, "id = ?", playerId)
   ```

2. **批量更新**: 大量数据更新时减少开销
   ```go
   // 批量更新商品库存
   for _, item := range items {
       record := dbkit.NewRecord().Set("stock", item.Stock)
       dbkit.UpdateFast("products", record, "id = ?", item.ID)
   }
   ```

3. 表本身不需要时间戳或乐观锁功能
   
   ```go
   // 更新配置表（不需要时间戳）
   record := dbkit.NewRecord().Set("value", "new_value")
   dbkit.UpdateFast("config", record, "key = ?", "app_version")
   ```
```
   
4. **已启用时间戳、乐观锁等功能但某些操作需要跳过**: 
   
   ```go
   
   dbkit.EnableTimestamp()
   
   // 但某些高频操作需要跳过
   record := dbkit.NewRecord().Set("view_count", viewCount)
   dbkit.UpdateFast("articles", record, "id = ?", articleId)
```

**性能对比:**
- 当时间戳 、 软删除、乐观锁等功能关闭时，`Update` 和 `UpdateFast` 性能相同
- 时间戳 、 软删除、乐观锁等功能`UpdateFast` 比 `Update` 快约 2-3 倍

**注意事项:**

- `UpdateFast` 不会自动更新 `updated_at` 字段
- `UpdateFast` 不会进行乐观锁版本检查
- 如果需要这些功能，请使用 `Update` 并启用相应的特性检查

### UpdateRecord
```go
func (db *DB) UpdateRecord(table string, record *Record) (int64, error)
func (tx *Tx) UpdateRecord(table string, record *Record) (int64, error)
```
根据 Record 中的主键更新记录。

### BatchInsert
```go
func BatchInsert(table string, records []*Record, batchSize int) (int64, error)
func (db *DB) BatchInsert(table string, records []*Record, batchSize int) (int64, error)
```
批量插入记录。

**参数:**
- `batchSize`: 每批插入的记录数

### BatchInsertDefault
```go
func BatchInsertDefault(table string, records []*Record) (int64, error)
func (db *DB) BatchInsertDefault(table string, records []*Record) (int64, error)
```
批量插入记录，默认每批100条。

---

## 删除操作

### Delete
```go
func Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (tx *Tx) Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
```
根据条件删除记录。如果表配置了软删除，则执行软删除（更新删除标记字段）。

### DeleteRecord
```go
func DeleteRecord(table string, record *Record) (int64, error)
func (db *DB) DeleteRecord(table string, record *Record) (int64, error)
func (tx *Tx) DeleteRecord(table string, record *Record) (int64, error)
```
根据 Record 中的主键删除记录。

---

## 软删除

软删除允许删除记录时只标记为已删除而非物理删除，便于数据恢复和审计。

**注意**: DBKit 默认关闭软删除功能以获得最佳性能。如需使用此功能，请先启用：

```go
// 启用软删除功能
dbkit.EnableSoftDelete()
```

### EnableSoftDelete
```go
func EnableSoftDelete()
func (db *DB) EnableSoftDelete() *DB
```
启用软删除功能。启用后，查询操作会自动过滤已软删除的记录。

**示例:**
```go
// 全局启用软删除功能
dbkit.EnableSoftDelete()

// 多数据库模式
dbkit.Use("main").EnableSoftDelete()
```

### 软删除类型
```go
const (
    SoftDeleteTimestamp SoftDeleteType = iota  // 时间戳类型 (deleted_at)
    SoftDeleteBool                              // 布尔类型 (is_deleted)
)
```

### ConfigSoftDelete
```go
func ConfigSoftDelete(table, field string)
func (db *DB) ConfigSoftDelete(table, field string) *DB
```
为表配置软删除（时间戳类型）。

**参数:**
- `table`: 表名
- `field`: 软删除字段名（如 "deleted_at"）

**示例:**
```go
// 配置软删除
dbkit.ConfigSoftDelete("users", "deleted_at")

// 多数据库模式
dbkit.Use("main").ConfigSoftDelete("users", "deleted_at")
```

### ConfigSoftDeleteWithType
```go
func ConfigSoftDeleteWithType(table, field string, deleteType SoftDeleteType)
func (db *DB) ConfigSoftDeleteWithType(table, field string, deleteType SoftDeleteType) *DB
```
为表配置软删除（指定类型）。

**示例:**
```go
// 使用布尔类型
dbkit.ConfigSoftDeleteWithType("posts", "is_deleted", dbkit.SoftDeleteBool)
```

### RemoveSoftDelete
```go
func RemoveSoftDelete(table string)
func (db *DB) RemoveSoftDelete(table string) *DB
```
移除表的软删除配置。

### HasSoftDelete
```go
func HasSoftDelete(table string) bool
func (db *DB) HasSoftDelete(table string) bool
```
检查表是否启用软删除。

### WithTrashed
```go
func (qb *QueryBuilder) WithTrashed() *QueryBuilder
```
查询时包含已删除的记录。

**示例:**
```go
// 查询所有用户（包括已删除）
users, err := dbkit.Table("users").WithTrashed().Find()
```

### OnlyTrashed
```go
func (qb *QueryBuilder) OnlyTrashed() *QueryBuilder
```
只查询已删除的记录。

**示例:**
```go
// 只查询已删除的用户
deletedUsers, err := dbkit.Table("users").OnlyTrashed().Find()
```

### ForceDelete
```go
func ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (tx *Tx) ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (qb *QueryBuilder) ForceDelete() (int64, error)
```
物理删除记录，绕过软删除配置。

**示例:**
```go
// 物理删除
dbkit.ForceDelete("users", "id = ?", 1)

// 链式调用
dbkit.Table("users").Where("id = ?", 1).ForceDelete()
```

### Restore
```go
func Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (tx *Tx) Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (qb *QueryBuilder) Restore() (int64, error)
```
恢复已软删除的记录。

**示例:**
```go
// 恢复记录
dbkit.Restore("users", "id = ?", 1)

// 链式调用
dbkit.Table("users").Where("id = ?", 1).Restore()
```

### 软删除完整示例
```go
// 1. 配置软删除
dbkit.ConfigSoftDelete("users", "deleted_at")

// 2. 插入数据
record := dbkit.NewRecord()
record.Set("name", "John")
dbkit.Insert("users", record)

// 3. 软删除（自动更新 deleted_at 字段）
dbkit.Delete("users", "id = ?", 1)

// 4. 普通查询（自动过滤已删除记录）
users, _ := dbkit.Table("users").Find()  // 不包含已删除

// 5. 查询包含已删除记录
allUsers, _ := dbkit.Table("users").WithTrashed().Find()

// 6. 只查询已删除记录
deletedUsers, _ := dbkit.Table("users").OnlyTrashed().Find()

// 7. 恢复已删除记录
dbkit.Restore("users", "id = ?", 1)

// 8. 物理删除（真正删除数据）
dbkit.ForceDelete("users", "id = ?", 1)
```

### DbModel 软删除方法

生成的 DbModel 自动包含软删除相关方法：

```go
// 软删除（如果配置了软删除）
user.Delete()

// 物理删除
user.ForceDelete()

// 恢复
user.Restore()

// 查询包含已删除
users, _ := user.FindWithTrashed("status = ?", "id DESC", "active")

// 只查询已删除
deletedUsers, _ := user.FindOnlyTrashed("", "id DESC")
```

---

## 自动时间戳

自动时间戳功能允许在插入和更新记录时自动填充时间戳字段，无需手动设置。

**注意:** DBKit 默认关闭自动时间戳功能以获得最佳性能。如需启用，请使用 `EnableTimestamps()`。

### EnableTimestamps
```go
func EnableTimestamps()
func (db *DB) EnableTimestamps() *DB
```
启用自动时间戳功能。启用后，Update 操作会检查表的时间戳配置并自动更新 `updated_at` 字段。

**示例:**
```go
// 全局启用时间戳自动更新
dbkit.EnableTimestamps()

// 多数据库模式
dbkit.Use("main").EnableTimestamps()
```

### ConfigTimestamps
```go
func ConfigTimestamps(table string)
func (db *DB) ConfigTimestamps(table string) *DB
```
为表配置自动时间戳，使用默认字段名 `created_at` 和 `updated_at`。

**示例:**
```go
// 配置自动时间戳
dbkit.ConfigTimestamps("users")

// 多数据库模式
dbkit.Use("main").ConfigTimestamps("users")
```

### ConfigTimestampsWithFields
```go
func ConfigTimestampsWithFields(table, createdAtField, updatedAtField string)
func (db *DB) ConfigTimestampsWithFields(table, createdAtField, updatedAtField string) *DB
```
为表配置自动时间戳，使用自定义字段名。

**参数:**
- `table`: 表名
- `createdAtField`: 创建时间字段名（如 "create_time"）
- `updatedAtField`: 更新时间字段名（如 "update_time"）

**示例:**
```go
// 使用自定义字段名
dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")
```

### ConfigCreatedAt
```go
func ConfigCreatedAt(table, field string)
func (db *DB) ConfigCreatedAt(table, field string) *DB
```
仅配置 created_at 字段。

**示例:**
```go
// 仅配置创建时间（适用于日志表等只需记录创建时间的场景）
dbkit.ConfigCreatedAt("logs", "log_time")
```

### ConfigUpdatedAt
```go
func ConfigUpdatedAt(table, field string)
func (db *DB) ConfigUpdatedAt(table, field string) *DB
```
仅配置 updated_at 字段。

**示例:**
```go
// 仅配置更新时间
dbkit.ConfigUpdatedAt("cache_data", "last_modified")
```

### RemoveTimestamps
```go
func RemoveTimestamps(table string)
func (db *DB) RemoveTimestamps(table string) *DB
```
移除表的时间戳配置。

### HasTimestamps
```go
func HasTimestamps(table string) bool
func (db *DB) HasTimestamps(table string) bool
```
检查表是否配置了自动时间戳。

### WithoutTimestamps
```go
func (qb *QueryBuilder) WithoutTimestamps() *QueryBuilder
```
临时禁用自动时间戳（用于 QueryBuilder 的 Update 操作）。

**示例:**
```go
// 更新时不自动填充 updated_at
dbkit.Table("users").Where("id = ?", 1).WithoutTimestamps().Update(record)
```

### 自动时间戳行为说明

- **Insert 操作**: 如果 `created_at` 字段未设置，自动填充当前时间
- **Update 操作**: 总是自动填充 `updated_at` 字段为当前时间
- **手动设置优先**: 如果 Record 中已设置 `created_at`，不会被覆盖

### 自动时间戳完整示例
```go
// 1. 配置自动时间戳
dbkit.ConfigTimestamps("users")

// 2. 插入数据（created_at 自动填充）
record := dbkit.NewRecord()
record.Set("name", "John")
record.Set("email", "john@example.com")
dbkit.Insert("users", record)
// created_at 自动设置为当前时间

// 3. 更新数据（updated_at 自动填充）
updateRecord := dbkit.NewRecord()
updateRecord.Set("name", "John Updated")
dbkit.Update("users", updateRecord, "id = ?", 1)
// updated_at 自动设置为当前时间

// 4. 插入时手动指定 created_at（不会被覆盖）
customTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
record2 := dbkit.NewRecord()
record2.Set("name", "Jane")
record2.Set("created_at", customTime)
dbkit.Insert("users", record2)
// created_at 保持为 2020-01-01

// 5. 临时禁用自动时间戳
dbkit.Table("users").Where("id = ?", 1).WithoutTimestamps().Update(record)
// updated_at 不会被自动更新

// 6. 使用自定义字段名
dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")

// 7. 仅配置 created_at（适用于日志表）
dbkit.ConfigCreatedAt("logs", "log_time")
```

### 与软删除配合使用

自动时间戳与软删除功能相互独立，可以同时使用：

```go
// 同时配置软删除和自动时间戳
dbkit.ConfigTimestamps("users")
dbkit.ConfigSoftDelete("users", "deleted_at")

// 软删除时，updated_at 也会自动更新
dbkit.Delete("users", "id = ?", 1)
// deleted_at 设置为当前时间，updated_at 也更新
```

---

## 乐观锁

乐观锁是一种并发控制机制，通过版本号字段检测并发更新冲突，防止数据被意外覆盖。

**注意:** DBKit 默认关闭乐观锁功能以获得最佳性能。如需启用，请使用 `EnableOptimisticLock()`。

### EnableOptimisticLock
```go
func EnableOptimisticLock()
func (db *DB) EnableOptimisticLock() *DB
```
启用乐观锁功能。启用后，Update 操作会检查表的乐观锁配置并自动进行版本检查。

**示例:**
```go
// 全局启用乐观锁功能
dbkit.EnableOptimisticLock()

// 多数据库模式
dbkit.Use("main").EnableOptimisticLock()
```

### 工作原理

1. **Insert**: 自动将版本字段初始化为 1
2. **Update**: 自动在 WHERE 条件中添加版本检查，并在 SET 中递增版本号
3. **冲突检测**: 如果更新影响 0 行（版本不匹配），返回 `ErrVersionMismatch` 错误

### ErrVersionMismatch
```go
var ErrVersionMismatch = fmt.Errorf("dbkit: optimistic lock conflict - record was modified by another transaction")
```
版本冲突时返回的错误。

### ConfigOptimisticLock
```go
func ConfigOptimisticLock(table string)
func (db *DB) ConfigOptimisticLock(table string) *DB
```
为表配置乐观锁，使用默认字段名 `version`。

**示例:**
```go
// 配置乐观锁
dbkit.ConfigOptimisticLock("products")

// 多数据库模式
dbkit.Use("main").ConfigOptimisticLock("products")
```

### ConfigOptimisticLockWithField
```go
func ConfigOptimisticLockWithField(table, versionField string)
func (db *DB) ConfigOptimisticLockWithField(table, versionField string) *DB
```
为表配置乐观锁，使用自定义版本字段名。

**示例:**
```go
// 使用自定义字段名
dbkit.ConfigOptimisticLockWithField("orders", "revision")
```

### RemoveOptimisticLock
```go
func RemoveOptimisticLock(table string)
func (db *DB) RemoveOptimisticLock(table string) *DB
```
移除表的乐观锁配置。

### HasOptimisticLock
```go
func HasOptimisticLock(table string) bool
func (db *DB) HasOptimisticLock(table string) bool
```
检查表是否配置了乐观锁。

### 版本字段处理规则

| version 字段值 | 行为 |
|---------------|------|
| 不存在 | 跳过版本检查，正常更新 |
| `nil` / `NULL` | 跳过版本检查，正常更新 |
| `""` (空字符串) | 跳过版本检查，正常更新 |
| `0`, `1`, `2`, ... | 进行版本检查 |
| `"123"` (数字字符串) | 进行版本检查（解析为数字） |

### 乐观锁完整示例

```go
// 1. 配置乐观锁
dbkit.ConfigOptimisticLock("products")

// 2. 插入数据（version 自动初始化为 1）
record := dbkit.NewRecord()
record.Set("name", "Laptop")
record.Set("price", 999.99)
dbkit.Insert("products", record)
// version 自动设置为 1

// 3. 正常更新（带版本号）
updateRecord := dbkit.NewRecord()
updateRecord.Set("version", int64(1))  // 当前版本
updateRecord.Set("price", 899.99)
rows, err := dbkit.Update("products", updateRecord, "id = ?", 1)
// 成功：version 自动递增为 2

// 4. 并发冲突检测（使用过期版本）
staleRecord := dbkit.NewRecord()
staleRecord.Set("version", int64(1))  // 过期版本！
staleRecord.Set("price", 799.99)
rows, err = dbkit.Update("products", staleRecord, "id = ?", 1)
if errors.Is(err, dbkit.ErrVersionMismatch) {
    fmt.Println("检测到并发冲突，记录已被其他事务修改")
}

// 5. 正确处理并发：先读取最新版本
latestRecord, _ := dbkit.Table("products").Where("id = ?", 1).FindFirst()
currentVersion := latestRecord.GetInt("version")

updateRecord2 := dbkit.NewRecord()
updateRecord2.Set("version", currentVersion)
updateRecord2.Set("price", 799.99)
dbkit.Update("products", updateRecord2, "id = ?", 1)

// 6. 不带版本字段更新（跳过版本检查）
noVersionRecord := dbkit.NewRecord()
noVersionRecord.Set("stock", 90)  // 没有设置 version
dbkit.Update("products", noVersionRecord, "id = ?", 1)
// 正常更新，不检查版本

// 7. 使用 UpdateRecord（自动从记录中提取版本）
product, _ := dbkit.Table("products").Where("id = ?", 1).FindFirst()
product.Set("name", "Gaming Laptop")
dbkit.Use("default").UpdateRecord("products", product)
// version 已在 product 中，自动进行版本检查

// 8. 事务中使用乐观锁
dbkit.Transaction(func(tx *dbkit.Tx) error {
    rec, _ := tx.Table("products").Where("id = ?", 1).FindFirst()
    currentVersion := rec.GetInt("version")
    
    updateRec := dbkit.NewRecord()
    updateRec.Set("version", currentVersion)
    updateRec.Set("stock", 80)
    _, err := tx.Update("products", updateRec, "id = ?", 1)
    return err  // 版本冲突时自动回滚
})
```

### 与其他功能配合使用

乐观锁可以与自动时间戳、软删除同时使用：

```go
// 同时配置多个功能
dbkit.ConfigOptimisticLock("products")
dbkit.ConfigTimestamps("products")
dbkit.ConfigSoftDelete("products", "deleted_at")

// Insert: version=1, created_at=now
// Update: version++, updated_at=now
// Delete: deleted_at=now, updated_at=now
```

### IOptimisticLockModel 接口

```go
type IOptimisticLockModel interface {
    IDbModel
    VersionField() string  // 返回版本字段名，空字符串表示不使用
}
```

生成的 DbModel 可以实现此接口来自动配置乐观锁。

---

## 事务处理

### Transaction
```go
func Transaction(fn func(*Tx) error) error
func (db *DB) Transaction(fn func(*Tx) error) error
```
自动事务处理。闭包返回 error 时自动回滚，否则自动提交。

**示例:**
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    _, err := tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE id = ?", 1)
    if err != nil {
        return err // 自动回滚
    }
    _, err = tx.Exec("UPDATE accounts SET balance = balance + 100 WHERE id = ?", 2)
    return err
})
```

### BeginTransaction
```go
func BeginTransaction() (*Tx, error)
```
开始手动事务。

### Tx.Commit
```go
func (tx *Tx) Commit() error
```
提交事务。

### Tx.Rollback
```go
func (tx *Tx) Rollback() error
```
回滚事务。

---

## Record 对象

### NewRecord
```go
func NewRecord() *Record
```
创建新的空 Record 对象。

### Record.Set
```go
func (r *Record) Set(column string, value interface{}) *Record
```
设置字段值，支持链式调用。

### Record.Get
```go
func (r *Record) Get(column string) interface{}
```
获取字段值。

### 类型安全获取方法
```go
func (r *Record) GetString(column string) string
func (r *Record) GetInt(column string) int
func (r *Record) GetInt64(column string) int64
func (r *Record) GetFloat(column string) float64
func (r *Record) GetBool(column string) bool
func (r *Record) GetTime(column string) time.Time

// 简写方法
func (r *Record) Str(column string) string
func (r *Record) Int(column string) int
func (r *Record) Int64(column string) int64
func (r *Record) Float(column string) float64
func (r *Record) Bool(column string) bool
```

### Record.Has
```go
func (r *Record) Has(column string) bool
```
检查字段是否存在。

### Record.Keys
```go
func (r *Record) Keys() []string
```
获取所有字段名。

### Record.Remove
```go
func (r *Record) Remove(column string)
```
删除字段。

### Record.Clear
```go
func (r *Record) Clear()
```
清空所有字段。

### Record.ToMap
```go
func (r *Record) ToMap() map[string]interface{}
```
转换为 map。

### Record.ToJson
```go
func (r *Record) ToJson() string
```
转换为 JSON 字符串。

### Record.FromJson
```go
func (r *Record) FromJson(jsonStr string) error
```
从 JSON 字符串解析。

### Record.ToStruct
```go
func (r *Record) ToStruct(dest interface{}) error
```
转换为结构体。

### Record.FromStruct
```go
func (r *Record) FromStruct(src interface{}) error
```
从结构体填充。

---

## 链式查询

### Table
```go
func Table(name string) *QueryBuilder
func (db *DB) Table(name string) *QueryBuilder
func (tx *Tx) Table(name string) *QueryBuilder
```
开始链式查询，指定表名。

### QueryBuilder 方法

```go
func (b *QueryBuilder) Select(columns string) *QueryBuilder    // 指定查询字段
func (b *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder  // WHERE 条件
func (b *QueryBuilder) And(condition string, args ...interface{}) *QueryBuilder    // AND 条件
func (b *QueryBuilder) OrderBy(orderBy string) *QueryBuilder   // 排序
func (b *QueryBuilder) Limit(limit int) *QueryBuilder          // 限制数量
func (b *QueryBuilder) Offset(offset int) *QueryBuilder        // 偏移量

// 执行方法
func (b *QueryBuilder) Find() ([]Record, error)                // 查询多条
func (b *QueryBuilder) Query() ([]Record, error)               // Find 的别名
func (b *QueryBuilder) FindFirst() (*Record, error)            // 查询第一条
func (b *QueryBuilder) QueryFirst() (*Record, error)           // FindFirst 的别名
func (b *QueryBuilder) FindToDbModel(dest interface{}) error   // 查询并映射到结构体切片
func (b *QueryBuilder) FindFirstToDbModel(dest interface{}) error // 查询第一条并映射到结构体
func (b *QueryBuilder) Delete() (int64, error)                 // 删除
func (b *QueryBuilder) Paginate(page, pageSize int) (*Page[Record], error) // 分页
```

**示例:**
```go
users, err := dbkit.Table("users").
    Select("id, name, age").
    Where("age > ?", 18).
    Where("status = ?", "active").
    OrderBy("created_at DESC").
    Limit(10).
    Find()
// SQL: SELECT id, name, age FROM users WHERE age > ? AND status = ? ORDER BY created_at DESC LIMIT 10
// Args: [18, "active"]
```

### Join 查询

支持多种 JOIN 类型的链式调用：

```go
func (b *QueryBuilder) Join(table, condition string, args ...interface{}) *QueryBuilder      // JOIN
func (b *QueryBuilder) LeftJoin(table, condition string, args ...interface{}) *QueryBuilder  // LEFT JOIN
func (b *QueryBuilder) RightJoin(table, condition string, args ...interface{}) *QueryBuilder // RIGHT JOIN
func (b *QueryBuilder) InnerJoin(table, condition string, args ...interface{}) *QueryBuilder // INNER JOIN
```

**示例:**
```go
// 简单 LEFT JOIN
records, err := dbkit.Table("users").
    Select("users.name, orders.total").
    LeftJoin("orders", "users.id = orders.user_id").
    Where("orders.status = ?", "completed").
    Find()
// SQL: SELECT users.name, orders.total FROM users LEFT JOIN orders ON users.id = orders.user_id WHERE orders.status = ?
// Args: ["completed"]

// 多表 INNER JOIN
records, err := dbkit.Table("orders").
    Select("orders.id, users.name, products.name as product_name").
    InnerJoin("users", "orders.user_id = users.id").
    InnerJoin("order_items", "orders.id = order_items.order_id").
    InnerJoin("products", "order_items.product_id = products.id").
    Where("orders.status = ?", "completed").
    OrderBy("orders.created_at DESC").
    Find()
// SQL: SELECT orders.id, users.name, products.name as product_name FROM orders 
//      INNER JOIN users ON orders.user_id = users.id 
//      INNER JOIN order_items ON orders.id = order_items.order_id 
//      INNER JOIN products ON order_items.product_id = products.id 
//      WHERE orders.status = ? ORDER BY orders.created_at DESC
// Args: ["completed"]

// 带参数的 JOIN 条件
records, err := dbkit.Table("users").
    Join("orders", "users.id = orders.user_id AND orders.status = ?", "active").
    Find()
// SQL: SELECT * FROM users JOIN orders ON users.id = orders.user_id AND orders.status = ?
// Args: ["active"]
```

### 子查询 (Subquery)

#### NewSubquery
```go
func NewSubquery() *Subquery
```
创建新的子查询构建器。

#### Subquery 方法
```go
func (s *Subquery) Table(name string) *Subquery                           // 设置表名
func (s *Subquery) Select(columns string) *Subquery                       // 设置查询字段
func (s *Subquery) Where(condition string, args ...interface{}) *Subquery // 添加条件
func (s *Subquery) OrderBy(orderBy string) *Subquery                      // 排序
func (s *Subquery) Limit(limit int) *Subquery                             // 限制数量
func (s *Subquery) ToSQL() (string, []interface{})                        // 生成 SQL
```

#### WHERE IN 子查询
```go
func (b *QueryBuilder) WhereIn(column string, sub *Subquery) *QueryBuilder    // WHERE column IN (subquery)
func (b *QueryBuilder) WhereNotIn(column string, sub *Subquery) *QueryBuilder // WHERE column NOT IN (subquery)
```

**示例:**
```go
// 查询有已完成订单的用户
activeUsersSub := dbkit.NewSubquery().
    Table("orders").
    Select("DISTINCT user_id").
    Where("status = ?", "completed")

users, err := dbkit.Table("users").
    Select("*").
    WhereIn("id", activeUsersSub).
    Find()
// SQL: SELECT * FROM users WHERE id IN (SELECT DISTINCT user_id FROM orders WHERE status = ?)
// Args: ["completed"]

// 查询没有被禁用的用户的订单
bannedUsersSub := dbkit.NewSubquery().
    Table("users").
    Select("id").
    Where("status = ?", "banned")

orders, err := dbkit.Table("orders").
    WhereNotIn("user_id", bannedUsersSub).
    Find()
// SQL: SELECT * FROM orders WHERE user_id NOT IN (SELECT id FROM users WHERE status = ?)
// Args: ["banned"]
```

#### FROM 子查询
```go
func (b *QueryBuilder) TableSubquery(sub *Subquery, alias string) *QueryBuilder
```
使用子查询作为 FROM 数据源（派生表）。

**示例:**
```go
// 从聚合子查询中查询
userTotalsSub := dbkit.NewSubquery().
    Table("orders").
    Select("user_id, SUM(total) as total_spent")

records, err := (&dbkit.QueryBuilder{}).
    TableSubquery(userTotalsSub, "user_totals").
    Select("user_id, total_spent").
    Where("total_spent > ?", 1000).
    Find()
// SQL: SELECT user_id, total_spent FROM (SELECT user_id, SUM(total) as total_spent FROM orders) AS user_totals WHERE total_spent > ?
// Args: [1000]
```

#### SELECT 子查询
```go
func (b *QueryBuilder) SelectSubquery(sub *Subquery, alias string) *QueryBuilder
```
在 SELECT 子句中添加子查询作为字段。

**示例:**
```go
// 为每个用户添加订单数量字段
orderCountSub := dbkit.NewSubquery().
    Table("orders").
    Select("COUNT(*)").
    Where("orders.user_id = users.id")

users, err := dbkit.Table("users").
    Select("users.id, users.name").
    SelectSubquery(orderCountSub, "order_count").
    Find()
// SQL: SELECT users.id, users.name, (SELECT COUNT(*) FROM orders WHERE orders.user_id = users.id) AS order_count FROM users
// Args: []
```

### 高级 WHERE 条件

#### OrWhere
```go
func (b *QueryBuilder) OrWhere(condition string, args ...interface{}) *QueryBuilder
```
添加 OR 条件到查询。当与 Where 组合使用时，AND 条件会被括号包裹以保持正确的优先级。

**示例:**
```go
// 查询状态为 active 或 priority 为 high 的订单
orders, err := dbkit.Table("orders").
    Where("status = ?", "active").
    OrWhere("priority = ?", "high").
    Find()
// SQL: SELECT * FROM orders WHERE (status = ?) OR priority = ?
// Args: ["active", "high"]

// 多个 OR 条件
orders, err := dbkit.Table("orders").
    OrWhere("status = ?", "pending").
    OrWhere("status = ?", "processing").
    OrWhere("status = ?", "shipped").
    Find()
// SQL: SELECT * FROM orders WHERE status = ? OR status = ? OR status = ?
// Args: ["pending", "processing", "shipped"]
```

#### WhereGroup / OrWhereGroup
```go
type WhereGroupFunc func(qb *QueryBuilder) *QueryBuilder

func (b *QueryBuilder) WhereGroup(fn WhereGroupFunc) *QueryBuilder
func (b *QueryBuilder) OrWhereGroup(fn WhereGroupFunc) *QueryBuilder
```
添加分组条件，支持嵌套括号。`WhereGroup` 使用 AND 连接，`OrWhereGroup` 使用 OR 连接。

**示例:**
```go
// OR 分组条件
records, err := dbkit.Table("table").
    Where("a = ?", 1).
    OrWhereGroup(func(qb *dbkit.QueryBuilder) *dbkit.QueryBuilder {
        return qb.Where("b = ?", 1).OrWhere("c = ?", 1)
    }).
    Find()
// SQL: SELECT * FROM table WHERE (a = ?) OR (b = ? OR c = ?)
// Args: [1, 1, 1]

// AND 分组条件
records, err := dbkit.Table("orders").
    Where("status = ?", "active").
    WhereGroup(func(qb *dbkit.QueryBuilder) *dbkit.QueryBuilder {
        return qb.Where("type = ?", "A").OrWhere("priority = ?", "high")
    }).
    Find()
// SQL: SELECT * FROM orders WHERE status = ? AND (type = ? OR priority = ?)
// Args: ["active", "A", "high"]

// 复杂嵌套
records, err := dbkit.Table("table").
    Where("a = ?", 1).
    WhereGroup(func(outer *dbkit.QueryBuilder) *dbkit.QueryBuilder {
        return outer.Where("b = ?", 2).
            OrWhereGroup(func(inner *dbkit.QueryBuilder) *dbkit.QueryBuilder {
                return inner.Where("c = ?", 3).Where("d = ?", 4)
            })
    }).
    Find()
// SQL: SELECT * FROM table WHERE a = ? AND (b = ? OR (c = ? AND d = ?))
// Args: [1, 2, 3, 4]
```

#### WhereInValues / WhereNotInValues
```go
func (b *QueryBuilder) WhereInValues(column string, values []interface{}) *QueryBuilder
func (b *QueryBuilder) WhereNotInValues(column string, values []interface{}) *QueryBuilder
```
使用值列表进行 IN/NOT IN 查询（与子查询版本 WhereIn/WhereNotIn 区分）。

**示例:**
```go
// 查询指定 ID 的用户
users, err := dbkit.Table("users").
    WhereInValues("id", []interface{}{1, 2, 3, 4, 5}).
    Find()
// SQL: SELECT * FROM users WHERE id IN (?, ?, ?, ?, ?)
// Args: [1, 2, 3, 4, 5]

// 排除指定状态的订单
orders, err := dbkit.Table("orders").
    WhereNotInValues("status", []interface{}{"cancelled", "refunded"}).
    Find()
// SQL: SELECT * FROM orders WHERE status NOT IN (?, ?)
// Args: ["cancelled", "refunded"]
```

#### WhereBetween / WhereNotBetween
```go
func (b *QueryBuilder) WhereBetween(column string, min, max interface{}) *QueryBuilder
func (b *QueryBuilder) WhereNotBetween(column string, min, max interface{}) *QueryBuilder
```
范围查询。

**示例:**
```go
// 查询年龄在 18-65 之间的用户
users, err := dbkit.Table("users").
    WhereBetween("age", 18, 65).
    Find()
// SQL: SELECT * FROM users WHERE age BETWEEN ? AND ?
// Args: [18, 65]

// 查询价格不在 100-500 之间的产品
products, err := dbkit.Table("products").
    WhereNotBetween("price", 100, 500).
    Find()
// SQL: SELECT * FROM products WHERE price NOT BETWEEN ? AND ?
// Args: [100, 500]

// 日期范围查询
orders, err := dbkit.Table("orders").
    WhereBetween("created_at", "2024-01-01", "2024-12-31").
    Find()
// SQL: SELECT * FROM orders WHERE created_at BETWEEN ? AND ?
// Args: ["2024-01-01", "2024-12-31"]
```

#### WhereNull / WhereNotNull
```go
func (b *QueryBuilder) WhereNull(column string) *QueryBuilder
func (b *QueryBuilder) WhereNotNull(column string) *QueryBuilder
```
NULL 值检查。

**示例:**
```go
// 查询没有邮箱的用户
users, err := dbkit.Table("users").
    WhereNull("email").
    Find()
// SQL: SELECT * FROM users WHERE email IS NULL
// Args: []

// 查询有手机号的用户
users, err := dbkit.Table("users").
    WhereNotNull("phone").
    Find()
// SQL: SELECT * FROM users WHERE phone IS NOT NULL
// Args: []
```

### 分组和聚合

#### GroupBy
```go
func (b *QueryBuilder) GroupBy(columns string) *QueryBuilder
```
添加 GROUP BY 子句。

#### Having
```go
func (b *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder
```
添加 HAVING 子句，用于过滤分组结果。

**示例:**
```go
// 按状态分组统计订单
stats, err := dbkit.Table("orders").
    Select("status, COUNT(*) as count, SUM(total) as total_amount").
    GroupBy("status").
    Find()
// SQL: SELECT status, COUNT(*) as count, SUM(total) as total_amount FROM orders GROUP BY status
// Args: []

// 查询订单数大于 5 的用户
users, err := dbkit.Table("orders").
    Select("user_id, COUNT(*) as order_count").
    GroupBy("user_id").
    Having("COUNT(*) > ?", 5).
    Find()
// SQL: SELECT user_id, COUNT(*) as order_count FROM orders GROUP BY user_id HAVING COUNT(*) > ?
// Args: [5]

// 多个 HAVING 条件
stats, err := dbkit.Table("orders").
    Select("user_id, COUNT(*) as cnt, SUM(total) as total").
    GroupBy("user_id").
    Having("COUNT(*) > ?", 3).
    Having("SUM(total) > ?", 1000).
    Find()
// SQL: SELECT user_id, COUNT(*) as cnt, SUM(total) as total FROM orders GROUP BY user_id HAVING COUNT(*) > ? AND SUM(total) > ?
// Args: [3, 1000]
```

### 复杂查询示例

```go
// 组合多种条件的复杂查询
results, err := dbkit.Table("orders").
    Select("status, COUNT(*) as cnt, SUM(total) as total_amount").
    Where("created_at > ?", "2024-01-01").
    Where("active = ?", 1).
    OrWhere("priority = ?", "high").
    WhereInValues("type", []interface{}{"A", "B", "C"}).
    WhereNotNull("customer_id").
    GroupBy("status").
    Having("COUNT(*) > ?", 10).
    OrderBy("total_amount DESC").
    Limit(20).
    Find()
// SQL: SELECT status, COUNT(*) as cnt, SUM(total) as total_amount FROM orders 
//      WHERE (created_at > ? AND active = ? AND type IN (?, ?, ?) AND customer_id IS NOT NULL) OR priority = ? 
//      GROUP BY status HAVING COUNT(*) > ? ORDER BY total_amount DESC LIMIT 20
// Args: ["2024-01-01", 1, "A", "B", "C", "high", 10]
```

---

## DbModel 操作

### GenerateDbModel
```go
func GenerateDbModel(tablename, outPath, structName string) error
func (db *DB) GenerateDbModel(tablename, outPath, structName string) error
```
根据数据表生成 Go 结构体代码。

**参数:**
- `tablename`: 表名
- `outPath`: 输出路径（目录或完整文件路径）
- `structName`: 结构体名称（空则自动生成）

### IDbModel 接口
```go
type IDbModel interface {
    TableName() string
    DatabaseName() string
}
```

### DbModel CRUD 函数
```go
func SaveDbModel(model IDbModel) (int64, error)
func InsertDbModel(model IDbModel) (int64, error)
func UpdateDbModel(model IDbModel) (int64, error)
func DeleteDbModel(model IDbModel) (int64, error)
func FindFirstToDbModel(model IDbModel, whereSql string, whereArgs ...interface{}) error
func FindToDbModel(dest interface{}, table, whereSql, orderBySql string, whereArgs ...interface{}) error
```

### 泛型辅助函数
```go
func FindModel[T IDbModel](model T, cache *ModelCache, whereSql, orderBySql string, whereArgs ...interface{}) ([]T, error)
func FindFirstModel[T IDbModel](model T, cache *ModelCache, whereSql string, whereArgs ...interface{}) (T, error)
func PaginateModel[T IDbModel](model T, cache *ModelCache, page, pageSize int, whereSql, orderBySql string, whereArgs ...interface{}) (*Page[T], error)
```

---

## 缓存操作

### SetCache
```go
func SetCache(c CacheProvider)
```
设置全局缓存提供者。

### GetCache
```go
func GetCache() CacheProvider
```
获取当前缓存提供者。

### SetLocalCacheConfig
```go
func SetLocalCacheConfig(cleanupInterval time.Duration)
```
配置本地缓存清理间隔。

### CreateCache
```go
func CreateCache(cacheName string, ttl time.Duration)
```
创建命名缓存并设置默认 TTL。

### CacheSet
```go
func CacheSet(cacheName, key string, value interface{}, ttl ...time.Duration)
```
设置缓存值。

### CacheGet
```go
func CacheGet(cacheName, key string) (interface{}, bool)
```
获取缓存值。

### CacheDelete
```go
func CacheDelete(cacheName, key string)
```
删除缓存键。

### CacheClear
```go
func CacheClear(cacheName string)
```
清空指定缓存。

### CacheStatus
```go
func CacheStatus() map[string]interface{}
```
获取缓存状态信息。

### Cache (链式调用)
```go
func Cache(name string, ttl ...time.Duration) *DB
func (db *DB) Cache(name string, ttl ...time.Duration) *DB
func (tx *Tx) Cache(name string, ttl ...time.Duration) *Tx
```
为查询启用缓存。

**示例:**
```go
records, err := dbkit.Cache("user_cache", 5*time.Minute).Query("SELECT * FROM users")
```

### CacheProvider 接口
```go
type CacheProvider interface {
    CacheGet(cacheName, key string) (interface{}, bool)
    CacheSet(cacheName, key string, value interface{}, ttl time.Duration)
    CacheDelete(cacheName, key string)
    CacheClear(cacheName string)
    Status() map[string]interface{}
}
```

---

## 日志配置

### SetDebugMode
```go
func SetDebugMode(enabled bool)
```
开启/关闭调试模式（输出 SQL 语句）。

### SetLogger
```go
func SetLogger(l Logger)
```
设置自定义日志记录器。

### InitLoggerWithFile
```go
func InitLoggerWithFile(level string, filePath string)
```
初始化文件日志。

### Logger 接口
```go
type Logger interface {
    Log(level LogLevel, msg string, fields map[string]interface{})
}
```

### 日志级别
```go
const (
    LevelDebug LogLevel = "debug"
    LevelInfo  LogLevel = "info"
    LevelWarn  LogLevel = "warn"
    LevelError LogLevel = "error"
)
```

### 日志函数
```go
func LogDebug(msg string, fields map[string]interface{})
func LogInfo(msg string, fields map[string]interface{})
func LogWarn(msg string, fields map[string]interface{})
func LogError(msg string, fields map[string]interface{})
```

---

## SQL 模板

DBKit 提供了强大的 SQL 模板功能，允许您将 SQL 语句配置化管理，支持动态参数、条件构建和多数据库执行。

### 配置文件结构

SQL 模板使用 JSON 格式的配置文件。以下是一个完整的配置文件格式模板：

#### 完整 JSON 格式模板

```json
{
  "version": "1.0",
  "description": "服务SQL配置文件描述",
  "namespace": "service_name",
  "sqls": [
    {
      "name": "sqlName",
      "description": "SQL语句描述",
      "sql": "SELECT * FROM table WHERE condition = :param",
      "type": "select",
      "order": "created_at DESC",
      "inparam": [
        {
          "name": "paramName",
          "type": "string",
          "desc": "参数描述",
          "sql": " AND column = :paramName"
        }
      ]
    }
  ]
}
```

#### 字段说明

**根级别字段：**
- `version` (string, 必需): 配置文件版本号
- `description` (string, 可选): 配置文件描述
- `namespace` (string, 可选): 命名空间，用于避免 SQL 名称冲突
- `sqls` (array, 必需): SQL 语句配置数组

**SQL 配置字段：**
- `name` (string, 必需): SQL 语句唯一标识符
- `description` (string, 可选): SQL 语句描述
- `sql` (string, 必需): SQL 语句模板
- `type` (string, 可选): SQL 类型 (`select`, `insert`, `update`, `delete`)
- `order` (string, 可选): 默认排序条件
- `inparam` (array, 可选): 输入参数定义（用于动态 SQL）

**输入参数字段 (inparam)：**
- `name` (string, 必需): 参数名称
- `type` (string, 必需): 参数类型
- `desc` (string, 可选): 参数描述
- `sql` (string, 必需): 当参数存在时追加的 SQL 片段

#### 实际配置示例

```json
{
  "version": "1.0",
  "description": "用户服务SQL配置",
  "namespace": "user_service",
  "sqls": [
    {
      "name": "findById",
      "description": "根据ID查找用户",
      "sql": "SELECT * FROM users WHERE id = :id",
      "type": "select"
    },
    {
      "name": "findUsers",
      "description": "动态查询用户列表",
      "sql": "SELECT * FROM users WHERE 1=1",
      "type": "select",
      "order": "created_at DESC",
      "inparam": [
        {
          "name": "status",
          "type": "int",
          "desc": "用户状态",
          "sql": " AND status = :status"
        },
        {
          "name": "name",
          "type": "string",
          "desc": "用户名模糊查询",
          "sql": " AND name LIKE CONCAT('%', :name, '%')"
        }
      ]
    }
  ]
}
```

### 参数类型支持

DBKit SQL 模板支持多种参数传递方式，提供灵活的使用体验：

#### 支持的参数类型

| 参数类型 | 适用场景 | SQL 占位符 | 示例 |
|---------|---------|-----------|------|
| `map[string]interface{}` | 命名参数 | `:name` | `map[string]interface{}{"id": 123}` |
| `[]interface{}` | 多个位置参数 | `?` | `[]interface{}{123, "John"}` |
| **单个简单类型** | 单个位置参数 | `?` | `123`, `"John"`, `true` |
| **可变参数** | 多个位置参数 | `?` | `SqlTemplate(name, 123, "John", true)` |

#### 单个简单类型支持

🆕 支持直接传递简单类型参数，无需包装成 map 或 slice：

- `string` - 字符串
- `int`, `int8`, `int16`, `int32`, `int64` - 整数类型
- `uint`, `uint8`, `uint16`, `uint32`, `uint64` - 无符号整数
- `float32`, `float64` - 浮点数
- `bool` - 布尔值

#### 可变参数支持

🆕 **新特性**：支持 Go 风格的可变参数 (`...interface{}`)，提供最自然的参数传递方式：

```go
// 可变参数方式 - 最直观和简洁
records, err := dbkit.SqlTemplate("findByIdAndStatus", 123, 1).Query()
records, err := dbkit.SqlTemplate("updateUser", "John", "john@example.com", 25, 123).Exec()
records, err := dbkit.SqlTemplate("findByAgeRange", 18, 65, 1).Query()
```

#### 参数匹配规则

| SQL 占位符 | 参数类型 | 结果 |
|-----------|---------|------|
| 单个 `?` | 单个简单类型 | ✅ 支持 |
| 单个 `?` | `map[string]interface{}` | ✅ 支持（向后兼容） |
| 单个 `?` | `[]interface{}{value}` | ✅ 支持（向后兼容） |
| 多个 `?` | `[]interface{}{v1, v2, ...}` | ✅ 支持 |
| 多个 `?` | **可变参数 `v1, v2, ...`** | ✅ 支持 🆕 |
| 多个 `?` | 单个简单类型 | ❌ 错误提示 |
| `:name` | `map[string]interface{}{"name": value}` | ✅ 支持 |
| `:name` | 单个简单类型 | ❌ 错误提示 |
| `:name` | 可变参数 | ❌ 错误提示 |

#### 参数数量验证

系统会自动验证参数数量与 SQL 占位符数量是否匹配：

```go
// SQL: "SELECT * FROM users WHERE id = ? AND status = ?"
// 正确：2个参数匹配2个占位符
records, err := dbkit.SqlTemplate("findByIdAndStatus", 123, 1).Query()

// 错误：参数不足
records, err := dbkit.SqlTemplate("findByIdAndStatus", 123).Query()
// 返回错误: parameter count mismatch: SQL has 2 '?' placeholders but got 1 parameters

// 错误：参数过多  
records, err := dbkit.SqlTemplate("findByIdAndStatus", 123, 1, 2).Query()
// 返回错误: parameter count mismatch: SQL has 2 '?' placeholders but got 3 parameters
```

#### 使用示例

```go
// 1. 单个简单参数（推荐用于单参数查询）
records, err := dbkit.SqlTemplate("user_service.findById", 123).Query()
records, err := dbkit.SqlTemplate("user_service.findByEmail", "user@example.com").Query()
records, err := dbkit.SqlTemplate("user_service.findActive", true).Query()

// 2. 可变参数（推荐用于多参数查询）
records, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 123, 1).Query()
records, err := dbkit.SqlTemplate("user_service.updateUser", "John", "john@example.com", 25, 123).Exec()
records, err := dbkit.SqlTemplate("user_service.findByAgeRange", 18, 65, 1).Query()

// 3. 命名参数（适用于复杂查询）
params := map[string]interface{}{
    "status": 1,
    "name": "John",
    "ageMin": 18,
}
records, err := dbkit.SqlTemplate("user_service.findUsers", params).Query()

// 4. 位置参数（向后兼容）
records, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 
    []interface{}{123, 1}).Query()
```

### 配置加载

#### LoadSqlConfig
```go
func LoadSqlConfig(configPath string) error
```
加载单个 SQL 配置文件。

**示例:**
```go
err := dbkit.LoadSqlConfig("config/user_service.json")
```

#### LoadSqlConfigs
```go
func LoadSqlConfigs(configPaths []string) error
```
批量加载多个 SQL 配置文件。

**示例:**
```go
configPaths := []string{
    "config/user_service.json",
    "config/order_service.json",
}
err := dbkit.LoadSqlConfigs(configPaths)
```

#### LoadSqlConfigDir
```go
func LoadSqlConfigDir(dirPath string) error
```
加载指定目录下的所有 JSON 配置文件。

**示例:**
```go
err := dbkit.LoadSqlConfigDir("config/")
```

#### ReloadSqlConfig
```go
func ReloadSqlConfig(configPath string) error
```
重新加载指定的配置文件。

#### ReloadAllSqlConfigs
```go
func ReloadAllSqlConfigs() error
```
重新加载所有已加载的配置文件。

### 配置信息查询

#### GetSqlConfigInfo
```go
func GetSqlConfigInfo() []ConfigInfo
```
获取所有已加载配置文件的信息。

**ConfigInfo 结构体:**
```go
type ConfigInfo struct {
    FilePath    string `json:"filePath"`
    Namespace   string `json:"namespace"`
    Description string `json:"description"`
    SqlCount    int    `json:"sqlCount"`
}
```

#### ListSqlItems
```go
func ListSqlItems() map[string]*SqlItem
```
列出所有可用的 SQL 模板项。

### SQL 模板执行

#### SqlTemplate (全局)
```go
func SqlTemplate(name string, params ...interface{}) *SqlTemplateBuilder
```
创建 SQL 模板构建器，使用默认数据库连接。

**参数:**
- `name`: SQL 模板名称（支持命名空间，如 "user_service.findById"）
- `params`: 可变参数，支持以下类型：
  - `map[string]interface{}` - 命名参数（`:name`）
  - `[]interface{}` - 位置参数数组（`?`）
  - **单个简单类型** - 单个位置参数（`?`），支持 `string`、`int`、`float`、`bool` 等基本类型
  - **🆕 可变参数** - 多个位置参数（`?`），直接传递多个值

**示例:**
```go
// 使用命名参数
records, err := dbkit.SqlTemplate("user_service.findById", 
    map[string]interface{}{"id": 123}).Query()

// 使用位置参数数组
records, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 
    []interface{}{123, 1}).Query()

// 🆕 使用单个简单参数（推荐用于单参数查询）
records, err := dbkit.SqlTemplate("user_service.findById", 123).Query()
records, err := dbkit.SqlTemplate("user_service.findByEmail", "user@example.com").Query()

// 🆕 使用可变参数（推荐用于多参数查询）
records, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 123, 1).Query()
records, err := dbkit.SqlTemplate("user_service.updateUser", "John", "john@example.com", 25, 123).Exec()
records, err := dbkit.SqlTemplate("user_service.findByAgeRange", 18, 65, 1).Query()
```

#### SqlTemplate (指定数据库)
```go
func (db *DB) SqlTemplate(name string, params ...interface{}) *SqlTemplateBuilder
```
在指定数据库上创建 SQL 模板构建器。

**示例:**
```go
// 传统方式
records, err := dbkit.Use("mysql").SqlTemplate("user_service.findById", 
    map[string]interface{}{"id": 123}).Query()

// 🆕 单个简单参数（更简洁）
records, err := dbkit.Use("mysql").SqlTemplate("user_service.findById", 123).Query()

// 🆕 可变参数（最简洁）
records, err := dbkit.Use("mysql").SqlTemplate("user_service.findByIdAndStatus", 123, 1).Query()
```

#### SqlTemplate (事务)
```go
func (tx *Tx) SqlTemplate(name string, params ...interface{}) *SqlTemplateBuilder
```
在事务中使用 SQL 模板。

**示例:**
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 使用可变参数
    result, err := tx.SqlTemplate("user_service.insertUser", "John", "john@example.com", 25).Exec()
    return err
})
```

### SqlTemplateBuilder 方法

#### Timeout
```go
func (b *SqlTemplateBuilder) Timeout(timeout time.Duration) *SqlTemplateBuilder
```
设置查询超时时间。

**示例:**
```go
records, err := dbkit.SqlTemplate("user_service.findUsers", params).
    Timeout(30 * time.Second).Query()
```

#### Query
```go
func (b *SqlTemplateBuilder) Query() ([]Record, error)
```
执行查询并返回多条记录。

#### QueryFirst
```go
func (b *SqlTemplateBuilder) QueryFirst() (*Record, error)
```
执行查询并返回第一条记录。

#### Exec
```go
func (b *SqlTemplateBuilder) Exec() (sql.Result, error)
```
执行 SQL 语句（INSERT、UPDATE、DELETE）。

### 动态 SQL 构建

通过 `inparam` 配置可以实现动态 SQL 条件构建：

```json
{
  "name": "searchUsers",
  "sql": "SELECT * FROM users WHERE 1=1",
  "inparam": [
    {
      "name": "status",
      "type": "int",
      "desc": "用户状态",
      "sql": " AND status = :status"
    },
    {
      "name": "ageMin",
      "type": "int", 
      "desc": "最小年龄",
      "sql": " AND age >= :ageMin"
    }
  ],
  "order": "created_at DESC"
}
```

**使用示例:**
```go
// 只传入部分参数，系统会自动构建相应的 SQL
params := map[string]interface{}{
    "status": 1,
    // ageMin 未提供，对应的条件不会被添加
}
records, err := dbkit.SqlTemplate("searchUsers", params).Query()
// 生成的 SQL: SELECT * FROM users WHERE 1=1 AND status = ? ORDER BY created_at DESC
```

### 参数处理

#### 命名参数
使用 `:paramName` 格式的命名参数：

```go
params := map[string]interface{}{
    "id": 123,
    "name": "张三",
}
records, err := dbkit.SqlTemplate("user_service.updateUser", params).Exec()
```

#### 位置参数
使用 `?` 占位符的位置参数：

```go
params := []interface{}{123}
records, err := dbkit.SqlTemplate("user_service.findById", params).Query()
```

### 错误处理

SQL 模板系统提供详细的错误信息：

```go
type SqlConfigError struct {
    Type    string // 错误类型：NotFoundError, ParameterError, ParseError 等
    Message string // 错误描述
    SqlName string // 相关的 SQL 名称
    Cause   error  // 原始错误
}
```

**常见错误类型:**
- `NotFoundError`: SQL 模板不存在
- `ParameterError`: 参数错误（缺失、类型不匹配等）
- `ParameterTypeMismatch`: 参数类型与 SQL 格式不匹配
- `ParseError`: 配置文件解析错误
- `DuplicateError`: 重复的 SQL 标识符

### 最佳实践

1. **命名规范**: 使用命名空间避免 SQL 名称冲突
2. **参数验证**: 系统会自动验证必需参数
3. **动态条件**: 使用 `inparam` 实现灵活的条件构建
4. **错误处理**: 捕获并处理 `SqlConfigError` 类型的错误
5. **性能优化**: 配置文件在首次加载后会被缓存

**完整示例:**
```go
// 1. 加载配置
err := dbkit.LoadSqlConfigDir("config/")
if err != nil {
    log.Fatal(err)
}

// 2. 执行查询
params := map[string]interface{}{
    "status": 1,
    "name": "张",
}

records, err := dbkit.Use("mysql").
    SqlTemplate("user_service.findUsers", params).
    Timeout(30 * time.Second).
    Query()

if err != nil {
    if sqlErr, ok := err.(*dbkit.SqlConfigError); ok {
        log.Printf("SQL 配置错误 [%s]: %s", sqlErr.Type, sqlErr.Message)
    } else {
        log.Printf("执行错误: %v", err)
    }
    return
}

// 3. 处理结果
for _, record := range records {
    fmt.Printf("用户: %s, 状态: %d\n", 
        record.GetString("name"), 
        record.GetInt("status"))
}
```

---

## 工具函数

### ToJson
```go
func ToJson(v interface{}) string
```
将任意值转换为 JSON 字符串。

### ToStruct
```go
func ToStruct(record *Record, dest interface{}) error
```
将 Record 转换为结构体。

### ToStructs
```go
func ToStructs(records []Record, dest interface{}) error
```
将 Record 切片转换为结构体切片。

### ToRecord
```go
func ToRecord(model interface{}) *Record
```
将结构体转换为 Record。

### FromStruct
```go
func FromStruct(src interface{}, record *Record) error
```
从结构体填充 Record。

### SnakeToCamel
```go
func SnakeToCamel(s string) string
```
蛇形命名转驼峰命名。

### ValidateTableName
```go
func ValidateTableName(table string) error
```
验证表名是否合法。

### GenerateCacheKey
```go
func GenerateCacheKey(dbName, sql string, args ...interface{}) string
```
生成缓存键。

### SupportedDrivers
```go
func SupportedDrivers() []DriverType
```
返回支持的数据库驱动列表。

### IsValidDriver
```go
func IsValidDriver(driver DriverType) bool
```
检查驱动是否支持。

---

## 数据库驱动类型

```go
const (
    MySQL      DriverType = "mysql"
    PostgreSQL DriverType = "postgres"
    SQLite3    DriverType = "sqlite3"
    Oracle     DriverType = "oracle"
    SQLServer  DriverType = "sqlserver"
)
```
