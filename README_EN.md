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

### 6. Logging Configuration

```go
// Enable debug mode
dbkit.SetDebugMode(true)

// Use file logging
logFile := filepath.Join(".", "logfile.log")
dbkit.InitLoggerWithFile("debug", logFile)

// Custom logger (implement Logger interface)
dbkit.SetLogger(myCustomLogger)
```

### 7. Connection Pool Configuration

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
