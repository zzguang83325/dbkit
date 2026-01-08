package dbkit

// 批量操作相关常量
const (
	// DefaultBatchSize 默认批量操作大小
	// 用于批量插入、更新、删除操作
	DefaultBatchSize = 100
)

// 缓存相关常量
const (
	// StmtCacheRepository 预编译语句缓存的内部仓库名称
	StmtCacheRepository = "__dbkit_stmt_cache__"

	// StmtCacheTTL 预编译语句缓存时间
	// 设置为 0 表示永不过期，只在数据库关闭或语句失效时清理
	StmtCacheTTL = 0
)

// 分页相关常量
const (
	// DefaultPage 默认页码
	DefaultPage = 1

	// DefaultPageSize 默认每页大小
	DefaultPageSize = 10

	// MinPageSize 最小每页大小
	MinPageSize = 1

	// MaxPageSize 最大每页大小（防止一次查询过多数据）
	MaxPageSize = 10000
)
