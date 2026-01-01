package dbkit

import (
	"fmt"
	"strings"
	"sync"
)

// ErrVersionMismatch is returned when an optimistic lock conflict is detected
var ErrVersionMismatch = fmt.Errorf("dbkit: optimistic lock conflict - record was modified by another transaction")

// OptimisticLockConfig holds the optimistic lock configuration for a table
type OptimisticLockConfig struct {
	VersionField string // Field name for version, e.g., "version", "revision"
}

// optimisticLockRegistry stores optimistic lock configurations per database
type optimisticLockRegistry struct {
	configs map[string]*OptimisticLockConfig // table -> config
	mu      sync.RWMutex
}

// newOptimisticLockRegistry creates a new optimistic lock registry
func newOptimisticLockRegistry() *optimisticLockRegistry {
	return &optimisticLockRegistry{
		configs: make(map[string]*OptimisticLockConfig),
	}
}

// set configures optimistic lock for a table
func (r *optimisticLockRegistry) set(table string, config *OptimisticLockConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs[strings.ToLower(table)] = config
}

// get returns the optimistic lock config for a table
func (r *optimisticLockRegistry) get(table string) *OptimisticLockConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.configs[strings.ToLower(table)]
}

// remove removes optimistic lock config for a table
func (r *optimisticLockRegistry) remove(table string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.configs, strings.ToLower(table))
}

// has checks if a table has optimistic lock configured
func (r *optimisticLockRegistry) has(table string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.configs[strings.ToLower(table)]
	return ok
}

// IOptimisticLockModel is an optional interface for models that support optimistic locking
type IOptimisticLockModel interface {
	IDbModel
	VersionField() string // Returns the version field name, empty string means not used
}

// --- Global Functions (for default database) ---

// ConfigOptimisticLock configures optimistic lock for a table using default field name "version"
func ConfigOptimisticLock(table string) {
	ConfigOptimisticLockWithField(table, "version")
}

// ConfigOptimisticLockWithField configures optimistic lock for a table with custom field name
func ConfigOptimisticLockWithField(table, versionField string) {
	db, err := defaultDB()
	if err != nil {
		return
	}
	db.ConfigOptimisticLockWithField(table, versionField)
}

// RemoveOptimisticLock removes optimistic lock configuration for a table
func RemoveOptimisticLock(table string) {
	db, err := defaultDB()
	if err != nil {
		return
	}
	db.RemoveOptimisticLock(table)
}

// HasOptimisticLock checks if a table has optimistic lock configured
func HasOptimisticLock(table string) bool {
	db, err := defaultDB()
	if err != nil {
		return false
	}
	return db.HasOptimisticLock(table)
}

// --- DB Methods ---

// ConfigOptimisticLock configures optimistic lock for a table using default field name "version"
func (db *DB) ConfigOptimisticLock(table string) *DB {
	return db.ConfigOptimisticLockWithField(table, "version")
}

// ConfigOptimisticLockWithField configures optimistic lock for a table with custom field name
func (db *DB) ConfigOptimisticLockWithField(table, versionField string) *DB {
	if db.lastErr != nil || db.dbMgr == nil {
		return db
	}
	db.dbMgr.setOptimisticLockConfig(table, &OptimisticLockConfig{
		VersionField: versionField,
	})
	return db
}

// RemoveOptimisticLock removes optimistic lock configuration for a table
func (db *DB) RemoveOptimisticLock(table string) *DB {
	if db.lastErr != nil || db.dbMgr == nil {
		return db
	}
	db.dbMgr.removeOptimisticLockConfig(table)
	return db
}

// HasOptimisticLock checks if a table has optimistic lock configured
func (db *DB) HasOptimisticLock(table string) bool {
	if db.lastErr != nil || db.dbMgr == nil {
		return false
	}
	return db.dbMgr.hasOptimisticLock(table)
}

// --- dbManager Methods ---

// setOptimisticLockConfig sets optimistic lock config for a table
func (mgr *dbManager) setOptimisticLockConfig(table string, config *OptimisticLockConfig) {
	if mgr.optimisticLocks == nil {
		mgr.optimisticLocks = newOptimisticLockRegistry()
	}
	mgr.optimisticLocks.set(table, config)
}

// getOptimisticLockConfig gets optimistic lock config for a table
func (mgr *dbManager) getOptimisticLockConfig(table string) *OptimisticLockConfig {
	if mgr.optimisticLocks == nil {
		return nil
	}
	return mgr.optimisticLocks.get(table)
}

// removeOptimisticLockConfig removes optimistic lock config for a table
func (mgr *dbManager) removeOptimisticLockConfig(table string) {
	if mgr.optimisticLocks == nil {
		return
	}
	mgr.optimisticLocks.remove(table)
}

// hasOptimisticLock checks if a table has optimistic lock configured
func (mgr *dbManager) hasOptimisticLock(table string) bool {
	if mgr.optimisticLocks == nil {
		return false
	}
	return mgr.optimisticLocks.has(table)
}

// applyVersionInit applies version initialization to a record if configured
// Sets version to 1 if not already set
func (mgr *dbManager) applyVersionInit(table string, record *Record) {
	config := mgr.getOptimisticLockConfig(table)
	if config == nil || config.VersionField == "" {
		return
	}
	// Only set if the field is not already set
	if !record.Has(config.VersionField) {
		record.Set(config.VersionField, int64(1))
	}
}

// getVersionFromRecord extracts the version value from a record
// Returns the version value and true if found, 0 and false otherwise
// Treats nil, empty string "", and non-numeric values as "no version"
func (mgr *dbManager) getVersionFromRecord(table string, record *Record) (int64, bool) {
	config := mgr.getOptimisticLockConfig(table)
	if config == nil || config.VersionField == "" {
		return 0, false
	}

	if !record.Has(config.VersionField) {
		return 0, false
	}

	val := record.Get(config.VersionField)
	if val == nil {
		return 0, false
	}

	// Convert to int64
	switch v := val.(type) {
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		// Empty string means no version
		if v == "" {
			return 0, false
		}
		// Try to parse string as number
		var n int64
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n, true
		}
		return 0, false
	}
	return 0, false
}
