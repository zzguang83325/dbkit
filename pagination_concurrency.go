package dbkit

import (
	"fmt"
	"sync"
)

// ConcurrentPaginationManager 并发安全的分页管理器
// 提供线程安全的分页操作，确保在高并发环境下的正确性
type ConcurrentPaginationManager struct {
	// 用于保护分页操作的读写锁
	// 虽然大部分操作是只读的，但为了确保一致性，使用读写锁
	mu sync.RWMutex

	// 分页配置缓存，避免重复创建配置对象
	configCache sync.Map // map[string]*PaginationConfig

	// SQL解析器池，复用解析器实例以提高性能
	parserPool sync.Pool

	// 适配器工厂池，复用工厂实例
	factoryPool sync.Pool
}

// NewConcurrentPaginationManager 创建并发安全的分页管理器
func NewConcurrentPaginationManager() *ConcurrentPaginationManager {
	mgr := &ConcurrentPaginationManager{}

	// 初始化SQL解析器池
	mgr.parserPool.New = func() interface{} {
		return NewSQLParser()
	}

	// 初始化适配器工厂池
	mgr.factoryPool.New = func() interface{} {
		return NewAdapterFactory()
	}

	return mgr
}

// GetPaginationConfig 线程安全地获取分页配置
// 使用缓存避免重复创建配置对象
func (mgr *ConcurrentPaginationManager) GetPaginationConfig(key string) *PaginationConfig {
	if config, ok := mgr.configCache.Load(key); ok {
		return config.(*PaginationConfig)
	}

	// 创建新配置并缓存
	newConfig := DefaultPaginationConfig()
	mgr.configCache.Store(key, newConfig)
	return newConfig
}

// GetSQLParser 从池中获取SQL解析器
// 使用对象池提高性能并确保线程安全
func (mgr *ConcurrentPaginationManager) GetSQLParser() SQLParser {
	return mgr.parserPool.Get().(SQLParser)
}

// PutSQLParser 将SQL解析器放回池中
func (mgr *ConcurrentPaginationManager) PutSQLParser(parser SQLParser) {
	mgr.parserPool.Put(parser)
}

// GetAdapterFactory 从池中获取适配器工厂
func (mgr *ConcurrentPaginationManager) GetAdapterFactory() *AdapterFactory {
	return mgr.factoryPool.Get().(*AdapterFactory)
}

// PutAdapterFactory 将适配器工厂放回池中
func (mgr *ConcurrentPaginationManager) PutAdapterFactory(factory *AdapterFactory) {
	mgr.factoryPool.Put(factory)
}

// SafePaginate 线程安全的分页查询
// 这是对现有Paginate函数的并发安全增强版本
func (mgr *ConcurrentPaginationManager) SafePaginate(
	db *DB,
	page int,
	pageSize int,
	querySQL string,
	args ...interface{},
) (*Page[Record], error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}

	// 使用读锁保护分页操作
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// 验证和规范化分页参数
	config := mgr.GetPaginationConfig("default")
	page, pageSize, err := ValidatePaginationParams(page, pageSize, config)
	if err != nil {
		return nil, err
	}

	// 从池中获取SQL解析器
	parser := mgr.GetSQLParser()
	defer mgr.PutSQLParser(parser)

	// 验证SQL安全性
	if err := parser.ValidateSQL(querySQL); err != nil {
		return nil, err
	}

	// 解析SQL语句
	parsedSQL, err := parser.ParseSQL(querySQL)
	if err != nil {
		return nil, err
	}

	// 从池中获取适配器工厂
	factory := mgr.GetAdapterFactory()
	defer mgr.PutAdapterFactory(factory)

	// 获取对应的分页适配器
	adapter := factory.CreateAdapter(string(db.dbMgr.config.Driver))

	// 执行分页查询
	return mgr.executePaginationQuery(db, adapter, parsedSQL, page, pageSize, args...)
}

// SafePaginateTx 线程安全的事务分页查询
func (mgr *ConcurrentPaginationManager) SafePaginateTx(
	tx *Tx,
	page int,
	pageSize int,
	querySQL string,
	args ...interface{},
) (*Page[Record], error) {
	// 使用读锁保护分页操作
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// 验证和规范化分页参数
	config := mgr.GetPaginationConfig("default")
	page, pageSize, err := ValidatePaginationParams(page, pageSize, config)
	if err != nil {
		return nil, err
	}

	// 从池中获取SQL解析器
	parser := mgr.GetSQLParser()
	defer mgr.PutSQLParser(parser)

	// 验证SQL安全性
	if err := parser.ValidateSQL(querySQL); err != nil {
		return nil, err
	}

	// 解析SQL语句
	parsedSQL, err := parser.ParseSQL(querySQL)
	if err != nil {
		return nil, err
	}

	// 从池中获取适配器工厂
	factory := mgr.GetAdapterFactory()
	defer mgr.PutAdapterFactory(factory)

	// 获取对应的分页适配器
	adapter := factory.CreateAdapter(string(tx.dbMgr.config.Driver))

	// 执行事务分页查询
	return mgr.executePaginationQueryTx(tx, adapter, parsedSQL, page, pageSize, args...)
}

// executePaginationQuery 执行分页查询的核心逻辑
func (mgr *ConcurrentPaginationManager) executePaginationQuery(
	db *DB,
	adapter PaginationAdapter,
	parsedSQL *ParsedSQL,
	page, pageSize int,
	args ...interface{},
) (*Page[Record], error) {
	// 构建计数SQL
	countSQL := adapter.BuildCountSQL(parsedSQL)

	// 执行计数查询
	ctx, cancel := db.getContext()
	defer cancel()

	var totalRow int64
	if db.cacheRepositoryName != "" {
		// 使用线程安全的缓存键生成
		countKey := GenerateCountCacheKey(db.dbMgr.name, parsedSQL, args...)
		if val, ok := GetCache().CacheGet(db.cacheRepositoryName, countKey); ok {
			if convertCacheValue(val, &totalRow) {
				// 缓存命中，继续执行分页查询
			} else {
				// 缓存值转换失败，重新查询
				countRecord, err := db.QueryFirst(countSQL, args...)
				if err != nil {
					return nil, fmt.Errorf("count query failed: %w", err)
				}
				if countRecord == nil {
					return nil, fmt.Errorf("count query returned empty result")
				}
				// 获取COUNT(*)的结果，通常是第一个字段
				for _, value := range countRecord.ToMap() {
					if count, ok := value.(int64); ok {
						totalRow = count
						break
					}
					// 处理其他数字类型
					if count, ok := value.(int); ok {
						totalRow = int64(count)
						break
					}
					if count, ok := value.(int32); ok {
						totalRow = int64(count)
						break
					}
				}
				GetCache().CacheSet(db.cacheRepositoryName, countKey, totalRow, getEffectiveTTL(db.cacheRepositoryName, db.cacheTTL))
			}
		} else {
			// 缓存未命中，执行查询
			countRecord, err := db.QueryFirst(countSQL, args...)
			if err != nil {
				return nil, fmt.Errorf("count query failed: %w", err)
			}
			if countRecord == nil {
				return nil, fmt.Errorf("count query returned empty result")
			}
			// 获取COUNT(*)的结果，通常是第一个字段
			for _, value := range countRecord.ToMap() {
				if count, ok := value.(int64); ok {
					totalRow = count
					break
				}
				// 处理其他数字类型
				if count, ok := value.(int); ok {
					totalRow = int64(count)
					break
				}
				if count, ok := value.(int32); ok {
					totalRow = int64(count)
					break
				}
			}
			GetCache().CacheSet(db.cacheRepositoryName, countKey, totalRow, getEffectiveTTL(db.cacheRepositoryName, db.cacheTTL))
		}
	} else {
		// 不使用缓存
		countRecord, err := db.QueryFirst(countSQL, args...)
		if err != nil {
			return nil, fmt.Errorf("count query failed: %w", err)
		}
		if countRecord == nil {
			return nil, fmt.Errorf("count query returned empty result")
		}
		// 获取COUNT(*)的结果，通常是第一个字段
		for _, value := range countRecord.ToMap() {
			if count, ok := value.(int64); ok {
				totalRow = count
				break
			}
			// 处理其他数字类型
			if count, ok := value.(int); ok {
				totalRow = int64(count)
				break
			}
			if count, ok := value.(int32); ok {
				totalRow = int64(count)
				break
			}
		}
	}

	// 如果总数为0，直接返回空结果
	if totalRow == 0 {
		return NewPage([]Record{}, page, pageSize, totalRow), nil
	}

	// 构建分页SQL
	paginationSQL := adapter.BuildPaginationSQL(parsedSQL, page, pageSize)

	// 执行分页查询
	var list []Record
	if db.cacheRepositoryName != "" {
		// 使用线程安全的缓存键生成
		paginationKey := GeneratePaginationCacheKey(db.dbMgr.name, parsedSQL, page, pageSize, args...)
		if val, ok := GetCache().CacheGet(db.cacheRepositoryName, paginationKey); ok {
			if convertCacheValue(val, &list) {
				// 缓存命中，直接返回结果
				return NewPage(list, page, pageSize, totalRow), nil
			}
		}

		// 缓存未命中或转换失败，执行查询
		list, err := db.dbMgr.queryWithContext(ctx, db.dbMgr.getDB(), paginationSQL, args...)
		if err != nil {
			return nil, fmt.Errorf("pagination query failed: %w", err)
		}

		// 将结果存入缓存
		GetCache().CacheSet(db.cacheRepositoryName, paginationKey, list, getEffectiveTTL(db.cacheRepositoryName, db.cacheTTL))
		return NewPage(list, page, pageSize, totalRow), nil
	} else {
		// 不使用缓存
		list, err := db.dbMgr.queryWithContext(ctx, db.dbMgr.getDB(), paginationSQL, args...)
		if err != nil {
			return nil, fmt.Errorf("pagination query failed: %w", err)
		}
		return NewPage(list, page, pageSize, totalRow), nil
	}
}

// executePaginationQueryTx 执行事务分页查询的核心逻辑
func (mgr *ConcurrentPaginationManager) executePaginationQueryTx(
	tx *Tx,
	adapter PaginationAdapter,
	parsedSQL *ParsedSQL,
	page, pageSize int,
	args ...interface{},
) (*Page[Record], error) {
	// 构建计数SQL
	countSQL := adapter.BuildCountSQL(parsedSQL)

	// 执行计数查询（在事务上下文中）
	ctx, cancel := tx.getContext()
	defer cancel()

	var totalRow int64
	if tx.cacheRepositoryName != "" {
		// 使用线程安全的缓存键生成
		countKey := GenerateCountCacheKey(tx.dbMgr.name, parsedSQL, args...)
		if val, ok := GetCache().CacheGet(tx.cacheRepositoryName, countKey); ok {
			if convertCacheValue(val, &totalRow) {
				// 缓存命中，继续执行分页查询
			} else {
				// 缓存值转换失败，重新查询
				countRecord, err := tx.QueryFirst(countSQL, args...)
				if err != nil {
					return nil, fmt.Errorf("transaction count query failed: %w", err)
				}
				if countRecord == nil {
					return nil, fmt.Errorf("transaction count query returned empty result")
				}
				// 获取COUNT(*)的结果，通常是第一个字段
				for _, value := range countRecord.ToMap() {
					if count, ok := value.(int64); ok {
						totalRow = count
						break
					}
					// 处理其他数字类型
					if count, ok := value.(int); ok {
						totalRow = int64(count)
						break
					}
					if count, ok := value.(int32); ok {
						totalRow = int64(count)
						break
					}
				}
				GetCache().CacheSet(tx.cacheRepositoryName, countKey, totalRow, getEffectiveTTL(tx.cacheRepositoryName, tx.cacheTTL))
			}
		} else {
			// 缓存未命中，执行查询
			countRecord, err := tx.QueryFirst(countSQL, args...)
			if err != nil {
				return nil, fmt.Errorf("transaction count query failed: %w", err)
			}
			if countRecord == nil {
				return nil, fmt.Errorf("transaction count query returned empty result")
			}
			// 获取COUNT(*)的结果，通常是第一个字段
			for _, value := range countRecord.ToMap() {
				if count, ok := value.(int64); ok {
					totalRow = count
					break
				}
				// 处理其他数字类型
				if count, ok := value.(int); ok {
					totalRow = int64(count)
					break
				}
				if count, ok := value.(int32); ok {
					totalRow = int64(count)
					break
				}
			}
			GetCache().CacheSet(tx.cacheRepositoryName, countKey, totalRow, getEffectiveTTL(tx.cacheRepositoryName, tx.cacheTTL))
		}
	} else {
		// 不使用缓存
		countRecord, err := tx.QueryFirst(countSQL, args...)
		if err != nil {
			return nil, fmt.Errorf("transaction count query failed: %w", err)
		}
		if countRecord == nil {
			return nil, fmt.Errorf("transaction count query returned empty result")
		}
		// 获取COUNT(*)的结果，通常是第一个字段
		for _, value := range countRecord.ToMap() {
			if count, ok := value.(int64); ok {
				totalRow = count
				break
			}
			// 处理其他数字类型
			if count, ok := value.(int); ok {
				totalRow = int64(count)
				break
			}
			if count, ok := value.(int32); ok {
				totalRow = int64(count)
				break
			}
		}
	}

	// 如果总数为0，直接返回空结果
	if totalRow == 0 {
		return NewPage([]Record{}, page, pageSize, totalRow), nil
	}

	// 构建分页SQL
	paginationSQL := adapter.BuildPaginationSQL(parsedSQL, page, pageSize)

	// 执行分页查询（在事务上下文中）
	var list []Record
	if tx.cacheRepositoryName != "" {
		// 使用线程安全的缓存键生成
		paginationKey := GeneratePaginationCacheKey(tx.dbMgr.name, parsedSQL, page, pageSize, args...)
		if val, ok := GetCache().CacheGet(tx.cacheRepositoryName, paginationKey); ok {
			if convertCacheValue(val, &list) {
				// 缓存命中，直接返回结果
				return NewPage(list, page, pageSize, totalRow), nil
			}
		}

		// 缓存未命中或转换失败，执行查询
		list, err := tx.dbMgr.queryWithContext(ctx, tx.tx, paginationSQL, args...)
		if err != nil {
			return nil, fmt.Errorf("transaction pagination query failed: %w", err)
		}

		// 将结果存入缓存
		GetCache().CacheSet(tx.cacheRepositoryName, paginationKey, list, getEffectiveTTL(tx.cacheRepositoryName, tx.cacheTTL))
		return NewPage(list, page, pageSize, totalRow), nil
	} else {
		// 不使用缓存
		list, err := tx.dbMgr.queryWithContext(ctx, tx.tx, paginationSQL, args...)
		if err != nil {
			return nil, fmt.Errorf("transaction pagination query failed: %w", err)
		}
		return NewPage(list, page, pageSize, totalRow), nil
	}
}

// 全局并发安全的分页管理器实例
var globalPaginationManager = NewConcurrentPaginationManager()

// SafePaginate 全局线程安全的分页函数
// 这是对现有Paginate函数的并发安全增强版本
func SafePaginate(page int, pageSize int, querySQL string, args ...interface{}) (*Page[Record], error) {
	db, err := defaultDB()
	if err != nil {
		return nil, err
	}
	return globalPaginationManager.SafePaginate(db, page, pageSize, querySQL, args...)
}

// SafePaginate 为DB结构体添加线程安全的分页方法
func (db *DB) SafePaginate(page int, pageSize int, querySQL string, args ...interface{}) (*Page[Record], error) {
	return globalPaginationManager.SafePaginate(db, page, pageSize, querySQL, args...)
}

// SafePaginate 为Tx结构体添加线程安全的分页方法
func (tx *Tx) SafePaginate(page int, pageSize int, querySQL string, args ...interface{}) (*Page[Record], error) {
	return globalPaginationManager.SafePaginateTx(tx, page, pageSize, querySQL, args...)
}
