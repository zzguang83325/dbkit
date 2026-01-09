package dbkit

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// TimestampConfig holds the timestamp configuration for a table
type TimestampConfig struct {
	CreatedAtField string // Field name for created_at, e.g., "created_at", "create_time"
	UpdatedAtField string // Field name for updated_at, e.g., "updated_at", "update_time"
}

// timestampRegistry stores timestamp configurations per database
type timestampRegistry struct {
	configs map[string]*TimestampConfig // table -> config
	mu      sync.RWMutex
}

// newTimestampRegistry creates a new timestamp registry
func newTimestampRegistry() *timestampRegistry {
	return &timestampRegistry{
		configs: make(map[string]*TimestampConfig),
	}
}

// set configures timestamps for a table
func (r *timestampRegistry) set(table string, config *TimestampConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs[strings.ToLower(table)] = config
}

// get returns the timestamp config for a table
func (r *timestampRegistry) get(table string) *TimestampConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.configs[strings.ToLower(table)]
}

// remove removes timestamp config for a table
func (r *timestampRegistry) remove(table string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.configs, strings.ToLower(table))
}

// has checks if a table has timestamps configured
func (r *timestampRegistry) has(table string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.configs[strings.ToLower(table)]
	return ok
}

// ITimestampModel is an optional interface for models that support auto timestamps
type ITimestampModel interface {
	IDbModel
	CreatedAtField() string // Returns the created_at field name, empty string means not used
	UpdatedAtField() string // Returns the updated_at field name, empty string means not used
}

// --- Global Functions (for default database) ---

// ConfigTimestamps configures auto timestamps for a table using default field names (created_at, updated_at)
func ConfigTimestamps(table string) {
	ConfigTimestampsWithFields(table, "created_at", "updated_at")
}

// ConfigTimestampsWithFields configures auto timestamps for a table with custom field names
func ConfigTimestampsWithFields(table, createdAtField, updatedAtField string) {
	db, err := defaultDB()
	if err != nil {
		return
	}
	db.ConfigTimestampsWithFields(table, createdAtField, updatedAtField)
}

// ConfigCreatedAt configures only the created_at field for a table
func ConfigCreatedAt(table, field string) {
	db, err := defaultDB()
	if err != nil {
		return
	}
	db.ConfigCreatedAt(table, field)
}

// ConfigUpdatedAt configures only the updated_at field for a table
func ConfigUpdatedAt(table, field string) {
	db, err := defaultDB()
	if err != nil {
		return
	}
	db.ConfigUpdatedAt(table, field)
}

// RemoveTimestamps removes timestamp configuration for a table
func RemoveTimestamps(table string) {
	db, err := defaultDB()
	if err != nil {
		return
	}
	db.RemoveTimestamps(table)
}

// HasTimestamps checks if a table has timestamps configured
func HasTimestamps(table string) bool {
	db, err := defaultDB()
	if err != nil {
		return false
	}
	return db.HasTimestamps(table)
}

// --- DB Methods ---

// ConfigTimestamps configures auto timestamps for a table using default field names
func (db *DB) ConfigTimestamps(table string) *DB {
	return db.ConfigTimestampsWithFields(table, "created_at", "updated_at")
}

// ConfigTimestampsWithFields configures auto timestamps for a table with custom field names
func (db *DB) ConfigTimestampsWithFields(table, createdAtField, updatedAtField string) *DB {
	if db.lastErr != nil || db.dbMgr == nil {
		return db
	}

	// 检查字段是否存在
	if createdAtField != "" && !db.dbMgr.checkTableColumn(table, createdAtField) {
		LogWarn(fmt.Sprintf("时间戳配置警告: 表 '%s' 中不存在字段 '%s'", table, createdAtField), map[string]interface{}{
			"db":    db.dbMgr.name,
			"table": table,
			"field": createdAtField,
		})
	}

	if updatedAtField != "" && !db.dbMgr.checkTableColumn(table, updatedAtField) {
		LogWarn(fmt.Sprintf("时间戳配置警告: 表 '%s' 中不存在字段 '%s'", table, updatedAtField), map[string]interface{}{
			"db":    db.dbMgr.name,
			"table": table,
			"field": updatedAtField,
		})
	}

	db.dbMgr.setTimestampConfig(table, &TimestampConfig{
		CreatedAtField: createdAtField,
		UpdatedAtField: updatedAtField,
	})
	return db
}

// ConfigCreatedAt configures only the created_at field for a table
func (db *DB) ConfigCreatedAt(table, field string) *DB {
	if db.lastErr != nil || db.dbMgr == nil {
		return db
	}
	existing := db.dbMgr.getTimestampConfig(table)
	if existing != nil {
		existing.CreatedAtField = field
		db.dbMgr.setTimestampConfig(table, existing)
	} else {
		db.dbMgr.setTimestampConfig(table, &TimestampConfig{
			CreatedAtField: field,
		})
	}
	return db
}

// ConfigUpdatedAt configures only the updated_at field for a table
func (db *DB) ConfigUpdatedAt(table, field string) *DB {
	if db.lastErr != nil || db.dbMgr == nil {
		return db
	}
	existing := db.dbMgr.getTimestampConfig(table)
	if existing != nil {
		existing.UpdatedAtField = field
		db.dbMgr.setTimestampConfig(table, existing)
	} else {
		db.dbMgr.setTimestampConfig(table, &TimestampConfig{
			UpdatedAtField: field,
		})
	}
	return db
}

// RemoveTimestamps removes timestamp configuration for a table
func (db *DB) RemoveTimestamps(table string) *DB {
	if db.lastErr != nil || db.dbMgr == nil {
		return db
	}
	db.dbMgr.removeTimestampConfig(table)
	return db
}

// HasTimestamps checks if a table has timestamps configured
func (db *DB) HasTimestamps(table string) bool {
	if db.lastErr != nil || db.dbMgr == nil {
		return false
	}
	return db.dbMgr.hasTimestamps(table)
}

// --- dbManager Methods ---

// setTimestampConfig sets timestamp config for a table
func (mgr *dbManager) setTimestampConfig(table string, config *TimestampConfig) {
	if mgr.timestamps == nil {
		mgr.timestamps = newTimestampRegistry()
	}
	mgr.timestamps.set(table, config)
}

// getTimestampConfig gets timestamp config for a table
func (mgr *dbManager) getTimestampConfig(table string) *TimestampConfig {
	if mgr.timestamps == nil {
		return nil
	}
	return mgr.timestamps.get(table)
}

// removeTimestampConfig removes timestamp config for a table
func (mgr *dbManager) removeTimestampConfig(table string) {
	if mgr.timestamps == nil {
		return
	}
	mgr.timestamps.remove(table)
}

// hasTimestamps checks if a table has timestamps configured
func (mgr *dbManager) hasTimestamps(table string) bool {
	if mgr.timestamps == nil {
		return false
	}
	return mgr.timestamps.has(table)
}

// applyCreatedAtTimestamp applies created_at timestamp to a record if configured
func (mgr *dbManager) applyCreatedAtTimestamp(table string, record *Record, skipTimestamps bool) {
	if skipTimestamps {
		return
	}
	config := mgr.getTimestampConfig(table)
	if config == nil || config.CreatedAtField == "" {
		return
	}

	// Check if field exists and has a valid value
	shouldSet := false
	if !record.Has(config.CreatedAtField) {
		shouldSet = true
	} else {
		// Check if the existing value is nil or a zero time
		val := record.Get(config.CreatedAtField)
		if val == nil {
			shouldSet = true
		} else if t, ok := val.(time.Time); ok && t.IsZero() {
			shouldSet = true
		} else if tp, ok := val.(*time.Time); ok && (tp == nil || tp.IsZero()) {
			shouldSet = true
		}
	}

	if shouldSet {
		record.Set(config.CreatedAtField, time.Now())
	}
}

// applyUpdatedAtTimestamp applies updated_at timestamp to a record if configured
func (mgr *dbManager) applyUpdatedAtTimestamp(table string, record *Record, skipTimestamps bool) {
	if skipTimestamps {
		return
	}
	config := mgr.getTimestampConfig(table)
	if config == nil || config.UpdatedAtField == "" {
		return
	}
	// Always update the updated_at field
	record.Set(config.UpdatedAtField, time.Now())
}
