package dbkit

import (
	"database/sql"
	"fmt"
	"strings"
)

// --- Global Functions (Operation on default database) ---

func Query(sql string, args ...interface{}) ([]Record, error) {
	return defaultDB().Query(sql, args...)
}

func QueryFirst(sql string, args ...interface{}) (*Record, error) {
	return defaultDB().QueryFirst(sql, args...)
}

func QueryMap(sql string, args ...interface{}) ([]map[string]interface{}, error) {
	return defaultDB().QueryMap(sql, args...)
}

func Exec(sql string, args ...interface{}) (sql.Result, error) {
	return defaultDB().Exec(sql, args...)
}

func Save(table string, record *Record) (int64, error) {
	return defaultDB().Save(table, record)
}

func Insert(table string, record *Record) (int64, error) {
	return defaultDB().Insert(table, record)
}

func Update(table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	return defaultDB().Update(table, record, where, whereArgs...)
}

func Delete(table string, where string, whereArgs ...interface{}) (int64, error) {
	return defaultDB().Delete(table, where, whereArgs...)
}

func BatchInsert(table string, records []*Record, batchSize int) (int64, error) {
	return defaultDB().BatchInsert(table, records, batchSize)
}

func BatchInsertDefault(table string, records []*Record) (int64, error) {
	return defaultDB().BatchInsertDefault(table, records)
}

func Count(table string, where string, whereArgs ...interface{}) (int64, error) {
	return defaultDB().Count(table, where, whereArgs...)
}

func Exists(table string, where string, whereArgs ...interface{}) bool {
	result, _ := defaultDB().Exists(table, where, whereArgs...)
	return result
}

func ExistsWithError(table string, where string, whereArgs ...interface{}) (bool, error) {
	return defaultDB().Exists(table, where, whereArgs...)
}

func Paginate(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) ([]Record, int64, error) {
	return defaultDB().Paginate(page, pageSize, selectSql, table, whereSql, orderBySql, args...)
}

func Transaction(fn func(*Tx) error) error {
	return defaultDB().Transaction(fn)
}

func Ping() error {
	dbMgr := GetCurrentDB()
	if dbMgr == nil {
		return fmt.Errorf("dbkit: database not initialized")
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
	dbMgr := GetCurrentDB()
	tx, err := dbMgr.getDB().Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, dbMgr: dbMgr}, nil
}

func ExecTx(tx *Tx, sql string, args ...interface{}) (sql.Result, error) {
	return tx.Exec(sql, args...)
}

func SaveTx(tx *Tx, table string, record *Record) (int64, error) {
	return tx.Save(table, record)
}

func UpdateTx(tx *Tx, table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	return tx.Update(table, record, where, whereArgs...)
}

func WithTransaction(fn func(*Tx) error) error {
	return Transaction(fn)
}

func FindAll(table string) ([]Record, error) {
	return defaultDB().FindAll(table)
}

// --- DB Methods (Operation on specific database instance) ---

func (db *DB) Query(sql string, args ...interface{}) ([]Record, error) {
	return db.dbMgr.query(db.dbMgr.getDB(), sql, args...)
}

func (db *DB) QueryFirst(sql string, args ...interface{}) (*Record, error) {
	return db.dbMgr.queryFirst(db.dbMgr.getDB(), sql, args...)
}

func (db *DB) QueryMap(sql string, args ...interface{}) ([]map[string]interface{}, error) {
	return db.dbMgr.queryMap(db.dbMgr.getDB(), sql, args...)
}

func (db *DB) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return db.dbMgr.exec(db.dbMgr.getDB(), sql, args...)
}

func (db *DB) Save(table string, record *Record) (int64, error) {
	return db.dbMgr.save(db.dbMgr.getDB(), table, record)
}

func (db *DB) Insert(table string, record *Record) (int64, error) {
	return db.dbMgr.insert(db.dbMgr.getDB(), table, record)
}

func (db *DB) Update(table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	return db.dbMgr.update(db.dbMgr.getDB(), table, record, where, whereArgs...)
}

func (db *DB) Delete(table string, where string, whereArgs ...interface{}) (int64, error) {
	return db.dbMgr.delete(db.dbMgr.getDB(), table, where, whereArgs...)
}

func (db *DB) BatchInsert(table string, records []*Record, batchSize int) (int64, error) {
	return db.dbMgr.batchInsert(db.dbMgr.getDB(), table, records, batchSize)
}

func (db *DB) BatchInsertDefault(table string, records []*Record) (int64, error) {
	return db.BatchInsert(table, records, 100)
}

func (db *DB) Count(table string, where string, whereArgs ...interface{}) (int64, error) {
	return db.dbMgr.count(db.dbMgr.getDB(), table, where, whereArgs...)
}

func (db *DB) Ping() error {
	return db.dbMgr.Ping()
}

func (db *DB) Exists(table string, where string, whereArgs ...interface{}) (bool, error) {
	return db.dbMgr.exists(db.dbMgr.getDB(), table, where, whereArgs...)
}

func (db *DB) Paginate(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) ([]Record, int64, error) {
	sql := selectSql
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(selectSql)), "SELECT ") {
		sql = "SELECT " + selectSql
	}

	if !strings.Contains(strings.ToUpper(sql), " FROM ") && table != "" {
		sql += " FROM " + table
	}
	if whereSql != "" {
		sql += " WHERE " + whereSql
	}
	if orderBySql != "" {
		sql += " ORDER BY " + orderBySql
	}
	return db.dbMgr.paginate(db.dbMgr.getDB(), sql, page, pageSize, args...)
}

func (db *DB) FindAll(table string) ([]Record, error) {
	return db.Query(fmt.Sprintf("SELECT * FROM %s", table))
}

// Transaction executes a function within a transaction
func (db *DB) Transaction(fn func(*Tx) error) error {
	tx, err := db.dbMgr.getDB().Begin()
	if err != nil {
		return err
	}

	dbtx := &Tx{tx: tx, dbMgr: db.dbMgr}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(dbtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// --- Tx Methods (Operation within a transaction) ---

func (tx *Tx) Query(sql string, args ...interface{}) ([]Record, error) {
	return tx.dbMgr.query(tx.tx, sql, args...)
}

func (tx *Tx) QueryFirst(sql string, args ...interface{}) (*Record, error) {
	return tx.dbMgr.queryFirst(tx.tx, sql, args...)
}

func (tx *Tx) QueryMap(sql string, args ...interface{}) ([]map[string]interface{}, error) {
	return tx.dbMgr.queryMap(tx.tx, sql, args...)
}

func (tx *Tx) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return tx.dbMgr.exec(tx.tx, sql, args...)
}

func (tx *Tx) Save(table string, record *Record) (int64, error) {
	return tx.dbMgr.save(tx.tx, table, record)
}

func (tx *Tx) Insert(table string, record *Record) (int64, error) {
	return tx.dbMgr.insert(tx.tx, table, record)
}

func (tx *Tx) Update(table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.update(tx.tx, table, record, where, whereArgs...)
}

func (tx *Tx) Delete(table string, where string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.delete(tx.tx, table, where, whereArgs...)
}

func (tx *Tx) BatchInsert(table string, records []*Record, batchSize int) (int64, error) {
	return tx.dbMgr.batchInsert(tx.tx, table, records, batchSize)
}

func (tx *Tx) BatchInsertDefault(table string, records []*Record) (int64, error) {
	return tx.BatchInsert(table, records, 100)
}

func (tx *Tx) Count(table string, where string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.count(tx.tx, table, where, whereArgs...)
}

func (tx *Tx) Exists(table string, where string, whereArgs ...interface{}) (bool, error) {
	return tx.dbMgr.exists(tx.tx, table, where, whereArgs...)
}

func (tx *Tx) Paginate(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) ([]Record, int64, error) {
	sql := selectSql
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(selectSql)), "SELECT ") {
		sql = "SELECT " + selectSql
	}

	if !strings.Contains(strings.ToUpper(sql), " FROM ") && table != "" {
		sql += " FROM " + table
	}
	if whereSql != "" {
		sql += " WHERE " + whereSql
	}
	if orderBySql != "" {
		sql += " ORDER BY " + orderBySql
	}
	return tx.dbMgr.paginate(tx.tx, sql, page, pageSize, args...)
}

func (tx *Tx) FindAll(table string) ([]Record, error) {
	return tx.Query(fmt.Sprintf("SELECT * FROM %s", table))
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}
