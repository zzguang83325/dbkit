package dbkit

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// SoftDeleteType represents the type of soft delete field
type SoftDeleteType int

const (
	// SoftDeleteTimestamp uses a timestamp field (e.g., deleted_at)
	// NULL means not deleted, non-NULL means deleted
	SoftDeleteTimestamp SoftDeleteType = iota
	// SoftDeleteBool uses a boolean field (e.g., is_deleted)
	// false means not deleted, true means deleted
	SoftDeleteBool
)

// SoftDeleteConfig holds the soft delete configuration for a table
type SoftDeleteConfig struct {
	Field string         // Field name, e.g., "deleted_at", "is_deleted"
	Type  SoftDeleteType // Field type: timestamp or boolean
}

// softDeleteRegistry stores soft delete configurations per database
type softDeleteRegistry struct {
	configs map[string]*SoftDeleteConfig // table -> config
	mu      sync.RWMutex
}

// newSoftDeleteRegistry creates a new soft delete registry
func newSoftDeleteRegistry() *softDeleteRegistry {
	return &softDeleteRegistry{
		configs: make(map[string]*SoftDeleteConfig),
	}
}

// set configures soft delete for a table
func (r *softDeleteRegistry) set(table string, config *SoftDeleteConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs[strings.ToLower(table)] = config
}

// get returns the soft delete config for a table
func (r *softDeleteRegistry) get(table string) *SoftDeleteConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.configs[strings.ToLower(table)]
}

// remove removes soft delete config for a table
func (r *softDeleteRegistry) remove(table string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.configs, strings.ToLower(table))
}

// has checks if a table has soft delete configured
func (r *softDeleteRegistry) has(table string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.configs[strings.ToLower(table)]
	return ok
}

// ISoftDeleteModel is an optional interface for models that support soft delete
type ISoftDeleteModel interface {
	IDbModel
	SoftDeleteField() string       // Returns the soft delete field name
	SoftDeleteType() SoftDeleteType // Returns the soft delete type
}

// --- Global Functions (for default database) ---

// ConfigSoftDelete configures soft delete for a table using timestamp type
func ConfigSoftDelete(table, field string) {
	ConfigSoftDeleteWithType(table, field, SoftDeleteTimestamp)
}

// ConfigSoftDeleteWithType configures soft delete for a table with specified type
func ConfigSoftDeleteWithType(table, field string, deleteType SoftDeleteType) {
	db, err := defaultDB()
	if err != nil {
		return
	}
	db.ConfigSoftDeleteWithType(table, field, deleteType)
}

// RemoveSoftDelete removes soft delete configuration for a table
func RemoveSoftDelete(table string) {
	db, err := defaultDB()
	if err != nil {
		return
	}
	db.RemoveSoftDelete(table)
}

// HasSoftDelete checks if a table has soft delete configured
func HasSoftDelete(table string) bool {
	db, err := defaultDB()
	if err != nil {
		return false
	}
	return db.HasSoftDelete(table)
}

// ForceDelete performs a physical delete, bypassing soft delete
func ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.ForceDelete(table, whereSql, whereArgs...)
}

// Restore restores soft-deleted records
func Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	db, err := defaultDB()
	if err != nil {
		return 0, err
	}
	return db.Restore(table, whereSql, whereArgs...)
}

// --- DB Methods ---

// ConfigSoftDelete configures soft delete for a table using timestamp type
func (db *DB) ConfigSoftDelete(table, field string) *DB {
	return db.ConfigSoftDeleteWithType(table, field, SoftDeleteTimestamp)
}

// ConfigSoftDeleteWithType configures soft delete for a table with specified type
func (db *DB) ConfigSoftDeleteWithType(table, field string, deleteType SoftDeleteType) *DB {
	if db.lastErr != nil || db.dbMgr == nil {
		return db
	}
	db.dbMgr.setSoftDeleteConfig(table, &SoftDeleteConfig{
		Field: field,
		Type:  deleteType,
	})
	return db
}

// RemoveSoftDelete removes soft delete configuration for a table
func (db *DB) RemoveSoftDelete(table string) *DB {
	if db.lastErr != nil || db.dbMgr == nil {
		return db
	}
	db.dbMgr.removeSoftDeleteConfig(table)
	return db
}

// HasSoftDelete checks if a table has soft delete configured
func (db *DB) HasSoftDelete(table string) bool {
	if db.lastErr != nil || db.dbMgr == nil {
		return false
	}
	return db.dbMgr.hasSoftDelete(table)
}

// ForceDelete performs a physical delete, bypassing soft delete
func (db *DB) ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.forceDelete(db.dbMgr.getDB(), table, whereSql, whereArgs...)
}

// Restore restores soft-deleted records
func (db *DB) Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	if db.lastErr != nil {
		return 0, db.lastErr
	}
	return db.dbMgr.restore(db.dbMgr.getDB(), table, whereSql, whereArgs...)
}

// --- Tx Methods ---

// ForceDelete performs a physical delete within a transaction
func (tx *Tx) ForceDelete(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.forceDelete(tx.tx, table, whereSql, whereArgs...)
}

// Restore restores soft-deleted records within a transaction
func (tx *Tx) Restore(table string, whereSql string, whereArgs ...interface{}) (int64, error) {
	return tx.dbMgr.restore(tx.tx, table, whereSql, whereArgs...)
}

// --- dbManager Methods ---

// setSoftDeleteConfig sets soft delete config for a table
func (mgr *dbManager) setSoftDeleteConfig(table string, config *SoftDeleteConfig) {
	if mgr.softDeletes == nil {
		mgr.softDeletes = newSoftDeleteRegistry()
	}
	mgr.softDeletes.set(table, config)
}

// getSoftDeleteConfig gets soft delete config for a table
func (mgr *dbManager) getSoftDeleteConfig(table string) *SoftDeleteConfig {
	if mgr.softDeletes == nil {
		return nil
	}
	return mgr.softDeletes.get(table)
}

// removeSoftDeleteConfig removes soft delete config for a table
func (mgr *dbManager) removeSoftDeleteConfig(table string) {
	if mgr.softDeletes == nil {
		return
	}
	mgr.softDeletes.remove(table)
}

// hasSoftDelete checks if a table has soft delete configured
func (mgr *dbManager) hasSoftDelete(table string) bool {
	if mgr.softDeletes == nil {
		return false
	}
	return mgr.softDeletes.has(table)
}

// softDelete performs a soft delete (UPDATE instead of DELETE)
func (mgr *dbManager) softDelete(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (int64, error) {
	config := mgr.getSoftDeleteConfig(table)
	if config == nil {
		return 0, fmt.Errorf("soft delete not configured for table %s", table)
	}

	var setValue string
	var setArgs []interface{}

	switch config.Type {
	case SoftDeleteTimestamp:
		setValue = fmt.Sprintf("%s = ?", config.Field)
		setArgs = append(setArgs, time.Now())
	case SoftDeleteBool:
		setValue = fmt.Sprintf("%s = ?", config.Field)
		setArgs = append(setArgs, true)
	}

	// Build UPDATE query
	querySQL := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, setValue, where)
	allArgs := append(setArgs, whereArgs...)

	querySQL = mgr.convertPlaceholder(querySQL, mgr.config.Driver)
	allArgs = mgr.sanitizeArgs(querySQL, allArgs)

	start := time.Now()
	result, err := executor.Exec(querySQL, allArgs...)
	mgr.logTrace(start, querySQL, allArgs, err)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// forceDelete performs a physical delete, bypassing soft delete
func (mgr *dbManager) forceDelete(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}
	if where == "" {
		return 0, fmt.Errorf("where condition is required for delete")
	}

	querySQL := fmt.Sprintf("DELETE FROM %s WHERE %s", table, where)
	querySQL, whereArgs = mgr.prepareQuerySQL(querySQL, whereArgs...)

	start := time.Now()
	result, err := executor.Exec(querySQL, whereArgs...)
	mgr.logTrace(start, querySQL, whereArgs, err)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// restore restores soft-deleted records
func (mgr *dbManager) restore(executor sqlExecutor, table string, where string, whereArgs ...interface{}) (int64, error) {
	if err := validateIdentifier(table); err != nil {
		return 0, err
	}

	config := mgr.getSoftDeleteConfig(table)
	if config == nil {
		return 0, fmt.Errorf("soft delete not configured for table %s", table)
	}

	var setValue string
	var setArgs []interface{}

	switch config.Type {
	case SoftDeleteTimestamp:
		setValue = fmt.Sprintf("%s = NULL", config.Field)
	case SoftDeleteBool:
		setValue = fmt.Sprintf("%s = ?", config.Field)
		setArgs = append(setArgs, false)
	}

	// Build UPDATE query
	querySQL := fmt.Sprintf("UPDATE %s SET %s", table, setValue)
	if where != "" {
		querySQL += " WHERE " + where
	}

	allArgs := append(setArgs, whereArgs...)
	querySQL = mgr.convertPlaceholder(querySQL, mgr.config.Driver)
	allArgs = mgr.sanitizeArgs(querySQL, allArgs)

	start := time.Now()
	result, err := executor.Exec(querySQL, allArgs...)
	mgr.logTrace(start, querySQL, allArgs, err)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// buildSoftDeleteCondition builds the WHERE condition for filtering soft-deleted records
func (mgr *dbManager) buildSoftDeleteCondition(table string, includeDeleted, onlyDeleted bool) string {
	config := mgr.getSoftDeleteConfig(table)
	if config == nil {
		return ""
	}

	// If including all records (withTrashed), no filter needed
	if includeDeleted && !onlyDeleted {
		return ""
	}

	switch config.Type {
	case SoftDeleteTimestamp:
		if onlyDeleted {
			return fmt.Sprintf("%s IS NOT NULL", config.Field)
		}
		return fmt.Sprintf("%s IS NULL", config.Field)
	case SoftDeleteBool:
		if onlyDeleted {
			return fmt.Sprintf("%s = true", config.Field)
		}
		return fmt.Sprintf("%s = false", config.Field)
	}
	return ""
}
