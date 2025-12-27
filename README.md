# DBKit - Go Database  Library

DBKit 是一个基于 Go 语言的高性能、轻量级数据库操作库，灵感来自 Java 的 JFinal 框架的 ActiveRecord 模式。它提供了极其简洁、直观的 API，通过 `Record` 对象和链式调用，让数据库操作变得像操作对象一样简单。

🔗 **项目链接**：[https://github.com/zzguang83325/dbkit.git](https://github.com/zzguang83325/dbkit.git)

## ✨ 特性

- 🚀 **数据库支持**：支持 MySQL、PostgreSQL、SQLite3、Oracle、SQL Server。
- 📦 **ActiveRecord 体验**：摆脱繁琐的 Struct 定义，使用灵活的 `Record` 对象进行 CRUD。
- 🎯 **多数据库管理**：支持同时连接多个数据库，并能轻松在它们之间切换。
- 📊 **内置分页**：针对不同数据库优化的分页查询实现。
- 🔄 **事务支持**：提供简单易用的事务包装器及底层事务控制。
- 📝 **调试友好**：内置 SQL 日志功能，支持多级日志输出。
- 🔗 **连接池管理**：自动管理数据库连接池，性能优异。

## 📦 安装

```bash
go get github.com/zzguang83325/dbkit
```

## 🚀 快速开始

```go
package main

import (
    "fmt"
    "log"
    "github.com/zzguang83325/dbkit"
)

func main() {
    // 1. 初始化数据库连接（默认注册为 "default"）
    dbkit.OpenDatabase(dbkit.MySQL, "root:password@tcp(127.0.0.1:3306)/test?charset=utf8mb4", 10)
    defer dbkit.Close()
    
    // 2. 插入数据
    user := dbkit.NewRecord().
        Set("name", "张三").
        Set("age", 25).
        Set("email", "zhangsan@example.com")
    
    id, err := dbkit.Save("users", user)
    if err == nil {
        fmt.Println("插入成功，ID:", id)
    }
    
    // 3. 查询数据
    users, err := dbkit.Query("SELECT * FROM users WHERE age > ?", 18)
    if err == nil {
        for _, u := range users {
            fmt.Printf("姓名: %s, 年龄: %d\n", u.Str("name"), u.Int("age"))
        }
    }
    
    // 4. 更新数据
    updateData := dbkit.NewRecord().Set("age", 26)
    _, err = dbkit.Update("users", updateData, "id = ?", id)
    
    // 5. 删除数据
    _, err = dbkit.Delete("users", "id = ?", id)
}
```

## 📁 示例目录

DBKit 提供了针对各种数据库的详细示例，您可以在 `examples/` 目录中找到：

- `examples/mysql/` - MySQL 数据库使用示例
- `examples/postgres/` - PostgreSQL 数据库使用示例
- `examples/sqlite/` - SQLite 数据库使用示例
- `examples/oracle/` - Oracle 数据库使用示例
- `examples/sqlserver/` - SQL Server 数据库使用示例
- `examples/multi_db/` - 多数据库同时使用示例

您可以通过运行以下命令来测试这些示例：

```bash
cd examples/mysql
go run main.go
```

## 📖 核心文档

### 1. 数据库初始化

#### 单数据库配置

```go
// 方式 1：快捷初始化
dsn:="root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
dbkit.OpenDatabase(dbkit.MySQL, dsn, 10)

// 方式 2：详细配置
config := &dbkit.Config{
    Driver:          dbkit.PostgreSQL,
    DSN:             "host=localhost port=5432 user=postgres dbname=test",
    MaxOpen:         50,
    MaxIdle:         25,
    ConnMaxLifetime: time.Hour,
}
dbkit.OpenDatabaseWithConfig(config)
```

#### 多数据库管理

```go
// 同时连接多个数据库
dbkit.OpenDatabaseWithDBName("main", dbkit.MySQL, "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 10)
dbkit.OpenDatabaseWithDBName("log_db", dbkit.SQLite3, "file:./logs.db", 5)
dbkit.OpenDatabaseWithDBName("oracle", dbkit.Oracle, "oracle://test:123456@127.0.0.1:1521/orcl", 25)
// SQL Server
dbkit.OpenDatabaseWithDBName("sqlserver", dbkit.SQLServer, "sqlserver://sa:123456@127.0.0.1:1433?database=test", 25)



// 使用 Use() 切换
dbkit.Use("main").Query("...")
dbkit.Use("main").Exec("...")
dbkit.Use("log_db").Save("logs", record)

// 获取特定库
db := dbkit.Use("main")
db.Query("...")
```

### 2. 查询操作

#### 基本查询

```go
// 操作默认数据库
users := dbkit.Query("SELECT * FROM users WHERE status = ?", "active")

// 返回第一条 Record (若无记录返回 nil)
user := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// 返回 []map[string]interface{}
data := dbkit.QueryMap("SELECT name, age FROM users")

// 统计记录
count, _ := dbkit.Count("users", "age > ?", 18)

// 检查是否存在
if dbkit.Exists("users", "name = ?", "张三") {
    // ...
}

//操作其它数据库用  dbkit.Use("main").Query("...")
```

#### 分页查询 (Paginate)

DBKit 的分页查询非常智能，它会自动分析 SQL 语句，并尝试优化 `COUNT(*)` 查询以提高性能。如果无法优化（如包含 `DISTINCT` 或 `GROUP BY`），则会自动降级为子查询模式。

```go
// 方式 1：操作默认数据库
// 参数：页码, 每页数量, SELECT 部分, 表名, WHERE 部分, ORDER BY 部分, 动态参数
// 返回：记录列表, 总记录数, 错误
users, total, err := dbkit.Paginate(1, 10, "id, name, age", "users", "age > ?", "id DESC", 18)

// 方式 2：指定数据库
// 参数：页码, 每页数量, SELECT 部分, 表名, WHERE 部分, ORDER BY 部分, 动态参数
db := dbkit.Use("default")
users, total, err := db.Paginate(1, 10, "SELECT *", "users", "age > ?", "id DESC", 18)
```

### 3. 插入与更新

#### Save (自动识别插入或更新)
`Save` 方法会自动识别主键（支持自动从数据库元数据获取主键名）。

- 如果 `Record` 中包含主键值且数据库中已存在该记录，则执行 `Update`。
- 如果不包含主键值或记录不存在，则执行 `Insert`。

```go
// 情况 1：插入新记录（无主键）
user := dbkit.NewRecord().Set("name", "张三").Set("age", 20)
id, err := dbkit.Save("users", user)

// 情况 2：更新记录（带主键）
user.Set("id", 1).Set("name", "张三-已更新")
affected, err := dbkit.Save("users", user)
```

#### Insert (强制插入)
`Insert` 始终执行 `INSERT` 语句，如果主键冲突会返回错误。

```go
user := dbkit.NewRecord().Set("name", "李四")
id, err := dbkit.Insert("users", user)
```

#### Update (显式更新)
```go
record := dbkit.NewRecord().Set("age", 26)
affected, err := dbkit.Update("users", record, "id = ?", 1)
```

#### Delete (删除数据)
```go
rows, err := dbkit.Delete("users", "id = ?", 10)
```

#### 批量插入

```go
var records []*dbkit.Record
// ... 填充 records
// 默认每批 100 条
dbkit.BatchInsertDefault("users", records)

// 自定义每批数量
dbkit.BatchInsert("users", records, 500)
```

### 4. Record 对象详解

`Record` 是 DBKit 的核心，它类似于一个增强版的 `map[string]interface{}`。

```go
r := dbkit.NewRecord()
r.Set("id", 1).Set("name", "王五")

// 类型安全获取
r.GetString("name") / r.Str("name")
r.GetInt("id")     / r.Int("id")
r.GetInt64("id")   / r.Int64("id")
r.GetFloat("price")/ r.Float("price")
r.GetBool("is_vip") / r.Bool("is_vip")

// 辅助方法
r.Has("email")      // 检查字段是否存在
r.Keys()            // 获取所有列名
r.ToMap()           // 转为 map
r.ToJson()          // 转为 JSON 字符串
r.FromJson(jsonStr) // 从 JSON 解析
```

### 5. 事务处理

#### 自动事务 

`Transaction` 函数会自动处理 `Commit` 和 `Rollback`。只要闭包返回 `error`，事务就会回滚。

```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // 注意：在事务中必须使用 tx 对象的方法
    _, err := tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE id = ?", 1)
    if err != nil {
        return err
    }
    
    record := dbkit.NewRecord().Set("amount", 100).Set("from_id", 1)
    _, err = tx.Save("transfer_logs", record)
    return err
})
```

#### 手动控制

```go
tx, err := dbkit.BeginTransaction()
// ... 执行操作
tx.Commit()   // 或 tx.Rollback()
```

### 6. 日志功能

DBKit 内置了强大的日志功能，基于 zap 日志库，支持多级日志输出、SQL 语句记录以及动态日志级别切换：

```go
// 1. 初始化文件日志（支持 debug, info, warn, error 级别）
logFilePath := filepath.Join(".", "log.log")
dbkit.InitLoggerWithFile("info", logFilePath)

// 2. 动态切换调试模式
// 开启调试模式后，所有的 SQL 执行详情（包括参数）都会输出到日志中
dbkit.SetDebugMode(true)

// 3. 也可以直接通过日志函数输出
dbkit.LogInfo("数据库初始化成功")
```

日志输出示例：
```
2025-12-27T15:44:54.898+0800    DEBUG   dbkit/logger.go:132     SQL executed    {"db": "default", "sql": "SELECT * FROM users ORDER BY id DESC OFFSET 0 ROWS FETCH NEXT 2 ROWS ONLY", "args": null}
```

### 7. 连接池配置

DBKit 自动管理数据库连接池，您可以通过 Config 结构体进行详细配置：

```go
config := &dbkit.Config{
    Driver:          dbkit.MySQL,
    DSN:             "root:password@tcp(127.0.0.1:3306)/test?charset=utf8mb4",
    MaxOpen:         50,    // 最大打开连接数
    MaxIdle:         25,    // 最大空闲连接数
    ConnMaxLifetime: time.Hour, // 连接最大生命周期
}

dbkit.OpenDatabaseWithConfig(config)
```

### 8. Record 对象高级用法

Record 对象提供了丰富的方法来处理数据：

```go
// 创建 Record 对象
record := dbkit.NewRecord().
    Set("name", "李四").
    Set("age", 30).
    Set("email", "lisi@example.com").
    Set("is_vip", true).
    Set("salary", 8000.50)

// 类型安全获取值
name := record.Str("name")       // 获取字符串
age := record.Int("age")         // 获取整数
email := record.Str("email")     // 获取字符串
isVIP := record.Bool("is_vip")   // 获取布尔值
salary := record.Float("salary") // 获取浮点数

// 检查字段是否存在
if record.Has("department") {
    department := record.Str("department")
}

// 获取所有键
keys := record.Keys() // []string{"name", "age", "email", "is_vip", "salary"}

// 转换为 map
recordMap := record.ToMap() // map[string]interface{}

// 转换为 JSON
jsonStr := record.ToJson() // 不返回错误，失败时返回 "{}"

// 从 JSON 创建 Record
newRecord := dbkit.NewRecord()
err := newRecord.FromJson(jsonStr) // 返回解析错误

// 删除字段
record.Remove("is_vip")

// 清空所有字段
record.Clear()
```

## 📚 API 文档

### 1. 数据库连接与管理

#### 初始化连接
```go
// 单数据库 快捷初始化
dbkit.OpenDatabase(driver DriverType, dsn string, maxOpen int)


// 多数据库初始化
dbkit.OpenDatabaseWithDBName(name string, driver DriverType, dsn string, maxOpen int)
```

#### 数据库切换与管理
```go

// 获取当前数据库
currentDB := dbkit.GetCurrentDB()

// 获取当前数据库名称
currentDBName := dbkit.GetCurrentDBName()

// 列出所有注册的数据库
allDBs := dbkit.ListDatabases()

// 关闭所有数据库连接
dbkit.Close()
```

### 2. 查询操作

#### 基本查询
```go
// 查询多条记录
records, err := dbkit.Query(sql string, args ...interface{}) ([]Record, error)

// 查询第一条记录
record, err := dbkit.QueryFirst(sql string, args ...interface{}) (*Record, error)

// 查询并返回 map 格式
resultMap, err := dbkit.QueryMap(sql string, args ...interface{}) ([]map[string]interface{}, error)

// 执行 SQL 语句
result, err := dbkit.Exec(sql string, args ...interface{}) (sql.Result, error)

// 统计记录数
count, err := dbkit.Count(table string, where string, whereArgs ...interface{}) (int64, error)

// 检查记录是否存在
exists := dbkit.Exists(table string, where string, whereArgs ...interface{}) bool
// 或者使用带错误返回的版本
exists, err := dbkit.ExistsWithError(table string, where string, whereArgs ...interface{}) (bool, error)
```

#### 分页查询
```go
// 分页查询
records, total, err := db.Paginate(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) ([]Record, int64, error)
```

### 3. CRUD 操作

#### 保存与更新
```go
// 保存记录（自动判断插入或更新）
id, err := dbkit.Save(table string, record *Record)

// 插入记录
id, err := dbkit.Insert(table string, record *Record)

// 更新记录
rowsAffected, err := dbkit.Update(table string, record *Record, where string, whereArgs ...interface{})

// 删除记录
rowsAffected, err := dbkit.Delete(table string, where string, whereArgs ...interface{})
```

#### 批量操作
```go
// 默认批量插入（每批 100 条）
totalRows, err := dbkit.BatchInsertDefault(table string, records []*Record)

// 自定义批量大小
totalRows, err := dbkit.BatchInsert(table string, records []*Record, batchSize int)
```

### 4. 事务操作

#### 自动事务
```go
// 自动提交和回滚的事务
err := dbkit.Transaction(func(tx *Tx) error {
    // 在事务中执行操作
    _, err := tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE id = ?", 1)
    if err != nil {
        return err // 发生错误时自动回滚
    }
    
    record := dbkit.NewRecord().Set("amount", 100).Set("from_id", 1)
    _, err = tx.Save("transfer_logs", record)
    return err // 成功时自动提交
})
```

#### 手动事务
```go
// 开始事务
tx, err := dbkit.BeginTransaction()

// 在事务中执行操作
_, err = tx.Exec(sql, args...)

// 提交事务
err = tx.Commit()

// 回滚事务
err = tx.Rollback()
```

### 5. 日志操作

#### 日志配置
```go

// 初始化文件日志
dbkit.InitLoggerWithFile(level string, logFilePath string)
```

#### 日志级别
```go
const (
    LogLevelDebug LogLevel = "debug"
    LogLevelInfo  LogLevel = "info"
    LogLevelWarn  LogLevel = "warn"
    LogLevelError LogLevel = "error"
)
```

#### 日志输出
```go
// 调试日志
dbkit.LogDebug(msg string, fields ...zap.Field)

// 信息日志
dbkit.LogInfo(msg string, fields ...zap.Field)

// 警告日志
dbkit.LogWarn(msg string, fields ...zap.Field)

// 错误日志
dbkit.LogError(msg string, fields ...zap.Field)
```

### 6. Record 对象

#### 创建与设置
```go
// 创建新 Record
record := dbkit.NewRecord()

// 链式设置字段
record.Set(column string, value interface{}) *Record
```

#### 类型安全获取
```go
// 获取字符串
strVal := record.GetString(column string) // 或 record.Str(column string)

// 获取整数
intVal := record.GetInt(column string)     // 或 record.Int(column string)

int64Val := record.GetInt64(column string) // 或 record.Int64(column string)

// 获取浮点数
floatVal := record.GetFloat(column string) // 或 record.Float(column string)

// 获取布尔值
boolVal := record.GetBool(column string)   // 或 record.Bool(column string)
```

#### 辅助方法
```go
// 获取原始值
val := record.Get(column string)

// 检查字段是否存在
has := record.Has(column string)

// 获取所有字段名
keys := record.Keys()

// 删除字段
record.Remove(column string)

// 清空所有字段
record.Clear()
```

#### 转换方法
```go
// 转换为 map
recordMap := record.ToMap() // 返回 map[string]interface{}

// 转换为 JSON
jsonStr := record.ToJson() // 返回 string

// 从 JSON 解析
err := record.FromJson(jsonStr) // 参数为 string，返回 error
```

## ⚖️ License

MIT License

## 🔗 项目链接

GitHub 仓库：[https://github.com/zzguang83325/dbkit.git](https://github.com/zzguang83325/dbkit.git)
