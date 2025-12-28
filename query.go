package dbkit

import (
	"database/sql"
	"fmt"
	"strings"
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

func Update(table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.Update(table, record, where, whereArgs...)
}

func Delete(table string, where string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.Delete(table, where, whereArgs...)
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

func Count(table string, where string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.Count(table, where, whereArgs...)
}

func Exists(table string, where string, whereArgs ...interface{}) bool {
	db, err := defaultDB()
	if err != nil {
		return false
	}
	result, _ := db.Exists(table, where, whereArgs...)
	return result
}

func ExistsWithError(table string, where string, whereArgs ...interface{}) (bool, error) {
	db, err := defaultDB()
	if err != nil {
		return false, err
	}
	return db.Exists(table, where, whereArgs...)
}

func Paginate(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) ([]Record, int64, error) {
	db, err := defaultDB()
	if err != nil {
		return nil, 0, err
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

func UpdateTx(tx *Tx, table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	return tx.Update(table, record, where, whereArgs...)
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

// --- DB Methods (Operation on specific database instance) ---

func (db *DB) Query(querySQL string, args ...interface{}) ([]Record, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	return db.dbMgr.query(db.dbMgr.getDB(), querySQL, args...)
}

func (db *DB) QueryFirst(querySQL string, args ...interface{}) (*Record, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
	}
	return db.dbMgr.queryFirst(db.dbMgr.getDB(), querySQL, args...)
}

func (db *DB) QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error) {
	if db.lastErr != nil {
		return nil, db.lastErr
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

func (db *DB) Update(table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.update(db.dbMgr.getDB(), table, record, where, whereArgs...)
}

func (db *DB) Delete(table string, where string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.delete(db.dbMgr.getDB(), table, where, whereArgs...)
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

func (db *DB) Count(table string, where string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.count(db.dbMgr.getDB(), table, where, whereArgs...)
}

func (db *DB) Ping() error {
	if db.lastErr != nil {
		return db.lastErr
	}
	return db.dbMgr.Ping()
}

func (db *DB) Exists(table string, where string, whereArgs ...interface{}) (bool, error) {
	if db.lastErr != nil {
		return false, db.lastErr
	}
	return db.dbMgr.exists(db.dbMgr.getDB(), table, where, whereArgs...)
}

func (db *DB) Paginate(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) ([]Record, int64, error) {
	if db.lastErr != nil {
		return nil, 0, db.lastErr
	}
	if table != "" {
		if err := ValidateTableName(table); err != nil {
			return nil, 0, err
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
	return db.dbMgr.paginate(db.dbMgr.getDB(), querySQL, page, pageSize, args...)
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

// Transaction executes a function within a transaction
func (db *DB) Transaction(fn func(*Tx) error) error {
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

func (tx *Tx) Query(querySQL string, args ...interface{}) ([]Record, error) {
	return tx.dbMgr.query(tx.tx, querySQL, args...)
}

func (tx *Tx) QueryFirst(querySQL string, args ...interface{}) (*Record, error) {
	return tx.dbMgr.queryFirst(tx.tx, querySQL, args...)
}

func (tx *Tx) QueryMap(querySQL string, args ...interface{}) ([]map[string]interface{}, error) {
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

func (tx *Tx) Update(table string, record *Record, where string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.update(tx.tx, table, record, where, whereArgs...)
}

func (tx *Tx) Delete(table string, where string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.delete(tx.tx, table, where, whereArgs...)
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

func (tx *Tx) Count(table string, where string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.count(tx.tx, table, where, whereArgs...)
}

func (tx *Tx) Exists(table string, where string, whereArgs ...interface{}) (bool, error) {
	return tx.dbMgr.exists(tx.tx, table, where, whereArgs...)
}

func (tx *Tx) Paginate(page int, pageSize int, selectSql string, table string, whereSql string, orderBySql string, args ...interface{}) ([]Record, int64, error) {
	if table != "" {
		if err := ValidateTableName(table); err != nil {
			return nil, 0, err
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
	return tx.dbMgr.paginate(tx.tx, querySQL, page, pageSize, args...)
}

func (tx *Tx) FindAll(table string) ([]Record, error) {
	if err := ValidateTableName(table); err != nil {
		return nil, err
	}
	return tx.Query(fmt.Sprintf("SELECT * FROM %s", table))
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}
