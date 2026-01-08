package dbkit

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	// _ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
	// _ "github.com/mattn/go-sqlite3"
	// _ "github.com/microsoft/go-mssqldb"
	// _ "github.com/sijms/go-ora/v2"
)

// DriverType represents the database driver type
type DriverType string

const (
	// MySQL database driver
	MySQL DriverType = "mysql"
	// PostgreSQL database driver
	PostgreSQL DriverType = "postgres"
	// SQLite3 database driver
	SQLite3 DriverType = "sqlite3"
	// Oracle database driver
	Oracle DriverType = "oracle"
	// SQL Server database driver
	SQLServer DriverType = "sqlserver"
)

// 预编译的正则表达式，用于 sanitizeArgs 函数
// 避免每次调用都重新编译，提升性能
var (
	postgresPlaceholderRe  = regexp.MustCompile(`\$(\d+)`)
	sqlserverPlaceholderRe = regexp.MustCompile(`@p(\d+)`)
	oraclePlaceholderRe    = regexp.MustCompile(`:(\d+)`)
)

// 预编译语句缓存相关常量
const (
	stmtCacheRepository = "__dbkit_stmt_cache__" // 内部使用的缓存名称
	stmtCacheTTL        = 30 * time.Minute       // 预编译语句缓存时间
)

// Config holds the database configuration
type Config struct {
	Driver          DriverType    // Database driver type (mysql, postgres, sqlite3)
	DSN             string        // Data source name (connection string)
	MaxOpen         int           // Maximum number of open connections
	MaxIdle         int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum connection lifetime
	QueryTimeout    time.Duration // Default query timeout (0 means no timeout)
}

// SupportedDrivers returns a list of all supported database drivers
func SupportedDrivers() []DriverType {
	return []DriverType{MySQL, PostgreSQL, SQLite3, Oracle, SQLServer}
}

// IsValidDriver checks if the given driver is supported
func IsValidDriver(driver DriverType) bool {
	for _, d := range SupportedDrivers() {
		if d == driver {
			return true
		}
	}
	return false
}

// DB represents a database connection with chainable methods
type DB struct {
	dbMgr               *dbManager
	lastErr             error
	cacheRepositoryName string
	cacheTTL            time.Duration
	timeout             time.Duration // Query timeout for this instance
	cacheProvider       CacheProvider // 指定的缓存提供者（nil 表示使用默认缓存）
}

// GetConfig returns the database configuration
func (db *DB) GetConfig() (*Config, error) {
	if db == nil || db.dbMgr == nil {
		return nil, fmt.Errorf("database or database manager is nil")
	}
	return db.dbMgr.GetConfig()
}

// getTimeout returns the effective timeout for this DB instance
func (db *DB) getTimeout() time.Duration {
	if db.timeout > 0 {
		return db.timeout
	}
	if db.dbMgr != nil && db.dbMgr.config != nil && db.dbMgr.config.QueryTimeout > 0 {
		return db.dbMgr.config.QueryTimeout
	}
	return 0
}

// getContext returns a context with timeout if configured
func (db *DB) getContext() (context.Context, context.CancelFunc) {
	timeout := db.getTimeout()
	if timeout > 0 {
		return context.WithTimeout(context.Background(), timeout)
	}
	return context.Background(), func() {}
}

// getEffectiveCache 获取当前有效的缓存提供者
// 如果 DB 实例指定了缓存提供者，则使用指定的；否则使用全局默认缓存
func (db *DB) getEffectiveCache() CacheProvider {
	if db.cacheProvider != nil {
		return db.cacheProvider
	}
	return GetCache()
}

// Tx represents a database transaction with chainable methods
type Tx struct {
	tx                  *sql.Tx
	dbMgr               *dbManager
	cacheRepositoryName string
	cacheTTL            time.Duration
	timeout             time.Duration // Query timeout for this transaction
	cacheProvider       CacheProvider // 指定的缓存提供者（nil 表示使用默认缓存）
}

// getEffectiveCache 获取当前有效的缓存提供者
// 如果 Tx 实例指定了缓存提供者，则使用指定的；否则使用全局默认缓存
func (tx *Tx) getEffectiveCache() CacheProvider {
	if tx.cacheProvider != nil {
		return tx.cacheProvider
	}
	return GetCache()
}

// sqlExecutor is an internal interface for executing SQL commands
type sqlExecutor interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// sqlExecutorContext is an internal interface for executing SQL commands with context
type sqlExecutorContext interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// dbManager manages database connections
type dbManager struct {
	name            string
	config          *Config
	db              *sql.DB
	mu              sync.RWMutex
	drivers         map[string]bool
	pkCache         map[string][]string     // Table name -> PK column names
	identityCache   map[string]string       // Table name -> Identity column name
	softDeletes     *softDeleteRegistry     // Soft delete configurations
	timestamps      *timestampRegistry      // Auto timestamp configurations
	optimisticLocks *optimisticLockRegistry // Optimistic lock configurations
	// Feature flags
	enableTimestampCheck      bool // Enable auto timestamp check in Update (default: false)
	enableOptimisticLockCheck bool // Enable optimistic lock check in Update (default: false)
	enableSoftDeleteCheck     bool // Enable soft delete check in queries (default: false)
}

// MultiDBManager manages multiple database connections
type MultiDBManager struct {
	databases map[string]*dbManager
	currentDB string
	defaultDB string
	mu        sync.RWMutex
}

var (
	multiMgr *MultiDBManager
)

// init initializes the multi-database manager
func init() {
	multiMgr = &MultiDBManager{
		databases: make(map[string]*dbManager),
	}
}

// createDefaultConfig creates a Config with default settings
func createDefaultConfig(driver DriverType, dsn string, maxOpen int) *Config {
	return &Config{
		Driver:          driver,
		DSN:             dsn,
		MaxOpen:         maxOpen,
		MaxIdle:         maxOpen / 2,
		ConnMaxLifetime: time.Hour,
	}
}

// OpenDatabaseWithConfig opens a database connection with custom configuration
// This is equivalent to registering a database named "default"
func OpenDatabaseWithConfig(config *Config) error {
	return Register("default", config)
}

// OpenDatabase opens a database connection with default settings
// This is equivalent to registering a database named "default"
func OpenDatabase(driver DriverType, dsn string, maxOpen int) error {
	config := createDefaultConfig(driver, dsn, maxOpen)
	return OpenDatabaseWithConfig(config)
}

// OpenDatabaseWithDBName opens a database connection with a name (multi-database mode)
func OpenDatabaseWithDBName(dbname string, driver DriverType, dsn string, maxOpen int) error {
	config := createDefaultConfig(driver, dsn, maxOpen)
	return Register(dbname, config)
}

// Register registers a database connection with a name (multi-database mode)
func Register(dbname string, config *Config) error {
	dbMgr := &dbManager{
		name:          dbname,
		config:        config,
		pkCache:       make(map[string][]string),
		identityCache: make(map[string]string),
	}

	if err := dbMgr.initDB(); err != nil {
		return err
	}

	multiMgr.mu.Lock()
	multiMgr.databases[dbname] = dbMgr
	// Set as default database if it's the first one
	if multiMgr.defaultDB == "" {
		multiMgr.defaultDB = dbname
		multiMgr.currentDB = dbname
	}
	multiMgr.mu.Unlock()

	return nil
}

// Use switches to a different database by name and returns a DB object for chainable calls
// This is a convenience method that avoids panicking for fluent API usage.
// If the database is not found or another error occurs, the error is stored in the returned DB object
// and will be returned by subsequent operations.
func Use(dbname string) *DB {
	db, err := UseWithError(dbname)
	if err != nil {
		return &DB{lastErr: err}
	}
	return db
}

// UseWithError returns a DB object for the specified database by name
func UseWithError(dbname string) (*DB, error) {
	multiMgr.mu.RLock()
	dbMgr, exists := multiMgr.databases[dbname]
	multiMgr.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("database '%s' not found", dbname)
	}

	return &DB{dbMgr: dbMgr}, nil
}

var (
	// ErrNotInitialized is returned when an operation is performed on an uninitialized database
	ErrNotInitialized = fmt.Errorf("dbkit: database not initialized. Please call dbkit.OpenDatabase() before using dbkit operations")
)

// defaultDB returns the default DB object (first registered database or single database mode)
func defaultDB() (*DB, error) {
	dbMgr, err := safeGetCurrentDB()
	if err != nil {
		return nil, err
	}
	return &DB{dbMgr: dbMgr}, nil
}

// --- Internal Helper Methods on dbManager to unify DB and Tx logic ---

func (mgr *dbManager) prepareQuerySQL(querySQL string, args ...interface{}) (string, []interface{}) {
	driver := mgr.config.Driver
	lowerSQL := strings.ToLower(querySQL)

	// 处理 Oracle 和 SQL Server 的 LIMIT 语法
	if driver == Oracle {
		if limitIndex := strings.LastIndex(lowerSQL, " limit "); limitIndex != -1 {
			limitStr := strings.TrimSpace(querySQL[limitIndex+6:])
			querySQL = fmt.Sprintf("SELECT * FROM (%s) WHERE ROWNUM <= %s", querySQL[:limitIndex], limitStr)
		}
	} else if driver == SQLServer {
		if limitIndex := strings.LastIndex(lowerSQL, " limit "); limitIndex != -1 {
			limitStr := strings.TrimSpace(querySQL[limitIndex+6:])
			sqlPart := querySQL[:limitIndex]
			if selectIndex := strings.Index(strings.ToLower(sqlPart), "select "); selectIndex != -1 {
				querySQL = fmt.Sprintf("SELECT TOP %s %s", limitStr, sqlPart[selectIndex+7:])
			}
		}
	}

	querySQL = mgr.convertPlaceholder(querySQL, driver)
	args = mgr.sanitizeArgs(querySQL, args)
	return querySQL, args
}

// getOrPrepareStmt 获取或创建预编译语句（内部方法）
// 返回值：stmt, fromCache, error
func (mgr *dbManager) getOrPrepareStmt(sqlStr string) (*sql.Stmt, bool, error) {
	// 构造缓存键：数据库名称 + SQL 语句
	// 这样可以确保不同数据库的预编译语句不会冲突
	cacheKey := mgr.name + ":" + sqlStr

	// 1. 尝试从 localCache 获取
	if cached, ok := GetLocalCacheInstance().CacheGet(stmtCacheRepository, cacheKey); ok {
		if stmt, ok := cached.(*sql.Stmt); ok {
			return stmt, true, nil // 从缓存获取
		}
	}

	// 2. 缓存未命中，创建新的预编译语句
	stmt, err := mgr.db.Prepare(sqlStr)
	if err != nil {
		return nil, false, err
	}

	// 3. 存入 localCache
	GetLocalCacheInstance().CacheSet(stmtCacheRepository, cacheKey, stmt, stmtCacheTTL)

	return stmt, false, nil // 新创建的
}

// clearStmtCache 清空预编译语句缓存（内部方法，用于数据库关闭时）
func (mgr *dbManager) clearStmtCache() {
	// 获取所有缓存的语句并关闭
	lc := GetLocalCacheInstance().(*localCache)
	if store, ok := lc.stores.Load(stmtCacheRepository); ok {
		s := store.(*sync.Map)
		s.Range(func(key, value interface{}) bool {
			if entry, ok := value.(cacheEntry); ok {
				if stmt, ok := entry.value.(*sql.Stmt); ok {
					stmt.Close()
				}
			}
			return true
		})
	}

	// 清空缓存
	GetLocalCacheInstance().CacheClearRepository(stmtCacheRepository)
}

// isStmtInvalidError 检查是否是语句失效错误
func isStmtInvalidError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "invalid connection") ||
		strings.Contains(errStr, "bad connection") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset")
}

func (mgr *dbManager) query(executor sqlExecutor, querySQL string, args ...interface{}) ([]Record, error) {
	return mgr.queryWithContext(context.Background(), executor, querySQL, args...)
}

func (mgr *dbManager) queryWithContext(ctx context.Context, executor sqlExecutor, querySQL string, args ...interface{}) ([]Record, error) {
	querySQL, args = mgr.prepareQuerySQL(querySQL, args...)
	start := time.Now()

	var rows *sql.Rows
	var err error

	// 只有当 executor 是 *sql.DB 时才使用预编译语句缓存
	// 事务（*sql.Tx）不使用缓存，因为事务有自己的生命周期
	if db, ok := executor.(*sql.DB); ok && db == mgr.db {
		// 使用缓存的预编译语句
		stmt, fromCache, stmtErr := mgr.getOrPrepareStmt(querySQL)
		if stmtErr != nil {
			mgr.logTrace(start, querySQL, args, stmtErr)
			return nil, stmtErr
		}

		// 执行查询（使用 context）
		rows, err = stmt.QueryContext(ctx, args...)

		// 如果执行失败且可能是语句失效，从缓存移除
		if err != nil && !fromCache {
			// 新创建的语句出错，不需要特殊处理
		} else if err != nil && isStmtInvalidError(err) {
			cacheKey := mgr.name + ":" + querySQL
			GetLocalCacheInstance().CacheDelete(stmtCacheRepository, cacheKey)
		}
	} else {
		// 事务或其他 executor，使用原有逻辑
		if execCtx, ok := executor.(sqlExecutorContext); ok {
			rows, err = execCtx.QueryContext(ctx, querySQL, args...)
		} else {
			rows, err = executor.Query(querySQL, args...)
		}
	}

	mgr.logTrace(start, querySQL, args, err)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := scanRecords(rows, mgr.config.Driver)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (mgr *dbManager) queryFirst(executor sqlExecutor, querySQL string, args ...interface{}) (*Record, error) {
	return mgr.queryFirstWithContext(context.Background(), executor, querySQL, args...)
}

func (mgr *dbManager) queryFirstWithContext(ctx context.Context, executor sqlExecutor, querySQL string, args ...interface{}) (*Record, error) {
	querySQL = mgr.addLimitOne(querySQL)
	return mgr.queryFirstInternalWithContext(ctx, executor, querySQL, args...)
}

func (mgr *dbManager) queryFirstInternal(executor sqlExecutor, querySQL string, args ...interface{}) (*Record, error) {
	return mgr.queryFirstInternalWithContext(context.Background(), executor, querySQL, args...)
}

func (mgr *dbManager) queryFirstInternalWithContext(ctx context.Context, executor sqlExecutor, querySQL string, args ...interface{}) (*Record, error) {
	records, err := mgr.queryWithContext(ctx, executor, querySQL, args...)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}

func (mgr *dbManager) queryMap(executor sqlExecutor, querySQL string, args ...interface{}) ([]map[string]interface{}, error) {
	return mgr.queryMapWithContext(context.Background(), executor, querySQL, args...)
}

func (mgr *dbManager) queryMapWithContext(ctx context.Context, executor sqlExecutor, querySQL string, args ...interface{}) ([]map[string]interface{}, error) {
	querySQL, args = mgr.prepareQuerySQL(querySQL, args...)
	start := time.Now()

	var rows *sql.Rows
	var err error

	// 只有当 executor 是 *sql.DB 时才使用预编译语句缓存
	if db, ok := executor.(*sql.DB); ok && db == mgr.db {
		// 使用缓存的预编译语句
		stmt, fromCache, stmtErr := mgr.getOrPrepareStmt(querySQL)
		if stmtErr != nil {
			mgr.logTrace(start, querySQL, args, stmtErr)
			return nil, stmtErr
		}

		// 执行查询（使用 context）
		rows, err = stmt.QueryContext(ctx, args...)

		// 如果执行失败且可能是语句失效，从缓存移除
		if err != nil && !fromCache {
			// 新创建的语句出错，不需要特殊处理
		} else if err != nil && isStmtInvalidError(err) {
			cacheKey := mgr.name + ":" + querySQL
			GetLocalCacheInstance().CacheDelete(stmtCacheRepository, cacheKey)
		}
	} else {
		// 事务或其他 executor，使用原有逻辑
		if execCtx, ok := executor.(sqlExecutorContext); ok {
			rows, err = execCtx.QueryContext(ctx, querySQL, args...)
		} else {
			rows, err = executor.Query(querySQL, args...)
		}
	}

	mgr.logTrace(start, querySQL, args, err)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := scanMaps(rows, mgr.config.Driver)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (mgr *dbManager) addLimitOne(querySQL string) string {
	driver := mgr.config.Driver
	lowerSQL := strings.ToLower(strings.TrimSpace(querySQL))

	// Check if already has limit
	if strings.Contains(lowerSQL, " limit ") ||
		strings.Contains(lowerSQL, " top ") ||
		strings.Contains(lowerSQL, " rownum ") ||
		strings.Contains(lowerSQL, " fetch first ") ||
		strings.Contains(lowerSQL, " fetch next ") {
		return querySQL
	}

	switch driver {
	case MySQL, PostgreSQL, SQLite3:
		return querySQL + " LIMIT 1"
	case Oracle:
		return fmt.Sprintf("SELECT * FROM (%s) WHERE ROWNUM <= 1", querySQL)
	case SQLServer:
		if strings.HasPrefix(lowerSQL, "select ") {
			// Basic SELECT TOP 1 implementation
			// Check for DISTINCT to avoid invalid syntax like "SELECT TOP 1 DISTINCT"
			if strings.HasPrefix(lowerSQL, "select distinct ") {
				return "SELECT DISTINCT TOP 1 " + querySQL[16:]
			}
			return "SELECT TOP 1 " + querySQL[7:]
		}
		return querySQL
	default:
		return querySQL
	}
}

func (mgr *dbManager) exec(executor sqlExecutor, querySQL string, args ...interface{}) (sql.Result, error) {
	return mgr.execWithContext(context.Background(), executor, querySQL, args...)
}

func (mgr *dbManager) execWithContext(ctx context.Context, executor sqlExecutor, querySQL string, args ...interface{}) (sql.Result, error) {
	querySQL = mgr.convertPlaceholder(querySQL, mgr.config.Driver)
	args = mgr.sanitizeArgs(querySQL, args)
	start := time.Now()

	var result sql.Result
	var err error

	// 只有当 executor 是 *sql.DB 时才使用预编译语句缓存
	if db, ok := executor.(*sql.DB); ok && db == mgr.db {
		// 使用缓存的预编译语句
		stmt, fromCache, stmtErr := mgr.getOrPrepareStmt(querySQL)
		if stmtErr != nil {
			mgr.logTrace(start, querySQL, args, stmtErr)
			return nil, stmtErr
		}

		// 执行命令（使用 context）
		result, err = stmt.ExecContext(ctx, args...)

		// 如果执行失败且可能是语句失效，从缓存移除
		if err != nil && !fromCache {
			// 新创建的语句出错，不需要特殊处理
		} else if err != nil && isStmtInvalidError(err) {
			cacheKey := mgr.name + ":" + querySQL
			GetLocalCacheInstance().CacheDelete(stmtCacheRepository, cacheKey)
		}
	} else {
		// 事务或其他 executor，使用原有逻辑
		if execCtx, ok := executor.(sqlExecutorContext); ok {
			result, err = execCtx.ExecContext(ctx, querySQL, args...)
		} else {
			result, err = executor.Exec(querySQL, args...)
		}
	}

	mgr.logTrace(start, querySQL, args, err)

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (mgr *dbManager) getIdentityColumn(executor sqlExecutor, table string) string {
	mgr.mu.RLock()
	if col, ok := mgr.identityCache[table]; ok {
		mgr.mu.RUnlock()
		return col
	}
	mgr.mu.RUnlock()

	var identityCol string
	driver := mgr.config.Driver

	if driver == SQLServer {
		// 查询 SQL Server 的标识列
		query := `
			SELECT name 
			FROM sys.columns 
			WHERE object_id = OBJECT_ID(?) AND is_identity = 1`
		records, err := mgr.query(executor, query, table)
		if err == nil && len(records) > 0 {
			identityCol = records[0].GetString("name")
		}
	} else if driver == MySQL {
		// 查询 MySQL 的自增列
		query := "SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND EXTRA = 'auto_increment'"
		records, err := mgr.query(executor, query, table)
		if err == nil && len(records) > 0 {
			identityCol = records[0].GetString("COLUMN_NAME")
		}
	} else if driver == PostgreSQL {
		// 查询 PostgreSQL 的自增列 (SERIAL 或 IDENTITY)
		query := `
			SELECT a.attname
			FROM pg_attribute a
			JOIN pg_class c ON a.attrelid = c.oid
			JOIN pg_namespace n ON c.relnamespace = n.oid
			WHERE c.relname = ? 
			  AND n.nspname = current_schema()
			  AND (a.attidentity != '' OR EXISTS (
				  SELECT 1 FROM pg_attrdef d WHERE d.adrelid = a.attrelid AND d.adnum = a.attnum AND d.adsrc LIKE 'nextval%'
			  ))`
		// 注意：PostgreSQL 的 adsrc 在新版本中可能不可用，使用 pg_get_expr
		query = `
			SELECT a.attname
			FROM pg_attribute a
			WHERE a.attrelid = ?::regclass 
			  AND a.attnum > 0 
			  AND NOT a.attisdropped
			  AND (a.attidentity IN ('a', 'd') OR EXISTS (
				  SELECT 1 FROM pg_attrdef d WHERE d.adrelid = a.attrelid AND d.adnum = a.attnum AND pg_get_expr(d.adbin, d.adrelid) LIKE 'nextval%'
			  ))`
		records, err := mgr.query(executor, query, table)
		if err == nil && len(records) > 0 {
			identityCol = records[0].GetString("attname")
		}
	} else if driver == Oracle {
		// 查询 Oracle 的自增列 (IDENTITY)
		// 尝试从 user_tab_cols 查询，这在 12c+ 中更通用
		// 注意：如果 Oracle 版本低于 12c，IDENTITY_COLUMN 可能不存在，会导致 ORA-00904
		query := "SELECT COLUMN_NAME FROM USER_TAB_COLS WHERE TABLE_NAME = ? AND IDENTITY_COLUMN = 'YES'"
		// 我们使用一个不打印错误的查询方式，或者捕获错误
		rows, err := mgr.db.Query(query, strings.ToUpper(table))
		if err == nil {
			defer rows.Close()
			if rows.Next() {
				var colName string
				if err := rows.Scan(&colName); err == nil {
					identityCol = colName
				}
			}
		}

		if identityCol == "" {
			// 如果上述查询失败或未找到，尝试查询 USER_TAB_IDENTITY_COLS
			query = "SELECT COLUMN_NAME FROM USER_TAB_IDENTITY_COLS WHERE TABLE_NAME = ?"
			rows, err := mgr.db.Query(query, strings.ToUpper(table))
			if err == nil {
				defer rows.Close()
				if rows.Next() {
					var colName string
					if err := rows.Scan(&colName); err == nil {
						identityCol = colName
					}
				}
			}
		}
	} else if driver == SQLite3 {
		// SQLite3 中，INTEGER PRIMARY KEY 自动就是自增的
		// 我们检查是否有 INTEGER 类型的 PK
		query := fmt.Sprintf("PRAGMA table_info(%s)", table)
		records, err := mgr.query(executor, query)
		if err == nil {
			for _, r := range records {
				if r.GetInt("pk") == 1 && strings.EqualFold(r.GetString("type"), "INTEGER") {
					identityCol = r.GetString("name")
					break
				}
			}
		}
	}

	mgr.mu.Lock()
	mgr.identityCache[table] = identityCol
	mgr.mu.Unlock()

	return identityCol
}

func (mgr *dbManager) getPrimaryKeys(executor sqlExecutor, table string) ([]string, error) {
	mgr.mu.RLock()
	if pks, ok := mgr.pkCache[table]; ok {
		mgr.mu.RUnlock()
		return pks, nil
	}
	mgr.mu.RUnlock()

	var pks []string
	driver := mgr.config.Driver

	switch driver {
	case MySQL:
		query := "SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE WHERE TABLE_SCHEMA = DATABASE() AND CONSTRAINT_NAME = 'PRIMARY' AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION"
		records, err := mgr.query(executor, query, table)
		if err == nil {
			for _, r := range records {
				if val := r.Get("COLUMN_NAME"); val != nil {
					pks = append(pks, fmt.Sprintf("%v", val))
				}
			}
		}
	case PostgreSQL:
		query := `
			SELECT a.attname 
			FROM pg_index i 
			JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey) 
			WHERE i.indrelid = ?::regclass AND i.indisprimary
			ORDER BY a.attnum`
		records, err := mgr.query(executor, query, table)
		if err == nil {
			for _, r := range records {
				if val := r.Get("attname"); val != nil {
					pks = append(pks, fmt.Sprintf("%v", val))
				}
			}
		}
	case SQLite3:
		query := fmt.Sprintf("PRAGMA table_info(%s)", table)
		records, err := mgr.query(executor, query)
		if err == nil {
			type pkInfo struct {
				name string
				pos  int
			}
			var infos []pkInfo
			for _, r := range records {
				is_pk := r.GetInt("pk")
				if is_pk > 0 {
					infos = append(infos, pkInfo{name: r.GetString("name"), pos: int(is_pk)})
				}
			}
			// SQLite pk column order is defined by is_pk value
			sort.Slice(infos, func(i, j int) bool {
				return infos[i].pos < infos[j].pos
			})
			for _, info := range infos {
				pks = append(pks, info.name)
			}
		}
	case Oracle:
		upperTable := strings.ToUpper(table)
		query := `
			SELECT cols.column_name
			FROM user_constraints cons
			JOIN user_cons_columns cols ON cons.constraint_name = cols.constraint_name
			WHERE cons.table_name = ? AND cons.constraint_type = 'P'
			ORDER BY cols.position`
		records, err := mgr.query(executor, query, upperTable)
		if err != nil || len(records) == 0 {
			// 如果查不到，再尝试从 all_constraints 查
			query = `
				SELECT cols.column_name 
				FROM all_constraints cons, all_cons_columns cols 
				WHERE cols.table_name = ?
				  AND cons.constraint_type = 'P' 
				  AND cons.constraint_name = cols.constraint_name 
				  AND cons.owner = cols.owner 
				  AND cons.owner = (SELECT user FROM dual)
				ORDER BY cols.position`
			records, _ = mgr.query(executor, query, upperTable)
		}
		for _, r := range records {
			if val := r.Get("COLUMN_NAME"); val != nil {
				pks = append(pks, fmt.Sprintf("%v", val))
			}
		}
	case SQLServer:
		query := `
			SELECT k.COLUMN_NAME, t.CONSTRAINT_TYPE
			FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE k
			JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS t 
			  ON k.CONSTRAINT_NAME = t.CONSTRAINT_NAME
			WHERE k.TABLE_NAME = ?`
		records, err := mgr.query(executor, query, table)
		if err == nil {
			for _, r := range records {
				consType := fmt.Sprintf("%v", r.Get("CONSTRAINT_TYPE"))
				if strings.EqualFold(consType, "PRIMARY KEY") {
					if val := r.Get("COLUMN_NAME"); val != nil {
						pks = append(pks, fmt.Sprintf("%v", val))
					}
				}
			}
		}
	}

	// 如果没有找到主键，则 pks 为空切片

	mgr.mu.Lock()
	mgr.pkCache[table] = pks
	mgr.mu.Unlock()

	return pks, nil
}

func (mgr *dbManager) getRecordID(record *Record, pks []string) (int64, bool) {
	if len(pks) == 0 || record == nil {
		return 0, false
	}

	firstPK := pks[0]
	for k, v := range record.columns {
		if strings.EqualFold(k, firstPK) {
			// 尝试多种方式转换主键值为 int64
			switch val := v.(type) {
			case int:
				return int64(val), true
			case int32:
				return int64(val), true
			case int64:
				return val, true
			case uint:
				return int64(val), true
			case uint32:
				return int64(val), true
			case uint64:
				return int64(val), true
			case float32:
				return int64(val), true
			case float64:
				return int64(val), true
			case string:
				if i, err := strconv.ParseInt(val, 10, 64); err == nil {
					return i, true
				}
			default:
				if i, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64); err == nil {
					return i, true
				}
			}
			break
		}
	}
	return 0, false
}

func (mgr *dbManager) save(executor sqlExecutor, table string, record *Record) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if record == nil || len(record.columns) == 0 {
		return 0, fmt.Errorf("record is empty")
	}

	pks, _ := mgr.getPrimaryKeys(executor, table)
	if len(pks) == 0 {
		// 没有主键，直接执行插入
		return mgr.insert(executor, table, record)
	}

	// 检查 Record 中是否包含所有主键字段
	pkConditions := []string{}
	pkValues := []interface{}{}
	allPKsFound := true

	for _, pk := range pks {
		found := false
		var val interface{}
		// 尝试大小写敏感查找
		if v, ok := record.columns[pk]; ok {
			val = v
			found = true
		} else {
			// 尝试不区分大小写查找
			for k, v := range record.columns {
				if strings.EqualFold(k, pk) {
					val = v
					found = true
					break
				}
			}
		}

		if !found || val == nil {
			allPKsFound = false
			break
		}
		pkConditions = append(pkConditions, fmt.Sprintf("%s = ?", pk))
		pkValues = append(pkValues, val)
	}

	if allPKsFound {
		// Check if optimistic lock is configured and record has version field
		// If so, we need to use update instead of upsert to properly check version
		config := mgr.getOptimisticLockConfig(table)
		if config != nil && config.VersionField != "" {
			if _, hasVersion := mgr.getVersionFromRecord(table, record); hasVersion {
				// Record has version field, use update with version check
				where := strings.Join(pkConditions, " AND ")
				updateRecord := NewRecord()
				columns, _ := mgr.getOrderedColumns(record)
				for _, k := range columns {
					v := record.columns[k]
					isPK := false
					for _, pk := range pks {
						if strings.EqualFold(k, pk) {
							isPK = true
							break
						}
					}
					if !isPK {
						updateRecord.Set(k, v)
					}
				}
				if len(updateRecord.columns) > 0 {
					return mgr.update(executor, table, updateRecord, where, pkValues...)
				}
				return 0, nil
			}
		}

		// 如果是 MySQL, PostgreSQL, SQLite, Oracle, SQLServer，使用原生的 Upsert 语法
		driver := mgr.config.Driver
		if driver == MySQL || driver == PostgreSQL || driver == SQLite3 || driver == Oracle || driver == SQLServer {
			return mgr.nativeUpsert(executor, table, record, pks)
		}

		// 所有主键字段都存在，检查记录是否存在
		where := strings.Join(pkConditions, " AND ")
		exists, err := mgr.exists(executor, table, where, pkValues...)
		if err == nil && exists {
			// 记录存在，执行更新
			updateRecord := NewRecord()
			columns, _ := mgr.getOrderedColumns(record)
			for _, k := range columns {
				v := record.columns[k]
				isPK := false
				for _, pk := range pks {
					if strings.EqualFold(k, pk) {
						isPK = true
						break
					}
				}
				if !isPK {
					updateRecord.Set(k, v)
				}
			}
			// 如果除了主键还有其他字段，则执行更新
			if len(updateRecord.columns) > 0 {
				return mgr.update(executor, table, updateRecord, where, pkValues...)
			}
			return 0, nil // 只有主键且已存在，无需更新
		}
	}

	// 记录不存在或不包含完整主键，执行插入
	return mgr.insert(executor, table, record)
}

// getOrderedColumns returns column names and their corresponding values from a record
// Note: columns are not sorted for better performance. SQL execution is not affected by column order.
func (mgr *dbManager) getOrderedColumns(record *Record) ([]string, []interface{}) {
	if record == nil || len(record.columns) == 0 {
		return nil, nil
	}
	columns := make([]string, 0, len(record.columns))
	values := make([]interface{}, 0, len(record.columns))
	for col, val := range record.columns {
		columns = append(columns, col)
		values = append(values, val)
	}
	return columns, values
}

func (mgr *dbManager) nativeUpsert(executor sqlExecutor, table string, record *Record, pks []string) (int64, error) {
	driver := mgr.config.Driver

	// 如果是 Oracle 或 SQL Server，使用 MERGE 语句
	if driver == Oracle || driver == SQLServer {
		return mgr.mergeUpsert(executor, table, record, pks)
	}

	// Apply version initialization for optimistic lock (for INSERT part of upsert)
	mgr.applyVersionInit(table, record)

	columns, values := mgr.getOrderedColumns(record)
	var placeholders []string
	for range columns {
		placeholders = append(placeholders, "?")
	}

	identityCol := mgr.getIdentityColumn(executor, table)
	_ = identityCol // 目前在 nativeUpsert 中仅作为保留，用于后续可能的扩展

	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, joinStrings(columns), joinStrings(placeholders))

	var updateClauses []string
	for _, col := range columns {
		isPK := false
		for _, pk := range pks {
			if strings.EqualFold(col, pk) {
				isPK = true
				break
			}
		}
		if !isPK {
			if driver == MySQL {
				updateClauses = append(updateClauses, fmt.Sprintf("%s = VALUES(%s)", col, col))
			} else { // PostgreSQL, SQLite
				updateClauses = append(updateClauses, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
			}
		}
	}

	// 如果有 ON DUPLICATE/CONFLICT 子句，我们需要确保在插入部分正确处理自增列
	// 对于 MySQL/PG/SQLite 的 nativeUpsert，如果 record 中包含自增列，
	// 数据库通常会自动处理（如果为 null 或 0 则自增，如果提供了值则使用该值）。
	// 这与 MERGE 语法强制要求排除 IDENTITY 不同。
	// 因此这里保持现状，允许 INSERT 部分包含所有 Record 字段。

	if len(updateClauses) > 0 {
		if driver == MySQL {
			sqlStr += " ON DUPLICATE KEY UPDATE " + joinStrings(updateClauses)
		} else { // PostgreSQL, SQLite
			sqlStr += fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s", joinStrings(pks), joinStrings(updateClauses))
		}
	} else {
		// 如果只有主键字段，执行一个无意义的更新以确保能返回 ID
		if driver == MySQL {
			sqlStr += fmt.Sprintf(" ON DUPLICATE KEY UPDATE %s = %s", pks[0], pks[0])
		} else {
			sqlStr += fmt.Sprintf(" ON CONFLICT (%s) DO UPDATE SET %s = EXCLUDED.%s", joinStrings(pks), pks[0], pks[0])
		}
	}

	sqlStr = mgr.convertPlaceholder(sqlStr, driver)
	values = mgr.sanitizeArgs(sqlStr, values)

	// 处理 PostgreSQL 的 ID 返回
	if driver == PostgreSQL {
		if len(pks) == 1 && strings.EqualFold(pks[0], "id") {
			sqlStr += " RETURNING id"
			var id int64
			start := time.Now()
			err := executor.QueryRow(sqlStr, values...).Scan(&id)
			mgr.logTrace(start, sqlStr, values, err)
			if err != nil {
				return 0, err
			}
			return id, nil
		}
	}

	start := time.Now()
	res, err := executor.Exec(sqlStr, values...)
	mgr.logTrace(start, sqlStr, values, err)
	if err != nil {
		return 0, err
	}

	// 1. 如果 Record 中已经包含了主键（通常是 Update 场景），优先返回它
	// 这样可以避免某些数据库（如 SQLite）在 Upsert 后 LastInsertId 返回不相关的值
	if id, ok := mgr.getRecordID(record, pks); ok {
		rows, _ := res.RowsAffected()
		if rows > 0 {
			return id, nil
		}
	}

	// 2. 否则对于 MySQL/SQLite 返回最后插入的 ID（通常是 Insert 场景）
	if driver == MySQL || driver == SQLite3 {
		id, _ := res.LastInsertId()
		if id > 0 {
			return id, nil
		}
	}

	return res.RowsAffected()
}

func (mgr *dbManager) mergeUpsert(executor sqlExecutor, table string, record *Record, pks []string) (int64, error) {
	driver := mgr.config.Driver

	// Apply version initialization for optimistic lock (for INSERT part of merge)
	mgr.applyVersionInit(table, record)

	columns, values := mgr.getOrderedColumns(record)

	// 构造 USING 子句
	var selectCols []string
	for _, col := range columns {
		selectCols = append(selectCols, "? AS "+col)
	}

	usingSQL := "SELECT " + strings.Join(selectCols, ", ")
	if driver == Oracle {
		usingSQL += " FROM DUAL"
	}

	// 构造 ON 子句
	var onClauses []string
	for _, pk := range pks {
		onClauses = append(onClauses, fmt.Sprintf("t.%s = s.%s", pk, pk))
	}

	// 构造 UPDATE 子句
	var updateClauses []string
	for _, col := range columns {
		isPK := false
		for _, pk := range pks {
			if strings.EqualFold(col, pk) {
				isPK = true
				break
			}
		}
		if !isPK {
			updateClauses = append(updateClauses, fmt.Sprintf("t.%s = s.%s", col, col))
		}
	}

	// 如果只有主键字段，执行一个无意义的更新以确保能触发更新逻辑
	if len(updateClauses) == 0 && len(pks) > 0 {
		updateClauses = append(updateClauses, fmt.Sprintf("t.%s = s.%s", pks[0], pks[0]))
	}

	sqlStr := fmt.Sprintf("MERGE INTO %s t USING (%s) s ON (%s)", table, usingSQL, strings.Join(onClauses, " AND "))

	if len(updateClauses) > 0 {
		sqlStr += " WHEN MATCHED THEN UPDATE SET " + strings.Join(updateClauses, ", ")
	}

	// 构造 INSERT 子句
	var insertCols []string
	var insertVals []string
	identityCol := mgr.getIdentityColumn(executor, table)

	for _, col := range columns {
		isIdentity := false
		// 对于支持 IDENTITY/自增的数据库，在 MERGE/Upsert 插入部分排除自增列
		// 这样数据库会自动生成值，或者避免违反 "GENERATED ALWAYS" 限制
		if identityCol != "" && strings.EqualFold(col, identityCol) {
			isIdentity = true
		}

		if !isIdentity {
			insertCols = append(insertCols, col)
			insertVals = append(insertVals, "s."+col)
		}
	}

	sqlStr += fmt.Sprintf(" WHEN NOT MATCHED THEN INSERT (%s) VALUES (%s)",
		strings.Join(insertCols, ", "),
		strings.Join(insertVals, ", "))

	if driver == SQLServer {
		sqlStr += ";" // SQL Server 的 MERGE 语句必须以分号结束
	}

	sqlStr = mgr.convertPlaceholder(sqlStr, driver)
	values = mgr.sanitizeArgs(sqlStr, values)

	// 对于 SQL Server，如果我们需要获取生成的 ID，可以使用 OUTPUT 子句
	// 但这会改变执行方式（从 Exec 变为 QueryRow），为了保持简单，我们先解决报错问题
	start := time.Now()
	res, err := executor.Exec(sqlStr, values...)
	mgr.logTrace(start, sqlStr, values, err)
	if err != nil {
		return 0, err
	}

	// 如果是 SQL Server 且执行的是 MERGE (Save)，RowsAffected 可能无法准确反映新生成的 ID
	// 但至少现在不会报错了。如果用户提供了主键值，我们返回它。
	if id, ok := mgr.getRecordID(record, pks); ok {
		return id, nil
	}

	return res.RowsAffected()
}

func (mgr *dbManager) insert(executor sqlExecutor, table string, record *Record) (int64, error) {
	return mgr.insertWithOptions(executor, table, record, false)
}

func (mgr *dbManager) insertWithOptions(executor sqlExecutor, table string, record *Record, skipTimestamps bool) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if record == nil || len(record.columns) == 0 {
		return 0, fmt.Errorf("record is empty")
	}

	// Apply created_at timestamp
	mgr.applyCreatedAtTimestamp(table, record, skipTimestamps)

	// Apply version initialization for optimistic lock
	mgr.applyVersionInit(table, record)

	columns, values := mgr.getOrderedColumns(record)
	var placeholders []string
	for range columns {
		placeholders = append(placeholders, "?")
	}

	driver := mgr.config.Driver

	querySQL := fmt.Sprintf("INSERT INTO %s (%s)", table, joinStrings(columns))

	if driver == PostgreSQL {
		pks, _ := mgr.getPrimaryKeys(executor, table)
		// 只有当存在单列主键且名为 id 时才使用 RETURNING id
		if len(pks) == 1 && strings.EqualFold(pks[0], "id") {
			querySQL += fmt.Sprintf(" VALUES (%s) RETURNING %s", joinStrings(placeholders), pks[0])
			querySQL = mgr.convertPlaceholder(querySQL, driver)
			values = mgr.sanitizeArgs(querySQL, values)
			var id int64
			start := time.Now()
			err := executor.QueryRow(querySQL, values...).Scan(&id)
			mgr.logTrace(start, querySQL, values, err)
			if err != nil {
				return 0, err
			}
			return id, nil
		}
		// 否则执行普通插入
		querySQL += fmt.Sprintf(" VALUES (%s)", joinStrings(placeholders))
		querySQL = mgr.convertPlaceholder(querySQL, driver)
		values = mgr.sanitizeArgs(querySQL, values)
		start := time.Now()
		res, err := executor.Exec(querySQL, values...)
		mgr.logTrace(start, querySQL, values, err)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	}

	if driver == SQLServer {
		pks, _ := mgr.getPrimaryKeys(executor, table)
		identityCol := mgr.getIdentityColumn(executor, table)
		// 只有当确定存在标识列且它是唯一主键时，才使用 SCOPE_IDENTITY
		if len(pks) == 1 && identityCol != "" && strings.EqualFold(pks[0], identityCol) {
			querySQL += fmt.Sprintf(" VALUES (%s); SELECT SCOPE_IDENTITY()", joinStrings(placeholders))
			querySQL = mgr.convertPlaceholder(querySQL, driver)
			values = mgr.sanitizeArgs(querySQL, values)
			var id int64
			start := time.Now()
			err := executor.QueryRow(querySQL, values...).Scan(&id)
			mgr.logTrace(start, querySQL, values, err)
			if err == nil {
				return id, nil
			}
		}

		querySQL += fmt.Sprintf(" VALUES (%s)", joinStrings(placeholders))
		querySQL = mgr.convertPlaceholder(querySQL, driver)
		values = mgr.sanitizeArgs(querySQL, values)
		start := time.Now()
		res, err := executor.Exec(querySQL, values...)
		mgr.logTrace(start, querySQL, values, err)
		if err != nil {
			return 0, err
		}

		// 如果主键存在且非自增，尝试返回主键值
		if id, ok := mgr.getRecordID(record, pks); ok {
			return id, nil
		}
		return res.RowsAffected()
	}

	if driver == Oracle {
		pks, _ := mgr.getPrimaryKeys(executor, table)

		// 1. 如果 Record 中已经包含了主键，优先执行并返回该主键
		if id, ok := mgr.getRecordID(record, pks); ok {
			querySQL += fmt.Sprintf(" VALUES (%s)", joinStrings(placeholders))
			querySQL = mgr.convertPlaceholder(querySQL, driver)
			values = mgr.sanitizeArgs(querySQL, values)
			start := time.Now()
			_, err := executor.Exec(querySQL, values...)
			mgr.logTrace(start, querySQL, values, err)
			if err != nil {
				return 0, err
			}
			return id, nil
		}

		// 2. 否则尝试使用 RETURNING 获取新生成的 ID
		if len(pks) == 1 {
			returningSql := querySQL + fmt.Sprintf(" VALUES (%s) RETURNING %s INTO ?", joinStrings(placeholders), pks[0])
			returningSql = mgr.convertPlaceholder(returningSql, driver)
			values = mgr.sanitizeArgs(returningSql, values)
			start := time.Now()

			var lastID int64
			argsWithOut := append(values, sql.Out{Dest: &lastID})
			_, err := executor.Exec(returningSql, argsWithOut...)
			mgr.logTrace(start, returningSql, values, err)
			if err == nil {
				return lastID, nil
			}
		}

		// 3. 最后退回到普通插入
		querySQL += fmt.Sprintf(" VALUES (%s)", joinStrings(placeholders))
		querySQL = mgr.convertPlaceholder(querySQL, driver)
		values = mgr.sanitizeArgs(querySQL, values)
		start := time.Now()
		res, err := executor.Exec(querySQL, values...)
		mgr.logTrace(start, querySQL, values, err)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	}

	querySQL += fmt.Sprintf(" VALUES (%s)", joinStrings(placeholders))
	querySQL = mgr.convertPlaceholder(querySQL, driver)
	values = mgr.sanitizeArgs(querySQL, values)
	start := time.Now()
	result, err := executor.Exec(querySQL, values...)
	mgr.logTrace(start, querySQL, values, err)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (mgr *dbManager) update(executor sqlExecutor, table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	// If both feature checks are disabled, use fast path
	if !mgr.enableTimestampCheck && !mgr.enableOptimisticLockCheck {
		return mgr.updateFast(executor, table, record, where, whereArgs...)
	}
	// Feature checks enabled, use full path
	return mgr.updateWithOptions(executor, table, record, where, false, whereArgs...)
}

// updateFast is a lightweight update that skips timestamp and optimistic lock checks for better performance
func (mgr *dbManager) updateFast(executor sqlExecutor, table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if record == nil || len(record.columns) == 0 {
		return 0, fmt.Errorf("record is empty")
	}

	columns, values := mgr.getOrderedColumns(record)
	var setClauses []string
	for _, col := range columns {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
	}

	values = append(values, whereArgs...)

	var querySQL string
	if where != "" {
		querySQL = fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, joinStrings(setClauses), where)
	} else {
		querySQL = fmt.Sprintf("UPDATE %s SET %s", table, joinStrings(setClauses))
	}

	querySQL = mgr.convertPlaceholder(querySQL, mgr.config.Driver)
	values = mgr.sanitizeArgs(querySQL, values)
	start := time.Now()
	result, err := executor.Exec(querySQL, values...)
	mgr.logTrace(start, querySQL, values, err)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (mgr *dbManager) updateWithOptions(executor sqlExecutor, table string, record *Record, where string, skipTimestamps bool, whereArgs ...interface{}) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if record == nil || len(record.columns) == 0 {
		return 0, fmt.Errorf("record is empty")
	}

	// Apply updated_at timestamp (only if feature is enabled)
	if mgr.enableTimestampCheck {
		mgr.applyUpdatedAtTimestamp(table, record, skipTimestamps)
	}

	// Check for optimistic lock (only if feature is enabled)
	versionChecked := false
	var currentVersion int64
	var config *OptimisticLockConfig
	if mgr.enableOptimisticLockCheck {
		config = mgr.getOptimisticLockConfig(table)
		if config != nil && config.VersionField != "" {
			if ver, ok := mgr.getVersionFromRecord(table, record); ok {
				currentVersion = ver
				versionChecked = true
				// Remove version from record so it's not in the regular SET clause
				// We'll add it separately with increment
				record.Remove(config.VersionField)
			}
		}
	}

	columns, values := mgr.getOrderedColumns(record)
	var setClauses []string
	for _, col := range columns {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
	}

	// Add version increment to SET clause if optimistic lock is enabled and version was found
	if versionChecked && config != nil {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", config.VersionField))
		values = append(values, currentVersion+1)
	}

	// Add version check to WHERE clause
	if versionChecked && config != nil {
		if where != "" {
			where = fmt.Sprintf("(%s) AND %s = ?", where, config.VersionField)
		} else {
			where = fmt.Sprintf("%s = ?", config.VersionField)
		}
		whereArgs = append(whereArgs, currentVersion)
	}

	values = append(values, whereArgs...)

	var querySQL string
	if where != "" {
		querySQL = fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, joinStrings(setClauses), where)
	} else {
		querySQL = fmt.Sprintf("UPDATE %s SET %s", table, joinStrings(setClauses))
	}

	querySQL = mgr.convertPlaceholder(querySQL, mgr.config.Driver)
	values = mgr.sanitizeArgs(querySQL, values)
	start := time.Now()
	result, err := executor.Exec(querySQL, values...)
	mgr.logTrace(start, querySQL, values, err)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	// If version was checked and no rows were affected, it's a version mismatch
	if versionChecked && rowsAffected == 0 {
		return 0, ErrVersionMismatch
	}

	return rowsAffected, nil
}

func (mgr *dbManager) delete(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if where == "" {
		return 0, fmt.Errorf("where condition is required for delete")
	}

	// Check if soft delete is configured for this table
	if mgr.hasSoftDelete(table) {
		return mgr.softDelete(executor, table, where, whereArgs...)
	}

	querySQL := fmt.Sprintf("DELETE FROM %s WHERE %s", table, where)
	querySQL, whereArgs = mgr.prepareQuerySQL(querySQL, whereArgs...)

	start := time.Now()
	result, err := executor.Exec(querySQL, whereArgs...)
	mgr.logTrace(start, querySQL, whereArgs, err)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// deleteRecord 根据 Record 中的主键字段删除记录
func (mgr *dbManager) deleteRecord(executor sqlExecutor, table string, record *Record) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if record == nil || len(record.columns) == 0 {
		return 0, fmt.Errorf("record is empty")
	}

	// 获取表的主键
	pks, err := mgr.getPrimaryKeys(executor, table)
	if err != nil {
		return 0, fmt.Errorf("failed to get primary keys: %v", err)
	}
	if len(pks) == 0 {
		return 0, fmt.Errorf("table %s has no primary key, cannot use DeleteRecord", table)
	}

	// 从 Record 中提取主键值构建 WHERE 条件
	var whereClauses []string
	var whereArgs []interface{}
	for _, pk := range pks {
		if !record.Has(pk) {
			return 0, fmt.Errorf("primary key '%s' not found in record", pk)
		}
		val := record.Get(pk)
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", pk))
		whereArgs = append(whereArgs, val)
	}

	where := strings.Join(whereClauses, " AND ")
	return mgr.delete(executor, table, where, whereArgs...)
}

// updateRecord 根据 Record 中的主键字段更新记录
func (mgr *dbManager) updateRecord(executor sqlExecutor, table string, record *Record) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if record == nil || len(record.columns) == 0 {
		return 0, fmt.Errorf("record is empty")
	}

	// 获取表的主键
	pks, err := mgr.getPrimaryKeys(executor, table)
	if err != nil {
		return 0, fmt.Errorf("failed to get primary keys: %v", err)
	}
	if len(pks) == 0 {
		return 0, fmt.Errorf("table %s has no primary key, cannot use updateRecord", table)
	}

	// 提取主键值构建 WHERE 条件，并从更新字段中排除主键
	var pkClauses []string
	var pkValues []interface{}

	updateRecord := NewRecord()
	columns, _ := mgr.getOrderedColumns(record)

	for _, col := range columns {
		isPK := false
		for _, pk := range pks {
			if strings.EqualFold(col, pk) {
				isPK = true
				pkClauses = append(pkClauses, fmt.Sprintf("%s = ?", col))
				pkValues = append(pkValues, record.Get(col))
				break
			}
		}
		if !isPK {
			updateRecord.Set(col, record.Get(col))
		}
	}

	if len(pkClauses) != len(pks) {
		return 0, fmt.Errorf("not all primary keys found in record")
	}

	if len(updateRecord.columns) == 0 {
		return 0, nil // 只有主键，无需更新
	}

	where := strings.Join(pkClauses, " AND ")
	return mgr.update(executor, table, updateRecord, where, pkValues...)
}

func (mgr *dbManager) count(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	var querySQL string
	if where != "" {
		querySQL = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where)
	} else {
		querySQL = fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	}
	querySQL = mgr.convertPlaceholder(querySQL, mgr.config.Driver)
	whereArgs = mgr.sanitizeArgs(querySQL, whereArgs)

	var count int64
	start := time.Now()
	err := executor.QueryRow(querySQL, whereArgs...).Scan(&count)
	mgr.logTrace(start, querySQL, whereArgs, err)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (mgr *dbManager) exists(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (bool, error) {
	count, err := mgr.count(executor, table, where, whereArgs...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (mgr *dbManager) batchInsert(executor sqlExecutor, table string, records []*Record, batchSize int) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if len(records) == 0 {
		return 0, fmt.Errorf("no records to insert")
	}

	var totalAffected int64
	driver := mgr.config.Driver

	// Extract and sort columns once for all batches
	firstRecord := records[0]
	firstRecord.mu.RLock()
	var columns []string
	for col := range firstRecord.columns {
		columns = append(columns, col)
	}
	firstRecord.mu.RUnlock()
	sort.Strings(columns)

	numCols := len(columns)
	colNamesJoined := joinStrings(columns)

	// Pre-generate row placeholders for drivers that support multi-row INSERT
	var rowPlaceholder string
	if driver != PostgreSQL && driver != SQLServer && driver != Oracle {
		placeholders := make([]string, numCols)
		for i := range placeholders {
			placeholders[i] = "?"
		}
		rowPlaceholder = "(" + joinStrings(placeholders) + ")"
	}

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		batchLen := len(batch)

		var querySQL string
		var flatArgs []interface{}

		if driver == PostgreSQL {
			flatArgs = make([]interface{}, 0, batchLen*numCols)
			var sb strings.Builder
			sb.WriteString("INSERT INTO ")
			sb.WriteString(table)
			sb.WriteString(" (")
			sb.WriteString(colNamesJoined)
			sb.WriteString(") VALUES ")

			for rowIdx, record := range batch {
				if rowIdx > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString("(")
				record.mu.RLock()
				for colIdx, col := range columns {
					if colIdx > 0 {
						sb.WriteString(", ")
					}
					placeholderIdx := rowIdx*numCols + colIdx + 1
					sb.WriteString("$")
					sb.WriteString(strconv.Itoa(placeholderIdx))
					flatArgs = append(flatArgs, record.columns[col])
				}
				record.mu.RUnlock()
				sb.WriteString(")")
			}
			querySQL = sb.String()
		} else if driver == SQLServer || driver == Oracle {
			// SQLServer/Oracle often perform better with prepared statements for batching
			var placeholders []string
			for range columns {
				placeholders = append(placeholders, "?")
			}
			querySQL = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, colNamesJoined, joinStrings(placeholders))
			querySQL = mgr.convertPlaceholder(querySQL, driver)

			if preparer, ok := executor.(interface {
				Prepare(query string) (*sql.Stmt, error)
			}); ok {
				stmt, err := preparer.Prepare(querySQL)
				if err != nil {
					for _, record := range batch {
						values := make([]interface{}, numCols)
						for j, col := range columns {
							values[j] = record.getValue(col)
						}
						values = mgr.sanitizeArgs(querySQL, values)
						start := time.Now()
						result, err := executor.Exec(querySQL, values...)
						mgr.logTrace(start, querySQL, values, err)
						if err != nil {
							return totalAffected, err
						}
						affected, _ := result.RowsAffected()
						totalAffected += affected
					}
					continue
				}
				defer stmt.Close()

				for _, record := range batch {
					values := make([]interface{}, numCols)
					for j, col := range columns {
						values[j] = record.getValue(col)
					}
					values = mgr.sanitizeArgs(querySQL, values)
					start := time.Now()
					result, err := stmt.Exec(values...)
					mgr.logTrace(start, querySQL, values, err)
					if err != nil {
						return totalAffected, err
					}
					affected, _ := result.RowsAffected()
					totalAffected += affected
				}
			} else {
				for _, record := range batch {
					values := make([]interface{}, numCols)
					for j, col := range columns {
						values[j] = record.getValue(col)
					}
					values = mgr.sanitizeArgs(querySQL, values)
					start := time.Now()
					result, err := executor.Exec(querySQL, values...)
					mgr.logTrace(start, querySQL, values, err)
					if err != nil {
						return totalAffected, err
					}
					affected, _ := result.RowsAffected()
					totalAffected += affected
				}
			}
			continue
		} else {
			// MySQL, SQLite and others
			flatArgs = make([]interface{}, 0, batchLen*numCols)
			var sb strings.Builder
			sb.WriteString("INSERT INTO ")
			sb.WriteString(table)
			sb.WriteString(" (")
			sb.WriteString(colNamesJoined)
			sb.WriteString(") VALUES ")

			for rowIdx, record := range batch {
				if rowIdx > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(rowPlaceholder)
				record.mu.RLock()
				for _, col := range columns {
					flatArgs = append(flatArgs, record.columns[col])
				}
				record.mu.RUnlock()
			}
			querySQL = sb.String()
		}

		start := time.Now()
		result, err := executor.Exec(querySQL, flatArgs...)
		mgr.logTrace(start, querySQL, flatArgs, err)
		if err != nil {
			return totalAffected, err
		}
		affected, _ := result.RowsAffected()
		totalAffected += affected
	}
	return totalAffected, nil
}

// batchUpdate 批量更新记录（根据主键）
func (mgr *dbManager) batchUpdate(executor sqlExecutor, table string, records []*Record, batchSize int) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if len(records) == 0 {
		return 0, fmt.Errorf("no records to update")
	}

	// 获取表的主键
	pks, err := mgr.getPrimaryKeys(executor, table)
	if err != nil {
		return 0, fmt.Errorf("failed to get primary keys: %v", err)
	}
	if len(pks) == 0 {
		return 0, fmt.Errorf("table %s has no primary key, cannot use BatchUpdate", table)
	}

	// Extract update columns from first record (excluding PKs)
	firstRecord := records[0]
	firstRecord.mu.RLock()
	var updateCols []string
	for col := range firstRecord.columns {
		isPK := false
		for _, pk := range pks {
			if strings.EqualFold(col, pk) {
				isPK = true
				break
			}
		}
		if !isPK {
			updateCols = append(updateCols, col)
		}
	}
	firstRecord.mu.RUnlock()

	if len(updateCols) == 0 {
		return 0, nil // Nothing to update besides PKs
	}

	// Build UPDATE SQL once
	var setClauses []string
	for _, col := range updateCols {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
	}
	var whereClauses []string
	for _, pk := range pks {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", pk))
	}

	querySQL := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		table, joinStrings(setClauses), joinStrings(whereClauses))
	querySQL = mgr.convertPlaceholder(querySQL, mgr.config.Driver)

	var totalAffected int64
	numUpdateCols := len(updateCols)
	numPKs := len(pks)
	numTotalArgs := numUpdateCols + numPKs

	// Try to use prepared statement for all batches
	var stmt *sql.Stmt
	if preparer, ok := executor.(interface {
		Prepare(query string) (*sql.Stmt, error)
	}); ok {
		stmt, _ = preparer.Prepare(querySQL)
	}
	if stmt != nil {
		defer stmt.Close()
	}

	// 分批处理
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		for _, record := range batch {
			values := make([]interface{}, numTotalArgs)
			record.mu.RLock()
			// SET values
			for j, col := range updateCols {
				values[j] = record.columns[col]
			}
			// WHERE values (PKs)
			for j, pk := range pks {
				values[numUpdateCols+j] = record.columns[pk]
			}
			record.mu.RUnlock()

			start := time.Now()
			var result sql.Result
			var err error
			if stmt != nil {
				result, err = stmt.Exec(values...)
			} else {
				sanitizedValues := mgr.sanitizeArgs(querySQL, values)
				result, err = executor.Exec(querySQL, sanitizedValues...)
			}
			mgr.logTrace(start, querySQL, values, err)
			if err != nil {
				return totalAffected, err
			}
			affected, _ := result.RowsAffected()
			totalAffected += affected
		}
	}

	return totalAffected, nil
}

// batchDelete 批量删除记录（根据主键）
func (mgr *dbManager) batchDelete(executor sqlExecutor, table string, records []*Record, batchSize int) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if len(records) == 0 {
		return 0, fmt.Errorf("no records to delete")
	}

	// 获取表的主键
	pks, err := mgr.getPrimaryKeys(executor, table)
	if err != nil {
		return 0, fmt.Errorf("failed to get primary keys: %v", err)
	}
	if len(pks) == 0 {
		return 0, fmt.Errorf("table %s has no primary key, cannot use BatchDelete", table)
	}

	var totalAffected int64
	driver := mgr.config.Driver

	// 分批处理
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]

		// 对于单主键，使用 IN 子句优化
		if len(pks) == 1 {
			pk := pks[0]
			var pkValues []interface{}
			var placeholders []string

			for idx, record := range batch {
				pkVal := record.Get(pk)
				if pkVal == nil {
					continue
				}
				pkValues = append(pkValues, pkVal)
				if driver == PostgreSQL {
					placeholders = append(placeholders, fmt.Sprintf("$%d", idx+1))
				} else if driver == SQLServer {
					placeholders = append(placeholders, fmt.Sprintf("@p%d", idx+1))
				} else if driver == Oracle {
					placeholders = append(placeholders, fmt.Sprintf(":%d", idx+1))
				} else {
					placeholders = append(placeholders, "?")
				}
			}

			if len(pkValues) == 0 {
				continue
			}

			querySQL := fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)",
				table, pk, strings.Join(placeholders, ", "))

			start := time.Now()
			result, err := executor.Exec(querySQL, pkValues...)
			mgr.logTrace(start, querySQL, pkValues, err)
			if err != nil {
				return totalAffected, err
			}
			affected, _ := result.RowsAffected()
			totalAffected += affected
		} else {
			// 复合主键，使用预处理语句逐条删除
			var whereClauses []string
			for _, pk := range pks {
				whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", pk))
			}

			querySQL := fmt.Sprintf("DELETE FROM %s WHERE %s",
				table, strings.Join(whereClauses, " AND "))
			querySQL = mgr.convertPlaceholder(querySQL, driver)

			// 尝试使用预处理语句
			if preparer, ok := executor.(interface {
				Prepare(query string) (*sql.Stmt, error)
			}); ok {
				stmt, err := preparer.Prepare(querySQL)
				if err == nil {
					defer stmt.Close()

					for _, record := range batch {
						var pkValues []interface{}
						for _, pk := range pks {
							pkValues = append(pkValues, record.Get(pk))
						}

						start := time.Now()
						result, err := stmt.Exec(pkValues...)
						mgr.logTrace(start, querySQL, pkValues, err)
						if err != nil {
							return totalAffected, err
						}
						affected, _ := result.RowsAffected()
						totalAffected += affected
					}
					continue
				}
			}

			// 回退到单条执行
			for _, record := range batch {
				var pkValues []interface{}
				for _, pk := range pks {
					pkValues = append(pkValues, record.Get(pk))
				}

				start := time.Now()
				result, err := executor.Exec(querySQL, pkValues...)
				mgr.logTrace(start, querySQL, pkValues, err)
				if err != nil {
					return totalAffected, err
				}
				affected, _ := result.RowsAffected()
				totalAffected += affected
			}
		}
	}

	return totalAffected, nil
}

// batchDeleteByIds 根据主键ID列表批量删除
func (mgr *dbManager) batchDeleteByIds(executor sqlExecutor, table string, ids []interface{}, batchSize int) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, fmt.Errorf("no ids to delete")
	}

	// 获取表的主键
	pks, err := mgr.getPrimaryKeys(executor, table)
	if err != nil {
		return 0, fmt.Errorf("failed to get primary keys: %v", err)
	}
	if len(pks) == 0 {
		return 0, fmt.Errorf("table %s has no primary key", table)
	}
	if len(pks) > 1 {
		return 0, fmt.Errorf("BatchDeleteByIds only supports single primary key tables")
	}

	pk := pks[0]
	var totalAffected int64
	driver := mgr.config.Driver

	// 分批处理
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}

		batch := ids[i:end]
		var placeholders []string

		for idx := range batch {
			if driver == PostgreSQL {
				placeholders = append(placeholders, fmt.Sprintf("$%d", idx+1))
			} else if driver == SQLServer {
				placeholders = append(placeholders, fmt.Sprintf("@p%d", idx+1))
			} else if driver == Oracle {
				placeholders = append(placeholders, fmt.Sprintf(":%d", idx+1))
			} else {
				placeholders = append(placeholders, "?")
			}
		}

		querySQL := fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)",
			table, pk, strings.Join(placeholders, ", "))

		start := time.Now()
		result, err := executor.Exec(querySQL, batch...)
		mgr.logTrace(start, querySQL, batch, err)
		if err != nil {
			return totalAffected, err
		}
		affected, _ := result.RowsAffected()
		totalAffected += affected
	}

	return totalAffected, nil
}

func (mgr *dbManager) paginate(executor sqlExecutor, querySQL string, page, pageSize int, args ...interface{}) ([]Record, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	driver := mgr.config.Driver
	lowerSQL := strings.ToLower(querySQL)
	baseSQL := querySQL
	if strings.Contains(lowerSQL, " order by ") {
		orderIdx := strings.Index(lowerSQL, " order by ")
		baseSQL = querySQL[:orderIdx]
	}

	var countSQL string
	// 尝试优化 COUNT 语句
	if optimized, ok := optimizeCountSQL(baseSQL); ok {
		countSQL = optimized
	} else {
		// 如果无法优化（含有 DISTINCT, GROUP BY 等），则使用子查询
		if driver == Oracle {
			countSQL = fmt.Sprintf("SELECT COUNT(*) FROM (%s) sub", baseSQL)
		} else {
			countSQL = fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS sub", baseSQL)
		}
	}

	countSQL = mgr.convertPlaceholder(countSQL, driver)
	args = mgr.sanitizeArgs(countSQL, args)

	var total int64
	startCount := time.Now()
	err := executor.QueryRow(countSQL, args...).Scan(&total)
	mgr.logTrace(startCount, countSQL, args, err)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var paginatedSQL string
	if driver == SQLServer {
		if strings.Contains(lowerSQL, " order by ") {
			paginatedSQL = fmt.Sprintf("%s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", querySQL, offset, pageSize)
		} else {
			paginatedSQL = fmt.Sprintf("%s ORDER BY (SELECT NULL) OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", querySQL, offset, pageSize)
		}
	} else if driver == Oracle {
		if strings.Contains(lowerSQL, " order by ") {
			paginatedSQL = fmt.Sprintf("SELECT a.* FROM (SELECT a.*, ROWNUM rn FROM (%s) a WHERE ROWNUM <= %d) a WHERE rn > %d", querySQL, offset+pageSize, offset)
		} else {
			paginatedSQL = fmt.Sprintf("SELECT a.* FROM (SELECT a.*, ROWNUM rn FROM (%s ORDER BY 1) a WHERE ROWNUM <= %d) a WHERE rn > %d", querySQL, offset+pageSize, offset)
		}
	} else {
		paginatedSQL = fmt.Sprintf("%s LIMIT %d OFFSET %d", querySQL, pageSize, offset)
	}

	paginatedSQL = mgr.convertPlaceholder(paginatedSQL, driver)

	startPaginate := time.Now()
	rows, err := executor.Query(paginatedSQL, args...)
	mgr.logTrace(startPaginate, paginatedSQL, args, err)
	if err != nil {
		return nil, total, err
	}
	defer rows.Close()

	results, err := scanRecords(rows, driver)
	if err != nil {
		return nil, total, err
	}
	return results, total, nil
}

// scanRows is a helper function to scan sql.Rows into a slice of maps
func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	numCols := len(columns)
	var results []map[string]interface{}
	// Reuse slices for each row to reduce allocations
	values := make([]interface{}, numCols)
	valuePtrs := make([]interface{}, numCols)
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		entry := make(map[string]interface{}, numCols)
		for i, col := range columns {
			val := values[i]

			// Handle []byte conversion for numeric/decimal types
			if b, ok := val.([]byte); ok {
				dbType := strings.ToUpper(columnTypes[i].DatabaseTypeName())

				if isNumericType(dbType) {
					if s := string(b); s != "" {
						entry[col] = s
					} else {
						entry[col] = nil
					}
					continue
				}

				if !isBinaryType(dbType) {
					entry[col] = string(b)
					continue
				}
				// Keep as []byte for binary types, but we must copy it
				// because the underlying buffer might be reused by the driver
				bCopy := make([]byte, len(b))
				copy(bCopy, b)
				entry[col] = bCopy
				continue
			}

			entry[col] = val
		}
		results = append(results, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func isNumericType(dbType string) bool {
	numericTypes := []string{"DECIMAL", "NUMERIC", "NUMBER", "MONEY", "SMALLMONEY", "DEC", "FIXED"}
	for _, t := range numericTypes {
		if strings.Contains(dbType, t) {
			return true
		}
	}
	return false
}

func isBinaryType(dbType string) bool {
	binaryTypes := []string{"BLOB", "BINARY", "VARBINARY", "BYTEA", "IMAGE", "RAW"}
	for _, t := range binaryTypes {
		if strings.Contains(dbType, t) {
			return true
		}
	}
	return false
}

// optimizeCountSQL 尝试将简单的 SELECT ... FROM ... 转换为 SELECT COUNT(*) FROM ...
func optimizeCountSQL(querySQL string) (string, bool) {
	lower := strings.ToLower(querySQL)

	// 如果包含以下关键字，不进行优化，使用子查询最安全
	if strings.Contains(lower, "distinct") ||
		strings.Contains(lower, "group by") ||
		strings.Contains(lower, "union") ||
		strings.Contains(lower, "having") ||
		strings.Contains(lower, "intersect") ||
		strings.Contains(lower, "except") {
		return "", false
	}

	// 寻找第一个 FROM
	fromIdx := strings.Index(lower, " from ")
	if fromIdx == -1 {
		return "", false
	}

	// 检查 FROM 之前是否有子查询（简单判断：是否有左括号）
	// 如果 SELECT 列表中包含子查询，优化可能会变得复杂，此时回退到安全模式
	if strings.Contains(querySQL[:fromIdx], "(") {
		return "", false
	}

	// 构建优化的 COUNT 语句
	optimized := "SELECT COUNT(*) " + querySQL[fromIdx:]
	return optimized, true
}

// scanRecords is a helper function to scan sql.Rows into a slice of Record
func scanRecords(rows *sql.Rows, driver DriverType) ([]Record, error) {
	maps, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	results := make([]Record, len(maps))
	for i, m := range maps {
		results[i] = Record{columns: m}
	}
	return results, nil
}

// scanMaps is a helper function to scan sql.Rows into a slice of map
func scanMaps(rows *sql.Rows, driver DriverType) ([]map[string]interface{}, error) {
	return scanRows(rows)
}

// GetDB returns the underlying database connection
func (db *DB) GetDB() *sql.DB {
	return db.dbMgr.getDB()
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.dbMgr.db != nil {
		return db.dbMgr.db.Close()
	}
	return nil
}

// SetCurrentDB switches the global default database by name
func SetCurrentDB(dbname string) error {
	multiMgr.mu.RLock()
	_, exists := multiMgr.databases[dbname]
	multiMgr.mu.RUnlock()

	if !exists {
		return fmt.Errorf("database '%s' not found", dbname)
	}

	multiMgr.mu.Lock()
	multiMgr.currentDB = dbname
	multiMgr.mu.Unlock()

	return nil
}

// safeGetCurrentDB returns the current database manager without panicking
func safeGetCurrentDB() (*dbManager, error) {
	if multiMgr == nil {
		return nil, ErrNotInitialized
	}

	multiMgr.mu.RLock()
	currentDB := multiMgr.currentDB
	multiMgr.mu.RUnlock()

	if currentDB == "" {
		return nil, ErrNotInitialized
	}

	dbMgr := GetDatabase(currentDB)
	if dbMgr == nil {
		return nil, ErrNotInitialized
	}

	return dbMgr, nil
}

// GetCurrentDB returns the current database manager
func GetCurrentDB() *dbManager {
	dbMgr, err := safeGetCurrentDB()
	if err != nil {
		panic(err.Error())
	}
	return dbMgr
}

// GetConfig returns the database configuration
func (mgr *dbManager) GetConfig() (*Config, error) {
	if mgr == nil {
		return nil, fmt.Errorf("database manager is nil")
	}
	return mgr.config, nil
}

// GetDatabase returns the database manager by name
func GetDatabase(dbname string) *dbManager {
	if multiMgr == nil {
		return nil
	}

	multiMgr.mu.RLock()
	defer multiMgr.mu.RUnlock()

	return multiMgr.databases[dbname]
}

// GetDB returns the underlying database connection of current database
func GetDB() *sql.DB {
	return GetCurrentDB().getDB()
}

// GetDBByName returns the database connection by name
func GetDBByName(dbname string) (*sql.DB, error) {
	multiMgr.mu.RLock()
	dbMgr, exists := multiMgr.databases[dbname]
	multiMgr.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("database '%s' not found", dbname)
	}

	return dbMgr.getDB(), nil
}

// Close closes the database connection
// Close closes all database connections
func Close() error {
	if multiMgr == nil {
		return nil
	}

	multiMgr.mu.Lock()
	defer multiMgr.mu.Unlock()

	for _, dbMgr := range multiMgr.databases {
		// 清理预编译语句缓存
		dbMgr.clearStmtCache()

		// 关闭数据库连接
		if dbMgr.db != nil {
			dbMgr.db.Close()
		}
	}
	multiMgr.databases = make(map[string]*dbManager)
	multiMgr.currentDB = ""
	multiMgr.defaultDB = ""

	return nil
}

// CloseDB closes a specific database connection by name
func CloseDB(dbname string) error {
	if multiMgr != nil {
		multiMgr.mu.Lock()
		defer multiMgr.mu.Unlock()

		if dbMgr, exists := multiMgr.databases[dbname]; exists {
			// 清理预编译语句缓存
			dbMgr.clearStmtCache()

			// 关闭数据库连接
			if dbMgr.db != nil {
				dbMgr.db.Close()
				dbMgr.db = nil
			}
			delete(multiMgr.databases, dbname)

			if multiMgr.currentDB == dbname {
				if multiMgr.defaultDB != "" && multiMgr.defaultDB != dbname {
					multiMgr.currentDB = multiMgr.defaultDB
				} else {
					multiMgr.currentDB = ""
				}
			}

			if multiMgr.defaultDB == dbname {
				multiMgr.defaultDB = ""
				for name := range multiMgr.databases {
					multiMgr.defaultDB = name
					break
				}
			}
		}
	}

	return nil
}

// ListDatabases returns the list of registered database names
func ListDatabases() []string {
	var databases []string
	if multiMgr != nil {
		multiMgr.mu.RLock()
		for name := range multiMgr.databases {
			databases = append(databases, name)
		}
		multiMgr.mu.RUnlock()
	}
	return databases
}

// GetCurrentDBName returns the name of the current database
func GetCurrentDBName() string {
	if multiMgr == nil {
		return ""
	}

	multiMgr.mu.RLock()
	defer multiMgr.mu.RUnlock()

	return multiMgr.currentDB
}

// EnableTimestamps enables auto timestamp in Update operations.
// When enabled, Update will check and apply auto timestamp configurations.
// Default is false (disabled) for better performance.
func EnableTimestamps() {
	mgr, err := defaultDB()
	if err != nil {
		return
	}
	mgr.dbMgr.mu.Lock()
	defer mgr.dbMgr.mu.Unlock()
	mgr.dbMgr.enableTimestampCheck = true
}

// EnableTimestamps enables auto timestamp for this database instance.
func (db *DB) EnableTimestamps() *DB {
	if db.lastErr != nil {
		return db
	}
	db.dbMgr.mu.Lock()
	defer db.dbMgr.mu.Unlock()
	db.dbMgr.enableTimestampCheck = true
	return db
}

// Deprecated: Use EnableTimestamps() instead
func EnableTimestampCheck() {
	EnableTimestamps()
}

// Deprecated: Use EnableTimestamps() instead
func (db *DB) EnableTimestampCheck() *DB {
	return db.EnableTimestamps()
}

// EnableOptimisticLock enables optimistic lock in Update operations.
// When enabled, Update will check and apply optimistic lock configurations.
// Default is false (disabled) for better performance.
func EnableOptimisticLock() {
	mgr, err := defaultDB()
	if err != nil {
		return
	}
	mgr.dbMgr.mu.Lock()
	defer mgr.dbMgr.mu.Unlock()
	mgr.dbMgr.enableOptimisticLockCheck = true
}

// EnableOptimisticLock enables optimistic lock for this database instance.
func (db *DB) EnableOptimisticLock() *DB {
	if db.lastErr != nil {
		return db
	}
	db.dbMgr.mu.Lock()
	defer db.dbMgr.mu.Unlock()
	db.dbMgr.enableOptimisticLockCheck = true
	return db
}

// Deprecated: Use EnableOptimisticLock() instead
func EnableOptimisticLockCheck() {
	EnableOptimisticLock()
}

// Deprecated: Use EnableOptimisticLock() instead
func (db *DB) EnableOptimisticLockCheck() *DB {
	return db.EnableOptimisticLock()
}

// EnableSoftDelete enables soft delete in query operations.
// When enabled, queries will automatically filter out soft-deleted records.
// Default is false (disabled) for better performance.
func EnableSoftDelete() {
	mgr, err := defaultDB()
	if err != nil {
		return
	}
	mgr.dbMgr.mu.Lock()
	defer mgr.dbMgr.mu.Unlock()
	mgr.dbMgr.enableSoftDeleteCheck = true
}

// EnableSoftDelete enables soft delete for this database instance.
func (db *DB) EnableSoftDelete() *DB {
	if db.lastErr != nil {
		return db
	}
	db.dbMgr.mu.Lock()
	defer db.dbMgr.mu.Unlock()
	db.dbMgr.enableSoftDeleteCheck = true
	return db
}

// Deprecated: Use EnableSoftDelete() instead
func EnableSoftDeleteCheck() {
	EnableSoftDelete()
}

// Deprecated: Use EnableSoftDelete() instead
func (db *DB) EnableSoftDeleteCheck() *DB {
	return db.EnableSoftDelete()
}

// initDB initializes the database connection
func (mgr *dbManager) initDB() error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	if mgr.db != nil {
		return nil
	}

	db, err := sql.Open(string(mgr.config.Driver), mgr.config.DSN)
	if err != nil {
		return err
	}

	// Configure connection pool
	db.SetMaxOpenConns(mgr.config.MaxOpen)
	db.SetMaxIdleConns(mgr.config.MaxIdle)
	db.SetConnMaxLifetime(mgr.config.ConnMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return err
	}

	mgr.db = db
	return nil
}

// getDB returns the database connection, initializing if necessary
func (mgr *dbManager) getDB() *sql.DB {
	if mgr == nil {
		panic("dbkit: database manager is nil. Please call dbkit.OpenDatabase()  before using dbkit operations")
	}
	if mgr.db == nil {
		if err := mgr.initDB(); err != nil {
			panic(fmt.Sprintf("dbkit: failed to initialize database: %v", err))
		}
	}
	return mgr.db
}

// Ping checks if the database connection is alive
func (mgr *dbManager) Ping() error {
	if mgr == nil {
		return fmt.Errorf("database manager not initialized. Please call dbkit.OpenDatabase()  before using dbkit operations")
	}
	if mgr.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return mgr.db.Ping()
}

// convertPlaceholder converts ? placeholders to $n for PostgreSQL, @param for SQL Server, or :n for Oracle
func (mgr *dbManager) convertPlaceholder(querySQL string, driver DriverType) string {
	return mgr.convertPlaceholderWithOffset(querySQL, driver, 0)
}

// convertPlaceholderWithOffset converts ? placeholders with an index offset
func (mgr *dbManager) convertPlaceholderWithOffset(querySQL string, driver DriverType, offset int) string {
	if driver == MySQL || driver == SQLite3 {
		return querySQL
	}

	var builder strings.Builder
	builder.Grow(len(querySQL) + 10)
	paramIndex := 1 + offset
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false

	for i := 0; i < len(querySQL); i++ {
		char := querySQL[i]

		// Handle escaping (mostly for single quotes but good practice)
		if i+1 < len(querySQL) && char == '\\' {
			builder.WriteByte(char)
			builder.WriteByte(querySQL[i+1])
			i++
			continue
		}

		// Handle string literals and identifiers
		if char == '\'' && !inDoubleQuote && !inBacktick {
			if i+1 < len(querySQL) && querySQL[i+1] == '\'' {
				builder.WriteByte('\'')
				builder.WriteByte('\'')
				i++
				continue
			}
			inSingleQuote = !inSingleQuote
			builder.WriteByte('\'')
			continue
		}

		if char == '"' && !inSingleQuote && !inBacktick {
			inDoubleQuote = !inDoubleQuote
			builder.WriteByte('"')
			continue
		}

		if char == '`' && !inSingleQuote && !inDoubleQuote {
			inBacktick = !inBacktick
			builder.WriteByte('`')
			continue
		}

		if char == '?' && !inSingleQuote && !inDoubleQuote && !inBacktick {
			switch driver {
			case PostgreSQL:
				builder.WriteString(fmt.Sprintf("$%d", paramIndex))
			case SQLServer:
				builder.WriteString(fmt.Sprintf("@p%d", paramIndex))
			case Oracle:
				builder.WriteString(fmt.Sprintf(":%d", paramIndex))
			default:
				builder.WriteByte('?')
			}
			paramIndex++
		} else {
			builder.WriteByte(char)
		}
	}
	return builder.String()
}

// sanitizeArgs 自动清理不必要的参数。如果用户误传了参数，则根据 SQL 中的占位符数量进行截断或清理。
func (mgr *dbManager) sanitizeArgs(querySQL string, args []interface{}) []interface{} {
	if len(args) == 0 {
		return args
	}

	placeholderCount := 0
	switch mgr.config.Driver {
	case PostgreSQL:
		// 使用预编译正则精确匹配 $1, $2...，避免 $1 匹配到 $10
		matches := postgresPlaceholderRe.FindAllStringSubmatch(querySQL, -1)
		for _, match := range matches {
			if len(match) > 1 {
				idx, _ := strconv.Atoi(match[1])
				if idx > placeholderCount {
					placeholderCount = idx
				}
			}
		}
	case SQLServer:
		// 使用预编译正则精确匹配 @p1, @p2...
		matches := sqlserverPlaceholderRe.FindAllStringSubmatch(querySQL, -1)
		for _, match := range matches {
			if len(match) > 1 {
				idx, _ := strconv.Atoi(match[1])
				if idx > placeholderCount {
					placeholderCount = idx
				}
			}
		}
	case Oracle:
		// 使用预编译正则精确匹配 :1, :2...
		matches := oraclePlaceholderRe.FindAllStringSubmatch(querySQL, -1)
		for _, match := range matches {
			if len(match) > 1 {
				idx, _ := strconv.Atoi(match[1])
				if idx > placeholderCount {
					placeholderCount = idx
				}
			}
		}
	case MySQL, SQLite3:
		// 统计 ? 的数量，需要跳过字符串常量中的问号
		count := 0
		inString := false
		var quoteChar rune
		for i, char := range querySQL {
			if (char == '\'' || char == '"' || char == '`') && (i == 0 || querySQL[i-1] != '\\') {
				if !inString {
					inString = true
					quoteChar = char
				} else if char == quoteChar {
					inString = false
				}
			}
			if char == '?' && !inString {
				count++
			}
		}
		placeholderCount = count
	}

	if placeholderCount == 0 {
		return args
	}

	if len(args) > placeholderCount {
		// 如果参数过多，截断多余部分
		return args[:placeholderCount]
	}

	return args
}

// logTrace 辅助函数，封装 SQL 日志记录逻辑
func (mgr *dbManager) logTrace(start time.Time, sql string, args []interface{}, err error) {
	duration := time.Since(start)
	cleanArgs := mgr.sanitizeArgs(sql, args)
	if err != nil {
		LogSQLError(mgr.name, sql, cleanArgs, duration, err)
	} else {
		LogSQL(mgr.name, sql, cleanArgs, duration)
	}
}

// joinStrings joins strings with commas
func joinStrings(strs []string) string {
	return strings.Join(strs, ", ")
}

// PoolStats represents database connection pool statistics
type PoolStats struct {
	// Database name
	DBName string `json:"db_name"`
	// Driver type
	Driver string `json:"driver"`
	// Maximum number of open connections (configured)
	MaxOpenConnections int `json:"max_open_connections"`
	// Current number of open connections (in use + idle)
	OpenConnections int `json:"open_connections"`
	// Number of connections currently in use
	InUse int `json:"in_use"`
	// Number of idle connections
	Idle int `json:"idle"`
	// Total number of connections waited for
	WaitCount int64 `json:"wait_count"`
	// Total time blocked waiting for a new connection
	WaitDuration time.Duration `json:"wait_duration"`
	// Total number of connections closed due to MaxIdleTime
	MaxIdleClosed int64 `json:"max_idle_closed"`
	// Total number of connections closed due to MaxLifetime
	MaxLifetimeClosed int64 `json:"max_lifetime_closed"`
}

// PoolStats returns the connection pool statistics for the DB instance
func (db *DB) PoolStats() *PoolStats {
	if db.lastErr != nil || db.dbMgr == nil || db.dbMgr.db == nil {
		return nil
	}
	return db.dbMgr.poolStats()
}

// poolStats returns the connection pool statistics
func (mgr *dbManager) poolStats() *PoolStats {
	if mgr == nil || mgr.db == nil {
		return nil
	}

	stats := mgr.db.Stats()
	return &PoolStats{
		DBName:             mgr.name,
		Driver:             string(mgr.config.Driver),
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
}

// GetPoolStats returns the connection pool statistics for the default database
func GetPoolStats() *PoolStats {
	db, err := defaultDB()
	if err != nil {
		return nil
	}
	return db.PoolStats()
}

// GetPoolStatsDB returns the connection pool statistics for a specific database
func GetPoolStatsDB(dbname string) *PoolStats {
	return Use(dbname).PoolStats()
}

// AllPoolStats returns the connection pool statistics for all registered databases
func AllPoolStats() map[string]*PoolStats {
	result := make(map[string]*PoolStats)

	if multiMgr == nil {
		return result
	}

	multiMgr.mu.RLock()
	defer multiMgr.mu.RUnlock()

	for name, mgr := range multiMgr.databases {
		if mgr != nil && mgr.db != nil {
			result[name] = mgr.poolStats()
		}
	}

	return result
}

// PoolStatsMap returns the connection pool statistics as a map (for JSON serialization)
func (ps *PoolStats) ToMap() map[string]interface{} {
	if ps == nil {
		return nil
	}
	return map[string]interface{}{
		"db_name":              ps.DBName,
		"driver":               ps.Driver,
		"max_open_connections": ps.MaxOpenConnections,
		"open_connections":     ps.OpenConnections,
		"in_use":               ps.InUse,
		"idle":                 ps.Idle,
		"wait_count":           ps.WaitCount,
		"wait_duration_ms":     ps.WaitDuration.Milliseconds(),
		"max_idle_closed":      ps.MaxIdleClosed,
		"max_lifetime_closed":  ps.MaxLifetimeClosed,
	}
}

// String returns a human-readable string representation of the pool stats
func (ps *PoolStats) String() string {
	if ps == nil {
		return "PoolStats: nil"
	}
	return fmt.Sprintf(
		"PoolStats[%s/%s]: Open=%d (InUse=%d, Idle=%d), MaxOpen=%d, WaitCount=%d, WaitDuration=%v",
		ps.DBName, ps.Driver,
		ps.OpenConnections, ps.InUse, ps.Idle,
		ps.MaxOpenConnections, ps.WaitCount, ps.WaitDuration,
	)
}

// PrometheusMetrics returns Prometheus-compatible metrics string
func (ps *PoolStats) PrometheusMetrics() string {
	if ps == nil {
		return ""
	}

	dbLabel := fmt.Sprintf(`db="%s",driver="%s"`, ps.DBName, ps.Driver)

	return fmt.Sprintf(`# HELP dbkit_pool_max_open_connections Maximum number of open connections to the database.
# TYPE dbkit_pool_max_open_connections gauge
dbkit_pool_max_open_connections{%s} %d

# HELP dbkit_pool_open_connections The number of established connections both in use and idle.
# TYPE dbkit_pool_open_connections gauge
dbkit_pool_open_connections{%s} %d

# HELP dbkit_pool_in_use The number of connections currently in use.
# TYPE dbkit_pool_in_use gauge
dbkit_pool_in_use{%s} %d

# HELP dbkit_pool_idle The number of idle connections.
# TYPE dbkit_pool_idle gauge
dbkit_pool_idle{%s} %d

# HELP dbkit_pool_wait_count_total The total number of connections waited for.
# TYPE dbkit_pool_wait_count_total counter
dbkit_pool_wait_count_total{%s} %d

# HELP dbkit_pool_wait_duration_seconds_total The total time blocked waiting for a new connection.
# TYPE dbkit_pool_wait_duration_seconds_total counter
dbkit_pool_wait_duration_seconds_total{%s} %f

# HELP dbkit_pool_max_idle_closed_total The total number of connections closed due to SetMaxIdleConns.
# TYPE dbkit_pool_max_idle_closed_total counter
dbkit_pool_max_idle_closed_total{%s} %d

# HELP dbkit_pool_max_lifetime_closed_total The total number of connections closed due to SetConnMaxLifetime.
# TYPE dbkit_pool_max_lifetime_closed_total counter
dbkit_pool_max_lifetime_closed_total{%s} %d
`,
		dbLabel, ps.MaxOpenConnections,
		dbLabel, ps.OpenConnections,
		dbLabel, ps.InUse,
		dbLabel, ps.Idle,
		dbLabel, ps.WaitCount,
		dbLabel, ps.WaitDuration.Seconds(),
		dbLabel, ps.MaxIdleClosed,
		dbLabel, ps.MaxLifetimeClosed,
	)
}

// AllPrometheusMetrics returns Prometheus metrics for all databases
func AllPrometheusMetrics() string {
	allStats := AllPoolStats()
	var result strings.Builder

	for _, stats := range allStats {
		result.WriteString(stats.PrometheusMetrics())
		result.WriteString("\n")
	}

	return result.String()
}
