package dbkit

import (
	"fmt"
	"strings"
)

// MySQLAdapter MySQL分页适配器
// 实现MySQL数据库的分页SQL生成逻辑
type MySQLAdapter struct{}

// NewMySQLAdapter 创建MySQL分页适配器实例
func NewMySQLAdapter() PaginationAdapter {
	return &MySQLAdapter{}
}

// BuildPaginationSQL 构建MySQL分页SQL
// MySQL使用 LIMIT offset, count 语法
func (m *MySQLAdapter) BuildPaginationSQL(parsedSQL *ParsedSQL, page, pageSize int) string {
	offset := (page - 1) * pageSize

	if parsedSQL.IsComplex {
		// 复杂查询使用子查询包装
		// 这样可以确保分页逻辑不会干扰原始查询的语义
		return fmt.Sprintf("SELECT * FROM (%s) AS subquery LIMIT %d, %d",
			parsedSQL.OriginalSQL, offset, pageSize)
	}

	// 简单查询直接添加LIMIT子句
	// 需要确保原SQL没有已存在的LIMIT子句
	cleanSQL := m.removeLimitClause(parsedSQL.OriginalSQL)
	return fmt.Sprintf("%s LIMIT %d, %d", cleanSQL, offset, pageSize)
}

// BuildCountSQL 构建MySQL计数SQL
// 用于获取总记录数，支持复杂查询的正确计数
func (m *MySQLAdapter) BuildCountSQL(parsedSQL *ParsedSQL) string {
	if parsedSQL.IsComplex || parsedSQL.GroupByClause != "" {
		// 对于复杂查询或包含GROUP BY的查询，使用子查询包装
		cleanSQL := m.removeLimitClause(parsedSQL.OriginalSQL)
		return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_subquery", cleanSQL)
	}

	// 简单查询直接替换SELECT子句
	countSQL := m.buildSimpleCountSQL(parsedSQL)
	return countSQL
}

// GetDatabaseType 获取数据库类型
func (m *MySQLAdapter) GetDatabaseType() string {
	return "mysql"
}

// removeLimitClause 移除SQL中已存在的LIMIT子句
func (m *MySQLAdapter) removeLimitClause(sql string) string {
	upperSQL := strings.ToUpper(sql)
	limitIndex := strings.LastIndex(upperSQL, "LIMIT")

	if limitIndex == -1 {
		return sql
	}

	// 确保LIMIT是一个完整的单词
	if limitIndex > 0 && isAlphaNumeric(upperSQL[limitIndex-1]) {
		return sql
	}
	if limitIndex+5 < len(upperSQL) && isAlphaNumeric(upperSQL[limitIndex+5]) {
		return sql
	}

	return strings.TrimSpace(sql[:limitIndex])
}

// buildSimpleCountSQL 为简单查询构建计数SQL
func (m *MySQLAdapter) buildSimpleCountSQL(parsedSQL *ParsedSQL) string {
	var parts []string

	// 构建COUNT查询
	parts = append(parts, "SELECT COUNT(*)")

	// 添加FROM子句
	if parsedSQL.FromClause != "" {
		parts = append(parts, "FROM", parsedSQL.FromClause)
	}

	// 添加WHERE子句
	if parsedSQL.WhereClause != "" {
		parts = append(parts, "WHERE", parsedSQL.WhereClause)
	}

	// 注意：COUNT查询不需要GROUP BY、HAVING、ORDER BY子句
	// 因为我们只需要总数

	return strings.Join(parts, " ")
}

// PostgreSQLAdapter PostgreSQL分页适配器
// 实现PostgreSQL数据库的分页SQL生成逻辑
type PostgreSQLAdapter struct{}

// NewPostgreSQLAdapter 创建PostgreSQL分页适配器实例
func NewPostgreSQLAdapter() PaginationAdapter {
	return &PostgreSQLAdapter{}
}

// BuildPaginationSQL 构建PostgreSQL分页SQL
// PostgreSQL使用 LIMIT count OFFSET offset 语法
func (p *PostgreSQLAdapter) BuildPaginationSQL(parsedSQL *ParsedSQL, page, pageSize int) string {
	offset := (page - 1) * pageSize

	if parsedSQL.IsComplex {
		// 复杂查询使用子查询包装
		return fmt.Sprintf("SELECT * FROM (%s) AS subquery LIMIT %d OFFSET %d",
			parsedSQL.OriginalSQL, pageSize, offset)
	}

	// 简单查询直接添加LIMIT和OFFSET子句
	cleanSQL := p.removeLimitClause(parsedSQL.OriginalSQL)
	return fmt.Sprintf("%s LIMIT %d OFFSET %d", cleanSQL, pageSize, offset)
}

// BuildCountSQL 构建PostgreSQL计数SQL
// 用于获取总记录数，支持复杂查询的正确计数
func (p *PostgreSQLAdapter) BuildCountSQL(parsedSQL *ParsedSQL) string {
	if parsedSQL.IsComplex || parsedSQL.GroupByClause != "" {
		// 对于复杂查询或包含GROUP BY的查询，使用子查询包装
		cleanSQL := p.removeLimitClause(parsedSQL.OriginalSQL)
		return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_subquery", cleanSQL)
	}

	// 简单查询直接替换SELECT子句
	countSQL := p.buildSimpleCountSQL(parsedSQL)
	return countSQL
}

// GetDatabaseType 获取数据库类型
func (p *PostgreSQLAdapter) GetDatabaseType() string {
	return "postgresql"
}

// removeLimitClause 移除SQL中已存在的LIMIT子句
func (p *PostgreSQLAdapter) removeLimitClause(sql string) string {
	upperSQL := strings.ToUpper(sql)

	// 查找LIMIT关键字
	limitIndex := strings.LastIndex(upperSQL, "LIMIT")
	if limitIndex == -1 {
		return sql
	}

	// 确保LIMIT是一个完整的单词
	if limitIndex > 0 && isAlphaNumeric(upperSQL[limitIndex-1]) {
		return sql
	}
	if limitIndex+5 < len(upperSQL) && isAlphaNumeric(upperSQL[limitIndex+5]) {
		return sql
	}

	// 查找OFFSET关键字（可能在LIMIT之后）
	offsetIndex := strings.Index(upperSQL[limitIndex:], "OFFSET")
	if offsetIndex != -1 {
		// 如果有OFFSET，移除整个LIMIT...OFFSET...部分
		return strings.TrimSpace(sql[:limitIndex])
	}

	// 只有LIMIT，移除LIMIT部分
	return strings.TrimSpace(sql[:limitIndex])
}

// buildSimpleCountSQL 为简单查询构建计数SQL
func (p *PostgreSQLAdapter) buildSimpleCountSQL(parsedSQL *ParsedSQL) string {
	var parts []string

	// 构建COUNT查询
	parts = append(parts, "SELECT COUNT(*)")

	// 添加FROM子句
	if parsedSQL.FromClause != "" {
		parts = append(parts, "FROM", parsedSQL.FromClause)
	}

	// 添加WHERE子句
	if parsedSQL.WhereClause != "" {
		parts = append(parts, "WHERE", parsedSQL.WhereClause)
	}

	// 注意：COUNT查询不需要GROUP BY、HAVING、ORDER BY子句
	// 因为我们只需要总数

	return strings.Join(parts, " ")
}

// SQLServerAdapter SQL Server分页适配器
// 实现SQL Server数据库的分页SQL生成逻辑
type SQLServerAdapter struct{}

// NewSQLServerAdapter 创建SQL Server分页适配器实例
func NewSQLServerAdapter() PaginationAdapter {
	return &SQLServerAdapter{}
}

// BuildPaginationSQL 构建SQL Server分页SQL
// SQL Server使用 OFFSET offset ROWS FETCH NEXT pageSize ROWS ONLY 语法
func (s *SQLServerAdapter) BuildPaginationSQL(parsedSQL *ParsedSQL, page, pageSize int) string {
	offset := (page - 1) * pageSize

	if parsedSQL.IsComplex {
		// 复杂查询使用子查询包装
		// SQL Server的OFFSET需要ORDER BY子句，如果子查询没有，需要添加默认排序
		subquery := parsedSQL.OriginalSQL
		if parsedSQL.OrderByClause == "" {
			subquery = fmt.Sprintf("%s ORDER BY (SELECT NULL)", parsedSQL.OriginalSQL)
		}
		return fmt.Sprintf("SELECT * FROM (%s) AS subquery ORDER BY (SELECT NULL) OFFSET %d ROWS FETCH NEXT %d ROWS ONLY",
			subquery, offset, pageSize)
	}

	// 简单查询直接添加OFFSET和FETCH子句
	cleanSQL := s.removeOffsetClause(parsedSQL.OriginalSQL)

	// SQL Server的OFFSET需要ORDER BY子句
	if parsedSQL.OrderByClause == "" {
		cleanSQL = fmt.Sprintf("%s ORDER BY (SELECT NULL)", cleanSQL)
	}

	return fmt.Sprintf("%s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", cleanSQL, offset, pageSize)
}

// BuildCountSQL 构建SQL Server计数SQL
// 用于获取总记录数，支持复杂查询的正确计数
func (s *SQLServerAdapter) BuildCountSQL(parsedSQL *ParsedSQL) string {
	if parsedSQL.IsComplex || parsedSQL.GroupByClause != "" {
		// 对于复杂查询或包含GROUP BY的查询，使用子查询包装
		cleanSQL := s.removeOffsetClause(parsedSQL.OriginalSQL)
		return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_subquery", cleanSQL)
	}

	// 简单查询直接替换SELECT子句
	countSQL := s.buildSimpleCountSQL(parsedSQL)
	return countSQL
}

// GetDatabaseType 获取数据库类型
func (s *SQLServerAdapter) GetDatabaseType() string {
	return "sqlserver"
}

// removeOffsetClause 移除SQL中已存在的OFFSET子句
func (s *SQLServerAdapter) removeOffsetClause(sql string) string {
	upperSQL := strings.ToUpper(sql)

	// 查找OFFSET关键字
	offsetIndex := strings.LastIndex(upperSQL, "OFFSET")
	if offsetIndex == -1 {
		return sql
	}

	// 确保OFFSET是一个完整的单词
	if offsetIndex > 0 && isAlphaNumeric(upperSQL[offsetIndex-1]) {
		return sql
	}
	if offsetIndex+6 < len(upperSQL) && isAlphaNumeric(upperSQL[offsetIndex+6]) {
		return sql
	}

	// 移除OFFSET...FETCH...部分
	return strings.TrimSpace(sql[:offsetIndex])
}

// buildSimpleCountSQL 为简单查询构建计数SQL
func (s *SQLServerAdapter) buildSimpleCountSQL(parsedSQL *ParsedSQL) string {
	var parts []string

	// 构建COUNT查询
	parts = append(parts, "SELECT COUNT(*)")

	// 添加FROM子句
	if parsedSQL.FromClause != "" {
		parts = append(parts, "FROM", parsedSQL.FromClause)
	}

	// 添加WHERE子句
	if parsedSQL.WhereClause != "" {
		parts = append(parts, "WHERE", parsedSQL.WhereClause)
	}

	// 注意：COUNT查询不需要GROUP BY、HAVING、ORDER BY子句
	// 因为我们只需要总数

	return strings.Join(parts, " ")
}

// OracleAdapter Oracle分页适配器
// 实现Oracle数据库的分页SQL生成逻辑
type OracleAdapter struct{}

// NewOracleAdapter 创建Oracle分页适配器实例
func NewOracleAdapter() PaginationAdapter {
	return &OracleAdapter{}
}

// BuildPaginationSQL 构建Oracle分页SQL
// Oracle使用 ROW_NUMBER() 窗口函数进行分页
func (o *OracleAdapter) BuildPaginationSQL(parsedSQL *ParsedSQL, page, pageSize int) string {
	offset := (page - 1) * pageSize
	endRow := offset + pageSize

	// Oracle分页需要使用ROW_NUMBER()窗口函数
	// 对于所有查询都使用子查询包装的方式
	orderBy := parsedSQL.OrderByClause
	if orderBy == "" {
		// 如果没有ORDER BY，使用默认排序
		orderBy = "ROWID"
	}

	// 构建带ROW_NUMBER()的内层查询
	innerQuery := fmt.Sprintf(`
		SELECT inner_query.*, ROW_NUMBER() OVER (ORDER BY %s) AS rn 
		FROM (%s) inner_query`,
		orderBy, parsedSQL.OriginalSQL)

	// 构建外层分页查询
	return fmt.Sprintf(`
		SELECT * FROM (%s) 
		WHERE rn > %d AND rn <= %d`,
		innerQuery, offset, endRow)
}

// BuildCountSQL 构建Oracle计数SQL
// 用于获取总记录数，支持复杂查询的正确计数
func (o *OracleAdapter) BuildCountSQL(parsedSQL *ParsedSQL) string {
	if parsedSQL.IsComplex || parsedSQL.GroupByClause != "" {
		// 对于复杂查询或包含GROUP BY的查询，使用子查询包装
		cleanSQL := o.removeRowNumClause(parsedSQL.OriginalSQL)
		return fmt.Sprintf("SELECT COUNT(*) FROM (%s)", cleanSQL)
	}

	// 简单查询直接替换SELECT子句
	countSQL := o.buildSimpleCountSQL(parsedSQL)
	return countSQL
}

// GetDatabaseType 获取数据库类型
func (o *OracleAdapter) GetDatabaseType() string {
	return "oracle"
}

// removeRowNumClause 移除SQL中已存在的ROWNUM相关子句
func (o *OracleAdapter) removeRowNumClause(sql string) string {
	upperSQL := strings.ToUpper(sql)

	// 查找WHERE子句中的ROWNUM条件
	// 这是一个简化的实现，实际情况可能更复杂
	if strings.Contains(upperSQL, "ROWNUM") {
		// 如果包含ROWNUM，可能需要更复杂的解析
		// 这里简化处理，直接返回原SQL
		return sql
	}

	return sql
}

// buildSimpleCountSQL 为简单查询构建计数SQL
func (o *OracleAdapter) buildSimpleCountSQL(parsedSQL *ParsedSQL) string {
	var parts []string

	// 构建COUNT查询
	parts = append(parts, "SELECT COUNT(*)")

	// 添加FROM子句
	if parsedSQL.FromClause != "" {
		parts = append(parts, "FROM", parsedSQL.FromClause)
	}

	// 添加WHERE子句
	if parsedSQL.WhereClause != "" {
		parts = append(parts, "WHERE", parsedSQL.WhereClause)
	}

	// 注意：COUNT查询不需要GROUP BY、HAVING、ORDER BY子句
	// 因为我们只需要总数

	return strings.Join(parts, " ")
}

// SQLiteAdapter SQLite分页适配器
// 实现SQLite数据库的分页SQL生成逻辑
type SQLiteAdapter struct{}

// NewSQLiteAdapter 创建SQLite分页适配器实例
func NewSQLiteAdapter() PaginationAdapter {
	return &SQLiteAdapter{}
}

// BuildPaginationSQL 构建SQLite分页SQL
// SQLite使用 LIMIT count OFFSET offset 语法（与PostgreSQL相同）
func (s *SQLiteAdapter) BuildPaginationSQL(parsedSQL *ParsedSQL, page, pageSize int) string {
	offset := (page - 1) * pageSize

	if parsedSQL.IsComplex {
		// 复杂查询使用子查询包装
		return fmt.Sprintf("SELECT * FROM (%s) AS subquery LIMIT %d OFFSET %d",
			parsedSQL.OriginalSQL, pageSize, offset)
	}

	// 简单查询直接添加LIMIT和OFFSET子句
	cleanSQL := s.removeLimitClause(parsedSQL.OriginalSQL)
	return fmt.Sprintf("%s LIMIT %d OFFSET %d", cleanSQL, pageSize, offset)
}

// BuildCountSQL 构建SQLite计数SQL
// 用于获取总记录数，支持复杂查询的正确计数
func (s *SQLiteAdapter) BuildCountSQL(parsedSQL *ParsedSQL) string {
	if parsedSQL.IsComplex || parsedSQL.GroupByClause != "" {
		// 对于复杂查询或包含GROUP BY的查询，使用子查询包装
		cleanSQL := s.removeLimitClause(parsedSQL.OriginalSQL)
		return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_subquery", cleanSQL)
	}

	// 简单查询直接替换SELECT子句
	countSQL := s.buildSimpleCountSQL(parsedSQL)
	return countSQL
}

// GetDatabaseType 获取数据库类型
func (s *SQLiteAdapter) GetDatabaseType() string {
	return "sqlite"
}

// removeLimitClause 移除SQL中已存在的LIMIT子句
func (s *SQLiteAdapter) removeLimitClause(sql string) string {
	upperSQL := strings.ToUpper(sql)

	// 查找LIMIT关键字
	limitIndex := strings.LastIndex(upperSQL, "LIMIT")
	if limitIndex == -1 {
		return sql
	}

	// 确保LIMIT是一个完整的单词
	if limitIndex > 0 && isAlphaNumeric(upperSQL[limitIndex-1]) {
		return sql
	}
	if limitIndex+5 < len(upperSQL) && isAlphaNumeric(upperSQL[limitIndex+5]) {
		return sql
	}

	// 查找OFFSET关键字（可能在LIMIT之后）
	offsetIndex := strings.Index(upperSQL[limitIndex:], "OFFSET")
	if offsetIndex != -1 {
		// 如果有OFFSET，移除整个LIMIT...OFFSET...部分
		return strings.TrimSpace(sql[:limitIndex])
	}

	// 只有LIMIT，移除LIMIT部分
	return strings.TrimSpace(sql[:limitIndex])
}

// buildSimpleCountSQL 为简单查询构建计数SQL
func (s *SQLiteAdapter) buildSimpleCountSQL(parsedSQL *ParsedSQL) string {
	var parts []string

	// 构建COUNT查询
	parts = append(parts, "SELECT COUNT(*)")

	// 添加FROM子句
	if parsedSQL.FromClause != "" {
		parts = append(parts, "FROM", parsedSQL.FromClause)
	}

	// 添加WHERE子句
	if parsedSQL.WhereClause != "" {
		parts = append(parts, "WHERE", parsedSQL.WhereClause)
	}

	// 注意：COUNT查询不需要GROUP BY、HAVING、ORDER BY子句
	// 因为我们只需要总数

	return strings.Join(parts, " ")
}

// AdapterFactory 适配器工厂
// 根据数据库类型创建相应的分页适配器
type AdapterFactory struct{}

// NewAdapterFactory 创建适配器工厂实例
func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{}
}

// CreateAdapter 根据数据库类型创建分页适配器
func (f *AdapterFactory) CreateAdapter(dbType string) PaginationAdapter {
	switch strings.ToLower(dbType) {
	case "mysql":
		return NewMySQLAdapter()
	case "postgresql", "postgres":
		return NewPostgreSQLAdapter()
	case "sqlserver", "mssql":
		return NewSQLServerAdapter()
	case "oracle":
		return NewOracleAdapter()
	case "sqlite", "sqlite3":
		return NewSQLiteAdapter()
	default:
		// 默认使用MySQL适配器
		return NewMySQLAdapter()
	}
}
