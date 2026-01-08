package dbkit

import "errors"

// Pagination related error definitions
var (
	ErrInvalidSQL      = errors.New("dbkit: invalid SQL statement")
	ErrUnsupportedSQL  = errors.New("dbkit: unsupported SQL type")
	ErrInvalidPage     = errors.New("dbkit: invalid page number")
	ErrInvalidPageSize = errors.New("dbkit: invalid page size")
	ErrSQLInjection    = errors.New("dbkit: potential SQL injection detected")
)

// SQLParser SQL解析器接口
// 负责解析和验证SQL语句，提取各个部分用于分页处理
type SQLParser interface {
	// ParseSQL 解析SQL语句，提取各个部分
	ParseSQL(sql string) (*ParsedSQL, error)

	// ValidateSQL 验证SQL语句的安全性和格式
	ValidateSQL(sql string) error
}

// PaginationAdapter 分页适配器接口
// 负责根据不同数据库类型生成相应的分页SQL语句
type PaginationAdapter interface {
	// BuildPaginationSQL 构建分页SQL
	BuildPaginationSQL(parsedSQL *ParsedSQL, page, pageSize int) string

	// BuildCountSQL 构建计数SQL
	BuildCountSQL(parsedSQL *ParsedSQL) string

	// GetDatabaseType 获取数据库类型
	GetDatabaseType() string
}

// ParsedSQL 解析后的SQL结构
// 存储解析后的SQL各个部分，用于后续的分页处理
type ParsedSQL struct {
	OriginalSQL   string        // 原始SQL语句
	SelectClause  string        // SELECT部分
	FromClause    string        // FROM部分
	WhereClause   string        // WHERE部分
	GroupByClause string        // GROUP BY部分
	HavingClause  string        // HAVING部分
	OrderByClause string        // ORDER BY部分
	IsComplex     bool          // 是否复杂查询（包含子查询、JOIN等）
	HasSubquery   bool          // 是否包含子查询
	HasJoin       bool          // 是否包含JOIN
	Parameters    []interface{} // SQL参数
}

// PaginationConfig 分页配置
// 定义分页相关的默认配置和限制
type PaginationConfig struct {
	DefaultPageSize int    // 默认页面大小
	MaxPageSize     int    // 最大页面大小
	DefaultOrderBy  string // 默认排序
}

// DefaultPaginationConfig 返回默认的分页配置
func DefaultPaginationConfig() *PaginationConfig {
	return &PaginationConfig{
		DefaultPageSize: DefaultPageSize,
		MaxPageSize:     MaxPageSize,
		DefaultOrderBy:  "",
	}
}

// ValidatePaginationParams 验证分页参数
func ValidatePaginationParams(page, pageSize int, config *PaginationConfig) (int, int, error) {
	if config == nil {
		config = DefaultPaginationConfig()
	}

	// 验证页码
	if page < MinPageSize {
		page = DefaultPage
	}

	// 验证页面大小
	if pageSize < MinPageSize {
		pageSize = config.DefaultPageSize
	}
	if pageSize > config.MaxPageSize {
		return 0, 0, ErrInvalidPageSize
	}

	return page, pageSize, nil
}
