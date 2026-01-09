# DBKit API Reference

[Chinese Version](api.md) | [README](README.md) | [English README](README_EN.md)

## Table of Contents

- [Database Initialization](#database-initialization)
- [Database Connection Monitoring](#database-connection-monitoring)
- [Query Operations](#query-operations)
- [Query Timeout Control](#query-timeout-control)
- [Insert and Update](#insert-and-update)
- [Delete Operations](#delete-operations)
- [Soft Delete](#soft-delete)
- [Automatic Timestamps](#automatic-timestamps)
- [Optimistic Lock](#optimistic-lock)
- [Transaction Processing](#transaction-processing)
- [Record Object](#record-object)
- [Chained Query](#chained-query)
- [DbModel Operations](#dbmodel-operations)
- [Cache Operations](#cache-operations)
- [SQL Templates](#sql-templates)
- [Log Configuration](#log-configuration)
- [Utility Functions](#utility-functions)

---

## Database Initialization

### OpenDatabase
```go
func OpenDatabase(driver DriverType, dsn string, maxOpen int) error
```
Open database connection with default configuration.

**Parameters:**
- `driver`: Database driver type (MySQL, PostgreSQL, SQLite3, Oracle, SQLServer)
- `dsn`: Data Source Name (connection string)
- `maxOpen`: Maximum number of open connections

**Example:**
```go
err := dbkit.OpenDatabase(dbkit.MySQL, "root:password@tcp(localhost:3306)/test", 10)
```

### OpenDatabaseWithConfig
```go
func OpenDatabaseWithConfig(config *Config) error
```
Open database connection with custom configuration.

**Config Struct:**
```go
type Config struct {
    Driver          DriverType    // Database driver type
    DSN             string        // Data Source Name
    MaxOpen         int           // Maximum open connections
    MaxIdle         int           // Maximum idle connections
    ConnMaxLifetime time.Duration // Maximum connection lifetime
    QueryTimeout    time.Duration // Default query timeout (0 means no limit)
    
    // Connection monitoring configuration (new)
    MonitorNormalInterval time.Duration // Normal check interval (default 60s, 0 disables monitoring)
    MonitorErrorInterval  time.Duration // Error check interval (default 10s)
}
```

### OpenDatabaseWithDBName
```go
func OpenDatabaseWithDBName(dbname string, driver DriverType, dsn string, maxOpen int) error
```
Open a database connection with a specified name (Multi-database mode).

### Register
```go
func Register(dbname string, config *Config) error
```
Register a named database with custom configuration.

### Use
```go
func Use(dbname string) *DB
```
Switch to the database with the specified name and return the DB object for chaining.

**Example:**
```go
db := dbkit.Use("main")
records, err := db.Query("SELECT * FROM users")
```

### Close
```go
func Close() error
func CloseDB(dbname string) error
```
Close database connections.

### Ping
```go
func Ping() error
func PingDB(dbname string) error
```
Test database connection.

---

## Database Connection Monitoring

DBKit provides automatic database connection monitoring functionality that periodically checks database connection status and automatically reconnects when connections are lost, ensuring application database connection stability.

### Features

- **Auto-enabled**: Connection monitoring is enabled by default, no additional configuration required
- **Smart frequency adjustment**: Checks every 60 seconds normally, switches to 10-second rapid retry when connection errors are detected
- **Multi-database support**: Each database is monitored independently, supports multi-database environments
- **Concurrency control**: Uses global lock to ensure only one database performs connection checks at a time, avoiding network congestion
- **Simplified logging**: Only logs when connection status changes, reducing log noise
- **Performance optimized**: Uses lightweight Ping operations with minimal CPU and memory usage

### Configuring Connection Monitoring

Connection monitoring is configured through two fields in the `Config` struct:

```go
type Config struct {
    // ... other fields ...
    
    // Connection monitoring configuration
    MonitorNormalInterval time.Duration // Normal check interval (default 60s, 0 disables monitoring)
    MonitorErrorInterval  time.Duration // Error check interval (default 10s)
}
```

#### Default Configuration (Recommended)

```go
// Use default configuration, monitoring is automatically enabled
err := dbkit.OpenDatabase(dbkit.MySQL, "user:pass@tcp(localhost:3306)/db", 10)

// Or use Config struct (monitoring fields use default values)
config := &dbkit.Config{
    Driver:  dbkit.MySQL,
    DSN:     "user:pass@tcp(localhost:3306)/db",
    MaxOpen: 10,
    // MonitorNormalInterval and MonitorErrorInterval use default values
}
dbkit.OpenDatabaseWithConfig(config)
```

#### Custom Monitoring Intervals

```go
config := &dbkit.Config{
    Driver:                dbkit.MySQL,
    DSN:                   "user:pass@tcp(localhost:3306)/db",
    MaxOpen:               10,
    MonitorNormalInterval: 30 * time.Second, // Custom normal check interval
    MonitorErrorInterval:  5 * time.Second,  // Custom error check interval
}
dbkit.OpenDatabaseWithConfig(config)
```

#### Disable Connection Monitoring

```go
config := &dbkit.Config{
    Driver:                dbkit.MySQL,
    DSN:                   "user:pass@tcp(localhost:3306)/db",
    MaxOpen:               10,
    MonitorNormalInterval: 0, // Set to 0 to disable monitoring
}
dbkit.OpenDatabaseWithConfig(config)
```

### Multi-Database Monitoring

```go
// Configure different monitoring strategies for different databases
config1 := &dbkit.Config{
    Driver:                dbkit.MySQL,
    DSN:                   "user:pass@tcp(localhost:3306)/main_db",
    MaxOpen:               20,
    MonitorNormalInterval: 60 * time.Second,
    MonitorErrorInterval:  10 * time.Second,
}
dbkit.Register("main", config1)

config2 := &dbkit.Config{
    Driver:                dbkit.PostgreSQL,
    DSN:                   "host=localhost port=5432 user=postgres dbname=log_db",
    MaxOpen:               10,
    MonitorNormalInterval: 30 * time.Second, // More frequent checks
    MonitorErrorInterval:  5 * time.Second,
}
dbkit.Register("logs", config2)

config3 := &dbkit.Config{
    Driver:                dbkit.SQLite3,
    DSN:                   "file:cache.db",
    MaxOpen:               5,
    MonitorNormalInterval: 0, // Disable SQLite monitoring
}
dbkit.Register("cache", config3)
```

### How Monitoring Works

1. **Periodic checks**: Monitor uses `database.Ping()` method to periodically check connection status
2. **Frequency adjustment**:
   - When connection is healthy: Uses `MonitorNormalInterval` (default 60s)
   - When connection fails: Switches to `MonitorErrorInterval` (default 10s) for rapid retry
   - After connection recovers: Automatically switches back to normal interval
3. **Concurrency control**: Global lock ensures only one database performs Ping operation at a time
4. **Logging**: Only logs when connection status changes

### Log Output Examples

```go
// Enable debug mode to see monitoring logs
dbkit.SetDebugMode(true)

// Log when connection fails
// [ERROR] Database connection failed {"database": "main", "error": "connection refused", "time": "2024-01-15T10:30:00Z"}

// Log when connection recovers
// [INFO] Database connection recovered {"database": "main", "time": "2024-01-15T10:31:00Z"}
```

### Performance Impact

Connection monitoring has minimal performance impact:

- **CPU usage**: About 6.94 nanoseconds per check, negligible
- **Memory usage**: A few hundred bytes per monitor
- **Network overhead**: Ping operations have minimal network overhead
- **Concurrency control**: Global lock prevents network burst traffic

### Best Practices

1. **Production environment**: Use default configuration (60s/10s) to balance performance and reliability
2. **Development environment**: Can use shorter intervals (30s/5s) for easier testing
3. **High availability environment**: Appropriately shorten intervals to improve response speed
4. **Local databases**: Consider disabling monitoring for databases like SQLite
5. **Multi-database**: Configure different monitoring strategies based on database importance

### Complete Example

```go
package main

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
)

func main() {
	// Enable debug mode to see monitoring logs
	dbkit.SetDebugMode(true)

	// Configure connection monitoring
	config := &dbkit.Config{
		Driver:                dbkit.MySQL,
		DSN:                   "user:pass@tcp(localhost:3306)/test",
		MaxOpen:               10,
		MonitorNormalInterval: 30 * time.Second, // 30s normal check
		MonitorErrorInterval:  5 * time.Second,  // 5s error retry
	}

	err := dbkit.OpenDatabaseWithConfig(config)
	if err != nil {
		fmt.Printf("Database connection failed: %v\n", err)
		return
	}
	defer dbkit.Close()

	fmt.Println("Database connected successfully, monitoring enabled")

	// Application continues running, monitoring works automatically in background
	// Monitor will:
	// 1. Check connection status every 30 seconds
	// 2. If connection issues detected, switch to 5-second retry
	// 3. After connection recovers, switch back to 30-second normal check
	// 4. Only log when status changes

	// Simulate application running
	time.Sleep(2 * time.Minute)
}
```

---

## Query Operations

### Query
```go
func Query(querySQL string, args ...interface{}) ([]Record, error)
func (db *DB) Query(querySQL string, args ...interface{}) ([]Record, error)
func (tx *Tx) Query(querySQL string, args ...interface{}) ([]Record, error)
```
Execute a query and return multiple records.

**Example:**
```go
records, err := dbkit.Query("SELECT * FROM users WHERE age > ?", 18)
```

### QueryFirst
```go
func QueryFirst(querySQL string, args ...interface{}) (*Record, error)
func (db *DB) QueryFirst(querySQL string, args ...interface{}) (*Record, error)
func (tx *Tx) QueryFirst(querySQL string, args ...interface{}) (*Record, error)
```
Execute a query and return the first record. Returns nil if no record found.

### QueryMap
```go
func QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error)
func (db *DB) QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error)
```
Execute a query and return a slice of maps.

### QueryToDbModel
```go
func QueryToDbModel(dest interface{}, querySQL string, args ...interface{}) error
func (db *DB) QueryToDbModel(dest interface{}, querySQL string, args ...interface{}) error
```
Execute a query and map the results to a struct slice.

### QueryFirstToDbModel
```go
func QueryFirstToDbModel(dest interface{}, querySQL string, args ...interface{}) error
func (db *DB) QueryFirstToDbModel(dest interface{}, querySQL string, args ...interface{}) error
```
Execute a query and map the first result to a struct.

### Count
```go
func Count(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Count(table string, whereSql string, whereArgs ...interface{}) (int64, error)
```
Count records matching the conditions.

**Example:**
```go
count, err := dbkit.Count("users", "age > ?", 18)
```

### Exists
```go
func Exists(table string, whereSql string, whereArgs ...interface{}) (bool, error)
func (db *DB) Exists(table string, whereSql string, whereArgs ...interface{}) (bool, error)
```
Check if records matching the conditions exist.

### FindAll
```go
func FindAll(table string) ([]Record, error)
func (db *DB) FindAll(table string) ([]Record, error)
```
Query all records in the table.

### Paginate
```go
func Paginate(page, pageSize int, querySQL string, args ...interface{}) (*Page[Record], error)
func (db *DB) Paginate(page, pageSize int, querySQL string, args ...interface{}) (*Page[Record], error)
```
Pagination query (recommended). Uses complete SQL statement for pagination, automatically parses SQL and generates appropriate pagination statements based on database type.

### PaginateBuilder
```go
func PaginateBuilder(page, pageSize int, selectSql, table, whereSql, orderBySql string, args ...interface{}) (*Page[Record], error)
func (db *DB) PaginateBuilder(page, pageSize int, selectSql, table, whereSql, orderBySql string, args ...interface{}) (*Page[Record], error)
```
Traditional builder-style pagination query. Performs pagination by separately specifying SELECT, table name, WHERE and ORDER BY clauses.

**Parameters:**
- `page`: Page number (starts from 1)
- `pageSize`: Number of records per page
- `selectSql`: SELECT part
- `table`: Table name
- `whereSql`: WHERE condition
- `orderBySql`: ORDER BY part
- `args`: Query parameters

**Returns Page Struct:**
```go
type Page[T any] struct {
    List       []T   // Data list
    PageNumber int   // Current page number
    PageSize   int   // Page size
    TotalPage  int   // Total pages
    TotalRow   int64 // Total records
}
```

---

## Query Timeout Control

DBKit supports global and single query timeout settings using Go's standard `context.Context`.

### Global Timeout Configuration
Set the `QueryTimeout` field in Config:
```go
config := &dbkit.Config{
    Driver:       dbkit.MySQL,
    DSN:          "root:password@tcp(localhost:3306)/test",
    MaxOpen:      10,
    QueryTimeout: 30 * time.Second,  // All queries default to 30s timeout
}
dbkit.OpenDatabaseWithConfig(config)
```

### Timeout (Global Function)
```go
func Timeout(d time.Duration) *DB
```
Returns a DB instance with the specified timeout.

**Example:**
```go
users, err := dbkit.Timeout(5 * time.Second).Query("SELECT * FROM users")
```

### DB.Timeout
```go
func (db *DB) Timeout(d time.Duration) *DB
```
Set query timeout for the DB instance.

**Example:**
```go
users, err := dbkit.Use("default").Timeout(5 * time.Second).Query("SELECT * FROM users")
```

### Tx.Timeout
```go
func (tx *Tx) Timeout(d time.Duration) *Tx
```
Set query timeout for the transaction.

**Example:**
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
Set timeout for chained queries.

**Example:**
```go
users, err := dbkit.Table("users").
    Where("age > ?", 18).
    Timeout(10 * time.Second).
    Find()
```

### Timeout Error Handling
Returns `context.DeadlineExceeded` error upon timeout:
```go
import "context"
import "errors"

users, err := dbkit.Timeout(1 * time.Second).Query("SELECT SLEEP(5)")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Query timeout")
    }
}
```

---

## Insert and Update

### Exec
```go
func Exec(querySQL string, args ...interface{}) (sql.Result, error)
func (db *DB) Exec(querySQL string, args ...interface{}) (sql.Result, error)
func (tx *Tx) Exec(querySQL string, args ...interface{}) (sql.Result, error)
```
Execute SQL statements (INSERT, UPDATE, DELETE, etc.).

### Save
```go
func Save(table string, record *Record) (int64, error)
func (db *DB) Save(table string, record *Record) (int64, error)
func (tx *Tx) Save(table string, record *Record) (int64, error)
```
Smart save record. Updates if primary key exists and record exists, otherwise inserts.

**Returns:** New ID for insert, rows affected for update.

### Insert
```go
func Insert(table string, record *Record) (int64, error)
func (db *DB) Insert(table string, record *Record) (int64, error)
func (tx *Tx) Insert(table string, record *Record) (int64, error)
```
Force insert a new record.

**Returns:** ID of the newly inserted record.

### Update
```go
func Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
func (tx *Tx) Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
```
Update records based on conditions.

**Returns:** Number of rows affected.

**Note:** DBKit disables timestamp auto-update, optimistic lock, and soft delete by default for optimal performance. To enable these features, use `EnableTimestamps()`, `EnableOptimisticLock()`, and `EnableSoftDelete()` respectively.

### UpdateFast
```go
func UpdateFast(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) UpdateFast(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
```
Lightweight update that always skips timestamps and optimistic lock checks, providing best performance.

**Returns:** Number of rows affected.

**Use Cases:**

1. **High-frequency update scenarios**: High concurrency updates requiring extreme performance
   ```go
   // Game server updating player score
   record := dbkit.NewRecord().Set("score", newScore)
   dbkit.UpdateFast("players", record, "id = ?", playerId)
   ```

2. **Batch updates**: Reducing overhead during mass data updates
   ```go
   // Batch update product stock
   for _, item := range items {
       record := dbkit.NewRecord().Set("stock", item.Stock)
       dbkit.UpdateFast("products", record, "id = ?", item.ID)
   }
   ```

3. Tables that do not need timestamp or optimistic lock features
   
   ```go
   // Update config table (no timestamp needed)
   record := dbkit.NewRecord().Set("value", "new_value")
   dbkit.UpdateFast("config", record, "key = ?", "app_version")
   ```
   
4. **Features enabled but need to be skipped for specific operations**: 
   
   ```go
   
   dbkit.EnableTimestamp()
   
   // But skip for some high-frequency operations
   record := dbkit.NewRecord().Set("view_count", viewCount)
   dbkit.UpdateFast("articles", record, "id = ?", articleId)
   ```

**Performance Comparison:**
- When timestamps, soft delete, optimistic lock, etc. are disabled, `Update` and `UpdateFast` have the same performance.
- When these features are enabled, `UpdateFast` is about 2-3 times faster than `Update`.

**Notes:**

- `UpdateFast` will NOT automatically update `updated_at` field.
- `UpdateFast` will NOT perform optimistic lock version checks.
- If you need these features, use `Update` and enable the corresponding feature checks.

### UpdateRecord
```go
func (db *DB) UpdateRecord(table string, record *Record) (int64, error)
func (tx *Tx) UpdateRecord(table string, record *Record) (int64, error)
```
Update record based on the primary key in Record.

### BatchInsert
```go
func BatchInsert(table string, records []*Record, batchSize int) (int64, error)
func (db *DB) BatchInsert(table string, records []*Record, batchSize int) (int64, error)
```
Batch insert records.

**Parameters:**
- `batchSize`: Number of records per batch

### BatchInsertDefault
```go
func BatchInsertDefault(table string, records []*Record) (int64, error)
func (db *DB) BatchInsertDefault(table string, records []*Record) (int64, error)
```
Batch insert records, default batch size is 100.

---

## Delete Operations

### Delete
```go
func Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (tx *Tx) Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
```
Delete records based on conditions. If soft delete is configured for the table, performs soft delete (updates the deletion marker field).

### DeleteRecord
```go
func DeleteRecord(table string, record *Record) (int64, error)
func (db *DB) DeleteRecord(table string, record *Record) (int64, error)
func (tx *Tx) DeleteRecord(table string, record *Record) (int64, error)
```
Delete record based on the primary key in Record.

---

## Soft Delete

Soft delete marks records as deleted instead of physically removing them, facilitating data recovery and auditing.

**Note**: DBKit disables soft delete by default for best performance. To use this feature, enable it first:

```go
// Enable soft delete feature
dbkit.EnableSoftDelete()
```

### EnableSoftDelete
```go
func EnableSoftDelete()
func (db *DB) EnableSoftDelete() *DB
```
Enable soft delete feature. Once enabled, query operations will automatically filter out soft-deleted records.

**Example:**
```go
// Global enable soft delete
dbkit.EnableSoftDelete()

// Multi-database mode
dbkit.Use("main").EnableSoftDelete()
```

### Soft Delete Types
```go
const (
    SoftDeleteTimestamp SoftDeleteType = iota  // Timestamp style (deleted_at)
    SoftDeleteBool                              // Boolean style (is_deleted)
)
```

### ConfigSoftDelete
```go
func ConfigSoftDelete(table, field string)
func (db *DB) ConfigSoftDelete(table, field string) *DB
```
Configure soft delete for a table (timestamp type).

**Parameters:**
- `table`: Table name
- `field`: Soft delete field name (e.g., "deleted_at")

**Example:**
```go
// Configure soft delete
dbkit.ConfigSoftDelete("users", "deleted_at")

// Multi-database mode
dbkit.Use("main").ConfigSoftDelete("users", "deleted_at")
```

### ConfigSoftDeleteWithType
```go
func ConfigSoftDeleteWithType(table, field string, deleteType SoftDeleteType)
func (db *DB) ConfigSoftDeleteWithType(table, field string, deleteType SoftDeleteType) *DB
```
Configure soft delete for a table (specified type).

**Example:**
```go
// Use boolean type
dbkit.ConfigSoftDeleteWithType("posts", "is_deleted", dbkit.SoftDeleteBool)
```

### RemoveSoftDelete
```go
func RemoveSoftDelete(table string)
func (db *DB) RemoveSoftDelete(table string) *DB
```
Remove soft delete configuration for a table.

### HasSoftDelete
```go
func HasSoftDelete(table string) bool
func (db *DB) HasSoftDelete(table string) bool
```
Check if soft delete is enabled for a table.

### WithTrashed
```go
func (qb *QueryBuilder) WithTrashed() *QueryBuilder
```
Include deleted records in the query.

**Example:**
```go
// Query all users (including deleted ones)
users, err := dbkit.Table("users").WithTrashed().Find()
```

### OnlyTrashed
```go
func (qb *QueryBuilder) OnlyTrashed() *QueryBuilder
```
Query only deleted records.

**Example:**
```go
// Query only deleted users
deletedUsers, err := dbkit.Table("users").OnlyTrashed().Find()
```

### ForceDelete
```go
func ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (tx *Tx) ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (qb *QueryBuilder) ForceDelete() (int64, error)
```
Physically delete records, bypassing soft delete configuration.

**Example:**
```go
// Physical delete
dbkit.ForceDelete("users", "id = ?", 1)

// Chain call
dbkit.Table("users").Where("id = ?", 1).ForceDelete()
```

### Restore
```go
func Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (tx *Tx) Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (qb *QueryBuilder) Restore() (int64, error)
```
Restore soft-deleted records.

**Example:**
```go
// Restore record
dbkit.Restore("users", "id = ?", 1)

// Chain call
dbkit.Table("users").Where("id = ?", 1).Restore()
```

### QueryWithOutTrashed
```go
func QueryWithOutTrashed(querySQL string, args ...interface{}) ([]Record, error)
func (db *DB) QueryWithOutTrashed(querySQL string, args ...interface{}) ([]Record, error)
func (tx *Tx) QueryWithOutTrashed(querySQL string, args ...interface{}) ([]Record, error)
```
Execute raw SQL queries with automatic soft delete filtering. Automatically detects tables involved in SQL and adds filtering conditions for tables configured with soft delete.

**Parameters:**
- `querySQL`: Raw SQL query statement
- `args`: SQL parameters

**Returns:**
- `[]Record`: List of query result records
- `error`: Error information

**Example:**
```go
// Single table query with automatic soft delete filtering
users, err := dbkit.QueryWithOutTrashed("SELECT * FROM users WHERE age > ?", 18)
// Original SQL: SELECT * FROM users WHERE age > ?
// Auto-converted: SELECT * FROM users WHERE age > ? AND deleted_at IS NULL

// Multi-table JOIN query
posts, err := dbkit.QueryWithOutTrashed(`
    SELECT p.*, u.name as author_name 
    FROM posts p 
    JOIN users u ON p.user_id = u.id 
    WHERE p.status = ?
`, "published")

// Multi-database mode
users, err := dbkit.Use("main").QueryWithOutTrashed("SELECT * FROM users")

// Use in transactions
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    users, err := tx.QueryWithOutTrashed("SELECT * FROM users WHERE active = ?", true)
    return err
})
```

### QueryFirstWithOutTrashed
```go
func QueryFirstWithOutTrashed(querySQL string, args ...interface{}) (*Record, error)
func (db *DB) QueryFirstWithOutTrashed(querySQL string, args ...interface{}) (*Record, error)
func (tx *Tx) QueryFirstWithOutTrashed(querySQL string, args ...interface{}) (*Record, error)
```
Execute raw SQL query and return the first non-soft-deleted record. Based on `QueryWithOutTrashed` implementation.

**Parameters:**
- `querySQL`: Raw SQL query statement
- `args`: SQL parameters

**Returns:**
- `*Record`: First record (returns nil if no records found)
- `error`: Error information

**Example:**
```go
// Query first record
user, err := dbkit.QueryFirstWithOutTrashed("SELECT * FROM users WHERE email = ?", "test@example.com")
if user != nil {
    fmt.Printf("User: %s\n", user.Str("name"))
}

// Multi-database mode
user, err := dbkit.Use("main").QueryFirstWithOutTrashed("SELECT * FROM users ORDER BY id DESC")

// Use in transactions
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    user, err := tx.QueryFirstWithOutTrashed("SELECT * FROM users WHERE id = ?", 1)
    if user == nil {
        return fmt.Errorf("user not found")
    }
    return nil
})
```

### Complete Soft Delete Example
```go
// 1. Configure soft delete
dbkit.ConfigSoftDelete("users", "deleted_at")

// 2. Insert data
record := dbkit.NewRecord()
record.Set("name", "John")
dbkit.Insert("users", record)

// 3. Soft delete (automatically updates deleted_at field)
dbkit.Delete("users", "id = ?", 1)

// 4. Normal query (automatically filters deleted records)
users, _ := dbkit.Table("users").Find()  // Does not include deleted

// 5. Query including deleted records
allUsers, _ := dbkit.Table("users").WithTrashed().Find()

// 6. Query only deleted records
deletedUsers, _ := dbkit.Table("users").OnlyTrashed().Find()

// 7. Restore deleted record
dbkit.Restore("users", "id = ?", 1)

// 8. Physical delete (truly delete data)
dbkit.ForceDelete("users", "id = ?", 1)
```

### DbModel Soft Delete Methods

Generated DbModels automatically include soft delete related methods:

```go
// Soft delete (if soft delete is configured)
user.Delete()

// Physical delete
user.ForceDelete()

// Restore
user.Restore()

// Query including deleted
users, _ := user.FindWithTrashed("status = ?", "id DESC", "active")

// Query only deleted
deletedUsers, _ := user.FindOnlyTrashed("", "id DESC")
```

---

## Automatic Timestamps

Automatic timestamps feature allows automatically populating timestamp fields during record insertion and update, without manual setting.

**Note:** DBKit disables automatic timestamps by default for best performance. To enable, use `EnableTimestamps()`.

### EnableTimestamps
```go
func EnableTimestamps()
func (db *DB) EnableTimestamps() *DB
```
Enable automatic timestamps feature. Once enabled, Update operations will check table timestamp configuration and automatically update `updated_at` field.

**Example:**
```go
// Global enable timestamp auto-update
dbkit.EnableTimestamps()

// Multi-database mode
dbkit.Use("main").EnableTimestamps()
```

### ConfigTimestamps
```go
func ConfigTimestamps(table string)
func (db *DB) ConfigTimestamps(table string) *DB
```
Configure automatic timestamps for a table, using default field names `created_at` and `updated_at`.

**Example:**
```go
// Configure automatic timestamps
dbkit.ConfigTimestamps("users")

// Multi-database mode
dbkit.Use("main").ConfigTimestamps("users")
```

### ConfigTimestampsWithFields
```go
func ConfigTimestampsWithFields(table, createdAtField, updatedAtField string)
func (db *DB) ConfigTimestampsWithFields(table, createdAtField, updatedAtField string) *DB
```
Configure automatic timestamps for a table using custom field names.

**Parameters:**
- `table`: Table name
- `createdAtField`: Create time field name (e.g., "create_time")
- `updatedAtField`: Update time field name (e.g., "update_time")

**Example:**
```go
// Use custom field names
dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")
```

### ConfigCreatedAt
```go
func ConfigCreatedAt(table, field string)
func (db *DB) ConfigCreatedAt(table, field string) *DB
```
Configure only `created_at` field.

**Example:**
```go
// Only configure create time (suitable for logs tables that only need to record creation time)
dbkit.ConfigCreatedAt("logs", "log_time")
```

### ConfigUpdatedAt
```go
func ConfigUpdatedAt(table, field string)
func (db *DB) ConfigUpdatedAt(table, field string) *DB
```
Configure only `updated_at` field.

**Example:**
```go
// Only configure update time
dbkit.ConfigUpdatedAt("cache_data", "last_modified")
```

### RemoveTimestamps
```go
func RemoveTimestamps(table string)
func (db *DB) RemoveTimestamps(table string) *DB
```
Remove timestamp configuration for a table.

### HasTimestamps
```go
func HasTimestamps(table string) bool
func (db *DB) HasTimestamps(table string) bool
```
Check if a table has automatic timestamps configured.

### WithoutTimestamps
```go
func (qb *QueryBuilder) WithoutTimestamps() *QueryBuilder
```
Temporarily disable automatic timestamps (for QueryBuilder Update operations).

**Example:**
```go
// Do not auto-fill updated_at during update
dbkit.Table("users").Where("id = ?", 1).WithoutTimestamps().Update(record)
```

### Automatic Timestamp Behavior Explanation

- **Insert Operation**: If `created_at` field is not set, automatically populate with current time.
- **Update Operation**: Always automatically populate `updated_at` field with current time.
- **Manual Setting Priority**: If `created_at` is already set in the Record, it will not be overwritten.

### Complete Automatic Timestamp Example
```go
// 1. Configure automatic timestamps
dbkit.ConfigTimestamps("users")

// 2. Insert data (created_at auto-filled)
record := dbkit.NewRecord()
record.Set("name", "John")
record.Set("email", "john@example.com")
dbkit.Insert("users", record)
// created_at automatically set to current time

// 3. Update data (updated_at auto-filled)
updateRecord := dbkit.NewRecord()
updateRecord.Set("name", "John Updated")
dbkit.Update("users", updateRecord, "id = ?", 1)
// updated_at automatically set to current time

// 4. Manually specify created_at during insert (will not be overwritten)
customTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
record2 := dbkit.NewRecord()
record2.Set("name", "Jane")
record2.Set("created_at", customTime)
dbkit.Insert("users", record2)
// created_at remains 2020-01-01

// 5. Temporarily disable automatic timestamps
dbkit.Table("users").Where("id = ?", 1).WithoutTimestamps().Update(record)
// updated_at will not be automatically updated

// 6. Use custom field names
dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")

// 7. Only configure created_at (suitable for logging)
dbkit.ConfigCreatedAt("logs", "log_time")
```

### Use with Soft Delete

Automatic timestamps and soft delete features are independent and can be used together:

```go
// Configure both soft delete and automatic timestamps
dbkit.ConfigTimestamps("users")
dbkit.ConfigSoftDelete("users", "deleted_at")

// During soft delete, updated_at will also be updated
dbkit.Delete("users", "id = ?", 1)
// deleted_at set to current time, updated_at also updated
```

---

## Optimistic Lock

Optimistic lock is a concurrency control mechanism that uses a version number field to detect concurrent update conflicts and prevent data from being accidentally overwritten.

**Note:** DBKit disables optimistic locking by default for best performance. To enable, use `EnableOptimisticLock()`.

### EnableOptimisticLock
```go
func EnableOptimisticLock()
func (db *DB) EnableOptimisticLock() *DB
```
Enable optimistic lock feature. Once enabled, Update operations will check the table's optimistic lock configuration and automatically perform version checks.

**Example:**
```go
// Global enable optimistic lock
dbkit.EnableOptimisticLock()

// Multi-database mode
dbkit.Use("main").EnableOptimisticLock()
```

### How It Works

1. **Insert**: Automatically initialize version field to 1.
2. **Update**: Automatically add version check to WHERE condition and increment version number in SET.
3. **Conflict Detection**: If update affects 0 rows (version mismatch), return `ErrVersionMismatch` error.

### ErrVersionMismatch
```go
var ErrVersionMismatch = fmt.Errorf("dbkit: optimistic lock conflict - record was modified by another transaction")
```
Error returned when version conflict occurs.

### ConfigOptimisticLock
```go
func ConfigOptimisticLock(table string)
func (db *DB) ConfigOptimisticLock(table string) *DB
```
Configure optimistic lock for a table, using default field name `version`.

**Example:**
```go
// Configure optimistic lock
dbkit.ConfigOptimisticLock("products")

// Multi-database mode
dbkit.Use("main").ConfigOptimisticLock("products")
```

### ConfigOptimisticLockWithField
```go
func ConfigOptimisticLockWithField(table, versionField string)
func (db *DB) ConfigOptimisticLockWithField(table, versionField string) *DB
```
Configure optimistic lock for a table using a custom version field name.

**Example:**
```go
// Use custom field name
dbkit.ConfigOptimisticLockWithField("orders", "revision")
```

### RemoveOptimisticLock
```go
func RemoveOptimisticLock(table string)
func (db *DB) RemoveOptimisticLock(table string) *DB
```
Remove optimistic lock configuration for a table.

### HasOptimisticLock
```go
func HasOptimisticLock(table string) bool
func (db *DB) HasOptimisticLock(table string) bool
```
Check if a table has optimistic lock configured.

### Version Field Handling Rules

| Version Field Value | Behavior |
|-------------------|----------|
| Not present | Skip version check, normal update |
| `nil` / `NULL` | Skip version check, normal update |
| `""` (Empty string) | Skip version check, normal update |
| `0`, `1`, `2`, ... | Perform version check |
| `"123"` (Numeric string) | Perform version check (parsed as number) |

### Complete Optimistic Lock Example

```go
// 1. Configure optimistic lock
dbkit.ConfigOptimisticLock("products")

// 2. Insert data (version auto-intialized to 1)
record := dbkit.NewRecord()
record.Set("name", "Laptop")
record.Set("price", 999.99)
dbkit.Insert("products", record)
// version automatically set to 1

// 3. Normal update (with version number)
updateRecord := dbkit.NewRecord()
updateRecord.Set("version", int64(1))  // Current version
updateRecord.Set("price", 899.99)
rows, err := dbkit.Update("products", updateRecord, "id = ?", 1)
// Success: version auto-incremented to 2

// 4. Concurrent conflict detection (using stale version)
staleRecord := dbkit.NewRecord()
staleRecord.Set("version", int64(1))  // Stale version!
staleRecord.Set("price", 799.99)
rows, err = dbkit.Update("products", staleRecord, "id = ?", 1)
if errors.Is(err, dbkit.ErrVersionMismatch) {
    fmt.Println("Concurrent conflict detected, record modified by another transaction")
}

// 5. Correctly handle concurrency: read latest version first
latestRecord, _ := dbkit.Table("products").Where("id = ?", 1).FindFirst()
currentVersion := latestRecord.GetInt("version")

updateRecord2 := dbkit.NewRecord()
updateRecord2.Set("version", currentVersion)
updateRecord2.Set("price", 799.99)
dbkit.Update("products", updateRecord2, "id = ?", 1)

// 6. Update without version field (skip version check)
noVersionRecord := dbkit.NewRecord()
noVersionRecord.Set("stock", 90)  // No version set
dbkit.Update("products", noVersionRecord, "id = ?", 1)
// Normal update, no version check

// 7. Use UpdateRecord (auto-extract version from record)
product, _ := dbkit.Table("products").Where("id = ?", 1).FindFirst()
product.Set("name", "Gaming Laptop")
dbkit.Use("default").UpdateRecord("products", product)
// version is in product, auto perform version check

// 8. Use optimistic lock in transaction
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

### Use with Other Features

Optimistic lock can be used simultaneously with automatic timestamps and soft delete:

```go
// Configure multiple features simultaneously
dbkit.ConfigOptimisticLock("products")
dbkit.ConfigTimestamps("products")
dbkit.ConfigSoftDelete("products", "deleted_at")

// Insert: version=1, created_at=now
// Update: version++, updated_at=now
// Delete: deleted_at=now, updated_at=now
```

### IOptimisticLockModel Interface

```go
type IOptimisticLockModel interface {
    IDbModel
    VersionField() string  // Return version field name, empty string means not used
}
```

Generated DbModels can implement this interface to automatically configure optimistic locking.

---

## Transaction Processing

### Transaction
```go
func Transaction(fn func(*Tx) error) error
func (db *DB) Transaction(fn func(*Tx) error) error
```
Automatic transaction processing. Automatically rolls back if the closure returns an error, otherwise automatically commits.

**Example:**
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    _, err := tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE id = ?", 1)
    if err != nil {
        return err // Auto rollback
    }
    _, err = tx.Exec("UPDATE accounts SET balance = balance + 100 WHERE id = ?", 2)
    return err
})
```

### BeginTransaction
```go
func BeginTransaction() (*Tx, error)
```
Start a manual transaction.

### Tx.Commit
```go
func (tx *Tx) Commit() error
```
Commit transaction.

### Tx.Rollback
```go
func (tx *Tx) Rollback() error
```
Rollback transaction.

---

## Record Object

### NewRecord
```go
func NewRecord() *Record
```
Create a new empty Record object.

### Record.Set
```go
func (r *Record) Set(column string, value interface{}) *Record
```
Set field value, supports chaining.

### Record.Get
```go
func (r *Record) Get(column string) interface{}
```
Get field value.

### Type-Safe Get Methods
```go
func (r *Record) GetString(column string) string
func (r *Record) GetInt(column string) int
func (r *Record) GetInt64(column string) int64
func (r *Record) GetFloat(column string) float64
func (r *Record) GetBool(column string) bool
func (r *Record) GetTime(column string) time.Time

// Short methods
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
Check if field exists.

### Record.Keys
```go
func (r *Record) Keys() []string
```
Get all field names.

### Record.Remove
```go
func (r *Record) Remove(column string)
```
Remove a field.

### Record.Clear
```go
func (r *Record) Clear()
```
Clear all fields.

### Record.ToMap
```go
func (r *Record) ToMap() map[string]interface{}
```
Convert to map.

### Record.ToJson
```go
func (r *Record) ToJson() string
```
Convert to JSON string.

### Record.FromJson
```go
func (r *Record) FromJson(jsonStr string) error
```
Parse from JSON string.

### Record.ToStruct
```go
func (r *Record) ToStruct(dest interface{}) error
```
Convert to struct.

### Record.FromStruct
```go
func (r *Record) FromStruct(src interface{}) error
```
Populate from struct.

---

## Chained Query

### Table
```go
func Table(name string) *QueryBuilder
func (db *DB) Table(name string) *QueryBuilder
func (tx *Tx) Table(name string) *QueryBuilder
```
Start a chained query, specifying the table name.

### QueryBuilder Methods

```go
func (b *QueryBuilder) Select(columns string) *QueryBuilder    // Specify query columns
func (b *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder  // WHERE condition
func (b *QueryBuilder) And(condition string, args ...interface{}) *QueryBuilder    // AND condition
func (b *QueryBuilder) OrderBy(orderBy string) *QueryBuilder   // Sort
func (b *QueryBuilder) Limit(limit int) *QueryBuilder          // Limit quantity
func (b *QueryBuilder) Offset(offset int) *QueryBuilder        // Offset

// Execution Methods
func (b *QueryBuilder) Find() ([]Record, error)                // Query multiple
func (b *QueryBuilder) Query() ([]Record, error)               // Alias for Find
func (b *QueryBuilder) FindFirst() (*Record, error)            // Query first
func (b *QueryBuilder) QueryFirst() (*Record, error)           // Alias for FindFirst
func (b *QueryBuilder) FindToDbModel(dest interface{}) error   // Query and map to struct slice
func (b *QueryBuilder) FindFirstToDbModel(dest interface{}) error // Query first and map to struct
func (b *QueryBuilder) Delete() (int64, error)                 // Delete
func (b *QueryBuilder) Paginate(page, pageSize int) (*Page[Record], error) // Pagination
```

**Example:**
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

### Join Queries

Supports chaining various JOIN types:

```go
func (b *QueryBuilder) Join(table, condition string, args ...interface{}) *QueryBuilder      // JOIN
func (b *QueryBuilder) LeftJoin(table, condition string, args ...interface{}) *QueryBuilder  // LEFT JOIN
func (b *QueryBuilder) RightJoin(table, condition string, args ...interface{}) *QueryBuilder // RIGHT JOIN
func (b *QueryBuilder) InnerJoin(table, condition string, args ...interface{}) *QueryBuilder // INNER JOIN
```

**Example:**
```go
// Simple LEFT JOIN
records, err := dbkit.Table("users").
    Select("users.name, orders.total").
    LeftJoin("orders", "users.id = orders.user_id").
    Where("orders.status = ?", "completed").
    Find()
// SQL: SELECT users.name, orders.total FROM users LEFT JOIN orders ON users.id = orders.user_id WHERE orders.status = ?
// Args: ["completed"]

// Multi-table INNER JOIN
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

// JOIN condition with parameters
records, err := dbkit.Table("users").
    Join("orders", "users.id = orders.user_id AND orders.status = ?", "active").
    Find()
// SQL: SELECT * FROM users JOIN orders ON users.id = orders.user_id AND orders.status = ?
// Args: ["active"]
```

### Subquery

#### NewSubquery
```go
func NewSubquery() *Subquery
```
Create a new subquery builder.

#### Subquery Methods
```go
func (s *Subquery) Table(name string) *Subquery                           // Set table name
func (s *Subquery) Select(columns string) *Subquery                       // Set select columns
func (s *Subquery) Where(condition string, args ...interface{}) *Subquery // Add condition
func (s *Subquery) OrderBy(orderBy string) *Subquery                      // Sort
func (s *Subquery) Limit(limit int) *Subquery                             // Limit quantity
func (s *Subquery) ToSQL() (string, []interface{})                        // Generate SQL
```

#### WHERE IN Subquery
```go
func (b *QueryBuilder) WhereIn(column string, sub *Subquery) *QueryBuilder    // WHERE column IN (subquery)
func (b *QueryBuilder) WhereNotIn(column string, sub *Subquery) *QueryBuilder // WHERE column NOT IN (subquery)
```

**Example:**
```go
// Query users who have completed orders
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

// Query orders of users who are not banned
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

#### FROM Subquery
```go
func (b *QueryBuilder) TableSubquery(sub *Subquery, alias string) *QueryBuilder
```
Use subquery as FROM data source (derived table).

**Example:**
```go
// Query from aggregate subquery
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

#### SELECT Subquery
```go
func (b *QueryBuilder) SelectSubquery(sub *Subquery, alias string) *QueryBuilder
```
Add subquery as a column in SELECT clause.

**Example:**
```go
// Add order count field for each user
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

### Advanced WHERE Conditions

#### OrWhere
```go
func (b *QueryBuilder) OrWhere(condition string, args ...interface{}) *QueryBuilder
```
Add OR condition to query. When used with Where, AND conditions will be wrapped in parentheses to maintain correct precedence.

**Example:**
```go
// Query orders with active status or high priority
orders, err := dbkit.Table("orders").
    Where("status = ?", "active").
    OrWhere("priority = ?", "high").
    Find()
// SQL: SELECT * FROM orders WHERE (status = ?) OR priority = ?
// Args: ["active", "high"]

// Multiple OR conditions
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
Add grouped conditions, supporting nested parentheses. `WhereGroup` uses AND connection, `OrWhereGroup` uses OR connection.

**Example:**
```go
// OR group condition
records, err := dbkit.Table("table").
    Where("a = ?", 1).
    OrWhereGroup(func(qb *dbkit.QueryBuilder) *dbkit.QueryBuilder {
        return qb.Where("b = ?", 1).OrWhere("c = ?", 1)
    }).
    Find()
// SQL: SELECT * FROM table WHERE (a = ?) OR (b = ? OR c = ?)
// Args: [1, 1, 1]

// AND group condition
records, err := dbkit.Table("orders").
    Where("status = ?", "active").
    WhereGroup(func(qb *dbkit.QueryBuilder) *dbkit.QueryBuilder {
        return qb.Where("type = ?", "A").OrWhere("priority = ?", "high")
    }).
    Find()
// SQL: SELECT * FROM orders WHERE status = ? AND (type = ? OR priority = ?)
// Args: ["active", "A", "high"]

// Complex nesting
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
Use value list for IN/NOT IN query (distinguished from subquery version WhereIn/WhereNotIn).

**Example:**
```go
// Query users with specific IDs
users, err := dbkit.Table("users").
    WhereInValues("id", []interface{}{1, 2, 3, 4, 5}).
    Find()
// SQL: SELECT * FROM users WHERE id IN (?, ?, ?, ?, ?)
// Args: [1, 2, 3, 4, 5]

// Exclude orders with specific statuses
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
Range query.

**Example:**
```go
// Query users aged between 18 and 65
users, err := dbkit.Table("users").
    WhereBetween("age", 18, 65).
    Find()
// SQL: SELECT * FROM users WHERE age BETWEEN ? AND ?
// Args: [18, 65]

// Query products with price not between 100 and 500
products, err := dbkit.Table("products").
    WhereNotBetween("price", 100, 500).
    Find()
// SQL: SELECT * FROM products WHERE price NOT BETWEEN ? AND ?
// Args: [100, 500]

// Date range query
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
NULL value check.

**Example:**
```go
// Query users without email
users, err := dbkit.Table("users").
    WhereNull("email").
    Find()
// SQL: SELECT * FROM users WHERE email IS NULL
// Args: []

// Query users with phone number
users, err := dbkit.Table("users").
    WhereNotNull("phone").
    Find()
// SQL: SELECT * FROM users WHERE phone IS NOT NULL
// Args: []
```

### Grouping and Aggregation

#### GroupBy
```go
func (b *QueryBuilder) GroupBy(columns string) *QueryBuilder
```
Add GROUP BY clause.

#### Having
```go
func (b *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder
```
Add HAVING clause to filter grouped results.

**Example:**
```go
// Group orders by status and count
stats, err := dbkit.Table("orders").
    Select("status, COUNT(*) as count, SUM(total) as total_amount").
    GroupBy("status").
    Find()
// SQL: SELECT status, COUNT(*) as count, SUM(total) as total_amount FROM orders GROUP BY status
// Args: []

// Query users with more than 5 orders
users, err := dbkit.Table("orders").
    Select("user_id, COUNT(*) as order_count").
    GroupBy("user_id").
    Having("COUNT(*) > ?", 5).
    Find()
// SQL: SELECT user_id, COUNT(*) as order_count FROM orders GROUP BY user_id HAVING COUNT(*) > ?
// Args: [5]

// Multiple HAVING conditions
stats, err := dbkit.Table("orders").
    Select("user_id, COUNT(*) as cnt, SUM(total) as total").
    GroupBy("user_id").
    Having("COUNT(*) > ?", 3).
    Having("SUM(total) > ?", 1000).
    Find()
// SQL: SELECT user_id, COUNT(*) as cnt, SUM(total) as total FROM orders GROUP BY user_id HAVING COUNT(*) > ? AND SUM(total) > ?
// Args: [3, 1000]
```

### Complex Query Example

```go
// Complex query combining multiple conditions
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

## DbModel Operations

### GenerateDbModel
```go
func GenerateDbModel(tablename, outPath, structName string) error
func (db *DB) GenerateDbModel(tablename, outPath, structName string) error
```
Generate Go struct code from database table.

**Parameters:**
- `tablename`: Table name
- `outPath`: Output path (directory or full file path)
- `structName`: Struct name (auto-generated if empty)

### IDbModel Interface
```go
type IDbModel interface {
    TableName() string
    DatabaseName() string
}
```

### DbModel CRUD Functions
```go
func SaveDbModel(model IDbModel) (int64, error)
func InsertDbModel(model IDbModel) (int64, error)
func UpdateDbModel(model IDbModel) (int64, error)
func DeleteDbModel(model IDbModel) (int64, error)
func FindFirstToDbModel(model IDbModel, whereSql string, whereArgs ...interface{}) error
func FindToDbModel(dest interface{}, table, whereSql, orderBySql string, whereArgs ...interface{}) error
```

### Generic Helper Functions
```go
func FindModel[T IDbModel](model T, cache *ModelCache, whereSql, orderBySql string, whereArgs ...interface{}) ([]T, error)
func FindFirstModel[T IDbModel](model T, cache *ModelCache, whereSql string, whereArgs ...interface{}) (T, error)
func PaginateModel[T IDbModel](model T, cache *ModelCache, page, pageSize int, whereSql, orderBySql string, whereArgs ...interface{}) (*Page[T], error)
func PaginateModel_FullSql[T IDbModel](model T, cache *ModelCache, page, pageSize int, querySQL string, whereArgs ...interface{}) (*Page[T], error)
```

### ModelCache Methods
```go
func (c *ModelCache) SetCache(cacheRepositoryName string, ttl ...time.Duration)  // Set cache name and TTL
func (c *ModelCache) WithCountCache(ttl time.Duration) *ModelCache               // Enable pagination count cache
func (c *ModelCache) GetCache() *ModelCache                                      // Get cache configuration
```

**Example:**
```go
// Create user model and use method chaining to set cache
user := &User{}

// Method 1: Using method chaining (recommended)
page, err := user.Cache("user_cache", 5*time.Minute).
    WithCountCache(5*time.Minute).
    PaginateBuilder(1, 10, "age > ?", "name ASC", 18)

// Method 2: Using PaginateModel function
// user.SetCache("user_cache", 5*time.Minute)
// user.WithCountCache(5*time.Minute)
// page, err := dbkit.PaginateModel(user, user.GetCache(), 1, 10, 
//     "age > ?", "name ASC", 18)
```

---

## Cache Operations

DBKit provides flexible caching strategies, supporting both local cache and Redis cache.

### Cache Initialization

#### InitLocalCache
```go
func InitLocalCache(cleanupInterval time.Duration)
```
Initialize local cache instance with cleanup interval.

**Example:**
```go
dbkit.InitLocalCache(1 * time.Minute)
```

#### InitRedisCache
```go
func InitRedisCache(provider CacheProvider)
```
Initialize Redis cache instance.

**Example:**
```go
import "github.com/zzguang83325/dbkit/redis"

rc, err := redis.NewRedisCache("localhost:6379", "", "password", 0)
if err != nil {
    panic(err)
}
dbkit.InitRedisCache(rc)
```

#### SetDefaultCache
```go
func SetDefaultCache(c CacheProvider)
```
Set default cache provider.

**Example:**
```go
dbkit.SetDefaultCache(rc) // Switch default cache to Redis
```

#### GetCache
```go
func GetCache() CacheProvider
```
Get current default cache provider.

#### GetLocalCacheInstance
```go
func GetLocalCacheInstance() CacheProvider
```
Get local cache instance.

#### GetRedisCacheInstance
```go
func GetRedisCacheInstance() CacheProvider
```
Get Redis cache instance.

### Cache Query (Chain Call)

#### Cache
```go
func Cache(name string, ttl ...time.Duration) *DB
func (db *DB) Cache(name string, ttl ...time.Duration) *DB
func (tx *Tx) Cache(name string, ttl ...time.Duration) *Tx
```
Create query builder using default cache.

**Example:**
```go
records, err := dbkit.Cache("user_cache", 5*time.Minute).Query("SELECT * FROM users")
```

#### LocalCache
```go
func LocalCache(cacheRepositoryName string, ttl ...time.Duration) *DB
```
Create query builder using local cache.

**Example:**
```go
user, _ := dbkit.LocalCache("user_cache").QueryFirst("SELECT * FROM users WHERE id = ?", 1)
```

#### RedisCache
```go
func RedisCache(cacheRepositoryName string, ttl ...time.Duration) *DB
```
Create query builder using Redis cache.

**Example:**
```go
orders, _ := dbkit.RedisCache("order_cache").Query("SELECT * FROM orders WHERE user_id = ?", userId)
```

#### WithCountCache
```go
func (db *DB) WithCountCache(ttl time.Duration) *DB
func (tx *Tx) WithCountCache(ttl time.Duration) *Tx
func (qb *QueryBuilder) WithCountCache(ttl time.Duration) *QueryBuilder
func (b *SqlTemplateBuilder) WithCountCache(ttl time.Duration) *SqlTemplateBuilder
```
Enable pagination count cache to avoid repeated COUNT queries.

**Example:**
```go
// SQL Template with count cache
page, err := dbkit.SqlTemplate("getUserList", params).
    Cache("sql_cache", 5*time.Minute).
    WithCountCache(5*time.Minute).
    Paginate(1, 10)

// Table query with count cache
page, err := dbkit.Table("users").
    Where("age > ?", 30).
    Cache("table_cache", 5*time.Minute).
    WithCountCache(5*time.Minute).
    Paginate(1, 10)

// DbModel with count cache
user := &User{}
page, err := user.Cache("user_cache", 5*time.Minute).
    WithCountCache(5*time.Minute).
    PaginateBuilder(1, 10, "age > ?", "name ASC", 18)
```

### Default Cache Operations

These functions operate on the current default cache (switchable via `SetDefaultCache()`).

#### CreateCache
```go
func CreateCache(cacheRepositoryName string, ttl time.Duration)
```
Create named cache repository and set default TTL.

#### CacheSet
```go
func CacheSet(cacheRepositoryName, key string, value interface{}, ttl ...time.Duration)
```
Store value in default cache.

**Example:**
```go
dbkit.CacheSet("user_cache", "user:1", userData, 5*time.Minute)
```

#### CacheGet
```go
func CacheGet(cacheRepositoryName, key string) (interface{}, bool)
```
Get value from default cache.

**Example:**
```go
val, ok := dbkit.CacheGet("user_cache", "user:1")
```

#### CacheDelete
```go
func CacheDelete(cacheRepositoryName, key string)
```
Delete specific key from default cache.

**Example:**
```go
dbkit.CacheDelete("user_cache", "user:1")
```

#### CacheClearRepository
```go
func CacheClearRepository(cacheRepositoryName string)
```
Clear specific repository in default cache.

**Example:**
```go
dbkit.CacheClearRepository("user_cache")
```

#### ClearAllCaches
```go
func ClearAllCaches()
```
Clear all repositories in default cache.

**Example:**
```go
dbkit.ClearAllCaches()
```

#### CacheStatus
```go
func CacheStatus() map[string]interface{}
```
Get default cache status information.

**Example:**
```go
status := dbkit.CacheStatus()
fmt.Printf("Cache type: %v\n", status["type"])
fmt.Printf("Total items: %v\n", status["total_items"])
```

### Local Cache Operations

These functions directly operate on local cache, unaffected by `SetDefaultCache()`.

#### LocalCacheSet
```go
func LocalCacheSet(cacheRepositoryName, key string, value interface{}, ttl ...time.Duration)
```
Store value in local cache.

**Example:**
```go
dbkit.LocalCacheSet("config_cache", "app_name", "MyApp", 10*time.Minute)
```

#### LocalCacheGet
```go
func LocalCacheGet(cacheRepositoryName, key string) (interface{}, bool)
```
Get value from local cache.

**Example:**
```go
val, ok := dbkit.LocalCacheGet("config_cache", "app_name")
```

#### LocalCacheDelete
```go
func LocalCacheDelete(cacheRepositoryName, key string)
```
Delete specific key from local cache.

**Example:**
```go
dbkit.LocalCacheDelete("config_cache", "app_name")
```

#### LocalCacheClearRepository
```go
func LocalCacheClearRepository(cacheRepositoryName string)
```
Clear specific repository in local cache.

**Example:**
```go
dbkit.LocalCacheClearRepository("config_cache")
```

#### LocalCacheClearAll
```go
func LocalCacheClearAll()
```
Clear all repositories in local cache.

**Example:**
```go
dbkit.LocalCacheClearAll()
```

#### LocalCacheStatus
```go
func LocalCacheStatus() map[string]interface{}
```
Get local cache status information.

**Example:**
```go
status := dbkit.LocalCacheStatus()
fmt.Printf("Local cache type: %v\n", status["type"])
fmt.Printf("Memory usage: %v\n", status["estimated_memory_human"])
```

### Redis Cache Operations

These functions directly operate on Redis cache, unaffected by `SetDefaultCache()`.

**Note:** Must call `InitRedisCache()` to initialize before use.

#### RedisCacheSet
```go
func RedisCacheSet(cacheRepositoryName, key string, value interface{}, ttl ...time.Duration) error
```
Store value in Redis cache.

**Example:**
```go
err := dbkit.RedisCacheSet("session_cache", "session:abc123", sessionData, 30*time.Minute)
if err != nil {
    log.Printf("Store failed: %v", err)
}
```

#### RedisCacheGet
```go
func RedisCacheGet(cacheRepositoryName, key string) (interface{}, bool, error)
```
Get value from Redis cache.

**Example:**
```go
val, ok, err := dbkit.RedisCacheGet("session_cache", "session:abc123")
if err != nil {
    log.Printf("Get failed: %v", err)
} else if ok {
    fmt.Println("Session data:", val)
}
```

#### RedisCacheDelete
```go
func RedisCacheDelete(cacheRepositoryName, key string) error
```
Delete specific key from Redis cache.

**Example:**
```go
err := dbkit.RedisCacheDelete("session_cache", "session:abc123")
if err != nil {
    log.Printf("Delete failed: %v", err)
}
```

#### RedisCacheClearRepository
```go
func RedisCacheClearRepository(cacheRepositoryName string) error
```
Clear specific repository in Redis cache.

**Example:**
```go
err := dbkit.RedisCacheClearRepository("session_cache")
if err != nil {
    log.Printf("Clear failed: %v", err)
}
```

#### RedisCacheClearAll
```go
func RedisCacheClearAll() error
```
Clear all dbkit-related caches in Redis.

**Example:**
```go
err := dbkit.RedisCacheClearAll()
if err != nil {
    log.Printf("Clear failed: %v", err)
}
```

#### RedisCacheStatus
```go
func RedisCacheStatus() (map[string]interface{}, error)
```
Get Redis cache status information.

**Example:**
```go
status, err := dbkit.RedisCacheStatus()
if err != nil {
    log.Printf("Get status failed: %v", err)
} else {
    fmt.Printf("Redis address: %v\n", status["address"])
    fmt.Printf("Database size: %v\n", status["db_size"])
}
```

### CacheProvider Interface
```go
type CacheProvider interface {
    CacheGet(cacheRepositoryName, key string) (interface{}, bool)
    CacheSet(cacheRepositoryName, key string, value interface{}, ttl time.Duration)
    CacheDelete(cacheRepositoryName, key string)
    CacheClearRepository(cacheRepositoryName string)
    Status() map[string]interface{}
}
```

### Usage Scenarios

#### Scenario 1: Using Default Cache Only
```go
// Use default local cache
dbkit.CacheSet("user_cache", "user:1", userData)
val, _ := dbkit.CacheGet("user_cache", "user:1")

// Or switch to Redis
rc, _ := redis.NewRedisCache("localhost:6379", "", "", 0)
dbkit.SetDefaultCache(rc)

// Now operating on Redis
dbkit.CacheSet("user_cache", "user:2", userData)
```

#### Scenario 2: Using Both Local Cache and Redis
```go
// Initialize Redis
rc, _ := redis.NewRedisCache("localhost:6379", "", "", 0)
dbkit.InitRedisCache(rc)

// Store config data locally (fast access)
dbkit.LocalCacheSet("config", "app_name", "MyApp")

// Store session data in Redis (distributed sharing)
dbkit.RedisCacheSet("session", "session:123", sessionData)

// Two caches are independent
config, _ := dbkit.LocalCacheGet("config", "app_name")
session, _, _ := dbkit.RedisCacheGet("session", "session:123")
```

#### Scenario 3: Layered Cache Management
```go
// L1 Cache: Local cache (config data)
func GetConfig(key string) string {
    val, ok := dbkit.LocalCacheGet("config", key)
    if ok {
        return val.(string)
    }
    
    // Load from database
    value := loadFromDB(key)
    dbkit.LocalCacheSet("config", key, value, 1*time.Hour)
    return value
}

// L2 Cache: Redis cache (business data)
func GetUser(userID int) (*User, error) {
    key := fmt.Sprintf("user:%d", userID)
    val, ok, err := dbkit.RedisCacheGet("users", key)
    if err != nil {
        return nil, err
    }
    if ok {
        return val.(*User), nil
    }
    
    // Load from database
    user, err := loadUserFromDB(userID)
    if err != nil {
        return nil, err
    }
    
    dbkit.RedisCacheSet("users", key, user, 5*time.Minute)
    return user, nil
}
```

---

## Log Configuration

### SetDebugMode
```go
func SetDebugMode(enabled bool)
```
Enable/Disable debug mode (outputs SQL statements).

### SetLogger
```go
func SetLogger(l Logger)
```
Set custom logger.

### InitLoggerWithFile
```go
func InitLoggerWithFile(level string, filePath string)
```
Initialize file logging.

### Logger Interface
```go
type Logger interface {
    Log(level LogLevel, msg string, fields map[string]interface{})
}
```

### Log Levels
```go
const (
    LevelDebug LogLevel = "debug"
    LevelInfo  LogLevel = "info"
    LevelWarn  LogLevel = "warn"
    LevelError LogLevel = "error"
)
```

### Log Functions
```go
func LogDebug(msg string, fields map[string]interface{})
func LogInfo(msg string, fields map[string]interface{})
func LogWarn(msg string, fields map[string]interface{})
func LogError(msg string, fields map[string]interface{})
```

---

## SQL Templates

DBKit provides powerful SQL template functionality that allows you to manage SQL statements through configuration, supporting dynamic parameters, conditional building, and multi-database execution.

### Configuration File Structure

SQL templates use JSON format configuration files. Here's a complete configuration file format template:

#### Complete JSON Format Template

```json
{
  "version": "1.0",
  "description": "Service SQL configuration file description",
  "namespace": "service_name",
  "sqls": [
    {
      "name": "sqlName",
      "description": "SQL statement description",
      "sql": "SELECT * FROM table WHERE condition = :param",
      "type": "select",
      "order": "created_at DESC",
      "inparam": [
        {
          "name": "paramName",
          "type": "string",
          "desc": "Parameter description",
          "sql": " AND column = :paramName"
        }
      ]
    }
  ]
}
```

#### Field Descriptions

**Root Level Fields:**
- `version` (string, required): Configuration file version number
- `description` (string, optional): Configuration file description
- `namespace` (string, optional): Namespace to avoid SQL name conflicts
- `sqls` (array, required): Array of SQL statement configurations

**SQL Configuration Fields:**
- `name` (string, required): Unique identifier for the SQL statement
- `description` (string, optional): SQL statement description
- `sql` (string, required): SQL statement template
- `type` (string, optional): SQL type (`select`, `insert`, `update`, `delete`)
- `order` (string, optional): Default sorting condition
- `inparam` (array, optional): Input parameter definitions (for dynamic SQL)

**Input Parameter Fields (inparam):**
- `name` (string, required): Parameter name
- `type` (string, required): Parameter type
- `desc` (string, optional): Parameter description
- `sql` (string, required): SQL fragment to append when parameter exists

#### Practical Configuration Example

```json
{
  "version": "1.0",
  "description": "User service SQL configuration",
  "namespace": "user_service",
  "sqls": [
    {
      "name": "findById",
      "description": "Find user by ID",
      "sql": "SELECT * FROM users WHERE id = :id",
      "type": "select"
    },
    {
      "name": "findUsers",
      "description": "Dynamic user query",
      "sql": "SELECT * FROM users WHERE 1=1",
      "type": "select",
      "order": "created_at DESC",
      "inparam": [
        {
          "name": "status",
          "type": "int",
          "desc": "User status",
          "sql": " AND status = :status"
        },
        {
          "name": "name",
          "type": "string",
          "desc": "Name fuzzy search",
          "sql": " AND name LIKE CONCAT('%', :name, '%')"
        }
      ]
    }
  ]
}
```

### Parameter Type Support

DBKit SQL templates support multiple parameter passing methods, providing flexible usage experience:

#### Supported Parameter Types

| Parameter Type | Use Case | SQL Placeholder | Example |
|---------------|----------|-----------------|---------|
| `map[string]interface{}` | Named parameters | `:name` | `map[string]interface{}{"id": 123}` |
| `[]interface{}` | Multiple positional parameters | `?` | `[]interface{}{123, "John"}` |
| **Single simple types** | Single positional parameter | `?` | `123`, `"John"`, `true` |
| **Variadic parameters** | Multiple positional parameters | `?` | `SqlTemplate(name, 123, "John", true)` |

#### Single Simple Type Support

🆕 **New Feature**: Support for directly passing single simple type parameters without wrapping in map or slice:

- `string` - String values
- `int`, `int8`, `int16`, `int32`, `int64` - Integer types
- `uint`, `uint8`, `uint16`, `uint32`, `uint64` - Unsigned integers
- `float32`, `float64` - Floating point numbers
- `bool` - Boolean values

#### Variadic Parameter Support

🆕 **New Feature**: Support for Go-style variadic parameters (`...interface{}`), providing the most natural parameter passing method:

```go
// Variadic parameter approach - most intuitive and concise
records, err := dbkit.SqlTemplate("findByIdAndStatus", 123, 1).Query()
records, err := dbkit.SqlTemplate("updateUser", "John", "john@example.com", 25, 123).Exec()
records, err := dbkit.SqlTemplate("findByAgeRange", 18, 65, 1).Query()
```

#### Parameter Matching Rules

| SQL Placeholder | Parameter Type | Result |
|----------------|---------------|--------|
| Single `?` | Single simple type | ✅ Supported |
| Single `?` | `map[string]interface{}` | ✅ Supported (backward compatible) |
| Single `?` | `[]interface{}{value}` | ✅ Supported (backward compatible) |
| Multiple `?` | `[]interface{}{v1, v2, ...}` | ✅ Supported |
| Multiple `?` | **Variadic parameters `v1, v2, ...`** | ✅ Supported 🆕 |
| Multiple `?` | Single simple type | ❌ Error message |
| `:name` | `map[string]interface{}{"name": value}` | ✅ Supported |
| `:name` | Single simple type | ❌ Error message |
| `:name` | Variadic parameters | ❌ Error message |

#### Parameter Count Validation

🆕 **Enhanced Feature**: The system automatically validates that parameter count matches SQL placeholder count:

```go
// SQL: "SELECT * FROM users WHERE id = ? AND status = ?"
// Correct: 2 parameters match 2 placeholders
records, err := dbkit.SqlTemplate("findByIdAndStatus", 123, 1).Query()

// Error: insufficient parameters
records, err := dbkit.SqlTemplate("findByIdAndStatus", 123).Query()
// Returns error: parameter count mismatch: SQL has 2 '?' placeholders but got 1 parameters

// Error: too many parameters
records, err := dbkit.SqlTemplate("findByIdAndStatus", 123, 1, 2).Query()
// Returns error: parameter count mismatch: SQL has 2 '?' placeholders but got 3 parameters
```
| Multiple `?` | Single simple type | ❌ Error message |
| `:name` | `map[string]interface{}{"name": value}` | ✅ Supported |
| `:name` | Single simple type | ❌ Error message |

#### Usage Examples

```go
// 1. Single simple parameter (recommended for single-parameter queries)
records, err := dbkit.SqlTemplate("user_service.findById", 123).Query()
records, err := dbkit.SqlTemplate("user_service.findByEmail", "user@example.com").Query()
records, err := dbkit.SqlTemplate("user_service.findActive", true).Query()

// 2. Named parameters (suitable for complex queries)
params := map[string]interface{}{
    "status": 1,
    "name": "John",
    "ageMin": 18,
}
records, err := dbkit.SqlTemplate("user_service.findUsers", params).Query()

// 3. Positional parameters (suitable for multi-parameter queries)
records, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 
    []interface{}{123, 1}).Query()
```

### Configuration Loading

#### LoadSqlConfig
```go
func LoadSqlConfig(configPath string) error
```
Load a single SQL configuration file.

**Example:**
```go
err := dbkit.LoadSqlConfig("config/user_service.json")
```

#### LoadSqlConfigs
```go
func LoadSqlConfigs(configPaths []string) error
```
Load multiple SQL configuration files in batch.

**Example:**
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
Load all JSON configuration files in the specified directory.

**Example:**
```go
err := dbkit.LoadSqlConfigDir("config/")
```

#### ReloadSqlConfig
```go
func ReloadSqlConfig(configPath string) error
```
Reload the specified configuration file.

#### ReloadAllSqlConfigs
```go
func ReloadAllSqlConfigs() error
```
Reload all loaded configuration files.

### Configuration Information Query

#### GetSqlConfigInfo
```go
func GetSqlConfigInfo() []ConfigInfo
```
Get information about all loaded configuration files.

**ConfigInfo Structure:**
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
List all available SQL template items.

### SQL Template Execution

#### SqlTemplate (Global)
```go
func SqlTemplate(name string, params ...interface{}) *SqlTemplateBuilder
```
Create SQL template builder using the default database connection.

**Parameters:**
- `name`: SQL template name (supports namespace, e.g., "user_service.findById")
- `params`: Variadic parameters, supports the following types:
  - `map[string]interface{}` - Named parameters (`:name`)
  - `[]interface{}` - Positional parameter array (`?`)
  - **Single simple types** - Single positional parameter (`?`), supports `string`, `int`, `float`, `bool`, etc.
  - **🆕 Variadic parameters** - Multiple positional parameters (`?`), directly pass multiple values

**Examples:**
```go
// Using named parameters
records, err := dbkit.SqlTemplate("user_service.findById", 
    map[string]interface{}{"id": 123}).Query()

// Using positional parameter array
records, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 
    []interface{}{123, 1}).Query()

// 🆕 Using single simple parameter (recommended for single parameter queries)
records, err := dbkit.SqlTemplate("user_service.findById", 123).Query()
records, err := dbkit.SqlTemplate("user_service.findByEmail", "user@example.com").Query()

// 🆕 Using variadic parameters (recommended for multi-parameter queries)
records, err := dbkit.SqlTemplate("user_service.findByIdAndStatus", 123, 1).Query()
records, err := dbkit.SqlTemplate("user_service.updateUser", "John", "john@example.com", 25, 123).Exec()
records, err := dbkit.SqlTemplate("user_service.findByAgeRange", 18, 65, 1).Query()
```

#### SqlTemplate (Database Specific)
```go
func (db *DB) SqlTemplate(name string, params ...interface{}) *SqlTemplateBuilder
```
Create SQL template builder on a specific database.

**Examples:**
```go
// Traditional approach
records, err := dbkit.Use("mysql").SqlTemplate("user_service.findById", 
    map[string]interface{}{"id": 123}).Query()

// 🆕 Single simple parameter (more concise)
records, err := dbkit.Use("mysql").SqlTemplate("user_service.findById", 123).Query()

// 🆕 Variadic parameters (most concise)
records, err := dbkit.Use("mysql").SqlTemplate("user_service.findByIdAndStatus", 123, 1).Query()
```

#### SqlTemplate (Transaction)
```go
func (tx *Tx) SqlTemplate(name string, params ...interface{}) *SqlTemplateBuilder
```
Use SQL templates within transactions.

**Examples:**
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    // Using variadic parameters
    result, err := tx.SqlTemplate("user_service.insertUser", "John", "john@example.com", 25).Exec()
    return err
})
```

**Example:**
```go
// Using named parameters
records, err := dbkit.SqlTemplate("user_service.findById", 
    map[string]interface{}{"id": 123}).Query()

// Using positional parameters
records, err := dbkit.SqlTemplate("user_service.findById", 
    []interface{}{123}).Query()

// 🆕 Using single simple parameter (recommended for single-parameter queries)
records, err := dbkit.SqlTemplate("user_service.findById", 123).Query()
records, err := dbkit.SqlTemplate("user_service.findByEmail", "user@example.com").Query()
records, err := dbkit.SqlTemplate("user_service.findActive", true).Query()
```

#### SqlTemplate (Database Specific)
```go
func (db *DB) SqlTemplate(name string, params interface{}) *SqlTemplateBuilder
```
Create SQL template builder on a specific database.

**Example:**
```go
// Traditional way
records, err := dbkit.Use("mysql").SqlTemplate("user_service.findById", 
    map[string]interface{}{"id": 123}).Query()

// 🆕 Single simple parameter (more concise)
records, err := dbkit.Use("mysql").SqlTemplate("user_service.findById", 123).Query()
```

#### SqlTemplate (Transaction)
```go
func (tx *Tx) SqlTemplate(name string, params interface{}) *SqlTemplateBuilder
```
Use SQL templates within transactions.

**Example:**
```go
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    result, err := tx.SqlTemplate("user_service.insertUser", userParams).Exec()
    return err
})
```

### SqlTemplateBuilder Methods

#### Timeout
```go
func (b *SqlTemplateBuilder) Timeout(timeout time.Duration) *SqlTemplateBuilder
```
Set query timeout.

**Example:**
```go
records, err := dbkit.SqlTemplate("user_service.findUsers", params).
    Timeout(30 * time.Second).Query()
```

#### Query
```go
func (b *SqlTemplateBuilder) Query() ([]Record, error)
```
Execute query and return multiple records.

#### QueryFirst
```go
func (b *SqlTemplateBuilder) QueryFirst() (*Record, error)
```
Execute query and return the first record.

#### Exec
```go
func (b *SqlTemplateBuilder) Exec() (sql.Result, error)
```
Execute SQL statement (INSERT, UPDATE, DELETE).

#### Paginate
```go
func (b *SqlTemplateBuilder) Paginate(page int, pageSize int) (*Page[Record], error)
```
Execute SQL template and return paginated results. Uses complete SQL statement for pagination query, automatically parses SQL and generates corresponding pagination statements based on database type.

**Parameters:**
- `page`: Page number (starting from 1)
- `pageSize`: Number of records per page

**Returns:**
- `*Page[Record]`: Pagination result object
- `error`: Error information

**Examples:**
```go
// Basic pagination query
pageObj, err := dbkit.SqlTemplate("user_service.findActiveUsers", 1).
    Paginate(1, 10)

// Pagination with parameters
pageObj, err := dbkit.SqlTemplate("user_service.findByStatus", "active", 18).
    Paginate(2, 20)

// Execute pagination on specific database
pageObj, err := dbkit.Use("mysql").SqlTemplate("findUsers", params).
    Paginate(1, 15)

// Execute pagination in transaction
err := dbkit.Transaction(func(tx *dbkit.Tx) error {
    pageObj, err := tx.SqlTemplate("findOrders", userId).Paginate(1, 10)
    // Process pagination results...
    return err
})

// Pagination with timeout
pageObj, err := dbkit.SqlTemplate("complexQuery", params).
    Timeout(30 * time.Second).
    Paginate(1, 50)

// Access pagination results
if err == nil {
    fmt.Printf("Page %d of %d, Total: %d\n", 
        pageObj.PageNumber, pageObj.TotalPage, pageObj.TotalRow)
    
    for _, record := range pageObj.List {
        fmt.Printf("User: %s, Age: %d\n", 
            record.Str("name"), record.Int("age"))
    }
}
```

### Dynamic SQL Building

Dynamic SQL condition building can be achieved through `inparam` configuration:

```json
{
  "name": "searchUsers",
  "sql": "SELECT * FROM users WHERE 1=1",
  "inparam": [
    {
      "name": "status",
      "type": "int",
      "desc": "User status",
      "sql": " AND status = :status"
    },
    {
      "name": "ageMin",
      "type": "int", 
      "desc": "Minimum age",
      "sql": " AND age >= :ageMin"
    }
  ],
  "order": "created_at DESC"
}
```

**Usage Example:**
```go
// Only pass partial parameters, system will automatically build corresponding SQL
params := map[string]interface{}{
    "status": 1,
    // ageMin not provided, corresponding condition won't be added
}
records, err := dbkit.SqlTemplate("searchUsers", params).Query()
// Generated SQL: SELECT * FROM users WHERE 1=1 AND status = ? ORDER BY created_at DESC
```

### Parameter Processing

#### Named Parameters
Use `:paramName` format for named parameters:

```go
params := map[string]interface{}{
    "id": 123,
    "name": "John",
}
records, err := dbkit.SqlTemplate("user_service.updateUser", params).Exec()
```

#### Positional Parameters
Use `?` placeholders for positional parameters:

```go
params := []interface{}{123}
records, err := dbkit.SqlTemplate("user_service.findById", params).Query()
```

### Error Handling

The SQL template system provides detailed error information:

```go
type SqlConfigError struct {
    Type    string // Error type: NotFoundError, ParameterError, ParseError, etc.
    Message string // Error description
    SqlName string // Related SQL name
    Cause   error  // Original error
}
```

**Common Error Types:**
- `NotFoundError`: SQL template not found
- `ParameterError`: Parameter error (missing, type mismatch, etc.)
- `ParameterTypeMismatch`: Parameter type doesn't match SQL format
- `ParseError`: Configuration file parsing error
- `DuplicateError`: Duplicate SQL identifier

### Best Practices

1. **Naming Convention**: Use namespaces to avoid SQL name conflicts
2. **Parameter Validation**: System automatically validates required parameters
3. **Dynamic Conditions**: Use `inparam` for flexible condition building
4. **Error Handling**: Catch and handle `SqlConfigError` type errors
5. **Performance Optimization**: Configuration files are cached after first load

**Complete Example:**
```go
// 1. Load configuration
err := dbkit.LoadSqlConfigDir("config/")
if err != nil {
    log.Fatal(err)
}

// 2. Execute query
params := map[string]interface{}{
    "status": 1,
    "name": "John",
}

records, err := dbkit.Use("mysql").
    SqlTemplate("user_service.findUsers", params).
    Timeout(30 * time.Second).
    Query()

if err != nil {
    if sqlErr, ok := err.(*dbkit.SqlConfigError); ok {
        log.Printf("SQL config error [%s]: %s", sqlErr.Type, sqlErr.Message)
    } else {
        log.Printf("Execution error: %v", err)
    }
    return
}

// 3. Process results
for _, record := range records {
    fmt.Printf("User: %s, Status: %d\n", 
        record.GetString("name"), 
        record.GetInt("status"))
}
```

---

## Utility Functions

### ToJson
```go
func ToJson(v interface{}) string
```
Convert any value to JSON string.

### ToStruct
```go
func ToStruct(record *Record, dest interface{}) error
```
Convert Record to struct.

### ToStructs
```go
func ToStructs(records []Record, dest interface{}) error
```
Convert Record slice to struct slice.

### ToRecord
```go
func ToRecord(model interface{}) *Record
```
Convert struct to Record.

### FromStruct
```go
func FromStruct(src interface{}, record *Record) error
```
Populate Record from struct.

### SnakeToCamel
```go
func SnakeToCamel(s string) string
```
Snake case to Camel case.

### ValidateTableName
```go
func ValidateTableName(table string) error
```
Validate if table name is legal.

### GenerateCacheKey
```go
func GenerateCacheKey(dbName, sql string, args ...interface{}) string
```
Generate cache key.

### SupportedDrivers
```go
func SupportedDrivers() []DriverType
```
Return list of supported database drivers.

### IsValidDriver
```go
func IsValidDriver(driver DriverType) bool
```
Check if driver is supported.

---

## Database Driver Types

```go
const (
    MySQL      DriverType = "mysql"
    PostgreSQL DriverType = "postgres"
    SQLite3    DriverType = "sqlite3"
    Oracle     DriverType = "oracle"
    SQLServer  DriverType = "sqlserver"
)
```
