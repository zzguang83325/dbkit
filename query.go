package dbkit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// --- Global Functions (Operation on default database) ---

func Query(querySQL string, args ...interface{}) ([]Record, error) {
	db, err := defaultDB()
	if err != nil {
		return nil, err
	}
	return db.Query(querySQL, args...)
}

func QueryFirst(querySQL string, args ...interface{}) (*Record, error) {
	db, err := defaultDB()
	if err != nil {
		return nil, err
	}
	return db.QueryFirst(querySQL, args...)
}

func QueryToDbModel(dest interface{}, querySQL string, args ...interface{}) error {
	records, err := Query(querySQL, args...)
	if err != nil {
		return err
	}
	return ToStructs(records, dest)
}

func QueryFirstToDbModel(dest interface{}, querySQL string, args ...interface{}) error {
	record, err := QueryFirst(querySQL, args...)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("dbkit: no record found")
	}
	return ToStruct(record, dest)
}

func QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error) {
	db, err := defaultDB()
	if err != nil {
		return nil, err
	}
	return db.QueryMap(querySQL, args...)
}

func Exec(querySQL string, args ...interface{}) (sql.Result, error) {
	db, err := defaultDB()
	if err != nil {
		return nil, err
	}
	return db.Exec(querySQL, args...)
}

func Save(table string, record *Record) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.Save(table, record)
}

func Insert(table string, record *Record) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.Insert(table, record)
}

func Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	// 如果特性检查都关闭，直接使用快速路径
	if !db.dbMgr.enableTimestampCheck && !db.dbMgr.enableOptimisticLockCheck {
		return db.dbMgr.updateFast(db.dbMgr.getDB(), table, record, whereSql, whereArgs...)
	}
	return db.dbMgr.update(db.dbMgr.getDB(), table, record, whereSql, whereArgs...)
}

// UpdateFast is a lightweight update that always skips timestamp and optimistic lock checks.
// Use this when you need maximum performance and don't need these features.
func UpdateFast(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.dbMgr.updateFast(db.dbMgr.getDB(), table, record, whereSql, whereArgs...)
}

func Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.Delete(table, whereSql, whereArgs...)
}

func DeleteRecord(table string, record *Record) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.DeleteRecord(table, record)
}

func BatchInsert(table string, records []*Record, batchSize int) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.BatchInsert(table, records, batchSize)
}

func BatchInsertDefault(table string, records []*Record) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.BatchInsertDefault(table, records)
}

// BatchUpdate updates multiple records by primary key
func BatchUpdate(table string, records []*Record, batchSize int) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.BatchUpdate(table, records, batchSize)
}

// BatchUpdateDefault updates multiple records with default batch size (100)
func BatchUpdateDefault(table string, records []*Record) (int64, error) {
	return BatchUpdate(table, records, 100)
}

// BatchDelete deletes multiple records by primary key
func BatchDelete(table string, records []*Record, batchSize int) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.BatchDelete(table, records, batchSize)
}

// BatchDeleteDefault deletes multiple records with default batch size (100)
func BatchDeleteDefault(table string, records []*Record) (int64, error) {
	return BatchDelete(table, records, 100)
}

// BatchDeleteByIds deletes records by primary key IDs
func BatchDeleteByIds(table string, ids []interface{}, batchSize int) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.BatchDeleteByIds(table, ids, batchSize)
}

// BatchDeleteByIdsDefault deletes records by IDs with default batch size (100)
func BatchDeleteByIdsDefault(table string, ids []interface{}) (int64, error) {
	return BatchDeleteByIds(table, ids, 100)
}

func Count(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.Count(table, whereSql, whereArgs...)
}

func Exists(table string, whereSql string, whereArgs ...interface{}) (bool, error) {
	db, err := defaultDB()
	if err != nil {
		return false, err
	}
	return db.Exists(table, whereSql, whereArgs...)

}

func PaginateBuilder(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) (*Page[Record], error) {
	db, err := defaultDB()
	if err != nil {
		return nil, err
	}
	return db.PaginateBuilder(page, pageSize, selectSql, table, whereSql, orderBySql, args...)
}

// Paginate 全局分页函数，使用完整SQL语句进行分页查询
// 自动解析SQL并根据数据库类型生成相应的分页语句
func Paginate(page int, pageSize int, querySQL string, args ...interface{}) (*Page[Record], error) {
	db, err := defaultDB()
	if err != nil {
		return nil, err
	}
	return db.Paginate(page, pageSize, querySQL, args...)
}

func Transaction(fn func(*Tx) error) error {
	db, err := defaultDB()
	if err != nil {
		return err
	}
	return db.Transaction(fn)
}

func Ping() error {
	dbMgr, err := safeGetCurrentDB()
	if err != nil {
		return err
	}
	return dbMgr.Ping()
}

// Timeout returns a DB instance with the specified query timeout
func Timeout(d time.Duration) *DB {
	db, err := defaultDB()
	if err != nil {
		return &DB{lastErr: err}
	}
	db.timeout = d
	return db
}

// PingDB pings a specific database by name
func PingDB(dbname string) error {
	dbMgr := GetDatabase(dbname)
	if dbMgr == nil {
		return fmt.Errorf("dbkit: database '%s' not found", dbname)
	}
	return dbMgr.Ping()
}

func BeginTransaction() (*Tx, error) {
	dbMgr, err := safeGetCurrentDB()
	if err != nil {
		return nil, err
	}
	tx, err := dbMgr.getDB().Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, dbMgr: dbMgr}, nil
}

func ExecTx(tx *Tx, querySQL string, args ...interface{}) (sql.Result, error) {
	return tx.dbMgr.exec(tx.tx, querySQL, args...)
}

func SaveTx(tx *Tx, table string, record *Record) (int64, error) {
	return tx.Save(table, record)
}

func UpdateTx(tx *Tx, table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error) {
	return tx.Update(table, record, whereSql, whereArgs...)
}

func WithTransaction(fn func(*Tx) error) error {
	return Transaction(fn)
}

func FindAll(table string) ([]Record, error) {
	db, err := defaultDB()
	if err != nil {
		return nil, err
	}
	return db.FindAll(table)
}

// --- Struct Methods (Operation on models implementing IDbModel) ---

func SaveDbModel(model IDbModel) (int64, error) {
	return Use(model.DatabaseName()).SaveDbModel(model)
}

func InsertDbModel(model IDbModel) (int64, error) {
	return Use(model.DatabaseName()).InsertDbModel(model)
}

func UpdateDbModel(model IDbModel) (int64, error) {
	return Use(model.DatabaseName()).UpdateDbModel(model)
}

func DeleteDbModel(model IDbModel) (int64, error) {
	return Use(model.DatabaseName()).DeleteDbModel(model)
}

func FindFirstToDbModel(model IDbModel, whereSql string, whereArgs ...interface{}) error {
	return Use(model.DatabaseName()).FindFirstToDbModel(model, whereSql, whereArgs...)
}

func FindToDbModel(dest interface{}, table string, whereSql string, orderBySql string, whereArgs ...interface{}) error {
	db, err := defaultDB()
	if err != nil {
		return err
	}
	return db.FindToDbModel(dest, table, whereSql, orderBySql, whereArgs...)
}

// --- DB Methods (Operation on specific database instance) ---

// Cache 使用默认缓存（可通过 SetDefaultCache 切换默认缓存）
func (db *DB) Cache(cacheRepositoryName string, ttl ...time.Duration) *DB {
	db.cacheRepositoryName = cacheRepositoryName
	db.cacheProvider = nil // 使用默认缓存
	if len(ttl) > 0 {
		db.cacheTTL = ttl[0]
	} else {
		db.cacheTTL = -1
	}
	return db
}

// LocalCache 使用本地缓存
func (db *DB) LocalCache(cacheRepositoryName string, ttl ...time.Duration) *DB {
	db.cacheRepositoryName = cacheRepositoryName
	db.cacheProvider = GetLocalCacheInstance()
	if len(ttl) > 0 {
		db.cacheTTL = ttl[0]
	} else {
		db.cacheTTL = -1
	}
	return db
}

// RedisCache 使用 Redis 缓存
func (db *DB) RedisCache(cacheRepositoryName string, ttl ...time.Duration) *DB {
	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		// 如果 Redis 缓存未初始化，记录错误但不中断链式调用
		LogError("Redis cache not initialized for DB", map[string]interface{}{
			"cacheRepositoryName": cacheRepositoryName,
		})
		return db
	}

	db.cacheRepositoryName = cacheRepositoryName
	db.cacheProvider = redisCache
	if len(ttl) > 0 {
		db.cacheTTL = ttl[0]
	} else {
		db.cacheTTL = -1
	}
	return db
}

// Timeout sets the query timeout for this DB instance
func (db *DB) Timeout(d time.Duration) *DB {
	db.timeout = d
	return db
}

func (db *DB) Query(querySQL string, args ...interface{}) ([]Record, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	ctx, cancel := db.getContext()
	defer cancel()

	if db.cacheRepositoryName != "" {
		cache := db.getEffectiveCache()
		key := GenerateCacheKey(db.dbMgr.name, querySQL, args...)
		if val, ok := cache.CacheGet(db.cacheRepositoryName, key); ok {
			var results []Record
			if convertCacheValue(val, &results) {
				return results, nil
			}
		}
		results, err := db.dbMgr.queryWithContext(ctx, db.dbMgr.getDB(), querySQL, args...)
		if err == nil {
			cache.CacheSet(db.cacheRepositoryName, key, results, getEffectiveTTL(db.cacheRepositoryName, db.cacheTTL))
		}
		return results, err
	}
	return db.dbMgr.queryWithContext(ctx, db.dbMgr.getDB(), querySQL, args...)
}

func (db *DB) QueryFirst(querySQL string, args ...interface{}) (*Record, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	ctx, cancel := db.getContext()
	defer cancel()

	if db.cacheRepositoryName != "" {
		cache := db.getEffectiveCache()
		key := GenerateCacheKey(db.dbMgr.name, querySQL, args...)
		if val, ok := cache.CacheGet(db.cacheRepositoryName, key); ok {
			var result *Record
			if convertCacheValue(val, &result) {
				return result, nil
			}
		}
		result, err := db.dbMgr.queryFirstWithContext(ctx, db.dbMgr.getDB(), querySQL, args...)
		if err == nil && result != nil {
			cache.CacheSet(db.cacheRepositoryName, key, result, getEffectiveTTL(db.cacheRepositoryName, db.cacheTTL))
		}
		return result, err
	}
	return db.dbMgr.queryFirstWithContext(ctx, db.dbMgr.getDB(), querySQL, args...)
}

func (db *DB) QueryToDbModel(dest interface{}, querySQL string, args ...interface{}) error {
	records, err := db.Query(querySQL, args...)
	if err != nil {
		return err
	}
	return ToStructs(records, dest)
}

func (db *DB) QueryFirstToDbModel(dest interface{}, querySQL string, args ...interface{}) error {
	record, err := db.QueryFirst(querySQL, args...)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("dbkit: no record found")
	}
	return ToStruct(record, dest)
}

func (db *DB) QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	ctx, cancel := db.getContext()
	defer cancel()

	if db.cacheRepositoryName != "" {
		cache := db.getEffectiveCache()
		key := GenerateCacheKey(db.dbMgr.name, querySQL, args...)
		if val, ok := cache.CacheGet(db.cacheRepositoryName, key); ok {
			var results []map[string]interface{}
			if convertCacheValue(val, &results) {
				return results, nil
			}
		}
		results, err := db.dbMgr.queryMapWithContext(ctx, db.dbMgr.getDB(), querySQL, args...)
		if err == nil {
			cache.CacheSet(db.cacheRepositoryName, key, results, getEffectiveTTL(db.cacheRepositoryName, db.cacheTTL))
		}
		return results, err
	}
	return db.dbMgr.queryMapWithContext(ctx, db.dbMgr.getDB(), querySQL, args...)
}

func (db *DB) Exec(querySQL string, args ...interface{}) (sql.Result, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	ctx, cancel := db.getContext()
	defer cancel()
	return db.dbMgr.execWithContext(ctx, db.dbMgr.getDB(), querySQL, args...)
}

func (db *DB) Save(table string, record *Record) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.save(db.dbMgr.getDB(), table, record)
}

func (db *DB) Insert(table string, record *Record) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.insert(db.dbMgr.getDB(), table, record)
}

func (db *DB) insertWithOptions(table string, record *Record, skipTimestamps bool) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.insertWithOptions(db.dbMgr.getDB(), table, record, skipTimestamps)
}

func (db *DB) Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	// If both feature checks are disabled, use fast path directly
	if !db.dbMgr.enableTimestampCheck && !db.dbMgr.enableOptimisticLockCheck {
		return db.dbMgr.updateFast(db.dbMgr.getDB(), table, record, whereSql, whereArgs...)
	}
	return db.dbMgr.update(db.dbMgr.getDB(), table, record, whereSql, whereArgs...)
}

// UpdateFast is a lightweight update that always skips timestamp and optimistic lock checks.
func (db *DB) UpdateFast(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.updateFast(db.dbMgr.getDB(), table, record, whereSql, whereArgs...)
}

func (db *DB) updateWithOptions(table string, record *Record, whereSql string, skipTimestamps bool, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.updateWithOptions(db.dbMgr.getDB(), table, record, whereSql, skipTimestamps, whereArgs...)
}

func (db *DB) UpdateRecord(table string, record *Record) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.updateRecord(db.dbMgr.getDB(), table, record)
}

func (db *DB) Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.delete(db.dbMgr.getDB(), table, whereSql, whereArgs...)
}

func (db *DB) DeleteRecord(table string, record *Record) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.deleteRecord(db.dbMgr.getDB(), table, record)
}

func (db *DB) BatchInsert(table string, records []*Record, batchSize int) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.batchInsert(db.dbMgr.getDB(), table, records, batchSize)
}

func (db *DB) BatchInsertDefault(table string, records []*Record) (int64, error) {
	return db.BatchInsert(table, records, 100)
}

// BatchUpdate updates multiple records by primary key
func (db *DB) BatchUpdate(table string, records []*Record, batchSize int) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.batchUpdate(db.dbMgr.getDB(), table, records, batchSize)
}

// BatchUpdateDefault updates multiple records with default batch size (100)
func (db *DB) BatchUpdateDefault(table string, records []*Record) (int64, error) {
	return db.BatchUpdate(table, records, 100)
}

// BatchDelete deletes multiple records by primary key
func (db *DB) BatchDelete(table string, records []*Record, batchSize int) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.batchDelete(db.dbMgr.getDB(), table, records, batchSize)
}

// BatchDeleteDefault deletes multiple records with default batch size (100)
func (db *DB) BatchDeleteDefault(table string, records []*Record) (int64, error) {
	return db.BatchDelete(table, records, 100)
}

// BatchDeleteByIds deletes records by primary key IDs
func (db *DB) BatchDeleteByIds(table string, ids []interface{}, batchSize int) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.batchDeleteByIds(db.dbMgr.getDB(), table, ids, batchSize)
}

// BatchDeleteByIdsDefault deletes records by IDs with default batch size (100)
func (db *DB) BatchDeleteByIdsDefault(table string, ids []interface{}) (int64, error) {
	return db.BatchDeleteByIds(table, ids, 100)
}

func (db *DB) Count(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	if db.cacheRepositoryName != "" {
		cache := db.getEffectiveCache()
		key := GenerateCacheKey(db.dbMgr.name, "COUNT:"+table+":"+whereSql, whereArgs...)
		if val, ok := cache.CacheGet(db.cacheRepositoryName, key); ok {
			var count int64
			if convertCacheValue(val, &count) {
				return count, nil
			}
		}
		count, err := db.dbMgr.count(db.dbMgr.getDB(), table, whereSql, whereArgs...)
		if err == nil {
			cache.CacheSet(db.cacheRepositoryName, key, count, getEffectiveTTL(db.cacheRepositoryName, db.cacheTTL))
		}
		return count, err
	}
	return db.dbMgr.count(db.dbMgr.getDB(), table, whereSql, whereArgs...)
}

func (db *DB) Ping() error {
	if db.lastErr != nil {
		return db.lastErr
	}
	return db.dbMgr.Ping()
}

func (db *DB) Exists(table string, whereSql string, whereArgs ...interface{}) (bool, error) {
	if db.lastErr != nil {
		return false, db.lastErr
	}
	return db.dbMgr.exists(db.dbMgr.getDB(), table, whereSql, whereArgs...)
}

func (db *DB) PaginateBuilder(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) (*Page[Record], error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	if table != "" {
		if err := ValidateTableName(table); err != nil {
			return nil, err
		}
	}
	querySQL := selectSql
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(selectSql)), "SELECT ") {
		querySQL = "SELECT " + selectSql
	}

	if !strings.Contains(strings.ToUpper(querySQL), " FROM") && table != "" {
		querySQL += " FROM " + table
	}
	if whereSql != "" {
		querySQL += " WHERE " + whereSql
	}
	if orderBySql != "" {
		querySQL += " ORDER BY " + orderBySql
	}

	if db.cacheRepositoryName != "" {
		cache := db.getEffectiveCache()
		key := GenerateCacheKey(db.dbMgr.name, "PAGINATE:"+querySQL, args...)
		if val, ok := cache.CacheGet(db.cacheRepositoryName, key); ok {
			var pageObj *Page[Record]
			if convertCacheValue(val, &pageObj) {
				return pageObj, nil
			}
		}
		list, totalRow, err := db.dbMgr.paginate(db.dbMgr.getDB(), querySQL, page, pageSize, args...)
		if err == nil {
			pageObj := NewPage(list, page, pageSize, totalRow)
			cache.CacheSet(db.cacheRepositoryName, key, pageObj, getEffectiveTTL(db.cacheRepositoryName, db.cacheTTL))
			return pageObj, nil
		}
		return nil, err
	}

	list, totalRow, err := db.dbMgr.paginate(db.dbMgr.getDB(), querySQL, page, pageSize, args...)
	if err != nil {
		return nil, err
	}
	return NewPage(list, page, pageSize, totalRow), nil
}

// Paginate DB实例分页方法，使用完整SQL语句进行分页查询
// 自动解析SQL并根据数据库类型生成相应的分页语句，支持缓存集成
func (db *DB) Paginate(page int, pageSize int, querySQL string, args ...interface{}) (*Page[Record], error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}

	if db.cacheRepositoryName != "" {
		cache := db.getEffectiveCache()
		key := GenerateCacheKey(db.dbMgr.name, "PAGINATE_SQL:"+querySQL, args...)
		if val, ok := cache.CacheGet(db.cacheRepositoryName, key); ok {
			var pageObj *Page[Record]
			if convertCacheValue(val, &pageObj) {
				return pageObj, nil
			}
		}
		list, totalRow, err := db.dbMgr.paginate(db.dbMgr.getDB(), querySQL, page, pageSize, args...)
		if err == nil {
			pageObj := NewPage(list, page, pageSize, totalRow)
			cache.CacheSet(db.cacheRepositoryName, key, pageObj, getEffectiveTTL(db.cacheRepositoryName, db.cacheTTL))
			return pageObj, nil
		}
		return nil, err
	}

	list, totalRow, err := db.dbMgr.paginate(db.dbMgr.getDB(), querySQL, page, pageSize, args...)
	if err != nil {
		return nil, err
	}
	return NewPage(list, page, pageSize, totalRow), nil
}

func (db *DB) FindAll(table string) ([]Record, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	if err := ValidateTableName(table); err != nil {
		return nil, err
	}
	return db.Query(fmt.Sprintf("SELECT * FROM %s", table))
}

// Struct methods for DB
func (db *DB) SaveDbModel(model IDbModel) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	record := ToRecord(model)
	// For Save, we also want to handle auto-increment PKs if they are 0
	pks, _ := db.dbMgr.getPrimaryKeys(db.dbMgr.getDB(), model.TableName())
	for _, pk := range pks {
		if val, ok := record.Get(pk).(int64); ok && val == 0 {
			record.Remove(pk)
		}
	}
	return db.Save(model.TableName(), record)
}

func (db *DB) InsertDbModel(model IDbModel) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	record := ToRecord(model)
	// Remove primary key if it's 0 to let DB auto-increment
	pks, _ := db.dbMgr.getPrimaryKeys(db.dbMgr.getDB(), model.TableName())
	for _, pk := range pks {
		if val, ok := record.Get(pk).(int64); ok && val == 0 {
			record.Remove(pk)
		}
	}
	return db.Insert(model.TableName(), record)
}

func (db *DB) UpdateDbModel(model IDbModel) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	record := ToRecord(model)
	return db.UpdateRecord(model.TableName(), record)
}

func (db *DB) DeleteDbModel(model IDbModel) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	record := ToRecord(model)
	return db.DeleteRecord(model.TableName(), record)
}

func (db *DB) FindFirstToDbModel(model IDbModel, whereSql string, whereArgs ...interface{}) error {
	if db.lastErr != nil {
		return db.lastErr
	}
	builder := db.Table(model.TableName())
	if whereSql != "" {
		builder.Where(whereSql, whereArgs...)
	}
	return builder.FindFirstToDbModel(model)
}

func (db *DB) FindToDbModel(dest interface{}, table string, whereSql string, orderBySql string, whereArgs ...interface{}) error {
	if db.lastErr != nil {
		return db.lastErr
	}
	builder := db.Table(table)
	if whereSql != "" {
		builder.Where(whereSql, whereArgs...)
	}
	if orderBySql != "" {
		builder.OrderBy(orderBySql)
	}
	return builder.FindToDbModel(dest)
}

// Transaction executes a function within a transaction
func (db *DB) Transaction(fn func(*Tx) error) (err error) {
	if db.lastErr != nil {
		return db.lastErr
	}
	tx, err := db.dbMgr.getDB().Begin()
	if err != nil {
		return err
	}

	dbtx := &Tx{tx: tx, dbMgr: db.dbMgr}

	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				LogError("transaction rollback failed on panic", map[string]interface{}{
					"rollback_error": rbErr.Error(),
				})
			}
			err = fmt.Errorf("transaction panic: %v", p)
		}
	}()

	if err = fn(dbtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			LogError("transaction rollback failed", map[string]interface{}{
				"original_error": err.Error(),
				"rollback_error": rbErr.Error(),
			})
		}
		return err
	}

	return tx.Commit()
}

// --- Tx Methods (Operation within a transaction) ---

// Cache 使用默认缓存创建事务查询（可通过 SetDefaultCache 切换默认缓存）
func (tx *Tx) Cache(name string, ttl ...time.Duration) *Tx {
	tx.cacheRepositoryName = name
	tx.cacheProvider = nil // 使用默认缓存
	if len(ttl) > 0 {
		tx.cacheTTL = ttl[0]
	} else {
		tx.cacheTTL = -1
	}
	return tx
}

// LocalCache 创建一个使用本地缓存的事务查询
func (tx *Tx) LocalCache(cacheRepositoryName string, ttl ...time.Duration) *Tx {
	tx.cacheRepositoryName = cacheRepositoryName
	tx.cacheProvider = GetLocalCacheInstance()
	if len(ttl) > 0 {
		tx.cacheTTL = ttl[0]
	} else {
		tx.cacheTTL = -1
	}
	return tx
}

// RedisCache 创建一个使用 Redis 缓存的事务查询
func (tx *Tx) RedisCache(cacheRepositoryName string, ttl ...time.Duration) *Tx {
	redisCache := GetRedisCacheInstance()
	if redisCache == nil {
		// 如果 Redis 缓存未初始化，记录错误但不中断链式调用
		LogError("Redis cache not initialized for transaction", map[string]interface{}{
			"cacheRepositoryName": cacheRepositoryName,
		})
		return tx
	}

	tx.cacheRepositoryName = cacheRepositoryName
	tx.cacheProvider = redisCache
	if len(ttl) > 0 {
		tx.cacheTTL = ttl[0]
	} else {
		tx.cacheTTL = -1
	}
	return tx
}

// Timeout sets the query timeout for this transaction
func (tx *Tx) Timeout(d time.Duration) *Tx {
	tx.timeout = d
	return tx
}

// getTimeout returns the effective timeout for this Tx instance
func (tx *Tx) getTimeout() time.Duration {
	if tx.timeout > 0 {
		return tx.timeout
	}
	if tx.dbMgr != nil && tx.dbMgr.config != nil && tx.dbMgr.config.QueryTimeout > 0 {
		return tx.dbMgr.config.QueryTimeout
	}
	return 0
}

// getContext returns a context with timeout if configured
func (tx *Tx) getContext() (context.Context, context.CancelFunc) {
	timeout := tx.getTimeout()
	if timeout > 0 {
		return context.WithTimeout(context.Background(), timeout)
	}
	return context.Background(), func() {}
}

func (tx *Tx) Query(querySQL string, args ...interface{}) ([]Record, error) {
	ctx, cancel := tx.getContext()
	defer cancel()

	if tx.cacheRepositoryName != "" {
		cache := tx.getEffectiveCache()
		key := GenerateCacheKey(tx.dbMgr.name, querySQL, args...)
		if val, ok := cache.CacheGet(tx.cacheRepositoryName, key); ok {
			var results []Record
			if convertCacheValue(val, &results) {
				return results, nil
			}
		}
		results, err := tx.dbMgr.queryWithContext(ctx, tx.tx, querySQL, args...)
		if err == nil {
			cache.CacheSet(tx.cacheRepositoryName, key, results, getEffectiveTTL(tx.cacheRepositoryName, tx.cacheTTL))
		}
		return results, err
	}
	return tx.dbMgr.queryWithContext(ctx, tx.tx, querySQL, args...)
}

func (tx *Tx) QueryFirst(querySQL string, args ...interface{}) (*Record, error) {
	ctx, cancel := tx.getContext()
	defer cancel()

	if tx.cacheRepositoryName != "" {
		cache := tx.getEffectiveCache()
		key := GenerateCacheKey(tx.dbMgr.name, querySQL, args...)
		if val, ok := cache.CacheGet(tx.cacheRepositoryName, key); ok {
			var result *Record
			if convertCacheValue(val, &result) {
				return result, nil
			}
		}
		result, err := tx.dbMgr.queryFirstWithContext(ctx, tx.tx, querySQL, args...)
		if err == nil && result != nil {
			cache.CacheSet(tx.cacheRepositoryName, key, result, getEffectiveTTL(tx.cacheRepositoryName, tx.cacheTTL))
		}
		return result, err
	}
	return tx.dbMgr.queryFirstWithContext(ctx, tx.tx, querySQL, args...)
}

func (tx *Tx) QueryToDbModel(dest interface{}, querySQL string, args ...interface{}) error {
	records, err := tx.Query(querySQL, args...)
	if err != nil {
		return err
	}
	return ToStructs(records, dest)
}

func (tx *Tx) QueryFirstToDbModel(dest interface{}, querySQL string, args ...interface{}) error {
	record, err := tx.QueryFirst(querySQL, args...)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("dbkit: no record found")
	}
	return ToStruct(record, dest)
}

func (tx *Tx) QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error) {
	ctx, cancel := tx.getContext()
	defer cancel()

	if tx.cacheRepositoryName != "" {
		cache := tx.getEffectiveCache()
		key := GenerateCacheKey(tx.dbMgr.name, querySQL, args...)
		if val, ok := cache.CacheGet(tx.cacheRepositoryName, key); ok {
			var results []map[string]interface{}
			if convertCacheValue(val, &results) {
				return results, nil
			}
		}
		results, err := tx.dbMgr.queryMapWithContext(ctx, tx.tx, querySQL, args...)
		if err == nil {
			cache.CacheSet(tx.cacheRepositoryName, key, results, getEffectiveTTL(tx.cacheRepositoryName, tx.cacheTTL))
		}
		return results, err
	}
	return tx.dbMgr.queryMapWithContext(ctx, tx.tx, querySQL, args...)
}

func (tx *Tx) Exec(querySQL string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := tx.getContext()
	defer cancel()
	return tx.dbMgr.execWithContext(ctx, tx.tx, querySQL, args...)
}

func (tx *Tx) Save(table string, record *Record) (int64, error) {
	return tx.dbMgr.save(tx.tx, table, record)
}

func (tx *Tx) Insert(table string, record *Record) (int64, error) {
	return tx.dbMgr.insert(tx.tx, table, record)
}

func (tx *Tx) insertWithOptions(table string, record *Record, skipTimestamps bool) (int64, error) {
	return tx.dbMgr.insertWithOptions(tx.tx, table, record, skipTimestamps)
}

func (tx *Tx) Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.update(tx.tx, table, record, whereSql, whereArgs...)
}

func (tx *Tx) updateWithOptions(table string, record *Record, whereSql string, skipTimestamps bool, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.updateWithOptions(tx.tx, table, record, whereSql, skipTimestamps, whereArgs...)
}

func (tx *Tx) UpdateRecord(table string, record *Record) (int64, error) {
	return tx.dbMgr.updateRecord(tx.tx, table, record)
}

func (tx *Tx) Delete(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.delete(tx.tx, table, whereSql, whereArgs...)
}

func (tx *Tx) DeleteRecord(table string, record *Record) (int64, error) {
	return tx.dbMgr.deleteRecord(tx.tx, table, record)
}

func (tx *Tx) BatchInsert(table string, records []*Record, batchSize int) (int64, error) {
	return tx.dbMgr.batchInsert(tx.tx, table, records, batchSize)
}

func (tx *Tx) BatchInsertDefault(table string, records []*Record) (int64, error) {
	return tx.BatchInsert(table, records, 100)
}

// BatchUpdate updates multiple records by primary key within transaction
func (tx *Tx) BatchUpdate(table string, records []*Record, batchSize int) (int64, error) {
	return tx.dbMgr.batchUpdate(tx.tx, table, records, batchSize)
}

// BatchUpdateDefault updates multiple records with default batch size (100)
func (tx *Tx) BatchUpdateDefault(table string, records []*Record) (int64, error) {
	return tx.BatchUpdate(table, records, 100)
}

// BatchDelete deletes multiple records by primary key within transaction
func (tx *Tx) BatchDelete(table string, records []*Record, batchSize int) (int64, error) {
	return tx.dbMgr.batchDelete(tx.tx, table, records, batchSize)
}

// BatchDeleteDefault deletes multiple records with default batch size (100)
func (tx *Tx) BatchDeleteDefault(table string, records []*Record) (int64, error) {
	return tx.BatchDelete(table, records, 100)
}

// BatchDeleteByIds deletes records by primary key IDs within transaction
func (tx *Tx) BatchDeleteByIds(table string, ids []interface{}, batchSize int) (int64, error) {
	return tx.dbMgr.batchDeleteByIds(tx.tx, table, ids, batchSize)
}

// BatchDeleteByIdsDefault deletes records by IDs with default batch size (100)
func (tx *Tx) BatchDeleteByIdsDefault(table string, ids []interface{}) (int64, error) {
	return tx.BatchDeleteByIds(table, ids, 100)
}

func (tx *Tx) Count(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	if tx.cacheRepositoryName != "" {
		cache := tx.getEffectiveCache()
		key := GenerateCacheKey(tx.dbMgr.name, "COUNT:"+table+":"+whereSql, whereArgs...)
		if val, ok := cache.CacheGet(tx.cacheRepositoryName, key); ok {
			var count int64
			if convertCacheValue(val, &count) {
				return count, nil
			}
		}
		count, err := tx.dbMgr.count(tx.tx, table, whereSql, whereArgs...)
		if err == nil {
			cache.CacheSet(tx.cacheRepositoryName, key, count, getEffectiveTTL(tx.cacheRepositoryName, tx.cacheTTL))
		}
		return count, err
	}
	return tx.dbMgr.count(tx.tx, table, whereSql, whereArgs...)
}

func (tx *Tx) Exists(table string, whereSql string, whereArgs ...interface{}) (bool, error) {
	return tx.dbMgr.exists(tx.tx, table, whereSql, whereArgs...)
}

func (tx *Tx) PaginateBuilder(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) (*Page[Record], error) {
	if table != "" {
		if err := ValidateTableName(table); err != nil {
			return nil, err
		}
	}
	querySQL := selectSql
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(selectSql)), "SELECT ") {
		querySQL = "SELECT " + selectSql
	}

	if !strings.Contains(strings.ToUpper(querySQL), " FROM ") && table != "" {
		querySQL += " FROM " + table
	}
	if whereSql != "" {
		querySQL += " WHERE " + whereSql
	}
	if orderBySql != "" {
		querySQL += " ORDER BY " + orderBySql
	}

	if tx.cacheRepositoryName != "" {
		cache := tx.getEffectiveCache()
		key := GenerateCacheKey(tx.dbMgr.name, "PAGINATE:"+querySQL, args...)
		if val, ok := cache.CacheGet(tx.cacheRepositoryName, key); ok {
			var pageObj *Page[Record]
			if convertCacheValue(val, &pageObj) {
				return pageObj, nil
			}
		}
		list, totalRow, err := tx.dbMgr.paginate(tx.tx, querySQL, page, pageSize, args...)
		if err == nil {
			pageObj := NewPage(list, page, pageSize, totalRow)
			cache.CacheSet(tx.cacheRepositoryName, key, pageObj, getEffectiveTTL(tx.cacheRepositoryName, tx.cacheTTL))
			return pageObj, nil
		}
		return nil, err
	}

	list, totalRow, err := tx.dbMgr.paginate(tx.tx, querySQL, page, pageSize, args...)
	if err != nil {
		return nil, err
	}
	return NewPage(list, page, pageSize, totalRow), nil
}

// Paginate 事务分页方法，使用完整SQL语句进行分页查询
// 在事务上下文中自动解析SQL并根据数据库类型生成相应的分页语句
func (tx *Tx) Paginate(page int, pageSize int, querySQL string, args ...interface{}) (*Page[Record], error) {
	if tx.cacheRepositoryName != "" {
		cache := tx.getEffectiveCache()
		key := GenerateCacheKey(tx.dbMgr.name, "PAGINATE_SQL:"+querySQL, args...)
		if val, ok := cache.CacheGet(tx.cacheRepositoryName, key); ok {
			var pageObj *Page[Record]
			if convertCacheValue(val, &pageObj) {
				return pageObj, nil
			}
		}
		list, totalRow, err := tx.dbMgr.paginate(tx.tx, querySQL, page, pageSize, args...)
		if err == nil {
			pageObj := NewPage(list, page, pageSize, totalRow)
			cache.CacheSet(tx.cacheRepositoryName, key, pageObj, getEffectiveTTL(tx.cacheRepositoryName, tx.cacheTTL))
			return pageObj, nil
		}
		return nil, err
	}

	list, totalRow, err := tx.dbMgr.paginate(tx.tx, querySQL, page, pageSize, args...)
	if err != nil {
		return nil, err
	}
	return NewPage(list, page, pageSize, totalRow), nil
}

func (tx *Tx) FindAll(table string) ([]Record, error) {
	if err := ValidateTableName(table); err != nil {
		return nil, err
	}
	return tx.Query(fmt.Sprintf("SELECT * FROM %s", table))
}

// Struct methods for Tx
func (tx *Tx) SaveDbModel(model IDbModel) (int64, error) {
	record := ToRecord(model)
	return tx.Save(model.TableName(), record)
}

func (tx *Tx) InsertDbModel(model IDbModel) (int64, error) {
	record := ToRecord(model)
	return tx.Insert(model.TableName(), record)
}

func (tx *Tx) UpdateDbModel(model IDbModel) (int64, error) {
	record := ToRecord(model)
	return tx.UpdateRecord(model.TableName(), record)
}

func (tx *Tx) DeleteDbModel(model IDbModel) (int64, error) {
	record := ToRecord(model)
	return tx.DeleteRecord(model.TableName(), record)
}

func (tx *Tx) FindFirstToDbModel(model IDbModel, whereSql string, whereArgs ...interface{}) error {
	builder := tx.Table(model.TableName())
	if whereSql != "" {
		builder.Where(whereSql, whereArgs...)
	}
	return builder.FindFirstToDbModel(model)
}

func (tx *Tx) FindToDbModel(dest interface{}, table string, whereSql string, orderBySql string, whereArgs ...interface{}) error {
	builder := tx.Table(table)
	if whereSql != "" {
		builder.Where(whereSql, whereArgs...)
	}
	if orderBySql != "" {
		builder.OrderBy(orderBySql)
	}
	return builder.FindToDbModel(dest)
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

// convertCacheValue 将缓存值转换为目标类型
// 优先使用类型断言（零开销），失败时才使用 JSON 序列化（兼容 RedisCache）
func convertCacheValue(val interface{}, dest interface{}) bool {
	if val == nil {
		return false
	}

	// 1. 优先尝试直接类型断言（LocalCache 零开销路径）
	switch d := dest.(type) {
	case *[]Record:
		// 处理 []Record
		if v, ok := val.([]Record); ok {
			*d = v
			return true
		}
		// 处理 []*Record -> []Record 的转换
		// 注意：Record 包含 mutex，不能直接解引用拷贝
		// 这种情况应该保持指针形式或使用 JSON 序列化
		if _, ok := val.([]*Record); ok {
			// 降级到 JSON 序列化处理
			break
		}

	case *[]*Record:
		// 处理 []*Record
		if v, ok := val.([]*Record); ok {
			*d = v
			return true
		}
		// []Record -> []*Record 的转换
		// 由于 Record 包含 mutex，不能直接拷贝，降级到 JSON 序列化
		if _, ok := val.([]Record); ok {
			break
		}

	case **Record:
		// 处理 *Record
		if v, ok := val.(*Record); ok {
			*d = v
			return true
		}

	case *[]map[string]interface{}:
		// 处理 []map[string]interface{}
		if v, ok := val.([]map[string]interface{}); ok {
			*d = v
			return true
		}

	case *int64:
		// 处理 int64
		if v, ok := val.(int64); ok {
			*d = v
			return true
		}
		// 处理其他整数类型
		if v, ok := val.(int); ok {
			*d = int64(v)
			return true
		}
		if v, ok := val.(int32); ok {
			*d = int64(v)
			return true
		}

	case *int:
		// 处理 int
		if v, ok := val.(int); ok {
			*d = v
			return true
		}
		if v, ok := val.(int64); ok {
			*d = int(v)
			return true
		}

	case **Page[Record]:
		// 处理 *Page[Record]
		if v, ok := val.(*Page[Record]); ok {
			*d = v
			return true
		}
	}

	// 2. 处理 RedisCache 返回的 JSON 字节数组（优化路径）
	if jsonBytes, ok := val.([]byte); ok {
		// 直接从字节数组反序列化，避免字符串转换
		return json.Unmarshal(jsonBytes, dest) == nil
	}

	// 3. 降级到 JSON 转换（用于其他复杂类型）
	// 注意：这是最后的手段，性能较差
	data, err := json.Marshal(val)
	if err != nil {
		return false
	}
	return json.Unmarshal(data, dest) == nil
}
