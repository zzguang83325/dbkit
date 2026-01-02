# DBKit API Reference

[中文版](api.md) | [README](README.md) | [English README](README_EN.md)

## Table of Contents

- [Database Initialization](#database-initialization)
- [Query Operations](#query-operations)
- [Insert & Update](#insert--update)
- [Delete Operations](#delete-operations)
- [Soft Delete](#soft-delete)
- [Auto Timestamps](#auto-timestamps)
- [Optimistic Lock](#optimistic-lock)
- [Transaction Handling](#transaction-handling)
- [Record Object](#record-object)
- [Chain Query](#chain-query)
- [DbModel Operations](#dbmodel-operations)
- [Cache Operations](#cache-operations)
- [Logging Configuration](#logging-configuration)
- [Utility Functions](#utility-functions)

---

## Database Initialization

### OpenDatabase
```go
func OpenDatabase(driver DriverType, dsn string, maxOpen int) error
```
Opens a database connection with default configuration.

**Parameters:**
- `driver`: Database driver type (MySQL, PostgreSQL, SQLite3, Oracle, SQLServer)
- `dsn`: Data source name (connection string)
- `maxOpen`: Maximum number of open connections

**Example:**
```go
err := dbkit.OpenDatabase(dbkit.MySQL, "root:password@tcp(localhost:3306)/test", 10)
```

### OpenDatabaseWithConfig
```go
func OpenDatabaseWithConfig(config *Config) error
```
Opens a database connection with custom configuration.

**Config struct:**
```go
type Config struct {
    Driver          DriverType    // Database driver type
    DSN             string        // Data source name
    MaxOpen         int           // Maximum open connections
    MaxIdle         int           // Maximum idle connections
    ConnMaxLifetime time.Duration // Maximum connection lifetime
}
```

### OpenDatabaseWithDBName
```go
func OpenDatabaseWithDBName(dbname string, driver DriverType, dsn string, maxOpen int) error
```
Opens a database connection with a specified name (multi-database mode).

### Register
```go
func Register(dbname string, config *Config) error
```
Registers a named database with custom configuration.

### Use
```go
func Use(dbname string) *DB
```
Switches to the specified database and returns a DB object for chaining.

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
Closes database connections.

### Ping
```go
func Ping() error
func PingDB(dbname string) error
```
Tests database connectivity.

---

## Query Operations

### Query
```go
func Query(querySQL string, args ...interface{}) ([]Record, error)
func (db *DB) Query(querySQL string, args ...interface{}) ([]Record, error)
func (tx *Tx) Query(querySQL string, args ...interface{}) ([]Record, error)
```
Executes a query and returns multiple records.

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
Executes a query and returns the first record, or nil if no records found.

### QueryMap
```go
func QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error)
func (db *DB) QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error)
```
Executes a query and returns a slice of maps.

### QueryToDbModel
```go
func QueryToDbModel(dest interface{}, querySQL string, args ...interface{}) error
func (db *DB) QueryToDbModel(dest interface{}, querySQL string, args ...interface{}) error
```
Executes a query and maps results to a struct slice.

### QueryFirstToDbModel
```go
func QueryFirstToDbModel(dest interface{}, querySQL string, args ...interface{}) error
func (db *DB) QueryFirstToDbModel(dest interface{}, querySQL string, args ...interface{}) error
```
Executes a query and maps the first result to a struct.

### Count
```go
func Count(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Count(table string, whereSql string, whereArgs ...interface{}) (int64, error)
```
Counts records matching the condition.

**Example:**
```go
count, err := dbkit.Count("users", "age > ?", 18)
```

### Exists
```go
func Exists(table string, whereSql string, whereArgs ...interface{}) (bool, error)
func (db *DB) Exists(table string, whereSql string, whereArgs ...interface{}) (bool, error)
```
Checks if records matching the condition exist.

### FindAll
```go
func FindAll(table string) ([]Record, error)
func (db *DB) FindAll(table string) ([]Record, error)
```
Retrieves all records from a table.

### Paginate
```go
func Paginate(page, pageSize int, selectSql, table, whereSql, orderBySql string, args ...interface{}) (*Page[Record], error)
func (db *DB) Paginate(page, pageSize int, selectSql, table, whereSql, orderBySql string, args ...interface{}) (*Page[Record], error)
```
Performs paginated query.

**Parameters:**
- `page`: Page number (starting from 1)
- `pageSize`: Records per page
- `selectSql`: SELECT clause
- `table`: Table name
- `whereSql`: WHERE condition
- `orderBySql`: ORDER BY clause
- `args`: Query parameters

**Returns Page struct:**
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

## Insert & Update

### Exec
```go
func Exec(querySQL string, args ...interface{}) (sql.Result, error)
func (db *DB) Exec(querySQL string, args ...interface{}) (sql.Result, error)
func (tx *Tx) Exec(querySQL string, args ...interface{}) (sql.Result, error)
```
Executes SQL statements (INSERT, UPDATE, DELETE, etc.).

### Save
```go
func Save(table string, record *Record) (int64, error)
func (db *DB) Save(table string, record *Record) (int64, error)
func (tx *Tx) Save(table string, record *Record) (int64, error)
```
Smart save: updates if primary key exists and record found, otherwise inserts.

**Returns:** New ID on insert, affected rows on update.

### Insert
```go
func Insert(table string, record *Record) (int64, error)
func (db *DB) Insert(table string, record *Record) (int64, error)
func (tx *Tx) Insert(table string, record *Record) (int64, error)
```
Forces insertion of a new record.

**Returns:** ID of the newly inserted record.

### Update
```go
func Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
func (tx *Tx) Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
```
Updates records matching the condition. Supports auto-timestamp and optimistic lock features.

**Returns:** Number of affected rows.

**Performance Note:** DBKit disables timestamp auto-update and optimistic lock checks by default for optimal performance. To enable these features, use `EnableTimestampCheck()` or `EnableOptimisticLockCheck()`.

### UpdateFast
```go
func UpdateFast(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) UpdateFast(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error)
```
Lightweight update that always skips timestamp and optimistic lock checks for maximum performance.

**Returns:** Number of affected rows.

**Use Cases:**

1. **High-frequency Updates**: High-concurrency update operations requiring extreme performance
   ```go
   // Game server updating player scores
   record := dbkit.NewRecord().Set("score", newScore)
   dbkit.UpdateFast("players", record, "id = ?", playerId)
   ```

2. **Batch Updates**: Reducing overhead when updating large amounts of data
   ```go
   // Batch update product inventory
   for _, item := range items {
       record := dbkit.NewRecord().Set("stock", item.Stock)
       dbkit.UpdateFast("products", record, "id = ?", item.ID)
   }
   ```

3. **Tables Without Feature Requirements**: Tables that don't need timestamp or optimistic lock features
   ```go
   // Update configuration table (no timestamp needed)
   record := dbkit.NewRecord().Set("value", "new_value")
   dbkit.UpdateFast("config", record, "key = ?", "app_version")
   ```

4. **Skip Checks When Features Enabled**: Feature checks enabled globally, but specific operations need maximum performance
   ```go
   // Feature checks enabled globally
   dbkit.EnableFeatureChecks()
   
   // But some high-frequency operations need to skip checks
   record := dbkit.NewRecord().Set("view_count", viewCount)
   dbkit.UpdateFast("articles", record, "id = ?", articleId)
   ```

**Performance Comparison:**
- When feature checks are disabled, `Update` and `UpdateFast` have the same performance
- When feature checks are enabled, `UpdateFast` is about 2-3x faster than `Update`

**Important Notes:**
- `UpdateFast` does not automatically update the `updated_at` field
- `UpdateFast` does not perform optimistic lock version checks
- If you need these features, use `Update` and enable the corresponding feature checks

### UpdateRecord
```go
func (db *DB) UpdateRecord(table string, record *Record) (int64, error)
func (tx *Tx) UpdateRecord(table string, record *Record) (int64, error)
```
Updates a record based on its primary key.

### BatchInsert
```go
func BatchInsert(table string, records []*Record, batchSize int) (int64, error)
func (db *DB) BatchInsert(table string, records []*Record, batchSize int) (int64, error)
```
Batch inserts records.

**Parameters:**
- `batchSize`: Number of records per batch

### BatchInsertDefault
```go
func BatchInsertDefault(table string, records []*Record) (int64, error)
func (db *DB) BatchInsertDefault(table string, records []*Record) (int64, error)
```
Batch inserts records with default batch size of 100.

---

## Delete Operations

### Delete
```go
func Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (db *DB) Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
func (tx *Tx) Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error)
```
Deletes records matching the condition. If soft delete is configured for the table, performs a soft delete (updates the delete marker field).

### DeleteRecord
```go
func DeleteRecord(table string, record *Record) (int64, error)
func (db *DB) DeleteRecord(table string, record *Record) (int64, error)
func (tx *Tx) DeleteRecord(table string, record *Record) (int64, error)
```
Deletes a record based on its primary key.

---

## Soft Delete

Soft delete allows marking records as deleted instead of physically removing them, enabling data recovery and auditing.

**Performance Note**: DBKit disables soft delete checks by default for optimal performance. To enable this feature, use:

```go
// Enable soft delete check
dbkit.EnableSoftDeleteCheck()
```

### EnableSoftDeleteCheck
```go
func EnableSoftDeleteCheck()
func (db *DB) EnableSoftDeleteCheck() *DB
```
Enables soft delete check feature. When enabled, query operations will automatically filter out soft-deleted records.

**Example:**
```go
// Enable soft delete check globally
dbkit.EnableSoftDeleteCheck()

// Multi-database mode
dbkit.Use("main").EnableSoftDeleteCheck()
```

### Soft Delete Types
```go
const (
    SoftDeleteTimestamp SoftDeleteType = iota  // Timestamp type (deleted_at)
    SoftDeleteBool                              // Boolean type (is_deleted)
)
```

### ConfigSoftDelete
```go
func ConfigSoftDelete(table, field string)
func (db *DB) ConfigSoftDelete(table, field string) *DB
```
Configures soft delete for a table (timestamp type).

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
Configures soft delete for a table with specified type.

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
Removes soft delete configuration for a table.

### HasSoftDelete
```go
func HasSoftDelete(table string) bool
func (db *DB) HasSoftDelete(table string) bool
```
Checks if soft delete is configured for a table.

### WithTrashed
```go
func (qb *QueryBuilder) WithTrashed() *QueryBuilder
```
Includes soft-deleted records in query results.

**Example:**
```go
// Query all users (including deleted)
users, err := dbkit.Table("users").WithTrashed().Find()
```

### OnlyTrashed
```go
func (qb *QueryBuilder) OnlyTrashed() *QueryBuilder
```
Returns only soft-deleted records.

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
Physically deletes records, bypassing soft delete configuration.

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
Restores soft-deleted records.

**Example:**
```go
// Restore record
dbkit.Restore("users", "id = ?", 1)

// Chain call
dbkit.Table("users").Where("id = ?", 1).Restore()
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
users, _ := dbkit.Table("users").Find()  // Excludes deleted

// 5. Query including deleted records
allUsers, _ := dbkit.Table("users").WithTrashed().Find()

// 6. Query only deleted records
deletedUsers, _ := dbkit.Table("users").OnlyTrashed().Find()

// 7. Restore deleted record
dbkit.Restore("users", "id = ?", 1)

// 8. Physical delete (permanently removes data)
dbkit.ForceDelete("users", "id = ?", 1)
```

### DbModel Soft Delete Methods

Generated DbModels automatically include soft delete methods:

```go
// Soft delete (if configured)
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

## Auto Timestamps

Auto timestamps feature automatically fills timestamp fields when inserting and updating records, without manual setting.

**Performance Note:** DBKit disables auto timestamp checks by default for optimal performance. To enable, use `EnableTimestampCheck()` or `EnableFeatureChecks()`.

### EnableTimestampCheck
```go
func EnableTimestampCheck()
func (db *DB) EnableTimestampCheck() *DB
```
Enables auto timestamp check feature. When enabled, Update operations will check table timestamp configuration and automatically update the `updated_at` field.

**Example:**
```go
// Enable timestamp check globally
dbkit.EnableTimestampCheck()

// Multi-database mode
dbkit.Use("main").EnableTimestampCheck()
```

### ConfigTimestamps
```go
func ConfigTimestamps(table string)
func (db *DB) ConfigTimestamps(table string) *DB
```
Configures auto timestamps for a table using default field names `created_at` and `updated_at`.

**Example:**
```go
// Configure auto timestamps
dbkit.ConfigTimestamps("users")

// Multi-database mode
dbkit.Use("main").ConfigTimestamps("users")
```

### ConfigTimestampsWithFields
```go
func ConfigTimestampsWithFields(table, createdAtField, updatedAtField string)
func (db *DB) ConfigTimestampsWithFields(table, createdAtField, updatedAtField string) *DB
```
Configures auto timestamps for a table with custom field names.

**Parameters:**
- `table`: Table name
- `createdAtField`: Created at field name (e.g., "create_time")
- `updatedAtField`: Updated at field name (e.g., "update_time")

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
Configures only the created_at field.

**Example:**
```go
// Configure only created_at (suitable for log tables)
dbkit.ConfigCreatedAt("logs", "log_time")
```

### ConfigUpdatedAt
```go
func ConfigUpdatedAt(table, field string)
func (db *DB) ConfigUpdatedAt(table, field string) *DB
```
Configures only the updated_at field.

**Example:**
```go
// Configure only updated_at
dbkit.ConfigUpdatedAt("cache_data", "last_modified")
```

### RemoveTimestamps
```go
func RemoveTimestamps(table string)
func (db *DB) RemoveTimestamps(table string) *DB
```
Removes timestamp configuration for a table.

### HasTimestamps
```go
func HasTimestamps(table string) bool
func (db *DB) HasTimestamps(table string) bool
```
Checks if auto timestamps are configured for a table.

### WithoutTimestamps
```go
func (qb *QueryBuilder) WithoutTimestamps() *QueryBuilder
```
Temporarily disables auto timestamps (for QueryBuilder Update operations).

**Example:**
```go
// Update without auto-filling updated_at
dbkit.Table("users").Where("id = ?", 1).WithoutTimestamps().Update(record)
```

### Auto Timestamps Behavior

- **Insert operation**: If `created_at` field is not set, automatically fills with current time
- **Update operation**: Always automatically fills `updated_at` field with current time
- **Manual setting priority**: If `created_at` is already set in Record, it won't be overwritten

### Complete Auto Timestamps Example
```go
// 1. Configure auto timestamps
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

// 4. Insert with manual created_at (won't be overwritten)
customTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
record2 := dbkit.NewRecord()
record2.Set("name", "Jane")
record2.Set("created_at", customTime)
dbkit.Insert("users", record2)
// created_at remains 2020-01-01

// 5. Temporarily disable auto timestamps
dbkit.Table("users").Where("id = ?", 1).WithoutTimestamps().Update(record)
// updated_at won't be auto-updated

// 6. Use custom field names
dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")

// 7. Configure only created_at (suitable for log tables)
dbkit.ConfigCreatedAt("logs", "log_time")
```

### Using with Soft Delete

Auto timestamps and soft delete features are independent and can be used together:

```go
// Configure both soft delete and auto timestamps
dbkit.ConfigTimestamps("users")
dbkit.ConfigSoftDelete("users", "deleted_at")

// When soft deleting, updated_at is also auto-updated
dbkit.Delete("users", "id = ?", 1)
// deleted_at set to current time, updated_at also updated
```

---

## Optimistic Lock

Optimistic lock is a concurrency control mechanism that detects concurrent update conflicts through a version field, preventing data from being accidentally overwritten.

**Performance Note:** DBKit disables optimistic lock checks by default for optimal performance. To enable, use `EnableOptimisticLockCheck()` or `EnableFeatureChecks()`.

### EnableOptimisticLockCheck
```go
func EnableOptimisticLockCheck()
func (db *DB) EnableOptimisticLockCheck() *DB
```
Enables optimistic lock check feature. When enabled, Update operations will check table optimistic lock configuration and automatically perform version checks.

**Example:**
```go
// Enable optimistic lock check globally
dbkit.EnableOptimisticLockCheck()

// Multi-database mode
dbkit.Use("main").EnableOptimisticLockCheck()
```

### EnableFeatureChecks
```go
func EnableFeatureChecks()
func (db *DB) EnableFeatureChecks() *DB
```
Enables both timestamp check and optimistic lock check features.

**Example:**
```go
// Enable all feature checks globally
dbkit.EnableFeatureChecks()

// Multi-database mode
dbkit.Use("main").EnableFeatureChecks()
```

### How It Works

1. **Insert**: Automatically initializes the version field to 1
2. **Update**: Automatically adds version check to WHERE clause and increments version in SET clause
3. **Conflict Detection**: If update affects 0 rows (version mismatch), returns `ErrVersionMismatch` error

### ErrVersionMismatch
```go
var ErrVersionMismatch = fmt.Errorf("dbkit: optimistic lock conflict - record was modified by another transaction")
```
Error returned when version conflict is detected.

### ConfigOptimisticLock
```go
func ConfigOptimisticLock(table string)
func (db *DB) ConfigOptimisticLock(table string) *DB
```
Configures optimistic lock for a table using default field name `version`.

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
Configures optimistic lock for a table with custom version field name.

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
Removes optimistic lock configuration for a table.

### HasOptimisticLock
```go
func HasOptimisticLock(table string) bool
func (db *DB) HasOptimisticLock(table string) bool
```
Checks if optimistic lock is configured for a table.

### Version Field Handling Rules

| version field value | Behavior |
|---------------------|----------|
| Not present | Skip version check, normal update |
| `nil` / `NULL` | Skip version check, normal update |
| `""` (empty string) | Skip version check, normal update |
| `0`, `1`, `2`, ... | Perform version check |
| `"123"` (numeric string) | Perform version check (parsed as number) |

### Complete Optimistic Lock Example

```go
// 1. Configure optimistic lock
dbkit.ConfigOptimisticLock("products")

// 2. Insert data (version auto-initialized to 1)
record := dbkit.NewRecord()
record.Set("name", "Laptop")
record.Set("price", 999.99)
dbkit.Insert("products", record)
// version automatically set to 1

// 3. Normal update (with version)
updateRecord := dbkit.NewRecord()
updateRecord.Set("version", int64(1))  // current version
updateRecord.Set("price", 899.99)
rows, err := dbkit.Update("products", updateRecord, "id = ?", 1)
// Success: version auto-incremented to 2

// 4. Concurrent conflict detection (using stale version)
staleRecord := dbkit.NewRecord()
staleRecord.Set("version", int64(1))  // stale version!
staleRecord.Set("price", 799.99)
rows, err = dbkit.Update("products", staleRecord, "id = ?", 1)
if errors.Is(err, dbkit.ErrVersionMismatch) {
    fmt.Println("Concurrent conflict detected, record was modified by another transaction")
}

// 5. Correct way to handle concurrency: read latest version first
latestRecord, _ := dbkit.Table("products").Where("id = ?", 1).FindFirst()
currentVersion := latestRecord.GetInt("version")

updateRecord2 := dbkit.NewRecord()
updateRecord2.Set("version", currentVersion)
updateRecord2.Set("price", 799.99)
dbkit.Update("products", updateRecord2, "id = ?", 1)

// 6. Update without version field (skips version check)
noVersionRecord := dbkit.NewRecord()
noVersionRecord.Set("stock", 90)  // no version set
dbkit.Update("products", noVersionRecord, "id = ?", 1)
// Normal update, no version check

// 7. Using UpdateRecord (auto-extracts version from record)
product, _ := dbkit.Table("products").Where("id = ?", 1).FindFirst()
product.Set("name", "Gaming Laptop")
dbkit.Use("default").UpdateRecord("products", product)
// version is already in product, auto version check

// 8. Using optimistic lock in transaction
dbkit.Transaction(func(tx *dbkit.Tx) error {
    rec, _ := tx.Table("products").Where("id = ?", 1).FindFirst()
    currentVersion := rec.GetInt("version")
    
    updateRec := dbkit.NewRecord()
    updateRec.Set("version", currentVersion)
    updateRec.Set("stock", 80)
    _, err := tx.Update("products", updateRec, "id = ?", 1)
    return err  // auto rollback on version conflict
})
```

### Using with Other Features

Optimistic lock can be used together with auto timestamps and soft delete:

```go
// Configure multiple features together
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
    VersionField() string  // Returns version field name, empty string means not using
}
```

Generated DbModels can implement this interface to auto-configure optimistic lock.

---

## Transaction Handling

### Transaction
```go
func Transaction(fn func(*Tx) error) error
func (db *DB) Transaction(fn func(*Tx) error) error
```
Automatic transaction handling. Rolls back if closure returns error, commits otherwise.

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
Begins a manual transaction.

### Tx.Commit
```go
func (tx *Tx) Commit() error
```
Commits the transaction.

### Tx.Rollback
```go
func (tx *Tx) Rollback() error
```
Rolls back the transaction.

---

## Record Object

### NewRecord
```go
func NewRecord() *Record
```
Creates a new empty Record object.

### Record.Set
```go
func (r *Record) Set(column string, value interface{}) *Record
```
Sets a field value, supports chaining.

### Record.Get
```go
func (r *Record) Get(column string) interface{}
```
Gets a field value.

### Type-safe Getter Methods
```go
func (r *Record) GetString(column string) string
func (r *Record) GetInt(column string) int
func (r *Record) GetInt64(column string) int64
func (r *Record) GetFloat(column string) float64
func (r *Record) GetBool(column string) bool
func (r *Record) GetTime(column string) time.Time

// Shorthand methods
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
Checks if a field exists.

### Record.Keys
```go
func (r *Record) Keys() []string
```
Gets all field names.

### Record.Remove
```go
func (r *Record) Remove(column string)
```
Removes a field.

### Record.Clear
```go
func (r *Record) Clear()
```
Clears all fields.

### Record.ToMap
```go
func (r *Record) ToMap() map[string]interface{}
```
Converts to map.

### Record.ToJson
```go
func (r *Record) ToJson() string
```
Converts to JSON string.

### Record.FromJson
```go
func (r *Record) FromJson(jsonStr string) error
```
Parses from JSON string.

### Record.ToStruct
```go
func (r *Record) ToStruct(dest interface{}) error
```
Converts to struct.

### Record.FromStruct
```go
func (r *Record) FromStruct(src interface{}) error
```
Populates from struct.

---

## Chain Query

### Table
```go
func Table(name string) *QueryBuilder
func (db *DB) Table(name string) *QueryBuilder
func (tx *Tx) Table(name string) *QueryBuilder
```
Starts a chain query, specifying the table name.

### QueryBuilder Methods

```go
func (b *QueryBuilder) Select(columns string) *QueryBuilder    // Specify columns
func (b *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder  // WHERE condition
func (b *QueryBuilder) And(condition string, args ...interface{}) *QueryBuilder    // AND condition
func (b *QueryBuilder) OrderBy(orderBy string) *QueryBuilder   // Order by
func (b *QueryBuilder) Limit(limit int) *QueryBuilder          // Limit
func (b *QueryBuilder) Offset(offset int) *QueryBuilder        // Offset

// Execution methods
func (b *QueryBuilder) Find() ([]Record, error)                // Find multiple
func (b *QueryBuilder) Query() ([]Record, error)               // Alias for Find
func (b *QueryBuilder) FindFirst() (*Record, error)            // Find first
func (b *QueryBuilder) QueryFirst() (*Record, error)           // Alias for FindFirst
func (b *QueryBuilder) FindToDbModel(dest interface{}) error   // Find and map to struct slice
func (b *QueryBuilder) FindFirstToDbModel(dest interface{}) error // Find first and map to struct
func (b *QueryBuilder) Delete() (int64, error)                 // Delete
func (b *QueryBuilder) Paginate(page, pageSize int) (*Page[Record], error) // Paginate
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

### Join Query

Supports multiple JOIN types with chain calls:

```go
func (b *QueryBuilder) Join(table, condition string, args ...interface{}) *QueryBuilder      // JOIN
func (b *QueryBuilder) LeftJoin(table, condition string, args ...interface{}) *QueryBuilder  // LEFT JOIN
func (b *QueryBuilder) RightJoin(table, condition string, args ...interface{}) *QueryBuilder // RIGHT JOIN
func (b *QueryBuilder) InnerJoin(table, condition string, args ...interface{}) *QueryBuilder // INNER JOIN
```

**Examples:**
```go
// Simple LEFT JOIN
records, err := dbkit.Table("users").
    Select("users.name, orders.total").
    LeftJoin("orders", "users.id = orders.user_id").
    Where("orders.status = ?", "completed").
    Find()
// SQL: SELECT users.name, orders.total FROM users LEFT JOIN orders ON users.id = orders.user_id WHERE orders.status = ?
// Args: ["completed"]

// Multiple INNER JOINs
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

// JOIN with parameterized condition
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
Creates a new subquery builder.

#### Subquery Methods
```go
func (s *Subquery) Table(name string) *Subquery                           // Set table name
func (s *Subquery) Select(columns string) *Subquery                       // Set columns
func (s *Subquery) Where(condition string, args ...interface{}) *Subquery // Add condition
func (s *Subquery) OrderBy(orderBy string) *Subquery                      // Order by
func (s *Subquery) Limit(limit int) *Subquery                             // Limit
func (s *Subquery) ToSQL() (string, []interface{})                        // Generate SQL
```

#### WHERE IN Subquery
```go
func (b *QueryBuilder) WhereIn(column string, sub *Subquery) *QueryBuilder    // WHERE column IN (subquery)
func (b *QueryBuilder) WhereNotIn(column string, sub *Subquery) *QueryBuilder // WHERE column NOT IN (subquery)
```

**Examples:**
```go
// Find users with completed orders
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

// Find orders from non-banned users
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
Uses a subquery as the FROM source (derived table).

**Example:**
```go
// Query from aggregated subquery
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
Adds a subquery as a field in the SELECT clause.

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
Adds an OR condition to the query. When combined with Where, AND conditions are wrapped in parentheses to maintain correct precedence.

**Examples:**
```go
// Find orders with status active OR priority high
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
Adds grouped conditions with nested parentheses. `WhereGroup` connects with AND, `OrWhereGroup` connects with OR.

**Examples:**
```go
// OR grouped condition
records, err := dbkit.Table("table").
    Where("a = ?", 1).
    OrWhereGroup(func(qb *dbkit.QueryBuilder) *dbkit.QueryBuilder {
        return qb.Where("b = ?", 1).OrWhere("c = ?", 1)
    }).
    Find()
// SQL: SELECT * FROM table WHERE (a = ?) OR (b = ? OR c = ?)
// Args: [1, 1, 1]

// AND grouped condition
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
IN/NOT IN query with a list of values (distinct from subquery versions WhereIn/WhereNotIn).

**Examples:**
```go
// Find users with specific IDs
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
Range queries.

**Examples:**
```go
// Find users aged between 18-65
users, err := dbkit.Table("users").
    WhereBetween("age", 18, 65).
    Find()
// SQL: SELECT * FROM users WHERE age BETWEEN ? AND ?
// Args: [18, 65]

// Find products with price not between 100-500
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
NULL value checks.

**Examples:**
```go
// Find users without email
users, err := dbkit.Table("users").
    WhereNull("email").
    Find()
// SQL: SELECT * FROM users WHERE email IS NULL
// Args: []

// Find users with phone number
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
Adds a GROUP BY clause.

#### Having
```go
func (b *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder
```
Adds a HAVING clause to filter grouped results.

**Examples:**
```go
// Group orders by status
stats, err := dbkit.Table("orders").
    Select("status, COUNT(*) as count, SUM(total) as total_amount").
    GroupBy("status").
    Find()
// SQL: SELECT status, COUNT(*) as count, SUM(total) as total_amount FROM orders GROUP BY status
// Args: []

// Find users with more than 5 orders
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
Generates Go struct code from database table.

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
```

---

## Cache Operations

### SetCache
```go
func SetCache(c CacheProvider)
```
Sets the global cache provider.

### GetCache
```go
func GetCache() CacheProvider
```
Gets the current cache provider.

### SetLocalCacheConfig
```go
func SetLocalCacheConfig(cleanupInterval time.Duration)
```
Configures local cache cleanup interval.

### CreateCache
```go
func CreateCache(cacheName string, ttl time.Duration)
```
Creates a named cache with default TTL.

### CacheSet
```go
func CacheSet(cacheName, key string, value interface{}, ttl ...time.Duration)
```
Sets a cache value.

### CacheGet
```go
func CacheGet(cacheName, key string) (interface{}, bool)
```
Gets a cache value.

### CacheDelete
```go
func CacheDelete(cacheName, key string)
```
Deletes a cache key.

### CacheClear
```go
func CacheClear(cacheName string)
```
Clears a specific cache.

### CacheStatus
```go
func CacheStatus() map[string]interface{}
```
Gets cache status information.

### Cache (Chaining)
```go
func Cache(name string, ttl ...time.Duration) *DB
func (db *DB) Cache(name string, ttl ...time.Duration) *DB
func (tx *Tx) Cache(name string, ttl ...time.Duration) *Tx
```
Enables caching for queries.

**Example:**
```go
records, err := dbkit.Cache("user_cache", 5*time.Minute).Query("SELECT * FROM users")
```

### CacheProvider Interface
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

## Logging Configuration

### SetDebugMode
```go
func SetDebugMode(enabled bool)
```
Enables/disables debug mode (SQL output).

### SetLogger
```go
func SetLogger(l Logger)
```
Sets a custom logger.

### InitLoggerWithFile
```go
func InitLoggerWithFile(level string, filePath string)
```
Initializes file logging.

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

### Logging Functions
```go
func LogDebug(msg string, fields map[string]interface{})
func LogInfo(msg string, fields map[string]interface{})
func LogWarn(msg string, fields map[string]interface{})
func LogError(msg string, fields map[string]interface{})
```

---

## Utility Functions

### ToJson
```go
func ToJson(v interface{}) string
```
Converts any value to JSON string.

### ToStruct
```go
func ToStruct(record *Record, dest interface{}) error
```
Converts Record to struct.

### ToStructs
```go
func ToStructs(records []Record, dest interface{}) error
```
Converts Record slice to struct slice.

### ToRecord
```go
func ToRecord(model interface{}) *Record
```
Converts struct to Record.

### FromStruct
```go
func FromStruct(src interface{}, record *Record) error
```
Populates Record from struct.

### SnakeToCamel
```go
func SnakeToCamel(s string) string
```
Converts snake_case to CamelCase.

### ValidateTableName
```go
func ValidateTableName(table string) error
```
Validates table name.

### GenerateCacheKey
```go
func GenerateCacheKey(dbName, sql string, args ...interface{}) string
```
Generates cache key.

### SupportedDrivers
```go
func SupportedDrivers() []DriverType
```
Returns list of supported database drivers.

### IsValidDriver
```go
func IsValidDriver(driver DriverType) bool
```
Checks if driver is supported.

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
