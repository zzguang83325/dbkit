package dbkit

import (
	"fmt"
	"strings"
	"time"
)

// QueryBuilder represents a fluent interface for building SQL queries
type QueryBuilder struct {
	db        *DB
	tx        *Tx
	table     string
	selectSql string
	whereSql  []string
	whereArgs []interface{}
	orderBy   string
	limit     int
	offset    int
	cacheName string
	cacheTTL  time.Duration
	lastErr   error
}

// Table starts a new query builder for the default database
func Table(name string) *QueryBuilder {

	if err := validateIdentifier(name); err != nil {
		return &QueryBuilder{lastErr: err}
	}

	db, err := defaultDB()
	if err != nil {
		return &QueryBuilder{lastErr: err}
	}
	return &QueryBuilder{
		db:        db,
		table:     name,
		selectSql: "*",
	}
}

// Table method for DB instance
func (db *DB) Table(name string) *QueryBuilder {

	if err := validateIdentifier(name); err != nil {
		return &QueryBuilder{lastErr: err}
	}
	return &QueryBuilder{
		db:        db,
		table:     name,
		selectSql: "*",
		cacheName: db.cacheName,
		cacheTTL:  db.cacheTTL,
		lastErr:   db.lastErr,
	}
}

// Table method for Tx instance
func (tx *Tx) Table(name string) *QueryBuilder {
	if err := validateIdentifier(name); err != nil {
		return &QueryBuilder{lastErr: err}
	}

	return &QueryBuilder{
		tx:        tx,
		table:     name,
		selectSql: "*",
		cacheName: tx.cacheName,
		cacheTTL:  tx.cacheTTL,
	}
}

// Select specifies the columns to select
func (qb *QueryBuilder) Select(columns string) *QueryBuilder {
	qb.selectSql = columns
	return qb
}

// Where adds a where clause to the query
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.whereSql = append(qb.whereSql, condition)
	qb.whereArgs = append(qb.whereArgs, args...)
	return qb
}

// And is an alias for Where
func (qb *QueryBuilder) And(condition string, args ...interface{}) *QueryBuilder {
	return qb.Where(condition, args...)
}

// OrderBy adds an order by clause to the query
func (qb *QueryBuilder) OrderBy(orderBy string) *QueryBuilder {
	qb.orderBy = orderBy
	return qb
}

// Limit adds a limit clause to the query
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset adds an offset clause to the query
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Cache enables caching for the query
func (qb *QueryBuilder) Cache(name string, ttl ...time.Duration) *QueryBuilder {
	qb.cacheName = name
	if len(ttl) > 0 {
		qb.cacheTTL = ttl[0]
	} else {
		qb.cacheTTL = -1
	}
	return qb
}

// buildSelectSql constructs the final SELECT SQL string
func (qb *QueryBuilder) buildSelectSql() (string, []interface{}) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("SELECT %s FROM %s", qb.selectSql, qb.table))

	if len(qb.whereSql) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(qb.whereSql, " AND "))
	}

	if qb.orderBy != "" {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(qb.orderBy)
	}

	if qb.limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	if qb.offset > 0 {
		sb.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return sb.String(), qb.whereArgs
}

// Query executes the query and returns a slice of Records
func (qb *QueryBuilder) Query() ([]Record, error) {
	if qb.lastErr != nil {
		return nil, qb.lastErr
	}
	sql, args := qb.buildSelectSql()

	// Handle caching
	if qb.cacheName != "" && qb.tx == nil {
		cacheKey := qb.generateCacheKey(sql, args)
		if val, ok := CacheGet(qb.cacheName, cacheKey); ok {
			if records, ok := val.([]Record); ok {
				return records, nil
			}
		}
		// If not in cache, query and store
		records, err := qb.db.Query(sql, args...)
		if err == nil {
			CacheSet(qb.cacheName, cacheKey, records, qb.cacheTTL)
		}
		return records, err
	}

	if qb.tx != nil {
		return qb.tx.Query(sql, args...)
	}
	return qb.db.Query(sql, args...)
}

// generateCacheKey creates a unique key for the query and its arguments
func (qb *QueryBuilder) generateCacheKey(sql string, args []interface{}) string {
	dbName := ""
	if qb.db != nil {
		dbName = qb.db.dbMgr.name
	} else if qb.tx != nil {
		dbName = qb.tx.dbMgr.name
	}
	return GenerateCacheKey(dbName, sql, args...)
}

// Find is an alias for Query
func (qb *QueryBuilder) Find() ([]Record, error) {
	return qb.Query()
}

// FindToDbModel executes the query and converts the results to the provided slice pointer
func (qb *QueryBuilder) FindToDbModel(dest interface{}) error {
	records, err := qb.Find()
	if err != nil {
		return err
	}
	return ToStructs(records, dest)
}

func (qb *QueryBuilder) QueryToDbModel(dest interface{}) error {
	records, err := qb.Find()
	if err != nil {
		return err
	}
	return ToStructs(records, dest)
}

// QueryFirst executes the query and returns the first Record
func (qb *QueryBuilder) QueryFirst() (*Record, error) {
	if qb.lastErr != nil {
		return nil, qb.lastErr
	}
	// Temporarily set limit to 1 if not set or set to something else
	oldLimit := qb.limit
	qb.limit = 1
	sql, args := qb.buildSelectSql()
	qb.limit = oldLimit

	// Handle caching
	if qb.cacheName != "" && qb.tx == nil {
		cacheKey := qb.generateCacheKey(sql, args) + "_first"
		if val, ok := CacheGet(qb.cacheName, cacheKey); ok {
			if record, ok := val.(*Record); ok {
				return record, nil
			}
		}
		// If not in cache, query and store
		record, err := qb.db.QueryFirst(sql, args...)
		if err == nil && record != nil {
			CacheSet(qb.cacheName, cacheKey, record, qb.cacheTTL)
		}
		return record, err
	}

	if qb.tx != nil {
		return qb.tx.QueryFirst(sql, args...)
	}
	return qb.db.QueryFirst(sql, args...)
}

// FindFirst is an alias for QueryFirst
func (qb *QueryBuilder) FindFirst() (*Record, error) {
	return qb.QueryFirst()
}

// FindFirstToDbModel executes the query and converts the first result to the provided struct pointer
func (qb *QueryBuilder) FindFirstToDbModel(dest interface{}) error {
	record, err := qb.FindFirst()
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("dbkit: no record found")
	}
	return ToStruct(record, dest)
}

// Paginate executes the query with pagination and returns a Page object
func (qb *QueryBuilder) Paginate(pageNumber, pageSize int) (*Page[Record], error) {
	if qb.lastErr != nil {
		return nil, qb.lastErr
	}

	whereSql := ""
	if len(qb.whereSql) > 0 {
		whereSql = strings.Join(qb.whereSql, " AND ")
	}

	// Handle caching
	if qb.cacheName != "" && qb.tx == nil {
		sql, args := qb.buildSelectSql()
		cacheKey := qb.generateCacheKey(sql, args) + fmt.Sprintf("_p%d_s%d", pageNumber, pageSize)
		if val, ok := CacheGet(qb.cacheName, cacheKey); ok {
			var pageObj *Page[Record]
			if convertCacheValue(val, &pageObj) {
				return pageObj, nil
			}
		}

		// If not in cache, query and store
		pageObj, err := qb.db.Paginate(pageNumber, pageSize, qb.selectSql, qb.table, whereSql, qb.orderBy, qb.whereArgs...)
		if err == nil {
			CacheSet(qb.cacheName, cacheKey, pageObj, qb.cacheTTL)
		}
		return pageObj, err
	}

	if qb.tx != nil {
		return qb.tx.Paginate(pageNumber, pageSize, qb.selectSql, qb.table, whereSql, qb.orderBy, qb.whereArgs...)
	}
	return qb.db.Paginate(pageNumber, pageSize, qb.selectSql, qb.table, whereSql, qb.orderBy, qb.whereArgs...)
}

// Update executes an update query with the criteria in the builder
func (qb *QueryBuilder) Update(record *Record) (int64, error) {
	if qb.lastErr != nil {
		return 0, qb.lastErr
	}

	whereSql := ""
	if len(qb.whereSql) > 0 {
		whereSql = strings.Join(qb.whereSql, " AND ")
	}

	if qb.tx != nil {
		return qb.tx.Update(qb.table, record, whereSql, qb.whereArgs...)
	}
	return qb.db.Update(qb.table, record, whereSql, qb.whereArgs...)
}

// Delete executes a delete query with the criteria in the builder
func (qb *QueryBuilder) Delete() (int64, error) {
	if qb.lastErr != nil {
		return 0, qb.lastErr
	}
	if qb.table == "" {
		return 0, fmt.Errorf("dbkit: table name is required for Delete")
	}
	if len(qb.whereSql) == 0 {
		return 0, fmt.Errorf("dbkit: Delete operation requires at least one Where condition for safety")
	}

	whereSql := strings.Join(qb.whereSql, " AND ")

	if qb.tx != nil {
		return qb.tx.Delete(qb.table, whereSql, qb.whereArgs...)
	}
	return qb.db.Delete(qb.table, whereSql, qb.whereArgs...)
}

// Count returns the number of records matching the criteria
func (qb *QueryBuilder) Count() (int64, error) {
	if qb.lastErr != nil {
		return 0, qb.lastErr
	}

	whereSql := ""
	if len(qb.whereSql) > 0 {
		whereSql = strings.Join(qb.whereSql, " AND ")
	}

	// Handle caching
	if qb.cacheName != "" && qb.tx == nil {
		sql, args := qb.buildSelectSql()
		cacheKey := qb.generateCacheKey(sql, args) + "_count"
		if val, ok := CacheGet(qb.cacheName, cacheKey); ok {
			if count, ok := val.(int64); ok {
				return count, nil
			}
		}

		// If not in cache, query and store
		count, err := qb.db.Count(qb.table, whereSql, qb.whereArgs...)
		if err == nil {
			CacheSet(qb.cacheName, cacheKey, count, qb.cacheTTL)
		}
		return count, err
	}

	if qb.tx != nil {
		return qb.tx.Count(qb.table, whereSql, qb.whereArgs...)
	}
	return qb.db.Count(qb.table, whereSql, qb.whereArgs...)
}
