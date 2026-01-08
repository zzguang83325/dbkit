package dbkit

import (
	"fmt"
	"strings"
	"time"
)

// Error definitions for model operations
var (
	ErrNoPrimaryKey       = fmt.Errorf("dbkit: table has no primary key")
	ErrPrimaryKeyNotFound = fmt.Errorf("dbkit: primary key not found in record")
)

// ModelCache 用于在 Model 中存储缓存配置，可嵌入到生成的 Model 中
type ModelCache struct {
	CacheRepositoryName string
	CacheTTL            time.Duration
}

// SetCache 设置缓存名称和TTL
func (c *ModelCache) SetCache(cacheRepositoryName string, ttl ...time.Duration) {
	c.CacheRepositoryName = cacheRepositoryName
	if len(ttl) > 0 {
		c.CacheTTL = ttl[0]
	} else {
		c.CacheTTL = -1
	}
}

// GetCache 获取缓存配置，如果未设置则返回 nil
func (c *ModelCache) GetCache() *ModelCache {
	if c.CacheRepositoryName == "" {
		return nil
	}
	return c
}

// FindModel 查询多条记录并映射到 DbModel 切片
func FindModel[T IDbModel](model T, cache *ModelCache, whereSql, orderBySql string, whereArgs ...interface{}) ([]T, error) {
	var results []T
	db := Use(model.DatabaseName())
	if cache != nil && cache.CacheRepositoryName != "" {
		db = db.Cache(cache.CacheRepositoryName, cache.CacheTTL)
	}
	err := db.Table(model.TableName()).Where(whereSql, whereArgs...).OrderBy(orderBySql).FindToDbModel(&results)
	return results, err
}

// FindFirstModel 查询第一条记录并映射到 DbModel
func FindFirstModel[T IDbModel](model T, cache *ModelCache, whereSql string, whereArgs ...interface{}) (T, error) {
	db := Use(model.DatabaseName())
	if cache != nil && cache.CacheRepositoryName != "" {
		db = db.Cache(cache.CacheRepositoryName, cache.CacheTTL)
	}
	err := db.Table(model.TableName()).Where(whereSql, whereArgs...).FindFirstToDbModel(model)
	return model, err
}

// PaginateModel 分页查询并映射到 DbModel
func PaginateModel[T IDbModel](model T, cache *ModelCache, page, pageSize int, whereSql, orderBySql string, whereArgs ...interface{}) (*Page[T], error) {
	db := Use(model.DatabaseName())
	if cache != nil && cache.CacheRepositoryName != "" {
		db = db.Cache(cache.CacheRepositoryName, cache.CacheTTL)
	}
	recordsPage, err := db.Table(model.TableName()).Where(whereSql, whereArgs...).OrderBy(orderBySql).Paginate(page, pageSize)
	if err != nil {
		return nil, err
	}
	return RecordPageToDbModelPage[T](recordsPage)
}

func PaginateModel_FullSql[T IDbModel](model T, cache *ModelCache, page, pageSize int, querySQL string, whereArgs ...interface{}) (*Page[T], error) {
	db := Use(model.DatabaseName())
	if cache != nil && cache.CacheRepositoryName != "" {
		db = db.Cache(cache.CacheRepositoryName, cache.CacheTTL)
	}
	recordsPage, err := db.Paginate(page, pageSize, querySQL, whereArgs...)
	if err != nil {
		return nil, err
	}
	return RecordPageToDbModelPage[T](recordsPage)
}

// --- Soft Delete Model Helpers ---

// ForceDeleteModel performs a physical delete on a soft-delete enabled model
func ForceDeleteModel(model IDbModel) (int64, error) {
	record := ToRecord(model)
	db := Use(model.DatabaseName())

	// Get primary keys

	sdb, err := db.dbMgr.getDB()
	if err != nil {
		return 0, err
	}

	pks, err := db.dbMgr.getPrimaryKeys(sdb, model.TableName())
	if err != nil {
		return 0, err
	}
	if len(pks) == 0 {
		return 0, ErrNoPrimaryKey
	}

	// Build WHERE clause from primary keys
	var whereClauses []string
	var whereArgs []interface{}
	for _, pk := range pks {
		if !record.Has(pk) {
			return 0, ErrPrimaryKeyNotFound
		}
		whereClauses = append(whereClauses, pk+" = ?")
		whereArgs = append(whereArgs, record.Get(pk))
	}

	return db.ForceDelete(model.TableName(), strings.Join(whereClauses, " AND "), whereArgs...)
}

// RestoreModel restores a soft-deleted model
func RestoreModel(model IDbModel) (int64, error) {
	record := ToRecord(model)
	db := Use(model.DatabaseName())
	sdb, err := db.dbMgr.getDB()
	if err != nil {
		return 0, err
	}
	// Get primary keys
	pks, err := db.dbMgr.getPrimaryKeys(sdb, model.TableName())
	if err != nil {
		return 0, err
	}
	if len(pks) == 0 {
		return 0, ErrNoPrimaryKey
	}

	// Build WHERE clause from primary keys
	var whereClauses []string
	var whereArgs []interface{}
	for _, pk := range pks {
		if !record.Has(pk) {
			return 0, ErrPrimaryKeyNotFound
		}
		whereClauses = append(whereClauses, pk+" = ?")
		whereArgs = append(whereArgs, record.Get(pk))
	}

	return db.Restore(model.TableName(), strings.Join(whereClauses, " AND "), whereArgs...)
}

// FindModelWithTrashed queries records including soft-deleted ones
func FindModelWithTrashed[T IDbModel](model T, cache *ModelCache, whereSql, orderBySql string, whereArgs ...interface{}) ([]T, error) {
	var results []T
	db := Use(model.DatabaseName())
	if cache != nil && cache.CacheRepositoryName != "" {
		db = db.Cache(cache.CacheRepositoryName, cache.CacheTTL)
	}
	err := db.Table(model.TableName()).WithTrashed().Where(whereSql, whereArgs...).OrderBy(orderBySql).FindToDbModel(&results)
	return results, err
}

// FindModelOnlyTrashed queries only soft-deleted records
func FindModelOnlyTrashed[T IDbModel](model T, cache *ModelCache, whereSql, orderBySql string, whereArgs ...interface{}) ([]T, error) {
	var results []T
	db := Use(model.DatabaseName())
	if cache != nil && cache.CacheRepositoryName != "" {
		db = db.Cache(cache.CacheRepositoryName, cache.CacheTTL)
	}
	err := db.Table(model.TableName()).OnlyTrashed().Where(whereSql, whereArgs...).OrderBy(orderBySql).FindToDbModel(&results)
	return results, err
}
