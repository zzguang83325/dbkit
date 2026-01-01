# DBKit API 手册

[English Version](api_en.md) | [README](README.md) | [English README](README_EN.md)

## 目录

- [数据库初始化](#数据库初始化)
- [查询操作](#查询操作)
- [插入与更新](#插入与更新)
- [删除操作](#删除操作)
- [事务处理](#事务处理)
- [Record 对象](#record-对象)
- [链式查询](#链式查询)
- [DbModel 操作](#dbmodel-操作)
- [缓存操作](#缓存操作)
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
func Paginate(page, pageSize int, selectSql, table, whereSql, orderBySql string, args ...interface{}) (*Page[Record], error)
func (db *DB) Paginate(page, pageSize int, selectSql, table, whereSql, orderBySql string, args ...interface{}) (*Page[Record], error)
```
分页查询。

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
根据条件删除记录。

### DeleteRecord
```go
func DeleteRecord(table string, record *Record) (int64, error)
func (db *DB) DeleteRecord(table string, record *Record) (int64, error)
func (tx *Tx) DeleteRecord(table string, record *Record) (int64, error)
```
根据 Record 中的主键删除记录。

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
