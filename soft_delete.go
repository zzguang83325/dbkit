package dbkit

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// --- 内部错误变量（不导出） ---

var (
	errSQLParsingFailed         = fmt.Errorf("dbkit: SQL parsing failed")
	errConditionInjectionFailed = fmt.Errorf("dbkit: condition injection failed")
	errUnsupportedSQLType       = fmt.Errorf("dbkit: unsupported SQL type for soft delete filtering")
	errInvalidSoftDeleteConfig  = fmt.Errorf("dbkit: invalid soft delete configuration")
)

// --- 性能优化：正则表达式缓存 ---

var (
	// 正则表达式缓存，避免重复编译
	regexCache = make(map[string]*regexp.Regexp)
	regexMu    sync.RWMutex
)

// getCompiledRegex 获取编译后的正则表达式，使用缓存提高性能
func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	regexMu.RLock()
	if regex, exists := regexCache[pattern]; exists {
		regexMu.RUnlock()
		return regex, nil
	}
	regexMu.RUnlock()

	// 编译新的正则表达式
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	// 缓存编译结果
	regexMu.Lock()
	regexCache[pattern] = regex
	regexMu.Unlock()

	return regex, nil
}

// SoftDeleteType represents the type of soft delete field
type SoftDeleteType int

const (
	// SoftDeleteTimestamp uses a timestamp field (e.g., deleted_at)
	// NULL means not deleted, non-NULL means deleted
	SoftDeleteTimestamp SoftDeleteType = iota
	// SoftDeleteBool uses a boolean field (e.g., is_deleted)
	// false means not deleted, true means deleted
	SoftDeleteBool
)

// SoftDeleteConfig holds the soft delete configuration for a table
type SoftDeleteConfig struct {
	Field string         // Field name, e.g., "deleted_at", "is_deleted"
	Type  SoftDeleteType // Field type: timestamp or boolean
}

// softDeleteRegistry stores soft delete configurations per database
type softDeleteRegistry struct {
	configs map[string]*SoftDeleteConfig // table -> config
	mu      sync.RWMutex
}

// newSoftDeleteRegistry creates a new soft delete registry
func newSoftDeleteRegistry() *softDeleteRegistry {
	return &softDeleteRegistry{
		configs: make(map[string]*SoftDeleteConfig),
	}
}

// set configures soft delete for a table
func (r *softDeleteRegistry) set(table string, config *SoftDeleteConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs[strings.ToLower(table)] = config
}

// get returns the soft delete config for a table
func (r *softDeleteRegistry) get(table string) *SoftDeleteConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.configs[strings.ToLower(table)]
}

// remove removes soft delete config for a table
func (r *softDeleteRegistry) remove(table string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.configs, strings.ToLower(table))
}

// has checks if a table has soft delete configured
func (r *softDeleteRegistry) has(table string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.configs[strings.ToLower(table)]
	return ok
}

// ISoftDeleteModel is an optional interface for models that support soft delete
type ISoftDeleteModel interface {
	IDbModel
	SoftDeleteField() string        // Returns the soft delete field name
	SoftDeleteType() SoftDeleteType // Returns the soft delete type
}

// --- Global Functions (for default database) ---

// ConfigSoftDelete configures soft delete for a table using default field name "deleted_at" and timestamp type
func ConfigSoftDelete(table string, field ...string) {
	fieldName := "deleted_at" // 默认字段名
	if len(field) > 0 && field[0] != "" {
		fieldName = field[0]
	}
	ConfigSoftDeleteWithType(table, fieldName, SoftDeleteTimestamp)
}

// ConfigSoftDeleteWithType configures soft delete for a table with specified type
func ConfigSoftDeleteWithType(table, field string, deleteType SoftDeleteType) {
	db, err := defaultDB()
	if err != nil {
		return
	}
	db.ConfigSoftDeleteWithType(table, field, deleteType)
}

// RemoveSoftDelete removes soft delete configuration for a table
func RemoveSoftDelete(table string) {
	db, err := defaultDB()
	if err != nil {
		return
	}
	db.RemoveSoftDelete(table)
}

// HasSoftDelete checks if a table has soft delete configured
func HasSoftDelete(table string) bool {
	db, err := defaultDB()
	if err != nil {
		return false
	}
	return db.HasSoftDelete(table)
}

// ForceDelete performs a physical delete, bypassing soft delete
func ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.ForceDelete(table, whereSql, whereArgs...)
}

// Restore restores soft-deleted records
func Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.Restore(table, whereSql, whereArgs...)
}

// --- DB Methods ---

// ConfigSoftDelete configures soft delete for a table using default field name "deleted_at" and timestamp type
func (db *DB) ConfigSoftDelete(table string, field ...string) *DB {
	fieldName := "deleted_at" // 默认字段名
	if len(field) > 0 && field[0] != "" {
		fieldName = field[0]
	}
	return db.ConfigSoftDeleteWithType(table, fieldName, SoftDeleteTimestamp)
}

// ConfigSoftDeleteWithType configures soft delete for a table with specified type
func (db *DB) ConfigSoftDeleteWithType(table, field string, deleteType SoftDeleteType) *DB {
	if db.lastErr != nil || db.dbMgr == nil {
		return db
	}

	// 检查软删除字段是否存在
	if field != "" && !db.dbMgr.checkTableColumn(table, field) {
		LogWarn(fmt.Sprintf("软删除配置警告: 表 '%s' 中不存在字段 '%s'", table, field), map[string]interface{}{
			"db":    db.dbMgr.name,
			"table": table,
			"field": field,
		})
	}

	db.dbMgr.setSoftDeleteConfig(table, &SoftDeleteConfig{
		Field: field,
		Type:  deleteType,
	})
	return db
}

// RemoveSoftDelete removes soft delete configuration for a table
func (db *DB) RemoveSoftDelete(table string) *DB {
	if db.lastErr != nil || db.dbMgr == nil {
		return db
	}
	db.dbMgr.removeSoftDeleteConfig(table)
	return db
}

// HasSoftDelete checks if a table has soft delete configured
func (db *DB) HasSoftDelete(table string) bool {
	if db.lastErr != nil || db.dbMgr == nil {
		return false
	}
	return db.dbMgr.hasSoftDelete(table)
}

// ForceDelete performs a physical delete, bypassing soft delete
func (db *DB) ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	sdb, err := db.dbMgr.getDB()
	if err != nil {
		return 0, err
	}
	return db.dbMgr.forceDelete(sdb, table, whereSql, whereArgs...)
}

// Restore restores soft-deleted records
func (db *DB) Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	sdb, err := db.dbMgr.getDB()
	if err != nil {
		return 0, err
	}
	return db.dbMgr.restore(sdb, table, whereSql, whereArgs...)
}

// --- Tx Methods ---

// ForceDelete performs a physical delete within a transaction
func (tx *Tx) ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.forceDelete(tx.tx, table, whereSql, whereArgs...)
}

// Restore restores soft-deleted records within a transaction
func (tx *Tx) Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.restore(tx.tx, table, whereSql, whereArgs...)
}

// --- dbManager Methods ---

// setSoftDeleteConfig sets soft delete config for a table
func (mgr *dbManager) setSoftDeleteConfig(table string, config *SoftDeleteConfig) {
	if mgr.softDeletes == nil {
		mgr.softDeletes = newSoftDeleteRegistry()
	}
	mgr.softDeletes.set(table, config)
}

// getSoftDeleteConfig gets soft delete config for a table
func (mgr *dbManager) getSoftDeleteConfig(table string) *SoftDeleteConfig {
	if mgr.softDeletes == nil {
		return nil
	}
	return mgr.softDeletes.get(table)
}

// removeSoftDeleteConfig removes soft delete config for a table
func (mgr *dbManager) removeSoftDeleteConfig(table string) {
	if mgr.softDeletes == nil {
		return
	}
	mgr.softDeletes.remove(table)
}

// hasSoftDelete checks if a table has soft delete configured
func (mgr *dbManager) hasSoftDelete(table string) bool {
	if mgr.softDeletes == nil {
		return false
	}
	return mgr.softDeletes.has(table)
}

// softDelete performs a soft delete (UPDATE instead of DELETE)
func (mgr *dbManager) softDelete(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (int64, error) {
	config := mgr.getSoftDeleteConfig(table)
	if config == nil {
		return 0, fmt.Errorf("soft delete not configured for table %s", table)
	}

	var setValue string
	var setArgs []interface{}

	switch config.Type {
	case SoftDeleteTimestamp:
		setValue = fmt.Sprintf("%s = ?", config.Field)
		setArgs = append(setArgs, time.Now())
	case SoftDeleteBool:
		setValue = fmt.Sprintf("%s = ?", config.Field)
		setArgs = append(setArgs, true)
	}

	// Build UPDATE query
	querySQL := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, setValue, where)
	allArgs := append(setArgs, whereArgs...)

	querySQL = mgr.convertPlaceholder(querySQL, mgr.config.Driver)
	allArgs = mgr.sanitizeArgs(querySQL, allArgs)

	start := time.Now()
	result, err := executor.Exec(querySQL, allArgs...)
	mgr.logTrace(start, querySQL, allArgs, err)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// forceDelete performs a physical delete, bypassing soft delete
func (mgr *dbManager) forceDelete(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if where == "" {
		return 0, fmt.Errorf("where condition is required for delete")
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

// restore restores soft-deleted records
func (mgr *dbManager) restore(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}

	config := mgr.getSoftDeleteConfig(table)
	if config == nil {
		return 0, fmt.Errorf("soft delete not configured for table %s", table)
	}

	var setValue string
	var setArgs []interface{}

	switch config.Type {
	case SoftDeleteTimestamp:
		setValue = fmt.Sprintf("%s = NULL", config.Field)
	case SoftDeleteBool:
		setValue = fmt.Sprintf("%s = ?", config.Field)
		setArgs = append(setArgs, false)
	}

	// Build UPDATE query
	querySQL := fmt.Sprintf("UPDATE %s SET %s", table, setValue)
	if where != "" {
		querySQL += " WHERE " + where
	}

	allArgs := append(setArgs, whereArgs...)
	querySQL = mgr.convertPlaceholder(querySQL, mgr.config.Driver)
	allArgs = mgr.sanitizeArgs(querySQL, allArgs)

	start := time.Now()
	result, err := executor.Exec(querySQL, allArgs...)
	mgr.logTrace(start, querySQL, allArgs, err)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// --- 内部数据结构（不导出） ---

// softDeleteDetectionResult 软删除条件检测结果（内部使用）
type softDeleteDetectionResult struct {
	tableName          string            // 表名
	hasCondition       bool              // 是否已有软删除条件
	existingCondition  string            // 现有条件（如果有）
	suggestedCondition string            // 建议的条件（如果需要注入）
	config             *SoftDeleteConfig // 软删除配置
}

// sqlAnalysisResult SQL分析结果（内部使用）
type sqlAnalysisResult struct {
	originalSQL      string                                // 原始SQL
	parsedSQL        *ParsedSQL                            // 解析后的SQL结构
	tables           []string                              // 涉及的表
	softDeleteTables []string                              // 配置了软删除的表
	detectionResults map[string]*softDeleteDetectionResult // 每个表的检测结果
	needsInjection   bool                                  // 是否需要注入条件
	modifiedSQL      string                                // 修改后的SQL（如果需要）
}

// --- 内部辅助方法（QueryWithOutTrashed 功能支持） ---

// extractTablesFromSQL 从 SQL 语句中提取所有涉及的表名
// 支持 FROM 子句、JOIN 查询、逗号分隔的多表查询
// 处理表别名、引号、schema.table 格式
func (mgr *dbManager) extractTablesFromSQL(sql string) []string {
	var tables []string

	// 1. 从 FROM 子句提取主表
	fromTables := mgr.extractTablesFromFromClause(sql)
	tables = append(tables, fromTables...)

	// 2. 从 JOIN 子句提取表
	joinTables := mgr.extractTablesFromJoinClause(sql)
	tables = append(tables, joinTables...)

	// 3. 去重并返回
	return mgr.uniqueTableNames(tables)
}

// extractTablesFromFromClause 从 FROM 子句中提取表名
func (mgr *dbManager) extractTablesFromFromClause(sql string) []string {
	var tables []string

	// 查找 FROM 关键字的位置
	upperSQL := strings.ToUpper(sql)
	fromIndex := strings.Index(upperSQL, "FROM")
	if fromIndex == -1 {
		return tables
	}

	// 查找 FROM 子句的结束位置
	endKeywords := []string{"WHERE", "GROUP BY", "HAVING", "ORDER BY", "LIMIT", "JOIN", "LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "OUTER JOIN", "FULL JOIN", "CROSS JOIN"}
	fromClause := mgr.extractClauseContent(sql, fromIndex+4, endKeywords)

	if fromClause == "" {
		return tables
	}

	// 处理逗号分隔的多表查询（传统的隐式 JOIN）
	parts := strings.Split(fromClause, ",")

	for _, part := range parts {
		tableName := mgr.extractSingleTableName(strings.TrimSpace(part))
		if tableName != "" {
			tables = append(tables, tableName)
		}
	}

	return tables
}

// extractTablesFromJoinClause 从 JOIN 子句中提取表名
func (mgr *dbManager) extractTablesFromJoinClause(sql string) []string {
	var tables []string

	// 匹配各种 JOIN 类型的正则表达式（使用缓存优化性能）
	joinPatterns := []string{
		`(?i)\bJOIN\s+([^\s]+)`,         // JOIN table
		`(?i)\bLEFT\s+JOIN\s+([^\s]+)`,  // LEFT JOIN table
		`(?i)\bRIGHT\s+JOIN\s+([^\s]+)`, // RIGHT JOIN table
		`(?i)\bINNER\s+JOIN\s+([^\s]+)`, // INNER JOIN table
		`(?i)\bOUTER\s+JOIN\s+([^\s]+)`, // OUTER JOIN table
		`(?i)\bFULL\s+JOIN\s+([^\s]+)`,  // FULL JOIN table
		`(?i)\bCROSS\s+JOIN\s+([^\s]+)`, // CROSS JOIN table
	}

	for _, pattern := range joinPatterns {
		regex, err := getCompiledRegex(pattern)
		if err != nil {
			// 如果正则表达式编译失败，跳过这个模式
			continue
		}

		matches := regex.FindAllStringSubmatch(sql, -1)

		for _, match := range matches {
			if len(match) > 1 {
				tableName := mgr.cleanTableName(match[1])
				if tableName != "" {
					tables = append(tables, tableName)
				}
			}
		}
	}

	return tables
}

// extractClauseContent 提取指定位置开始到结束关键字之间的内容
func (mgr *dbManager) extractClauseContent(sql string, startPos int, endKeywords []string) string {
	upperSQL := strings.ToUpper(sql)
	content := sql[startPos:]
	upperContent := upperSQL[startPos:]

	// 查找最近的结束关键字
	endPos := len(content)
	for _, keyword := range endKeywords {
		if pos := strings.Index(upperContent, keyword); pos != -1 && pos < endPos {
			endPos = pos
		}
	}

	return strings.TrimSpace(content[:endPos])
}

// extractSingleTableName 从单个表表达式中提取表名
func (mgr *dbManager) extractSingleTableName(tableExpr string) string {
	// 移除多余空格
	tableExpr = strings.TrimSpace(tableExpr)
	if tableExpr == "" {
		return ""
	}

	// 处理子查询情况 (SELECT ...) AS alias
	if strings.HasPrefix(tableExpr, "(") {
		return "" // 子查询不是真实表名
	}

	// 分割表名和别名 (table AS alias 或 table alias)
	parts := strings.Fields(tableExpr)
	if len(parts) == 0 {
		return ""
	}

	// 第一部分是表名
	tableName := parts[0]

	// 移除引号（如果有）
	tableName = mgr.cleanTableName(tableName)

	return tableName
}

// cleanTableName 清理表名，移除引号和其他装饰
func (mgr *dbManager) cleanTableName(tableName string) string {
	// 移除各种引号
	tableName = strings.Trim(tableName, "`\"'[]")

	// 处理 schema.table 格式，只返回表名部分
	if dotIndex := strings.LastIndex(tableName, "."); dotIndex != -1 {
		tableName = tableName[dotIndex+1:]
	}

	return tableName
}

// uniqueTableNames 去除重复的表名
func (mgr *dbManager) uniqueTableNames(tables []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, table := range tables {
		if table != "" && !seen[strings.ToLower(table)] {
			seen[strings.ToLower(table)] = true
			result = append(result, table)
		}
	}

	return result
}

// hasSoftDeleteConditionInSQL 检测 SQL 中是否已包含指定表的软删除条件
// 支持带表前缀和不带表前缀的条件检测
// 处理 SoftDeleteTimestamp 和 SoftDeleteBool 类型
func (mgr *dbManager) hasSoftDeleteConditionInSQL(sql, table string, config *SoftDeleteConfig) bool {
	if config == nil {
		return false
	}

	// 构建要检测的条件模式
	var patterns []string
	switch config.Type {
	case SoftDeleteTimestamp:
		patterns = []string{
			// 带表前缀的模式（推荐用于多表查询）
			fmt.Sprintf(`%s\.%s\s+IS\s+NULL`, regexp.QuoteMeta(table), regexp.QuoteMeta(config.Field)),
			// 不带表前缀的模式（单表查询）
			fmt.Sprintf(`\b%s\s+IS\s+NULL`, regexp.QuoteMeta(config.Field)),
			// 其他可能的 NULL 检查模式
			fmt.Sprintf(`%s\.%s\s*=\s*NULL`, regexp.QuoteMeta(table), regexp.QuoteMeta(config.Field)),
			fmt.Sprintf(`\b%s\s*=\s*NULL`, regexp.QuoteMeta(config.Field)),
			// 处理带引号的表名
			fmt.Sprintf("`%s`\\.%s\\s+IS\\s+NULL", regexp.QuoteMeta(table), regexp.QuoteMeta(config.Field)),
			fmt.Sprintf(`"%s"\.%s\s+IS\s+NULL`, regexp.QuoteMeta(table), regexp.QuoteMeta(config.Field)),
		}
	case SoftDeleteBool:
		patterns = []string{
			// 带表前缀的模式（推荐用于多表查询）
			fmt.Sprintf(`%s\.%s\s*=\s*(false|0)`, regexp.QuoteMeta(table), regexp.QuoteMeta(config.Field)),
			// 不带表前缀的模式（单表查询）
			fmt.Sprintf(`\b%s\s*=\s*(false|0)`, regexp.QuoteMeta(config.Field)),
			// 处理带引号的表名
			fmt.Sprintf("`%s`\\.%s\\s*=\\s*(false|0)", regexp.QuoteMeta(table), regexp.QuoteMeta(config.Field)),
			fmt.Sprintf(`"%s"\.%s\s*=\s*(false|0)`, regexp.QuoteMeta(table), regexp.QuoteMeta(config.Field)),
		}
	}

	// 使用不区分大小写的匹配（优化：使用缓存的正则表达式）
	for _, pattern := range patterns {
		// 注意：不能直接对整个模式使用 ToUpper，因为会破坏正则表达式的特殊字符
		// 我们需要使用不区分大小写的正则表达式标志
		caseInsensitivePattern := "(?i)" + pattern
		regex, err := getCompiledRegex(caseInsensitivePattern)
		if err != nil {
			// 如果正则表达式编译失败，跳过这个模式
			LogWarn("QueryWithOutTrashed: 正则表达式编译失败", map[string]interface{}{
				"pattern": caseInsensitivePattern,
				"error":   err.Error(),
			})
			continue
		}
		if regex.MatchString(sql) { // 使用原始SQL，因为正则表达式已经是不区分大小写的
			return true
		}
	}

	return false
}

// injectSoftDeleteConditions 向 SQL 中注入软删除过滤条件
// 智能处理 WHERE 子句的添加和追加
// 根据是否有 JOIN 决定是否使用表前缀
func (mgr *dbManager) injectSoftDeleteConditions(sql string, tableConfigs map[string]*SoftDeleteConfig) (string, error) {
	if len(tableConfigs) == 0 {
		return sql, nil
	}

	// 检测是否有 JOIN 查询
	hasJoin := mgr.hasJoinInSQL(sql)

	// 构建软删除条件
	var conditions []string
	for table, config := range tableConfigs {
		condition := mgr.buildConditionForTable(table, config, hasJoin)
		if condition != "" {
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return sql, nil
	}

	// 注入条件
	combinedCondition := strings.Join(conditions, " AND ")

	if mgr.hasWhereClause(sql) {
		// 有WHERE子句，使用AND连接
		return mgr.appendToWhereClause(sql, combinedCondition), nil
	} else {
		// 没有WHERE子句，添加WHERE
		return mgr.insertWhereClause(sql, combinedCondition), nil
	}
}

// hasJoinInSQL 检测 SQL 中是否包含 JOIN 查询
func (mgr *dbManager) hasJoinInSQL(sql string) bool {
	upperSQL := strings.ToUpper(sql)
	joinKeywords := []string{"JOIN", "LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "OUTER JOIN", "FULL JOIN", "CROSS JOIN"}

	for _, keyword := range joinKeywords {
		if strings.Contains(upperSQL, keyword) {
			return true
		}
	}

	return false
}

// hasWhereClause 检测 SQL 中是否已有 WHERE 子句
func (mgr *dbManager) hasWhereClause(sql string) bool {
	upperSQL := strings.ToUpper(sql)
	return strings.Contains(upperSQL, "WHERE")
}

// buildConditionForTable 为表构建软删除条件（考虑多表情况）
func (mgr *dbManager) buildConditionForTable(table string, config *SoftDeleteConfig, hasJoin bool) string {
	var condition string

	switch config.Type {
	case SoftDeleteTimestamp:
		// 简化处理：不使用表前缀，避免别名问题
		condition = fmt.Sprintf("%s IS NULL", config.Field)
	case SoftDeleteBool:
		// 简化处理：不使用表前缀，避免别名问题
		condition = fmt.Sprintf("%s = false", config.Field)
	default:
		return ""
	}

	return condition
}

// insertWhereClause 插入 WHERE 子句
func (mgr *dbManager) insertWhereClause(sql, condition string) string {
	// 查找插入位置（在 GROUP BY, HAVING, ORDER BY, LIMIT 之前）
	upperSQL := strings.ToUpper(sql)
	insertPos := len(sql)

	keywords := []string{"GROUP BY", "HAVING", "ORDER BY", "LIMIT"}
	for _, keyword := range keywords {
		if pos := strings.Index(upperSQL, keyword); pos != -1 && pos < insertPos {
			insertPos = pos
		}
	}

	// 插入 WHERE 子句，确保前后都有空格
	before := strings.TrimSpace(sql[:insertPos])
	after := sql[insertPos:]

	if after != "" {
		return before + " WHERE " + condition + " " + after
	} else {
		return before + " WHERE " + condition
	}
}

// appendToWhereClause 向现有 WHERE 子句追加条件
func (mgr *dbManager) appendToWhereClause(sql, newCondition string) string {
	// 查找现有 WHERE 子句的位置并追加条件
	upperSQL := strings.ToUpper(sql)
	wherePos := strings.Index(upperSQL, "WHERE")
	if wherePos == -1 {
		return sql // 不应该发生，因为已经检查过有 WHERE 子句
	}

	// 查找 WHERE 子句的结束位置
	whereStart := wherePos + 5 // "WHERE" 的长度
	whereEnd := mgr.findWhereClauseEnd(sql, whereStart)

	// 构建新的 WHERE 子句
	before := sql[:whereEnd]
	after := sql[whereEnd:]

	return before + " AND (" + newCondition + ")" + after
}

// findWhereClauseEnd 查找 WHERE 子句的结束位置
func (mgr *dbManager) findWhereClauseEnd(sql string, whereStart int) int {
	upperSQL := strings.ToUpper(sql)

	// 查找可能的结束关键字
	endKeywords := []string{"GROUP BY", "HAVING", "ORDER BY", "LIMIT"}
	endPos := len(sql)

	for _, keyword := range endKeywords {
		if pos := strings.Index(upperSQL[whereStart:], keyword); pos != -1 {
			candidatePos := whereStart + pos
			if candidatePos < endPos {
				endPos = candidatePos
			}
		}
	}

	return endPos
}

// analyzeSQLForSoftDelete 分析 SQL 语句以确定是否需要注入软删除条件
// 复用现有的 SQL 解析器，集成表名提取、条件检测和条件注入逻辑
func (mgr *dbManager) analyzeSQLForSoftDelete(sql string) (*sqlAnalysisResult, error) {
	result := &sqlAnalysisResult{
		originalSQL:      sql,
		detectionResults: make(map[string]*softDeleteDetectionResult),
	}

	// 输入验证
	if strings.TrimSpace(sql) == "" {
		LogWarn("QueryWithOutTrashed: 空 SQL 语句", map[string]interface{}{
			"db": mgr.name,
		})
		return result, nil
	}

	// 1. 提取表名
	tables := mgr.extractTablesFromSQL(sql)
	result.tables = tables

	if len(tables) == 0 {
		LogInfo("QueryWithOutTrashed: 未检测到表名", map[string]interface{}{
			"db":  mgr.name,
			"sql": sql,
		})
		return result, nil
	}

	// 2. 筛选配置了软删除的表
	softDeleteTables := mgr.filterSoftDeleteTables(tables)
	result.softDeleteTables = softDeleteTables

	// 3. 如果没有软删除表，直接返回
	if len(softDeleteTables) == 0 {
		LogDebug("QueryWithOutTrashed: 无软删除表配置", map[string]interface{}{
			"db":     mgr.name,
			"tables": tables,
		})
		return result, nil
	}

	// 4. 检测每个表的软删除条件
	needsInjection := false
	for _, table := range softDeleteTables {
		config := mgr.getSoftDeleteConfig(table)
		if config == nil {
			LogWarn("QueryWithOutTrashed: 软删除配置丢失", map[string]interface{}{
				"db":    mgr.name,
				"table": table,
			})
			continue
		}

		hasCondition := mgr.hasSoftDeleteConditionInSQL(sql, table, config)

		detectionResult := &softDeleteDetectionResult{
			tableName:    table,
			hasCondition: hasCondition,
			config:       config,
		}

		if !hasCondition {
			needsInjection = true
			// 构建建议的条件
			hasJoin := mgr.hasJoinInSQL(sql)
			detectionResult.suggestedCondition = mgr.buildConditionForTable(table, config, hasJoin)

			LogDebug("QueryWithOutTrashed: 需要注入软删除条件", map[string]interface{}{
				"db":        mgr.name,
				"table":     table,
				"condition": detectionResult.suggestedCondition,
			})
		}

		result.detectionResults[table] = detectionResult
	}

	result.needsInjection = needsInjection

	// 5. 如果需要注入条件，生成修改后的SQL
	if needsInjection {
		tableConfigs := make(map[string]*SoftDeleteConfig)
		for _, table := range softDeleteTables {
			if !result.detectionResults[table].hasCondition {
				tableConfigs[table] = mgr.getSoftDeleteConfig(table)
			}
		}

		modifiedSQL, err := mgr.injectSoftDeleteConditions(sql, tableConfigs)
		if err != nil {
			LogError("QueryWithOutTrashed: 条件注入失败", map[string]interface{}{
				"db":    mgr.name,
				"sql":   sql,
				"error": err.Error(),
			})
			return nil, fmt.Errorf("%w: %v", errConditionInjectionFailed, err)
		}
		result.modifiedSQL = modifiedSQL

		LogDebug("QueryWithOutTrashed: SQL 修改成功", map[string]interface{}{
			"db":          mgr.name,
			"originalSQL": sql,
			"modifiedSQL": modifiedSQL,
		})
	}

	return result, nil
}

// filterSoftDeleteTables 筛选配置了软删除的表
func (mgr *dbManager) filterSoftDeleteTables(tables []string) []string {
	var result []string
	for _, table := range tables {
		if mgr.hasSoftDelete(table) {
			result = append(result, table)
		}
	}
	return result
}

// buildSoftDeleteCondition builds the WHERE condition for filtering soft-deleted records
func (mgr *dbManager) buildSoftDeleteCondition(table string, includeDeleted, onlyDeleted bool) string {
	// Check if soft delete check is enabled
	if !mgr.enableSoftDeleteCheck {
		return ""
	}

	config := mgr.getSoftDeleteConfig(table)
	if config == nil {
		return ""
	}

	// If including all records (withTrashed), no filter needed
	if includeDeleted && !onlyDeleted {
		return ""
	}

	switch config.Type {
	case SoftDeleteTimestamp:
		if onlyDeleted {
			return fmt.Sprintf("%s IS NOT NULL", config.Field)
		}
		return fmt.Sprintf("%s IS NULL", config.Field)
	case SoftDeleteBool:
		if onlyDeleted {
			return fmt.Sprintf("%s = true", config.Field)
		}
		return fmt.Sprintf("%s = false", config.Field)
	}
	return ""
}
