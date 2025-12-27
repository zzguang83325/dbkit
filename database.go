package dbkit

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	// Import database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/sijms/go-ora/v2"
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

// Config holds the database configuration
type Config struct {
	Driver          DriverType    // Database driver type (mysql, postgres, sqlite3)
	DSN             string        // Data source name (connection string)
	MaxOpen         int           // Maximum number of open connections
	MaxIdle         int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum connection lifetime
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
	dbMgr *dbManager
}

// Tx represents a database transaction with chainable methods
type Tx struct {
	tx    *sql.Tx
	dbMgr *dbManager
}

// sqlExecutor is an internal interface for executing SQL commands
type sqlExecutor interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// dbManager manages database connections
type dbManager struct {
	name    string
	config  *Config
	db      *sql.DB
	mu      sync.RWMutex
	drivers map[string]bool
	pkCache map[string][]string // Table name -> PK column names
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
		name:    dbname,
		config:  config,
		pkCache: make(map[string][]string),
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
// This is a convenience method that panics on error for fluent API usage
// For error handling, use UseWithError instead
func Use(dbname string) *DB {
	db, err := UseWithError(dbname)
	if err != nil {
		panic(err)
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

// defaultDB returns the default DB object (first registered database or single database mode)
func defaultDB() *DB {
	dbMgr := GetCurrentDB()
	if dbMgr == nil {
		panic("dbkit: database not initialized. Please call dbkit.OpenDatabase() before using dbkit operations")
	}
	return &DB{dbMgr: dbMgr}
}

// --- Internal Helper Methods on dbManager to unify DB and Tx logic ---

func (mgr *dbManager) query(executor sqlExecutor, sql string, args ...interface{}) ([]Record, error) {
	driver := mgr.config.Driver
	sql = mgr.convertPlaceholder(sql, driver)
	args = mgr.sanitizeArgs(sql, args)
	LogSQL(mgr.name, sql, args)

	rows, err := executor.Query(sql, args...)
	if err != nil {
		LogSQLError(mgr.name, sql, args, err)
		return nil, err
	}
	defer rows.Close()

	results, err := scanRecords(rows, driver)
	if err != nil {
		LogSQLError(mgr.name, sql, args, err)
		return nil, err
	}
	return results, nil
}

func (mgr *dbManager) queryFirst(executor sqlExecutor, sql string, args ...interface{}) (*Record, error) {
	results, err := mgr.query(executor, sql, args...)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

func (mgr *dbManager) queryMap(executor sqlExecutor, sql string, args ...interface{}) ([]map[string]interface{}, error) {
	driver := mgr.config.Driver
	lowerSQL := strings.ToLower(sql)

	// 处理Oracle的LIMIT语法
	if driver == Oracle {
		// 检测并替换LIMIT子句
		if limitIndex := strings.LastIndex(lowerSQL, " limit "); limitIndex != -1 {
			// 提取LIMIT值
			limitStr := strings.TrimSpace(sql[limitIndex+6:])
			// 移除LIMIT部分
			sql = sql[:limitIndex]
			// 添加ROWNUM条件
			if strings.Contains(lowerSQL, " where ") {
				sql = fmt.Sprintf("SELECT * FROM (%s) WHERE ROWNUM <= %s", sql, limitStr)
			} else {
				sql = fmt.Sprintf("SELECT * FROM (%s) WHERE ROWNUM <= %s", sql, limitStr)
			}
		}
	} else if driver == SQLServer {
		// 处理SQL Server的LIMIT语法，转换为TOP
		if limitIndex := strings.LastIndex(lowerSQL, " limit "); limitIndex != -1 {
			// 提取LIMIT值
			limitStr := strings.TrimSpace(sql[limitIndex+6:])
			// 移除LIMIT部分
			sql = sql[:limitIndex]
			// 在SELECT关键字后添加TOP N
			if selectIndex := strings.Index(lowerSQL, "select "); selectIndex != -1 {
				// 提取SELECT后面的内容
				selectContent := sql[selectIndex+6:]
				// 构建新的SQL
				sql = fmt.Sprintf("SELECT TOP %s %s", limitStr, selectContent)
			}
		}
	}

	sql = mgr.convertPlaceholder(sql, driver)
	args = mgr.sanitizeArgs(sql, args)
	LogSQL(mgr.name, sql, args)

	rows, err := executor.Query(sql, args...)
	if err != nil {
		LogSQLError(mgr.name, sql, args, err)
		return nil, err
	}
	defer rows.Close()

	results, err := scanMaps(rows, driver)
	if err != nil {
		LogSQLError(mgr.name, sql, args, err)
		return nil, err
	}
	return results, nil
}

func (mgr *dbManager) exec(executor sqlExecutor, sql string, args ...interface{}) (sql.Result, error) {
	sql = mgr.convertPlaceholder(sql, mgr.config.Driver)
	args = mgr.sanitizeArgs(sql, args)
	LogSQL(mgr.name, sql, args)
	result, err := executor.Exec(sql, args...)
	if err != nil {
		LogSQLError(mgr.name, sql, args, err)
		return nil, err
	}
	return result, nil
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

func (mgr *dbManager) save(executor sqlExecutor, table string, record *Record) (int64, error) {
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
			for k, v := range record.columns {
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

func (mgr *dbManager) nativeUpsert(executor sqlExecutor, table string, record *Record, pks []string) (int64, error) {
	driver := mgr.config.Driver

	// 如果是 Oracle 或 SQL Server，使用 MERGE 语句
	if driver == Oracle || driver == SQLServer {
		return mgr.mergeUpsert(executor, table, record, pks)
	}

	var columns []string
	var values []interface{}
	var placeholders []string

	for col, val := range record.columns {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, "?")
	}

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
	LogSQL(mgr.name, sqlStr, values)

	// 处理 PostgreSQL 的 ID 返回
	if driver == PostgreSQL {
		if len(pks) == 1 && strings.EqualFold(pks[0], "id") {
			sqlStr += " RETURNING id"
			var id int64
			err := executor.QueryRow(sqlStr, values...).Scan(&id)
			if err != nil {
				return 0, err
			}
			return id, nil
		}
	}

	res, err := executor.Exec(sqlStr, values...)
	if err != nil {
		LogSQLError(mgr.name, sqlStr, values, err)
		return 0, err
	}

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
	columns := make([]string, 0, len(record.columns))
	values := make([]interface{}, 0, len(record.columns))

	// 保持列的顺序一致
	for col, val := range record.columns {
		columns = append(columns, col)
		values = append(values, val)
	}

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
	for _, col := range columns {
		insertCols = append(insertCols, col)
		insertVals = append(insertVals, "s."+col)
	}

	sqlStr += fmt.Sprintf(" WHEN NOT MATCHED THEN INSERT (%s) VALUES (%s)",
		strings.Join(insertCols, ", "),
		strings.Join(insertVals, ", "))

	if driver == SQLServer {
		sqlStr += ";" // SQL Server 的 MERGE 语句必须以分号结束
	}

	sqlStr = mgr.convertPlaceholder(sqlStr, driver)
	LogSQL(mgr.name, sqlStr, values)

	res, err := executor.Exec(sqlStr, values...)
	if err != nil {
		LogSQLError(mgr.name, sqlStr, values, err)
		return 0, err
	}

	return res.RowsAffected()
}

func (mgr *dbManager) insert(executor sqlExecutor, table string, record *Record) (int64, error) {
	if record == nil || len(record.columns) == 0 {
		return 0, fmt.Errorf("record is empty")
	}

	var columns []string
	var values []interface{}
	var placeholders []string

	driver := mgr.config.Driver

	for col, val := range record.columns {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, "?")
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s)", table, joinStrings(columns))

	if driver == PostgreSQL {
		pks, _ := mgr.getPrimaryKeys(executor, table)
		// 只有当存在单列主键且名为 id 时才使用 RETURNING id
		if len(pks) == 1 && strings.EqualFold(pks[0], "id") {
			sql += fmt.Sprintf(" VALUES (%s) RETURNING %s", joinStrings(placeholders), pks[0])
			sql = mgr.convertPlaceholder(sql, driver)
			LogSQL(mgr.name, sql, values)
			var id int64
			err := executor.QueryRow(sql, values...).Scan(&id)
			if err != nil {
				LogSQLError(mgr.name, sql, values, err)
				return 0, err
			}
			return id, nil
		}
		// 否则执行普通插入
		sql += fmt.Sprintf(" VALUES (%s)", joinStrings(placeholders))
		sql = mgr.convertPlaceholder(sql, driver)
		LogSQL(mgr.name, sql, values)
		res, err := executor.Exec(sql, values...)
		if err != nil {
			LogSQLError(mgr.name, sql, values, err)
			return 0, err
		}
		return res.RowsAffected()
	}

	if driver == SQLServer {
		pks, _ := mgr.getPrimaryKeys(executor, table)
		// 只有当可能存在自增列时才使用 SCOPE_IDENTITY
		if len(pks) == 1 {
			sql += fmt.Sprintf(" VALUES (%s); SELECT SCOPE_IDENTITY()", joinStrings(placeholders))
			sql = mgr.convertPlaceholder(sql, driver)
			LogSQL(mgr.name, sql, values)
			var id int64
			err := executor.QueryRow(sql, values...).Scan(&id)
			if err != nil {
				return 0, nil
			}
			return id, nil
		}

		sql += fmt.Sprintf(" VALUES (%s)", joinStrings(placeholders))
		sql = mgr.convertPlaceholder(sql, driver)
		LogSQL(mgr.name, sql, values)
		res, err := executor.Exec(sql, values...)
		if err != nil {
			LogSQLError(mgr.name, sql, values, err)
			return 0, err
		}
		return res.RowsAffected()
	}

	if driver == Oracle {
		pks, _ := mgr.getPrimaryKeys(executor, table)
		sql += fmt.Sprintf(" VALUES (%s)", joinStrings(placeholders))
		sql = mgr.convertPlaceholder(sql, driver)

		idVal := int64(0)
		if len(pks) > 0 {
			// 如果有主键，尝试从 record 中获取第一个主键的值作为返回值
			firstPK := pks[0]
			for k, v := range record.columns {
				if strings.EqualFold(k, firstPK) {
					if i, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64); err == nil {
						idVal = i
					}
					break
				}
			}
		}

		LogSQL(mgr.name, sql, values)
		_, err := executor.Exec(sql, values...)
		if err != nil {
			LogSQLError(mgr.name, sql, values, err)
			return 0, err
		}
		return idVal, nil
	}

	sql += fmt.Sprintf(" VALUES (%s)", joinStrings(placeholders))
	sql = mgr.convertPlaceholder(sql, driver)
	LogSQL(mgr.name, sql, values)
	result, err := executor.Exec(sql, values...)
	if err != nil {
		LogSQLError(mgr.name, sql, values, err)
		return 0, err
	}
	return result.LastInsertId()
}

func (mgr *dbManager) update(executor sqlExecutor, table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	if record == nil || len(record.columns) == 0 {
		return 0, fmt.Errorf("record is empty")
	}

	var setClauses []string
	var values []interface{}

	for col, val := range record.columns {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}
	values = append(values, whereArgs...)

	var sql string
	if where != "" {
		sql = fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, joinStrings(setClauses), where)
	} else {
		sql = fmt.Sprintf("UPDATE %s SET %s", table, joinStrings(setClauses))
	}

	sql = mgr.convertPlaceholder(sql, mgr.config.Driver)
	LogSQL(mgr.name, sql, values)
	result, err := executor.Exec(sql, values...)
	if err != nil {
		LogSQLError(mgr.name, sql, values, err)
		return 0, err
	}
	return result.RowsAffected()
}

func (mgr *dbManager) delete(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (int64, error) {
	if where == "" {
		return 0, fmt.Errorf("where condition is required for delete")
	}

	sql := fmt.Sprintf("DELETE FROM %s WHERE %s", table, where)
	sql = mgr.convertPlaceholder(sql, mgr.config.Driver)
	LogSQL(mgr.name, sql, whereArgs)

	result, err := executor.Exec(sql, whereArgs...)
	if err != nil {
		LogSQLError(mgr.name, sql, whereArgs, err)
		return 0, err
	}
	return result.RowsAffected()
}

func (mgr *dbManager) count(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (int64, error) {
	var sql string
	if where != "" {
		sql = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where)
	} else {
		sql = fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	}
	sql = mgr.convertPlaceholder(sql, mgr.config.Driver)
	whereArgs = mgr.sanitizeArgs(sql, whereArgs)
	LogSQL(mgr.name, sql, whereArgs)

	var count int64
	err := executor.QueryRow(sql, whereArgs...).Scan(&count)
	if err != nil {
		LogSQLError(mgr.name, sql, whereArgs, err)
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
	if len(records) == 0 {
		return 0, fmt.Errorf("no records to insert")
	}

	var totalAffected int64
	driver := mgr.config.Driver

	var columns []string
	for col := range records[0].columns {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		var allValues [][]interface{}
		for _, record := range batch {
			var values []interface{}
			for _, col := range columns {
				values = append(values, record.columns[col])
			}
			allValues = append(allValues, values)
		}

		var sqlStr string
		var flatArgs []interface{}

		if driver == PostgreSQL {
			var allPlaceholders []string
			for rowIdx, values := range allValues {
				var rowPlaceholders []string
				for colIdx := range columns {
					placeholderIdx := rowIdx*len(columns) + colIdx + 1
					rowPlaceholders = append(rowPlaceholders, fmt.Sprintf("$%d", placeholderIdx))
				}
				allPlaceholders = append(allPlaceholders, fmt.Sprintf("(%s)", joinStrings(rowPlaceholders)))
				flatArgs = append(flatArgs, values...)
			}
			sqlStr = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, joinStrings(columns), joinStrings(allPlaceholders))
		} else if driver == SQLServer || driver == Oracle {
			for _, values := range allValues {
				var placeholders []string
				for range columns {
					placeholders = append(placeholders, "?")
				}
				sqlStr = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, joinStrings(columns), joinStrings(placeholders))
				sqlStr = mgr.convertPlaceholder(sqlStr, driver)
				LogSQL(mgr.name, sqlStr, values)
				result, err := executor.Exec(sqlStr, values...)
				if err != nil {
					return totalAffected, err
				}
				affected, _ := result.RowsAffected()
				totalAffected += affected
			}
			continue
		} else {
			var valueLists []string
			var placeholders []string
			for range columns {
				placeholders = append(placeholders, "?")
			}
			for range batch {
				valueLists = append(valueLists, fmt.Sprintf("(%s)", joinStrings(placeholders)))
			}
			for _, values := range allValues {
				flatArgs = append(flatArgs, values...)
			}
			sqlStr = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, joinStrings(columns), joinStrings(valueLists))
		}

		LogSQL(mgr.name, sqlStr, flatArgs)
		result, err := executor.Exec(sqlStr, flatArgs...)
		if err != nil {
			return totalAffected, err
		}
		affected, _ := result.RowsAffected()
		totalAffected += affected
	}
	return totalAffected, nil
}

func (mgr *dbManager) paginate(executor sqlExecutor, sql string, page, pageSize int, args ...interface{}) ([]Record, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	driver := mgr.config.Driver
	lowerSQL := strings.ToLower(sql)
	baseSQL := sql
	if strings.Contains(lowerSQL, " order by ") {
		orderIdx := strings.Index(lowerSQL, " order by ")
		baseSQL = sql[:orderIdx]
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
	LogSQL(mgr.name, countSQL, args)

	var total int64
	err := executor.QueryRow(countSQL, args...).Scan(&total)
	if err != nil {
		LogSQLError(mgr.name, countSQL, args, err)
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var paginatedSQL string
	if driver == SQLServer {
		if strings.Contains(lowerSQL, " order by ") {
			paginatedSQL = fmt.Sprintf("%s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", sql, offset, pageSize)
		} else {
			paginatedSQL = fmt.Sprintf("%s ORDER BY (SELECT NULL) OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", sql, offset, pageSize)
		}
	} else if driver == Oracle {
		if strings.Contains(lowerSQL, " order by ") {
			paginatedSQL = fmt.Sprintf("SELECT a.* FROM (SELECT a.*, ROWNUM rn FROM (%s) a WHERE ROWNUM <= %d) a WHERE rn > %d", sql, offset+pageSize, offset)
		} else {
			paginatedSQL = fmt.Sprintf("SELECT a.* FROM (SELECT a.*, ROWNUM rn FROM (%s ORDER BY (SELECT NULL)) a WHERE ROWNUM <= %d) a WHERE rn > %d", sql, offset+pageSize, offset)
		}
	} else {
		paginatedSQL = fmt.Sprintf("%s LIMIT %d OFFSET %d", sql, pageSize, offset)
	}

	paginatedSQL = mgr.convertPlaceholder(paginatedSQL, driver)
	LogSQL(mgr.name, paginatedSQL, args)

	rows, err := executor.Query(paginatedSQL, args...)
	if err != nil {
		LogSQLError(mgr.name, paginatedSQL, args, err)
		return nil, total, err
	}
	defer rows.Close()

	results, err := scanRecords(rows, driver)
	if err != nil {
		LogSQLError(mgr.name, paginatedSQL, args, err)
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

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Handle []byte conversion for numeric/decimal types
			if b, ok := val.([]byte); ok {
				// Get database type info
				dbType := strings.ToUpper(columnTypes[i].DatabaseTypeName())

				// Check if it's a numeric type that should be a string or number
				// Common numeric types across DBs: DECIMAL, NUMERIC, NUMBER, MONEY, etc.
				if isNumericType(dbType) {
					// Try to convert to float64 if it looks like a decimal
					if s := string(b); s != "" {
						if _, err := strconv.ParseFloat(s, 64); err == nil {
							// If it's a whole number, maybe int64 is better?
							// But for precision, string or float64 is safer for decimals.
							// For now, let's use string to preserve precision for DECIMAL
							entry[col] = s
						} else {
							entry[col] = s
						}
					} else {
						entry[col] = nil
					}
					continue
				}

				// For other types, if it's []byte, convert to string by default for convenience
				// except if it's explicitly a BLOB/BINARY type (though DatabaseTypeName varies)
				if !isBinaryType(dbType) {
					entry[col] = string(b)
					continue
				}
			}

			entry[col] = val
		}
		results = append(results, entry)
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
func optimizeCountSQL(sql string) (string, bool) {
	lower := strings.ToLower(sql)

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
	if strings.Contains(sql[:fromIdx], "(") {
		return "", false
	}

	// 构建优化的 COUNT 语句
	optimized := "SELECT COUNT(*) " + sql[fromIdx:]
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

// GetCurrentDB returns the current database manager
func GetCurrentDB() *dbManager {
	if multiMgr == nil {
		panic("dbkit: database not initialized. Please call dbkit.OpenDatabase() or dbkit.Register() before using dbkit operations")
	}

	multiMgr.mu.RLock()
	defer multiMgr.mu.RUnlock()

	if multiMgr.currentDB == "" {
		panic("dbkit: no current database selected")
	}

	if dbMgr, exists := multiMgr.databases[multiMgr.currentDB]; exists {
		return dbMgr
	}

	panic(fmt.Sprintf("dbkit: current database '%s' not found", multiMgr.currentDB))
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
func (mgr *dbManager) convertPlaceholder(sql string, driver DriverType) string {
	return mgr.convertPlaceholderWithOffset(sql, driver, 0)
}

// convertPlaceholderWithOffset converts ? placeholders with an index offset
func (mgr *dbManager) convertPlaceholderWithOffset(sql string, driver DriverType, offset int) string {
	if driver == MySQL || driver == SQLite3 {
		return sql
	}

	var builder strings.Builder
	builder.Grow(len(sql) + 10)
	paramIndex := 1 + offset
	inString := false

	for i := 0; i < len(sql); i++ {
		char := sql[i]
		// Handle string literals to avoid replacing question marks inside them
		if char == '\'' {
			if i+1 < len(sql) && sql[i+1] == '\'' { // Handle escaped single quote ''
				builder.WriteByte('\'')
				builder.WriteByte('\'')
				i++
				continue
			}
			inString = !inString
			builder.WriteByte('\'')
			continue
		}

		if char == '?' && !inString {
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
func (mgr *dbManager) sanitizeArgs(sql string, args []interface{}) []interface{} {
	if len(args) == 0 {
		return args
	}

	placeholderCount := 0
	switch mgr.config.Driver {
	case PostgreSQL:
		// 统计 $1, $2... 的数量
		for i := 1; ; i++ {
			if strings.Contains(sql, fmt.Sprintf("$%d", i)) {
				placeholderCount = i
			} else {
				break
			}
		}
	case SQLServer:
		// 统计 @p1, @p2... 的数量
		for i := 1; ; i++ {
			if strings.Contains(sql, fmt.Sprintf("@p%d", i)) {
				placeholderCount = i
			} else {
				break
			}
		}
	case Oracle:
		// 统计 :1, :2... 的数量
		for i := 1; ; i++ {
			if strings.Contains(sql, fmt.Sprintf(":%d", i)) {
				placeholderCount = i
			} else {
				break
			}
		}
	case MySQL, SQLite3:
		// 统计 ? 的数量
		placeholderCount = strings.Count(sql, "?")
	}

	if placeholderCount == 0 {
		return nil
	}

	if len(args) > placeholderCount {
		// 如果参数过多，截断多余部分
		return args[:placeholderCount]
	}

	return args
}

// joinStrings joins strings with commas
func joinStrings(strs []string) string {
	return strings.Join(strs, ", ")
}
