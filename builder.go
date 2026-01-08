package dbkit

import (
	"fmt"
	"strings"
	"time"
)

// JoinClause represents a single JOIN clause in a query
type JoinClause struct {
	joinType  string        // "JOIN", "LEFT JOIN", "RIGHT JOIN", "INNER JOIN"
	table     string        // table name to join
	condition string        // join condition (e.g., "users.id = orders.user_id")
	args      []interface{} // arguments for parameterized conditions
}

// SelectSubquery represents a subquery used as a SELECT field
type SelectSubquery struct {
	subquery *Subquery
	alias    string
}

// QueryBuilder represents a fluent interface for building SQL queries
type QueryBuilder struct {
	db                  *DB
	tx                  *Tx
	table               string
	selectSql           string
	whereSql            []string
	whereArgs           []interface{}
	orWhereSql          []string      // OR conditions
	orWhereArgs         []interface{} // OR condition arguments
	orderBy             string
	groupBy             string        // GROUP BY clause
	havingSql           []string      // HAVING conditions
	havingArgs          []interface{} // HAVING arguments
	limit               int
	offset              int
	cacheRepositoryName string
	cacheTTL            time.Duration
	cacheProvider       CacheProvider // 指定的缓存提供者（nil 表示使用默认缓存）
	timeout             time.Duration
	lastErr             error
	withTrashed         bool             // Include soft-deleted records
	onlyTrashed         bool             // Only query soft-deleted records
	skipTimestamps      bool             // Skip auto timestamps for insert/update
	joins               []JoinClause     // JOIN clauses
	subqueryTable       *Subquery        // FROM subquery
	subqueryAlias       string           // FROM subquery alias
	selectSubqueries    []SelectSubquery // SELECT subqueries
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
		db:                  db,
		table:               name,
		selectSql:           "*",
		cacheRepositoryName: db.cacheRepositoryName,
		cacheTTL:            db.cacheTTL,
		cacheProvider:       db.cacheProvider, // 继承 DB 的缓存提供者
		lastErr:             db.lastErr,
	}
}

// Table method for Tx instance
func (tx *Tx) Table(name string) *QueryBuilder {
	if err := validateIdentifier(name); err != nil {
		return &QueryBuilder{lastErr: err}
	}

	return &QueryBuilder{
		tx:                  tx,
		table:               name,
		selectSql:           "*",
		cacheRepositoryName: tx.cacheRepositoryName,
		cacheTTL:            tx.cacheTTL,
		cacheProvider:       tx.cacheProvider, // 继承 Tx 的缓存提供者
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

// OrWhere adds an OR condition to the query
func (qb *QueryBuilder) OrWhere(condition string, args ...interface{}) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	qb.orWhereSql = append(qb.orWhereSql, condition)
	qb.orWhereArgs = append(qb.orWhereArgs, args...)
	return qb
}

// WhereGroupFunc is a function type for building grouped conditions
type WhereGroupFunc func(qb *QueryBuilder) *QueryBuilder

// WhereGroup adds a grouped AND condition: WHERE ... AND (grouped conditions)
func (qb *QueryBuilder) WhereGroup(fn WhereGroupFunc) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	// Create a temporary QueryBuilder to collect the grouped conditions
	tempQb := &QueryBuilder{table: qb.table, selectSql: "*"}
	fn(tempQb)

	// Build the grouped condition
	groupedCondition := buildGroupedCondition(tempQb)
	if groupedCondition != "" {
		qb.whereSql = append(qb.whereSql, "("+groupedCondition+")")
		qb.whereArgs = append(qb.whereArgs, tempQb.whereArgs...)
		qb.whereArgs = append(qb.whereArgs, tempQb.orWhereArgs...)
	}
	return qb
}

// OrWhereGroup adds a grouped OR condition: WHERE ... OR (grouped conditions)
func (qb *QueryBuilder) OrWhereGroup(fn WhereGroupFunc) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	// Create a temporary QueryBuilder to collect the grouped conditions
	tempQb := &QueryBuilder{table: qb.table, selectSql: "*"}
	fn(tempQb)

	// Build the grouped condition
	groupedCondition := buildGroupedCondition(tempQb)
	if groupedCondition != "" {
		qb.orWhereSql = append(qb.orWhereSql, "("+groupedCondition+")")
		qb.orWhereArgs = append(qb.orWhereArgs, tempQb.whereArgs...)
		qb.orWhereArgs = append(qb.orWhereArgs, tempQb.orWhereArgs...)
	}
	return qb
}

// buildGroupedCondition builds the condition string from a temporary QueryBuilder
func buildGroupedCondition(tempQb *QueryBuilder) string {
	var parts []string

	// Add AND conditions
	if len(tempQb.whereSql) > 0 {
		parts = append(parts, strings.Join(tempQb.whereSql, " AND "))
	}

	// Add OR conditions
	if len(tempQb.orWhereSql) > 0 {
		if len(parts) > 0 {
			// If we have both AND and OR, combine them
			andPart := parts[0]
			orPart := strings.Join(tempQb.orWhereSql, " OR ")
			return andPart + " OR " + orPart
		}
		parts = append(parts, strings.Join(tempQb.orWhereSql, " OR "))
	}

	if len(parts) == 0 {
		return ""
	}
	return parts[0]
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

// addJoin is an internal method to add a join clause
func (qb *QueryBuilder) addJoin(joinType, table, condition string, args ...interface{}) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	if err := validateIdentifier(table); err != nil {
		qb.lastErr = err
		return qb
	}
	qb.joins = append(qb.joins, JoinClause{
		joinType:  joinType,
		table:     table,
		condition: condition,
		args:      args,
	})
	return qb
}

// Join adds a JOIN clause to the query
func (qb *QueryBuilder) Join(table, condition string, args ...interface{}) *QueryBuilder {
	return qb.addJoin("JOIN", table, condition, args...)
}

// LeftJoin adds a LEFT JOIN clause to the query
func (qb *QueryBuilder) LeftJoin(table, condition string, args ...interface{}) *QueryBuilder {
	return qb.addJoin("LEFT JOIN", table, condition, args...)
}

// RightJoin adds a RIGHT JOIN clause to the query
func (qb *QueryBuilder) RightJoin(table, condition string, args ...interface{}) *QueryBuilder {
	return qb.addJoin("RIGHT JOIN", table, condition, args...)
}

// InnerJoin adds an INNER JOIN clause to the query
func (qb *QueryBuilder) InnerJoin(table, condition string, args ...interface{}) *QueryBuilder {
	return qb.addJoin("INNER JOIN", table, condition, args...)
}

// WhereIn adds a WHERE column IN (subquery) clause
func (qb *QueryBuilder) WhereIn(column string, sub *Subquery) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	if sub == nil {
		return qb
	}
	subSQL, subArgs := sub.ToSQL()
	if subSQL == "" {
		return qb
	}
	qb.whereSql = append(qb.whereSql, fmt.Sprintf("%s IN (%s)", column, subSQL))
	qb.whereArgs = append(qb.whereArgs, subArgs...)
	return qb
}

// WhereNotIn adds a WHERE column NOT IN (subquery) clause
func (qb *QueryBuilder) WhereNotIn(column string, sub *Subquery) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	if sub == nil {
		return qb
	}
	subSQL, subArgs := sub.ToSQL()
	if subSQL == "" {
		return qb
	}
	qb.whereSql = append(qb.whereSql, fmt.Sprintf("%s NOT IN (%s)", column, subSQL))
	qb.whereArgs = append(qb.whereArgs, subArgs...)
	return qb
}

// WhereInValues adds a WHERE column IN (?, ?, ...) clause with a list of values
func (qb *QueryBuilder) WhereInValues(column string, values []interface{}) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	if len(values) == 0 {
		return qb
	}
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}
	qb.whereSql = append(qb.whereSql, fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", ")))
	qb.whereArgs = append(qb.whereArgs, values...)
	return qb
}

// WhereNotInValues adds a WHERE column NOT IN (?, ?, ...) clause with a list of values
func (qb *QueryBuilder) WhereNotInValues(column string, values []interface{}) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	if len(values) == 0 {
		return qb
	}
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}
	qb.whereSql = append(qb.whereSql, fmt.Sprintf("%s NOT IN (%s)", column, strings.Join(placeholders, ", ")))
	qb.whereArgs = append(qb.whereArgs, values...)
	return qb
}

// WhereBetween adds a WHERE column BETWEEN ? AND ? clause
func (qb *QueryBuilder) WhereBetween(column string, min, max interface{}) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	qb.whereSql = append(qb.whereSql, fmt.Sprintf("%s BETWEEN ? AND ?", column))
	qb.whereArgs = append(qb.whereArgs, min, max)
	return qb
}

// WhereNotBetween adds a WHERE column NOT BETWEEN ? AND ? clause
func (qb *QueryBuilder) WhereNotBetween(column string, min, max interface{}) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	qb.whereSql = append(qb.whereSql, fmt.Sprintf("%s NOT BETWEEN ? AND ?", column))
	qb.whereArgs = append(qb.whereArgs, min, max)
	return qb
}

// WhereNull adds a WHERE column IS NULL clause
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	qb.whereSql = append(qb.whereSql, fmt.Sprintf("%s IS NULL", column))
	return qb
}

// WhereNotNull adds a WHERE column IS NOT NULL clause
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	qb.whereSql = append(qb.whereSql, fmt.Sprintf("%s IS NOT NULL", column))
	return qb
}

// GroupBy adds a GROUP BY clause to the query
func (qb *QueryBuilder) GroupBy(columns string) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	qb.groupBy = columns
	return qb
}

// Having adds a HAVING clause to the query
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	qb.havingSql = append(qb.havingSql, condition)
	qb.havingArgs = append(qb.havingArgs, args...)
	return qb
}

// TableSubquery sets a subquery as the FROM source
func (qb *QueryBuilder) TableSubquery(sub *Subquery, alias string) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	if alias == "" {
		qb.lastErr = fmt.Errorf("dbkit: alias is required for FROM subquery")
		return qb
	}
	qb.subqueryTable = sub
	qb.subqueryAlias = alias
	return qb
}

// SelectSubquery adds a subquery as a SELECT field
func (qb *QueryBuilder) SelectSubquery(sub *Subquery, alias string) *QueryBuilder {
	if qb.lastErr != nil {
		return qb
	}
	if sub == nil {
		return qb
	}
	qb.selectSubqueries = append(qb.selectSubqueries, SelectSubquery{
		subquery: sub,
		alias:    alias,
	})
	return qb
}

// Cache enables caching for the query
func (qb *QueryBuilder) Cache(cacheRepositoryName string, ttl ...time.Duration) *QueryBuilder {
	qb.cacheRepositoryName = cacheRepositoryName
	qb.cacheProvider = nil // 使用默认缓存
	if len(ttl) > 0 {
		qb.cacheTTL = ttl[0]
	} else {
		qb.cacheTTL = -1
	}
	return qb
}

// LocalCache 使用本地缓存
func (qb *QueryBuilder) LocalCache(cacheRepositoryName string, ttl ...time.Duration) *QueryBuilder {
	qb.cacheRepositoryName = cacheRepositoryName
	qb.cacheProvider = GetLocalCacheInstance()
	if len(ttl) > 0 {
		qb.cacheTTL = ttl[0]
	} else {
		qb.cacheTTL = -1
	}
	return qb
}

// RedisCache 使用 Redis 缓存
func (qb *QueryBuilder) RedisCache(cacheRepositoryName string, ttl ...time.Duration) *QueryBuilder {
	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		// 如果 Redis 缓存未初始化，记录错误但不中断链式调用
		LogError("Redis cache not initialized for QueryBuilder", map[string]interface{}{
			"table":               qb.table,
			"cacheRepositoryName": cacheRepositoryName,
		})
		return qb
	}

	qb.cacheRepositoryName = cacheRepositoryName
	qb.cacheProvider = redisCache
	if len(ttl) > 0 {
		qb.cacheTTL = ttl[0]
	} else {
		qb.cacheTTL = -1
	}
	return qb
}

// Timeout sets the query timeout
func (qb *QueryBuilder) Timeout(d time.Duration) *QueryBuilder {
	qb.timeout = d
	return qb
}

// getEffectiveCache 获取当前有效的缓存提供者
// 优先级: QueryBuilder.cacheProvider > DB/Tx.cacheProvider > 全局默认缓存
func (qb *QueryBuilder) getEffectiveCache() CacheProvider {
	if qb.cacheProvider != nil {
		return qb.cacheProvider
	}
	if qb.db != nil && qb.db.cacheProvider != nil {
		return qb.db.cacheProvider
	}
	if qb.tx != nil && qb.tx.cacheProvider != nil {
		return qb.tx.cacheProvider
	}
	return GetCache()
}

// buildSelectSql constructs the final SELECT SQL string
func (qb *QueryBuilder) buildSelectSql() (string, []interface{}) {
	var sb strings.Builder
	var allArgs []interface{}

	// Build SELECT clause with optional subqueries
	selectPart := qb.selectSql
	if len(qb.selectSubqueries) > 0 {
		for _, ss := range qb.selectSubqueries {
			subSQL, subArgs := ss.subquery.ToSQL()
			if subSQL != "" {
				if selectPart != "" && selectPart != "*" {
					selectPart += ", "
				} else if selectPart == "*" {
					selectPart += ", "
				}
				selectPart += fmt.Sprintf("(%s) AS %s", subSQL, ss.alias)
				allArgs = append(allArgs, subArgs...)
			}
		}
	}

	// Build FROM clause (table or subquery)
	var fromPart string
	if qb.subqueryTable != nil && qb.subqueryAlias != "" {
		subSQL, subArgs := qb.subqueryTable.ToSQL()
		fromPart = fmt.Sprintf("(%s) AS %s", subSQL, qb.subqueryAlias)
		allArgs = append(allArgs, subArgs...)
	} else {
		fromPart = qb.table
	}

	sb.WriteString(fmt.Sprintf("SELECT %s FROM %s", selectPart, fromPart))

	// Add JOIN clauses
	for _, join := range qb.joins {
		sb.WriteString(fmt.Sprintf(" %s %s ON %s", join.joinType, join.table, join.condition))
		allArgs = append(allArgs, join.args...)
	}

	// Collect all where conditions
	whereClauses := make([]string, 0, len(qb.whereSql)+1)
	whereClauses = append(whereClauses, qb.whereSql...)

	// Add soft delete filter if applicable
	softDeleteCondition := qb.getSoftDeleteCondition()
	if softDeleteCondition != "" {
		whereClauses = append(whereClauses, softDeleteCondition)
	}

	// Build WHERE clause with AND and OR conditions
	if len(whereClauses) > 0 || len(qb.orWhereSql) > 0 {
		sb.WriteString(" WHERE ")

		var wherePartBuilt bool
		if len(whereClauses) > 0 {
			// If we have both AND and OR conditions, group AND conditions with parentheses
			if len(qb.orWhereSql) > 0 {
				sb.WriteString("(")
				sb.WriteString(strings.Join(whereClauses, " AND "))
				sb.WriteString(")")
			} else {
				sb.WriteString(strings.Join(whereClauses, " AND "))
			}
			wherePartBuilt = true
		}

		// Add OR conditions
		if len(qb.orWhereSql) > 0 {
			if wherePartBuilt {
				sb.WriteString(" OR ")
			}
			sb.WriteString(strings.Join(qb.orWhereSql, " OR "))
		}
	}

	// Append WHERE args after JOIN args (AND args first, then OR args)
	allArgs = append(allArgs, qb.whereArgs...)
	allArgs = append(allArgs, qb.orWhereArgs...)

	// Add GROUP BY clause
	if qb.groupBy != "" {
		sb.WriteString(" GROUP BY ")
		sb.WriteString(qb.groupBy)
	}

	// Add HAVING clause
	if len(qb.havingSql) > 0 {
		sb.WriteString(" HAVING ")
		sb.WriteString(strings.Join(qb.havingSql, " AND "))
		allArgs = append(allArgs, qb.havingArgs...)
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

	return sb.String(), allArgs
}

// removeLimitOffset 移除SQL语句中的LIMIT和OFFSET子句
// 因为Paginate会自动处理分页逻辑
func removeLimitOffset(sql string) string {
	// 使用正则表达式移除LIMIT和OFFSET子句
	// 这里使用简单的字符串处理，因为buildSelectSql生成的SQL格式是可预测的

	// 移除 LIMIT 子句
	if idx := strings.LastIndex(strings.ToUpper(sql), " LIMIT "); idx != -1 {
		// 找到LIMIT后面的数字结束位置
		remaining := sql[idx+7:] // 跳过" LIMIT "
		var endIdx int
		for i, r := range remaining {
			if r < '0' || r > '9' {
				endIdx = i
				break
			}
		}
		if endIdx == 0 {
			endIdx = len(remaining)
		}
		sql = sql[:idx] + remaining[endIdx:]
	}

	// 移除 OFFSET 子句
	if idx := strings.LastIndex(strings.ToUpper(sql), " OFFSET "); idx != -1 {
		// 找到OFFSET后面的数字结束位置
		remaining := sql[idx+8:] // 跳过" OFFSET "
		var endIdx int
		for i, r := range remaining {
			if r < '0' || r > '9' {
				endIdx = i
				break
			}
		}
		if endIdx == 0 {
			endIdx = len(remaining)
		}
		sql = sql[:idx] + remaining[endIdx:]
	}

	return strings.TrimSpace(sql)
}

// getSoftDeleteCondition returns the soft delete filter condition
func (qb *QueryBuilder) getSoftDeleteCondition() string {
	var mgr *dbManager
	if qb.db != nil && qb.db.dbMgr != nil {
		mgr = qb.db.dbMgr
	} else if qb.tx != nil && qb.tx.dbMgr != nil {
		mgr = qb.tx.dbMgr
	}
	if mgr == nil {
		return ""
	}
	return mgr.buildSoftDeleteCondition(qb.table, qb.withTrashed, qb.onlyTrashed)
}

// Query executes the query and returns a slice of Records
func (qb *QueryBuilder) Query() ([]Record, error) {
	if qb.lastErr != nil {
		return nil, qb.lastErr
	}
	sql, args := qb.buildSelectSql()

	// Handle caching
	if qb.cacheRepositoryName != "" && qb.tx == nil {
		cache := qb.getEffectiveCache()
		cacheKey := qb.generateCacheKey(sql, args)
		if val, ok := cache.CacheGet(qb.cacheRepositoryName, cacheKey); ok {
			if records, ok := val.([]Record); ok {
				return records, nil
			}
		}
		// If not in cache, query and store
		db := qb.db
		if qb.timeout > 0 {
			db = &DB{dbMgr: qb.db.dbMgr, timeout: qb.timeout}
		}
		records, err := db.Query(sql, args...)
		if err == nil {
			cache.CacheSet(qb.cacheRepositoryName, cacheKey, records, qb.cacheTTL)
		}
		return records, err
	}

	if qb.tx != nil {
		if qb.timeout > 0 {
			tx := &Tx{tx: qb.tx.tx, dbMgr: qb.tx.dbMgr, timeout: qb.timeout}
			return tx.Query(sql, args...)
		}
		return qb.tx.Query(sql, args...)
	}

	if qb.timeout > 0 {
		db := &DB{dbMgr: qb.db.dbMgr, timeout: qb.timeout}
		return db.Query(sql, args...)
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
	if qb.cacheRepositoryName != "" && qb.tx == nil {
		cache := qb.getEffectiveCache()
		cacheKey := qb.generateCacheKey(sql, args) + "_first"
		if val, ok := cache.CacheGet(qb.cacheRepositoryName, cacheKey); ok {
			if record, ok := val.(*Record); ok {
				return record, nil
			}
		}
		// If not in cache, query and store
		db := qb.db
		if qb.timeout > 0 {
			db = &DB{dbMgr: qb.db.dbMgr, timeout: qb.timeout}
		}
		record, err := db.QueryFirst(sql, args...)
		if err == nil && record != nil {
			cache.CacheSet(qb.cacheRepositoryName, cacheKey, record, qb.cacheTTL)
		}
		return record, err
	}

	if qb.tx != nil {
		if qb.timeout > 0 {
			tx := &Tx{tx: qb.tx.tx, dbMgr: qb.tx.dbMgr, timeout: qb.timeout}
			return tx.QueryFirst(sql, args...)
		}
		return qb.tx.QueryFirst(sql, args...)
	}

	if qb.timeout > 0 {
		db := &DB{dbMgr: qb.db.dbMgr, timeout: qb.timeout}
		return db.QueryFirst(sql, args...)
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

	// 构建完整的SQL语句（不包含LIMIT和OFFSET，因为分页逻辑会处理）
	sql, args := qb.buildSelectSql()

	// 移除LIMIT和OFFSET子句，因为Paginate会处理分页
	sql = removeLimitOffset(sql)

	// 处理缓存
	if qb.cacheRepositoryName != "" && qb.tx == nil {
		cache := qb.getEffectiveCache()
		cacheKey := qb.generateCacheKey(sql, args) + fmt.Sprintf("_p%d_s%d", pageNumber, pageSize)
		if val, ok := cache.CacheGet(qb.cacheRepositoryName, cacheKey); ok {
			var pageObj *Page[Record]
			if convertCacheValue(val, &pageObj) {
				return pageObj, nil
			}
		}

		// 使用新的Paginate实现
		var pageObj *Page[Record]
		var err error
		if qb.tx != nil {
			pageObj, err = qb.tx.Paginate(pageNumber, pageSize, sql, args...)
		} else {
			pageObj, err = qb.db.Paginate(pageNumber, pageSize, sql, args...)
		}

		if err == nil {
			cache.CacheSet(qb.cacheRepositoryName, cacheKey, pageObj, qb.cacheTTL)
		}
		return pageObj, err
	}

	// 直接使用新的Paginate实现
	if qb.tx != nil {
		return qb.tx.Paginate(pageNumber, pageSize, sql, args...)
	}
	return qb.db.Paginate(pageNumber, pageSize, sql, args...)
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
		return qb.tx.updateWithOptions(qb.table, record, whereSql, qb.skipTimestamps, qb.whereArgs...)
	}
	return qb.db.updateWithOptions(qb.table, record, whereSql, qb.skipTimestamps, qb.whereArgs...)
}

// WithoutTimestamps disables auto timestamps for insert/update operations
func (qb *QueryBuilder) WithoutTimestamps() *QueryBuilder {
	qb.skipTimestamps = true
	return qb
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

	// Collect all where conditions including soft delete filter
	whereClauses := make([]string, 0, len(qb.whereSql)+1)
	whereClauses = append(whereClauses, qb.whereSql...)

	// Add soft delete filter if applicable
	softDeleteCondition := qb.getSoftDeleteCondition()
	if softDeleteCondition != "" {
		whereClauses = append(whereClauses, softDeleteCondition)
	}

	whereSql := ""
	if len(whereClauses) > 0 {
		whereSql = strings.Join(whereClauses, " AND ")
	}

	// Handle caching
	if qb.cacheRepositoryName != "" && qb.tx == nil {
		cache := qb.getEffectiveCache()
		sql, args := qb.buildSelectSql()
		cacheKey := qb.generateCacheKey(sql, args) + "_count"
		if val, ok := cache.CacheGet(qb.cacheRepositoryName, cacheKey); ok {
			if count, ok := val.(int64); ok {
				return count, nil
			}
		}

		// If not in cache, query and store
		count, err := qb.db.Count(qb.table, whereSql, qb.whereArgs...)
		if err == nil {
			cache.CacheSet(qb.cacheRepositoryName, cacheKey, count, qb.cacheTTL)
		}
		return count, err
	}

	if qb.tx != nil {
		return qb.tx.Count(qb.table, whereSql, qb.whereArgs...)
	}
	return qb.db.Count(qb.table, whereSql, qb.whereArgs...)
}

// WithTrashed includes soft-deleted records in the query results
func (qb *QueryBuilder) WithTrashed() *QueryBuilder {
	qb.withTrashed = true
	qb.onlyTrashed = false
	return qb
}

// OnlyTrashed returns only soft-deleted records
func (qb *QueryBuilder) OnlyTrashed() *QueryBuilder {
	qb.onlyTrashed = true
	qb.withTrashed = false
	return qb
}

// ForceDelete performs a physical delete, bypassing soft delete
func (qb *QueryBuilder) ForceDelete() (int64, error) {
	if qb.lastErr != nil {
		return 0, qb.lastErr
	}
	if qb.table == "" {
		return 0, fmt.Errorf("dbkit: table name is required for ForceDelete")
	}
	if len(qb.whereSql) == 0 {
		return 0, fmt.Errorf("dbkit: ForceDelete operation requires at least one Where condition for safety")
	}

	whereSql := strings.Join(qb.whereSql, " AND ")

	if qb.tx != nil {
		return qb.tx.ForceDelete(qb.table, whereSql, qb.whereArgs...)
	}
	return qb.db.ForceDelete(qb.table, whereSql, qb.whereArgs...)
}

// Restore restores soft-deleted records matching the criteria
func (qb *QueryBuilder) Restore() (int64, error) {
	if qb.lastErr != nil {
		return 0, qb.lastErr
	}
	if qb.table == "" {
		return 0, fmt.Errorf("dbkit: table name is required for Restore")
	}

	whereSql := ""
	if len(qb.whereSql) > 0 {
		whereSql = strings.Join(qb.whereSql, " AND ")
	}

	if qb.tx != nil {
		return qb.tx.Restore(qb.table, whereSql, qb.whereArgs...)
	}
	return qb.db.Restore(qb.table, whereSql, qb.whereArgs...)
}
