# DBKit API Reference

[中文版](api.md) | [README](README.md) | [English README](README_EN.md)

## Table of Contents

- [Database Initialization](#database-initialization)
- [Query Operations](#query-operations)
- [Insert & Update](#insert--update)
- [Delete Operations](#delete-operations)
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
Updates records matching the condition.

**Returns:** Number of affected rows.

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
Deletes records matching the condition.

### DeleteRecord
```go
func DeleteRecord(table string, record *Record) (int64, error)
func (db *DB) DeleteRecord(table string, record *Record) (int64, error)
func (tx *Tx) DeleteRecord(table string, record *Record) (int64, error)
```
Deletes a record based on its primary key.

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
