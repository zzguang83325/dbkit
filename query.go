package dbkit

import (
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
	return db.Update(table, record, whereSql, whereArgs...)
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

func Paginate(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) (*Page[Record], error) {
	db, err := defaultDB()
	if err != nil {
		return nil, err
	}
	return db.Paginate(page, pageSize, selectSql, table, whereSql, orderBySql, args...)
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

func (db *DB) Cache(name string, ttl ...time.Duration) *DB {
	db.cacheName = name
	if len(ttl) > 0 {
		db.cacheTTL = ttl[0]
	} else {
		db.cacheTTL = -1
	}
	return db
}

func (db *DB) Query(querySQL string, args ...interface{}) ([]Record, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	if db.cacheName != "" {
		key := GenerateCacheKey(db.dbMgr.name, querySQL, args...)
		if val, ok := GetCache().CacheGet(db.cacheName, key); ok {
			var results []Record
			if convertCacheValue(val, &results) {
				return results, nil
			}
		}
		results, err := db.dbMgr.query(db.dbMgr.getDB(), querySQL, args...)
		if err == nil {
			GetCache().CacheSet(db.cacheName, key, results, getEffectiveTTL(db.cacheName, db.cacheTTL))
		}
		return results, err
	}
	return db.dbMgr.query(db.dbMgr.getDB(), querySQL, args...)
}

func (db *DB) QueryFirst(querySQL string, args ...interface{}) (*Record, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	if db.cacheName != "" {
		key := GenerateCacheKey(db.dbMgr.name, querySQL, args...)
		if val, ok := GetCache().CacheGet(db.cacheName, key); ok {
			var result *Record
			if convertCacheValue(val, &result) {
				return result, nil
			}
		}
		result, err := db.dbMgr.queryFirst(db.dbMgr.getDB(), querySQL, args...)
		if err == nil && result != nil {
			GetCache().CacheSet(db.cacheName, key, result, getEffectiveTTL(db.cacheName, db.cacheTTL))
		}
		return result, err
	}
	return db.dbMgr.queryFirst(db.dbMgr.getDB(), querySQL, args...)
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
	if db.cacheName != "" {
		key := GenerateCacheKey(db.dbMgr.name, querySQL, args...)
		if val, ok := GetCache().CacheGet(db.cacheName, key); ok {
			var results []map[string]interface{}
			if convertCacheValue(val, &results) {
				return results, nil
			}
		}
		results, err := db.dbMgr.queryMap(db.dbMgr.getDB(), querySQL, args...)
		if err == nil {
			GetCache().CacheSet(db.cacheName, key, results, getEffectiveTTL(db.cacheName, db.cacheTTL))
		}
		return results, err
	}
	return db.dbMgr.queryMap(db.dbMgr.getDB(), querySQL, args...)
}

func (db *DB) Exec(querySQL string, args ...interface{}) (sql.Result, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	return db.dbMgr.exec(db.dbMgr.getDB(), querySQL, args...)
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

func (db *DB) Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.update(db.dbMgr.getDB(), table, record, whereSql, whereArgs...)
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

func (db *DB) Count(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	if db.cacheName != "" {
		key := GenerateCacheKey(db.dbMgr.name, "COUNT:"+table+":"+whereSql, whereArgs...)
		if val, ok := GetCache().CacheGet(db.cacheName, key); ok {
			var count int64
			if convertCacheValue(val, &count) {
				return count, nil
			}
		}
		count, err := db.dbMgr.count(db.dbMgr.getDB(), table, whereSql, whereArgs...)
		if err == nil {
			GetCache().CacheSet(db.cacheName, key, count, getEffectiveTTL(db.cacheName, db.cacheTTL))
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

func (db *DB) Paginate(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) (*Page[Record], error) {
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

	if db.cacheName != "" {
		key := GenerateCacheKey(db.dbMgr.name, "PAGINATE:"+querySQL, args...)
		if val, ok := GetCache().CacheGet(db.cacheName, key); ok {
			var pageObj *Page[Record]
			if convertCacheValue(val, &pageObj) {
				return pageObj, nil
			}
		}
		list, totalRow, err := db.dbMgr.paginate(db.dbMgr.getDB(), querySQL, page, pageSize, args...)
		if err == nil {
			pageObj := NewPage(list, page, pageSize, totalRow)
			GetCache().CacheSet(db.cacheName, key, pageObj, getEffectiveTTL(db.cacheName, db.cacheTTL))
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

func (tx *Tx) Cache(name string, ttl ...time.Duration) *Tx {
	tx.cacheName = name
	if len(ttl) > 0 {
		tx.cacheTTL = ttl[0]
	} else {
		tx.cacheTTL = -1
	}
	return tx
}

func (tx *Tx) Query(querySQL string, args ...interface{}) ([]Record, error) {
	if tx.cacheName != "" {
		key := GenerateCacheKey(tx.dbMgr.name, querySQL, args...)
		if val, ok := GetCache().CacheGet(tx.cacheName, key); ok {
			var results []Record
			if convertCacheValue(val, &results) {
				return results, nil
			}
		}
		results, err := tx.dbMgr.query(tx.tx, querySQL, args...)
		if err == nil {
			GetCache().CacheSet(tx.cacheName, key, results, getEffectiveTTL(tx.cacheName, tx.cacheTTL))
		}
		return results, err
	}
	return tx.dbMgr.query(tx.tx, querySQL, args...)
}

func (tx *Tx) QueryFirst(querySQL string, args ...interface{}) (*Record, error) {
	if tx.cacheName != "" {
		key := GenerateCacheKey(tx.dbMgr.name, querySQL, args...)
		if val, ok := GetCache().CacheGet(tx.cacheName, key); ok {
			var result *Record
			if convertCacheValue(val, &result) {
				return result, nil
			}
		}
		result, err := tx.dbMgr.queryFirst(tx.tx, querySQL, args...)
		if err == nil && result != nil {
			GetCache().CacheSet(tx.cacheName, key, result, getEffectiveTTL(tx.cacheName, tx.cacheTTL))
		}
		return result, err
	}
	return tx.dbMgr.queryFirst(tx.tx, querySQL, args...)
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
	if tx.cacheName != "" {
		key := GenerateCacheKey(tx.dbMgr.name, querySQL, args...)
		if val, ok := GetCache().CacheGet(tx.cacheName, key); ok {
			var results []map[string]interface{}
			if convertCacheValue(val, &results) {
				return results, nil
			}
		}
		results, err := tx.dbMgr.queryMap(tx.tx, querySQL, args...)
		if err == nil {
			GetCache().CacheSet(tx.cacheName, key, results, getEffectiveTTL(tx.cacheName, tx.cacheTTL))
		}
		return results, err
	}
	return tx.dbMgr.queryMap(tx.tx, querySQL, args...)
}

func (tx *Tx) Exec(querySQL string, args ...interface{}) (sql.Result, error) {
	return tx.dbMgr.exec(tx.tx, querySQL, args...)
}

func (tx *Tx) Save(table string, record *Record) (int64, error) {
	return tx.dbMgr.save(tx.tx, table, record)
}

func (tx *Tx) Insert(table string, record *Record) (int64, error) {
	return tx.dbMgr.insert(tx.tx, table, record)
}

func (tx *Tx) Update(table string, record *Record, whereSql string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.update(tx.tx, table, record, whereSql, whereArgs...)
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

func (tx *Tx) Count(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	if tx.cacheName != "" {
		key := GenerateCacheKey(tx.dbMgr.name, "COUNT:"+table+":"+whereSql, whereArgs...)
		if val, ok := GetCache().CacheGet(tx.cacheName, key); ok {
			var count int64
			if convertCacheValue(val, &count) {
				return count, nil
			}
		}
		count, err := tx.dbMgr.count(tx.tx, table, whereSql, whereArgs...)
		if err == nil {
			GetCache().CacheSet(tx.cacheName, key, count, getEffectiveTTL(tx.cacheName, tx.cacheTTL))
		}
		return count, err
	}
	return tx.dbMgr.count(tx.tx, table, whereSql, whereArgs...)
}

func (tx *Tx) Exists(table string, whereSql string, whereArgs ...interface{}) (bool, error) {
	return tx.dbMgr.exists(tx.tx, table, whereSql, whereArgs...)
}

func (tx *Tx) Paginate(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) (*Page[Record], error) {
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

	if tx.cacheName != "" {
		key := GenerateCacheKey(tx.dbMgr.name, "PAGINATE:"+querySQL, args...)
		if val, ok := GetCache().CacheGet(tx.cacheName, key); ok {
			var pageObj *Page[Record]
			if convertCacheValue(val, &pageObj) {
				return pageObj, nil
			}
		}
		list, totalRow, err := tx.dbMgr.paginate(tx.tx, querySQL, page, pageSize, args...)
		if err == nil {
			pageObj := NewPage(list, page, pageSize, totalRow)
			GetCache().CacheSet(tx.cacheName, key, pageObj, getEffectiveTTL(tx.cacheName, tx.cacheTTL))
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

func convertCacheValue(val interface{}, dest interface{}) bool {
	if val == nil {
		return false
	}

	// 1. Try direct type assertion first (for LocalCache)
	switch d := dest.(type) {
	case *[]Record:
		if v, ok := val.([]Record); ok {
			*d = v
			return true
		}
	case **Record:
		if v, ok := val.(*Record); ok {
			*d = v
			return true
		}
	case *[]map[string]interface{}:
		if v, ok := val.([]map[string]interface{}); ok {
			*d = v
			return true
		}
	case *int64:
		if v, ok := val.(int64); ok {
			*d = v
			return true
		}
	case **Page[Record]:
		if v, ok := val.(*Page[Record]); ok {
			*d = v
			return true
		}
	}

	// 2. Try JSON conversion (for RedisCache)
	data, err := json.Marshal(val)
	if err != nil {
		return false
	}
	return json.Unmarshal(data, dest) == nil
}
