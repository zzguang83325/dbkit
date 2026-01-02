# DBKit - Go Database Library

[中文文档](README.md) | [API Reference](api_en.md) | [中文 API 手册](api.md)

DBKit is a high-performance, lightweight database operation library for Go, inspired by the ActiveRecord pattern from Java's JFinal framework. It provides an extremely simple and intuitive API that makes database operations as easy as working with objects through `Record` and DbModel.

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

DBKit disables timestamp auto-update and optimistic lock checks by default for optimal performance. To enable:

```go
// Enable timestamp auto-update
dbkit.EnableTimestampCheck()

// Enable optimistic lock check
dbkit.EnableOptimisticLockCheck()

// Enable both features
dbkit.EnableFeatureChecks()
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

```go
// Use built-in LocalCache (memory)
dbkit.SetLocalCacheConfig(1 * time.Minute)

// Use Redis cache
import "github.com/zzguang83325/dbkit/redis"

rc, err := redis.NewRedisCache("localhost:6379", "username", "password", 1)
if err == nil {
    dbkit.SetCache(rc)
}

// Query with auto caching
dbkit.CreateCache("user_cache", 10*time.Minute)
records, err := dbkit.Cache("user_cache").Query("SELECT * FROM users")

// Manual cache operations
dbkit.CacheSet("my_store", "key1", "value1", 5*time.Minute)
val, ok := dbkit.CacheGet("my_store", "key1")
dbkit.CacheDelete("my_store", "key1")
dbkit.CacheClear("my_store")
```

### 6. Auto Timestamps

Auto timestamps automatically populate timestamp fields on insert and update operations without manual setting.

**Performance Note**: DBKit disables auto timestamp checks by default for optimal performance. Enable it when needed:

```go
// Enable timestamp auto-update
dbkit.EnableTimestampCheck()
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

### 9. Logging Configuration

```go
// Enable debug mode
dbkit.SetDebugMode(true)

// Use file logging
logFile := filepath.Join(".", "logfile.log")
dbkit.InitLoggerWithFile("debug", logFile)

// Custom logger (implement Logger interface)
dbkit.SetLogger(myCustomLogger)
```

### 10. Connection Pool Configuration

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
