# DBKit - Go Database Library

[中文文档](README.md) | [API Reference](api_en.md) | [中文 API 手册](api.md) | [SQL Template Guide](doc/en/SQL_TEMPLATE_GUIDE_EN.md) | [SQL 模板指南](doc/cn/SQL_TEMPLATE_GUIDE.md)

DBKit is a high-performance, lightweight database ORM library for Go, inspired  from Java's JFinal framework. It provides an extremely simple and intuitive API that makes database operations as easy as working with objects through `Record` and DbModel.

**Project Link**: https://github.com/zzguang83325/dbkit.git

## Features

- **Database Support**: MySQL, PostgreSQL, SQLite, SQL Server, Oracle
- **Multi-Database Management**: Connect to multiple databases simultaneously and switch between them easily
- **ActiveRecord Experience**: No need for tedious struct definitions, use flexible `Record` for CRUD operations
- **DbModel Experience**: Easily perform CRUD through auto-generated DbModel objects
- **Transaction Support**: Simple transaction wrappers and low-level transaction control
- **Auto Type Conversion**: Automatic handling of database types and Go types conversion
- **Parameterized Queries**: Automatic SQL parameter binding to prevent SQL injection
- **Pagination**: Optimized pagination implementation for different databases
- **Logging**: Built-in SQL logging with easy integration for various logging systems
- **Cache Support**: Built-in two-level cache supporting local memory and Redis cache with chain query caching
- **Connection Pool Management**: Built-in connection pool management for improved performance
- **Connection Pool Monitoring**: Connection pool statistics with Prometheus metrics export
- **Query Timeout Control**: Global and per-query timeout settings to prevent slow queries from blocking
- **Auto Timestamps**: Configurable auto timestamp fields, automatically populate created_at and updated_at on insert and update
- **Soft Delete Support**: Configurable soft delete fields, automatic filtering of deleted records, restore and force delete functions
- **Optimistic Lock Support**: Configurable version fields, automatic concurrent conflict detection, prevents data overwriting
- **SQL Templates**: SQL configuration management, dynamic parameter building, 🆕 variadic parameter support - [Detailed Guide](doc/en/SQL_TEMPLATE_GUIDE_EN.md)

## Performance Benchmark

DBKit outperforms GORM in most CRUD operations, with **15.1% better overall performance**.

MySQL-based performance test results (using separate tables to eliminate cache effects):

| Test | DBKit | GORM | Comparison |
|------|-------|------|------------|
| Single Insert | 440 ops/s | 356 ops/s | **DBKit 18.9% faster** |
| Batch Insert | 26,913 ops/s | 28,284 ops/s | GORM 4.8% faster |
| Single Query | 1,628 ops/s | 1,584 ops/s | **DBKit 2.7% faster** |
| Batch Query (100 rows) | 1,401 ops/s | 999 ops/s | **DBKit 28.7% faster** |
| Conditional Query | 1,413 ops/s | 1,409 ops/s | **DBKit 0.3% faster** |
| Update | 430 ops/s | 357 ops/s | **DBKit 17.1% faster** |
| Delete | 432 ops/s | 355 ops/s | **DBKit 17.9% faster** |
| **Total** | **6.03s** | **7.09s** | **DBKit 15.1% faster** |

**Key Advantages:**
- ✅ Batch query 28.7% faster (biggest advantage)
- ✅ Single insert 18.9% faster, delete 17.9% faster
- ✅ Update 17.1% faster
- ✅ Leads in 6 out of 7 test categories
- ✅ Record mode has no reflection overhead, excellent query performance

📊 **[View Full Performance Report](examples/benchmark/benchmark_report.md)**

**Test Methodology:**
- Separate tables (`benchmark_users_dbkit` and `benchmark_users_gorm`) to eliminate MySQL cache effects
- Same test conditions: data volume, batch size, test iterations
- Both use transactions for batch insert to ensure fair comparison
- Full benchmark code available in [examples/benchmark/](examples/benchmark/)

## Performance Optimization

DBKit disables timestamp auto-update, optimistic lock checks, and soft delete checks by default for optimal performance. To enable:

```go
// Enable timestamp auto-update
dbkit.EnableTimestamps()

// Enable optimistic lock
dbkit.EnableOptimisticLock()

// Enable soft delete
dbkit.EnableSoftDelete()
```

## Installation

```bash
go get github.com/zzguang83325/dbkit@latest
```

## Database Drivers

DBKit supports the following databases. Install the corresponding driver based on your database:

| Database   | Driver Package                   | Installation Command                      |
| ---------- | -------------------------------- | ----------------------------------------- |
| MySQL      | github.com/go-sql-driver/mysql   | `go get github.com/go-sql-driver/mysql`   |
| PostgreSQL | github.com/lib/pq                | `go get github.com/lib/pq`                |
| SQLite3    | github.com/mattn/go-sqlite3      | `go get github.com/mattn/go-sqlite3`      |
| SQL Server | github.com/denisenkom/go-mssqldb | `go get github.com/denisenkom/go-mssqldb` |
| Oracle     | github.com/sijms/go-ora/v2       | `go get github.com/sijms/go-ora/v2`       |

Import drivers in your code:

```go
// MySQL
import _ "github.com/go-sql-driver/mysql"

// PostgreSQL
import _ "github.com/lib/pq"

// SQLite3
import _ "github.com/mattn/go-sqlite3"

// SQL Server
import _ "github.com/denisenkom/go-mssqldb"

// Oracle
import _ "github.com/sijms/go-ora/v2"
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/zzguang83325/dbkit"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    // Initialize database connection
    err := dbkit.OpenDatabase(dbkit.MySQL, "root:password@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 10)
    if err != nil {
        log.Fatal(err)
    }
    defer dbkit.Close()

    // Test connection
    if err := dbkit.Ping(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Database connected successfully")

    // Create table
    dbkit.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        age INT NOT NULL,
        email VARCHAR(100) NOT NULL UNIQUE
    )`)

    // Create Record and insert data
    user := dbkit.NewRecord().
        Set("name", "John").
        Set("age", 25).
        Set("email", "john@example.com")
    
    id, err := dbkit.Save("users", user)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Insert successful, ID:", id)

    // Query data
    users, err := dbkit.Query("SELECT * FROM users WHERE age > ?", 18)
    if err != nil {
        log.Fatal(err)
    }
    for _, u := range users {
        fmt.Printf("ID: %d, Name: %s, Age: %d, Email: %s\n", 
            u.Int64("id"), u.Str("name"), u.Int("age"), u.Str("email"))
    }

    // Query single record
    record, _ := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", id)
    if record != nil {
        fmt.Printf("Name: %s, Age: %d\n", record.GetString("name"), record.GetInt("age"))
    }

    // Update data
    record.Set("age", 18)
    dbkit.Save("users", record)

    // Delete data
    rows, err := dbkit.Delete("users", "id = ?", id)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Delete successful, affected rows:", rows)

    // Pagination
    page := 1
    perPage := 10
    dataPage, err := dbkit.Paginate(page, perPage, "SELECT *", "users", "status=?", "id ASC", 1)
    if err != nil {
        log.Printf("Pagination failed: %v", err)
    } else {
        fmt.Printf("Page %d (per page %d), total: %d\n", page, perPage, dataPage.TotalRow)
    }
}
```

## DbModel Basic Usage

First call `GenerateDbModel` to generate the struct (automatically implements IDbModel interface):

```go
// Insert
user := &models.User{
    Name: "John",
    Age:  25,
}
id, err := user.Insert()  // or user.Save()

// Query
foundUser := &models.User{}
err := foundUser.FindFirst("id = ?", id)

// Update
foundUser.Age = 31
foundUser.Update()   // or foundUser.Save()

// Delete
foundUser.Delete()

// Query multiple
users, err := user.Find("id > ?", "id desc", 1)
for _, u := range users {
    fmt.Println(u.ToJson())
}

// Pagination
pageObj, err := foundUser.Paginate(1, 10, "id > ?", "id desc", 1)
if err != nil {
    return
}
fmt.Printf("Page %d (total %d pages), total records: %d\n", pageObj.PageNumber, pageObj.TotalPage, pageObj.TotalRow)
```

## 📁 Examples Directory

DBKit provides detailed examples for various databases in the `examples/` directory:

- `examples/mysql/` - MySQL usage examples
- `examples/postgres/` - PostgreSQL usage examples
- `examples/sqlite/` - SQLite usage examples
- `examples/oracle/` - Oracle usage examples
- `examples/sqlserver/` - SQL Server usage examples

Run examples with:

```bash
cd examples/mysql
go run main.go
```

## 📖 Documentation

### 1. Database Initialization

#### Single Database Configuration

```go
// Method 1: Quick initialization
dsn := "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
dbkit.OpenDatabase(dbkit.MySQL, dsn, 10)

// Method 2: Detailed configuration
config := &dbkit.Config{
    Driver:          dbkit.PostgreSQL,
    DSN:             "host=localhost port=5432 user=postgres dbname=test",
    MaxOpen:         50,
    MaxIdle:         25,
    ConnMaxLifetime: time.Hour,
}
dbkit.OpenDatabaseWithConfig(config)
```

#### Multi-Database Management

```go
// Connect to multiple databases
dbkit.OpenDatabaseWithDBName("main", dbkit.MySQL, "root:123456@tcp(localhost:3306)/test", 10)
dbkit.OpenDatabaseWithDBName("log_db", dbkit.SQLite3, "file:./logs.db", 5)
dbkit.OpenDatabaseWithDBName("oracle", dbkit.Oracle, "oracle://test:123456@127.0.0.1:1521/orcl", 25)

// Use specific database with chain calls
dbkit.Use("main").Query("...")
dbkit.Use("log_db").Save("logs", record)
```

### 2. Query Operations

#### Basic Queries

```go
// Query multiple records
users, _ := dbkit.Query("SELECT * FROM users WHERE status = ?", "active")

// Query first record (returns nil if not found)
user, _ := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)

// Return []map[string]interface{}
data, _ := dbkit.QueryMap("SELECT name, age FROM users")

// Count records
count, _ := dbkit.Count("users", "age > ?", 18)

// Check existence
if exists, _ := dbkit.Exists("users", "name = ?", "John"); exists {
    // ...
}
```

#### Chain Queries

```go
users, err := dbkit.Table("users").
    Where("age > ?", 18).
    Where("status = ?", "active").
    OrderBy("created_at DESC").
    Limit(10).
    Find()

// Single record
user, err := dbkit.Table("users").Where("id = ?", 1).FindFirst()

// Pagination
page, err := dbkit.Table("users").
    Where("age > ?", 18).
    OrderBy("id ASC").
    Paginate(1, 10)
```

##### Advanced WHERE Conditions

```go
// OrWhere - OR conditions
orders, err := dbkit.Table("orders").
    Where("status = ?", "active").
    OrWhere("priority = ?", "high").
    Find()
// Generates: WHERE (status = ?) OR priority = ?

// WhereInValues - IN query with value list
users, err := dbkit.Table("users").
    WhereInValues("id", []interface{}{1, 2, 3, 4, 5}).
    Find()
// Generates: WHERE id IN (?, ?, ?, ?, ?)

// WhereNotInValues - NOT IN query
orders, err := dbkit.Table("orders").
    WhereNotInValues("status", []interface{}{"cancelled", "refunded"}).
    Find()

// WhereBetween - Range query
users, err := dbkit.Table("users").
    WhereBetween("age", 18, 65).
    Find()
// Generates: WHERE age BETWEEN ? AND ?

// WhereNull / WhereNotNull - NULL checks
users, err := dbkit.Table("users").
    WhereNull("deleted_at").
    WhereNotNull("email").
    Find()
// Generates: WHERE deleted_at IS NULL AND email IS NOT NULL
```

##### Grouping and Aggregation

```go
// GroupBy + Having
stats, err := dbkit.Table("orders").
    Select("user_id, COUNT(*) as order_count, SUM(total) as total_amount").
    GroupBy("user_id").
    Having("COUNT(*) > ?", 5).
    Find()
// Generates: SELECT ... GROUP BY user_id HAVING COUNT(*) > ?
```

##### Complex Query Example

```go
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
```

##### Supported Methods

| Method | Description |
|--------|-------------|
| `Table(name)` | Specify table name |
| `Select(columns)` | Specify columns, default `*` |
| `Where(condition, args...)` | Add WHERE condition, multiple calls use `AND` |
| `And(condition, args...)` | Alias for `Where` |
| `OrWhere(condition, args...)` | Add OR condition |
| `WhereInValues(column, values)` | IN query with value list |
| `WhereNotInValues(column, values)` | NOT IN query with value list |
| `WhereBetween(column, min, max)` | BETWEEN range query |
| `WhereNotBetween(column, min, max)` | NOT BETWEEN range query |
| `WhereNull(column)` | IS NULL check |
| `WhereNotNull(column)` | IS NOT NULL check |
| `GroupBy(columns)` | GROUP BY clause |
| `Having(condition, args...)` | HAVING clause for filtering groups |
| `OrderBy(orderBy)` | Specify sort order |
| `Limit(limit)` | Limit number of records |
| `Offset(offset)` | Specify offset |
| `Find() / Query()` | Execute query and return results |
| `FindFirst() / QueryFirst()` | Execute query and return first record |
| `Delete()` | Delete with conditions (requires `Where`) |
| `Paginate(page, pageSize)` | Execute pagination query |

### 3. Insert & Update

```go
// Save (auto insert or update)
user := dbkit.NewRecord().Set("name", "John").Set("age", 20)
id, err := dbkit.Save("users", user)

// Insert
id, err := dbkit.Insert("users", user)

// Update
record := dbkit.NewRecord().Set("age", 26)
affected, err := dbkit.Update("users", record, "id = ?", 1)

// Batch insert
dbkit.BatchInsertDefault("users", records)
```

### 4. Transaction Handling

```go
// Automatic transaction
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    _, err := tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE id = ?", 1)
    if err != nil {
        return err // Auto rollback
    }
    
    record := dbkit.NewRecord().Set("amount", 100).Set("from_id", 1)
    _, err = tx.Save("transfer_logs", record)
    return err
})

// Manual transaction
tx, err := dbkit.BeginTransaction()
// ... operations
tx.Commit()   // or tx.Rollback()
```

### 5. Cache Support

DBKit uses built-in **LocalCache** (memory cache) by default. For Redis support, optionally import the `dbkit/redis` sub-package.

#### 1. Using Built-in LocalCache (Memory)
```go
// Default uses memory cache. Modify cache cleanup interval with the following function
// Expired cache data will be cleaned up periodically
dbkit.SetLocalCacheConfig(1 * time.Minute)
```

#### 2. Using Redis Cache (Optional)
First ensure your project imports the `dbkit/redis` sub-package, which will pull Redis dependencies.

```go
import "github.com/zzguang83325/dbkit/redis"

// Create Redis cache instance (parameters: address, username, password, DB)
rc, err := redis.NewRedisCache("localhost:6379", "username", "password", 1)
if err == nil {
    dbkit.SetCache(rc) // Switch global cache to Redis
}
```

#### 3. Query Data with Auto Caching
```go
// Create a cache store and set default expiration time
dbkit.CreateCache("user_cache", 10*time.Minute)

// Auto query cache: chain call Cache()
// If cache hits, return directly; otherwise query database and auto write to cache
records, err := dbkit.Cache("user_cache").Query("SELECT * FROM users")
```

#### 4. Manual Cache Operations

```go
// Store cache
dbkit.CacheSet("my_store", "key1", "value1", 5*time.Minute)

// Get cache
val, ok := dbkit.CacheGet("my_store", "key1")

// Delete specific key
dbkit.CacheDelete("my_store", "key1")

// Clear entire store
dbkit.CacheClear("my_store")
```

#### 5. View Cache Status

```go
status := dbkit.CacheStatus()
fmt.Printf("Type: %v\n", status["type"])
fmt.Printf("Total Items: %v\n", status["total_items"])
fmt.Printf("Estimated Memory: %v\n", status["estimated_memory_human"])
```

### 6. Auto Timestamps

Auto timestamps automatically populate timestamp fields on insert and update operations without manual setting.

**Note**: DBKit disables auto timestamp checks by default for optimal performance. Enable it when needed:

```go
// Enable timestamp auto-update
dbkit.EnableTimestamps()
```

**Configuration:**
```go
// Configure auto timestamps (default fields: created_at and updated_at)
dbkit.ConfigTimestamps("users")

// Use custom field names
dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")

// Configure only created_at
dbkit.ConfigCreatedAt("logs", "log_time")

// Configure only updated_at
dbkit.ConfigUpdatedAt("cache_data", "last_modified")
```

**Behavior:**
```go
// Insert: created_at auto-filled with current time
record := dbkit.NewRecord().Set("name", "John")
dbkit.Insert("users", record)

// Update: updated_at auto-filled with current time
updateRecord := dbkit.NewRecord().Set("name", "John Updated")
dbkit.Update("users", updateRecord, "id = ?", 1)

// Manual created_at (won't be overwritten)
customTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
record2 := dbkit.NewRecord().Set("name", "Jane").Set("created_at", customTime)
dbkit.Insert("users", record2)

// Temporarily disable auto timestamps
dbkit.Table("users").Where("id = ?", 1).WithoutTimestamps().Update(record)
```

### 7. Soft Delete

**Note**: DBKit disables soft delete checks by default for optimal performance. To enable this feature:

```go
// Enable soft delete
dbkit.EnableSoftDelete()
```

```go
// Configure soft delete
dbkit.ConfigSoftDelete("users", "deleted_at")

// Soft delete
dbkit.Delete("users", "id = ?", 1)

// Query (auto filters deleted)
users, _ := dbkit.Table("users").Find()

// Query including deleted
allUsers, _ := dbkit.Table("users").WithTrashed().Find()

// Restore
dbkit.Restore("users", "id = ?", 1)

// Force delete
dbkit.ForceDelete("users", "id = ?", 1)
```

### 8. Optimistic Lock

Optimistic lock detects concurrent update conflicts through version fields, preventing data from being accidentally overwritten.

```go
// Configure optimistic lock (default field: version)
dbkit.ConfigOptimisticLock("products")

// Custom field name
dbkit.ConfigOptimisticLockWithField("orders", "revision")

// Insert (version auto-initialized to 1)
record := dbkit.NewRecord().Set("name", "Laptop").Set("price", 999.99)
dbkit.Insert("products", record)

// Update with version
updateRecord := dbkit.NewRecord()
updateRecord.Set("version", int64(1))  // current version
updateRecord.Set("price", 899.99)
rows, err := dbkit.Update("products", updateRecord, "id = ?", 1)
// Success: version auto-incremented to 2

// Concurrent conflict detection (stale version)
staleRecord := dbkit.NewRecord()
staleRecord.Set("version", int64(1))  // stale version!
staleRecord.Set("price", 799.99)
rows, err = dbkit.Update("products", staleRecord, "id = ?", 1)
if errors.Is(err, dbkit.ErrVersionMismatch) {
    fmt.Println("Concurrent conflict detected")
}

// Correct way: read latest version first
latestRecord, _ := dbkit.Table("products").Where("id = ?", 1).FindFirst()
currentVersion := latestRecord.GetInt("version")
updateRecord2 := dbkit.NewRecord()
updateRecord2.Set("version", currentVersion)
updateRecord2.Set("price", 799.99)
dbkit.Update("products", updateRecord2, "id = ?", 1)
```

#### Using Optimistic Lock in Transactions
```go
dbkit.Transaction(func(tx *dbkit.Tx) error {
    rec, _ := tx.Table("products").Where("id = ?", 1).FindFirst()
    currentVersion := rec.GetInt("version")
    
    updateRec := dbkit.NewRecord()
    updateRec.Set("version", currentVersion)
    updateRec.Set("stock", 80)
    _, err := tx.Update("products", updateRec, "id = ?", 1)
    return err  // Auto rollback on version conflict
})
```

### 9. Query Timeout Control

DBKit supports global and per-query timeout settings using Go's standard `context.Context`, automatically canceling queries after timeout.

#### Global Default Timeout
```go
config := &dbkit.Config{
    Driver:       dbkit.MySQL,
    DSN:          "...",
    MaxOpen:      10,
    QueryTimeout: 30 * time.Second,  // All queries default to 30s timeout
}
dbkit.OpenDatabaseWithConfig(config)
```

#### Per-Query Timeout
```go
// Method 1: Global function
users, err := dbkit.Timeout(5 * time.Second).Query("SELECT * FROM users")

// Method 2: Specify database
users, err := dbkit.Use("default").Timeout(5 * time.Second).Query("SELECT * FROM users")

// Method 3: Chain query
users, err := dbkit.Table("users").
    Where("age > ?", 18).
    Timeout(10 * time.Second).
    Find()
```

#### Timeout in Transactions
```go
dbkit.Transaction(func(tx *dbkit.Tx) error {
    // Queries in transactions also support timeout
    _, err := tx.Timeout(5 * time.Second).Query("SELECT * FROM orders")
    return err
})
```

#### Timeout Error Handling
```go
import "context"

users, err := dbkit.Timeout(1 * time.Second).Query("SELECT SLEEP(5)")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Query timeout")
    }
}
```

### 10. Connection Pool Monitoring

DBKit provides connection pool monitoring to view real-time pool usage.

#### Get Pool Statistics
```go
// Get default database pool statistics
stats := dbkit.GetPoolStats()
fmt.Println(stats.String())
// Output: PoolStats[default/mysql]: Open=5 (InUse=2, Idle=3), MaxOpen=10, WaitCount=0, WaitDuration=0s

// Get specific database pool statistics
stats := dbkit.GetPoolStatsDB("postgresql")

// Get all database pool statistics
allStats := dbkit.AllPoolStats()
for name, stats := range allStats {
    fmt.Printf("%s: %s\n", name, stats.String())
}
```

#### PoolStats Structure
```go
type PoolStats struct {
    DBName             string        // Database name
    Driver             string        // Driver type
    MaxOpenConnections int           // Maximum connections (configured)
    OpenConnections    int           // Current open connections
    InUse              int           // Connections in use
    Idle               int           // Idle connections
    WaitCount          int64         // Total wait count
    WaitDuration       time.Duration // Total wait duration
    MaxIdleClosed      int64         // Connections closed due to max idle
    MaxLifetimeClosed  int64         // Connections closed due to max lifetime
}
```

#### Convert to Map (for JSON serialization)
```go
stats := dbkit.GetPoolStats()
statsMap := stats.ToMap()
jsonBytes, _ := json.Marshal(statsMap)
fmt.Println(string(jsonBytes))
```

#### Export Prometheus Metrics
```go
// Single database
stats := dbkit.GetPoolStats()
fmt.Println(stats.PrometheusMetrics())

// All databases
fmt.Println(dbkit.AllPrometheusMetrics())
```

Output example:
```
# HELP dbkit_pool_max_open_connections Maximum number of open connections to the database.
# TYPE dbkit_pool_max_open_connections gauge
dbkit_pool_max_open_connections{db="default",driver="mysql"} 10

# HELP dbkit_pool_open_connections The number of established connections both in use and idle.
# TYPE dbkit_pool_open_connections gauge
dbkit_pool_open_connections{db="default",driver="mysql"} 5

# HELP dbkit_pool_in_use The number of connections currently in use.
# TYPE dbkit_pool_in_use gauge
dbkit_pool_in_use{db="default",driver="mysql"} 2

# HELP dbkit_pool_idle The number of idle connections.
# TYPE dbkit_pool_idle gauge
dbkit_pool_idle{db="default",driver="mysql"} 3
```

### 11. Logging Configuration

DBKit uses the standard library `log` by default. For more powerful logging, you can optionally use the `zap` logging library.

#### 1. Output Logs to Console
```go
// Enable Debug mode to output SQL statements
dbkit.SetDebugMode(true)
```

#### 2. Using slog Logger

```go
logFile := filepath.Join(".", "logfile.log")
dbkit.InitLoggerWithFile("debug", logFile)
```

#### 3. Using Zap Logger Library

```go
type ZapAdapter struct {
    logger *zap.Logger
}

func (a *ZapAdapter) Log(level dbkit.LogLevel, msg string, fields map[string]interface{}) {
    var zapFields []zap.Field
    if len(fields) > 0 {
        zapFields = make([]zap.Field, 0, len(fields))
        for k, v := range fields {
            zapFields = append(zapFields, zap.Any(k, v))
        }
    }

    switch level {
    case dbkit.LevelDebug:
        a.logger.Debug(msg, zapFields...)
    case dbkit.LevelInfo:
        a.logger.Info(msg, zapFields...)
    case dbkit.LevelWarn:
        a.logger.Warn(msg, zapFields...)
    case dbkit.LevelError:
        a.logger.Error(msg, zapFields...)
    }
}

func main() {
    // 1. Initialize zap logger, output to both console and file
    cfg := zap.NewDevelopmentConfig()
    cfg.OutputPaths = []string{"stdout", "logfile.log"}

    zapLogger, _ := cfg.Build()
    defer zapLogger.Sync()

    // 2. Integrate zap into dbkit
    dbkit.SetLogger(&ZapAdapter{logger: zapLogger})
    dbkit.SetDebugMode(true) // Enable debug mode to see SQL traces
}
```

#### 4. Using zerolog
Simply implement the `dbkit.Logger` interface:
```go
type ZerologAdapter struct {
    logger zerolog.Logger
}

func (a *ZerologAdapter) Log(level dbkit.LogLevel, msg string, fields map[string]interface{}) {
    var event *zerolog.Event
    switch level {
    case dbkit.LevelDebug:
        event = a.logger.Debug()
    case dbkit.LevelInfo:
        event = a.logger.Info()
    case dbkit.LevelWarn:
        event = a.logger.Warn()
    case dbkit.LevelError:
        event = a.logger.Error()
    default:
        event = a.logger.Log()
    }

    if len(fields) > 0 {
        event.Fields(fields)
    }
    event.Msg(msg)
}

func main() {
    // 1. Initialize zerolog logger
    // Open log file
    logFile, _ := os.OpenFile("logfile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    defer logFile.Close()

    // 2. Chain create Logger: output to both console and file
    logger := zerolog.New(zerolog.MultiLevelWriter(
        zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339},
        logFile,
    )).With().Timestamp().Logger()

    // 3. Integrate zerolog into dbkit
    dbkit.SetLogger(&ZerologAdapter{logger: logger})
    dbkit.SetDebugMode(true) // Enable debug mode to see SQL
}
```

### 12. SQL Templates

DBKit provides powerful SQL template functionality that allows you to manage SQL statements through configuration, supporting dynamic parameters, conditional building, and multi-database execution.

#### Configuration File Structure

SQL templates use JSON format configuration files:

```json
{
  "version": "1.0",
  "description": "User Service SQL Configuration",
  "namespace": "user_service",
  "sqls": [
    {
      "name": "findById",
      "description": "Find user by ID",
      "sql": "SELECT * FROM users WHERE id = ?",
      "type": "select"
    },
    {
      "name": "findByIdAndStatus",
      "description": "Find user by ID and status",
      "sql": "SELECT * FROM users WHERE id = ? AND status = ?",
      "type": "select"
    },
    {
      "name": "updateUser",
      "description": "Update user information",
      "sql": "UPDATE users SET name = ?, email = ?, age = ? WHERE id = ?",
      "type": "update"
    }
  ]
}
```

#### Parameter Type Support

DBKit SQL templates support multiple parameter passing methods:

| Parameter Type | Use Case | SQL Placeholder | Example |
|---------------|----------|-----------------|---------|
| `map[string]interface{}` | Named parameters | `:name` | `map[string]interface{}{"id": 123}` |
| `[]interface{}` | Multiple positional parameters | `?` | `[]interface{}{123, "John"}` |
| Single simple types | Single positional parameter | `?` | `123`, `"John"`, `true` |
| **🆕 Variadic parameters** | **Multiple positional parameters** | `?` | `SqlTemplate(name, 123, "John", true)` |

#### Configuration Loading

```go
// Load single configuration file
err := dbkit.LoadSqlConfig("config/user_service.json")

// Load multiple configuration files
configPaths := []string{
    "config/user_service.json",
    "config/order_service.json",
}
err := dbkit.LoadSqlConfigs(configPaths)

// Load all JSON configuration files from directory
err := dbkit.LoadSqlConfigDir("config/")
```

#### SQL Template Execution

```go
// 1. Single simple parameter
user, err := dbkit.SqlTemplate("user_service.findById", 123).QueryFirst()

// 2. 🆕 Variadic parameters (recommended for multi-parameter queries)
users, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 123, 1).Query()

// 3. Update operations
result, err := dbkit.SqlTemplate("user_service.updateUser", 
    "John Doe", "john@example.com", 30, 123).Exec()

// 4. Named parameters (suitable for complex queries)
params := map[string]interface{}{
    "name": "John",
    "status": 1,
}
users, err := dbkit.SqlTemplate("user_service.findByNamedParams", params).Query()

// 5. Positional parameter array (backward compatible)
users, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 
    []interface{}{123, 1}).Query()
```

#### Multi-Database and Transaction Support

```go
// Execute on specific database
users, err := dbkit.Use("mysql").SqlTemplate("findUsers", 123, 1).Query()

// Use in transactions
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    result, err := tx.SqlTemplate("insertUser", "John", "john@example.com", 25).Exec()
    return err
})

// Set timeout
users, err := dbkit.SqlTemplate("findUsers", 123).
    Timeout(30 * time.Second).Query()
```

#### Parameter Count Validation

The system automatically validates that parameter count matches SQL placeholder count:

```go
// ✅ Correct: 2 parameters match 2 placeholders
users, err := dbkit.SqlTemplate("findByIdAndStatus", 123, 1).Query()

// ❌ Error: insufficient parameters
users, err := dbkit.SqlTemplate("findByIdAndStatus", 123).Query()
// Returns: parameter count mismatch: SQL has 2 '?' placeholders but got 1 parameters

// ❌ Error: too many parameters
users, err := dbkit.SqlTemplate("findByIdAndStatus", 123, 1, 2).Query()
// Returns: parameter count mismatch: SQL has 2 '?' placeholders but got 3 parameters
```

#### Dynamic SQL Building

Through `inparam` configuration, you can implement dynamic SQL condition building:

```json
{
  "name": "searchUsers",
  "sql": "SELECT * FROM users WHERE 1=1",
  "inparam": [
    {
      "name": "status",
      "type": "int",
      "desc": "User status",
      "sql": " AND status = ?"
    },
    {
      "name": "ageMin",
      "type": "int", 
      "desc": "Minimum age",
      "sql": " AND age >= ?"
    }
  ],
  "order": "created_at DESC"
}
```

```go
// Only pass partial parameters, system will automatically build corresponding SQL
params := map[string]interface{}{
    "status": 1,
    // ageMin not provided, corresponding condition won't be added
}
users, err := dbkit.SqlTemplate("searchUsers", params).Query()
// Generated SQL: SELECT * FROM users WHERE 1=1 AND status = ? ORDER BY created_at DESC
```

#### Best Practices

1. **Single parameter queries** - Use `?` placeholders with simple parameters
2. **Multi-parameter queries** - Use variadic parameters or named parameters
3. **Complex queries** - Use named parameters and dynamic SQL
4. **Parameter validation** - System automatically validates parameter count and type
5. **Error handling** - Catch and handle `SqlConfigError` type errors

### 13. Connection Pool Configuration

```go
config := &dbkit.Config{
    Driver:          dbkit.MySQL,
    DSN:             "root:password@tcp(127.0.0.1:3306)/test?charset=utf8mb4",
    MaxOpen:         50,    // Maximum open connections
    MaxIdle:         25,    // Maximum idle connections
    ConnMaxLifetime: time.Hour, // Maximum connection lifetime
}
dbkit.OpenDatabaseWithConfig(config)
```

## 🔗 Project Links

- GitHub Repository: [https://github.com/zzguang83325/dbkit.git](https://github.com/zzguang83325/dbkit.git)
- [API Reference (English)](api_en.md)
- [API 手册 (中文)](api.md)
