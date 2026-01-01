package dbkit

import "time"

// ModelCache 用于在 Model 中存储缓存配置，可嵌入到生成的 Model 中
type ModelCache struct {
	CacheName string
	CacheTTL  time.Duration
}

// SetCache 设置缓存名称和TTL
func (c *ModelCache) SetCache(name string, ttl ...time.Duration) {
	c.CacheName = name
	if len(ttl) > 0 {
		c.CacheTTL = ttl[0]
	} else {
		c.CacheTTL = -1
	}
}

// GetCache 获取缓存配置，如果未设置则返回 nil
func (c *ModelCache) GetCache() *ModelCache {
	if c.CacheName == "" {
		return nil
	}
	return c
}

// FindModel 查询多条记录并映射到 DbModel 切片
func FindModel[T IDbModel](model T, cache *ModelCache, whereSql, orderBySql string, whereArgs ...interface{}) ([]T, error) {
	var results []T
	db := Use(model.DatabaseName())
	if cache != nil && cache.CacheName != "" {
		db = db.Cache(cache.CacheName, cache.CacheTTL)
	}
	err := db.Table(model.TableName()).Where(whereSql, whereArgs...).OrderBy(orderBySql).FindToDbModel(&results)
	return results, err
}

// FindFirstModel 查询第一条记录并映射到 DbModel
func FindFirstModel[T IDbModel](model T, cache *ModelCache, whereSql string, whereArgs ...interface{}) (T, error) {
	db := Use(model.DatabaseName())
	if cache != nil && cache.CacheName != "" {
		db = db.Cache(cache.CacheName, cache.CacheTTL)
	}
	err := db.Table(model.TableName()).Where(whereSql, whereArgs...).FindFirstToDbModel(model)
	return model, err
}

// PaginateModel 分页查询并映射到 DbModel
func PaginateModel[T IDbModel](model T, cache *ModelCache, page, pageSize int, whereSql, orderBySql string, whereArgs ...interface{}) (*Page[T], error) {
	db := Use(model.DatabaseName())
	if cache != nil && cache.CacheName != "" {
		db = db.Cache(cache.CacheName, cache.CacheTTL)
	}
	recordsPage, err := db.Table(model.TableName()).Where(whereSql, whereArgs...).OrderBy(orderBySql).Paginate(page, pageSize)
	if err != nil {
		return nil, err
	}
	return RecordPageToDbModelPage[T](recordsPage)
}
