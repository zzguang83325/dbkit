package dbkit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Record represents a single record in the database, similar to JFinal's ActiveRecord
type Record struct {
	columns map[string]interface{}
	mu      sync.RWMutex
}

// NewRecord creates a new empty Record
func NewRecord() *Record {
	return &Record{
		columns: make(map[string]interface{}),
	}
}

// Set sets a column value in the Record with case-insensitive support for existing columns
func (r *Record) Set(column string, value interface{}) *Record {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if a column with different case already exists
	for k := range r.columns {
		if strings.EqualFold(k, column) {
			r.columns[k] = value
			return r
		}
	}
	r.columns[column] = value
	return r
}

// getValue gets a column value from the Record with case-insensitive support
func (r *Record) getValue(column string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if val, ok := r.columns[column]; ok {
		return val
	}
	// Case-insensitive fallback
	for k, v := range r.columns {
		if strings.EqualFold(k, column) {
			return v
		}
	}
	return nil
}

// Get gets a column value from the Record
func (r *Record) Get(column string) interface{} {
	return r.getValue(column)
}

// GetInt gets a column value as int
func (r *Record) GetInt(column string) int {
	val := r.getValue(column)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case []byte:
		if i, err := strconv.Atoi(string(v)); err == nil {
			return i
		}
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}

// GetInt64 gets a column value as int64
func (r *Record) GetInt64(column string) int64 {
	val := r.getValue(column)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case []byte:
		if i, err := strconv.ParseInt(string(v), 10, 64); err == nil {
			return i
		}
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

// GetFloat gets a column value as float64
func (r *Record) GetFloat(column string) float64 {
	val := r.getValue(column)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case []byte:
		if f, err := strconv.ParseFloat(string(v), 64); err == nil {
			return f
		}
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

// GetTime gets a column value as time.Time
func (r *Record) GetTime(column string) time.Time {
	val := r.getValue(column)
	if val == nil {
		return time.Time{}
	}
	switch v := val.(type) {
	case time.Time:
		return v
	case *time.Time:
		if v != nil {
			return *v
		}
	case string:
		// Try some common formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02",
		}
		for _, f := range formats {
			if t, err := time.Parse(f, v); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// GetString gets a column value as string
func (r *Record) GetString(column string) string {
	val := r.getValue(column)
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	}
	return fmt.Sprintf("%v", val)
}

// GetBool gets a column value as bool
func (r *Record) GetBool(column string) bool {
	val := r.getValue(column)
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case string:
		return v == "1" || v == "true" || v == "TRUE"
	case []byte:
		s := string(v)
		return s == "1" || s == "true" || s == "TRUE"
	}
	return false
}

// Has checks if a column exists in the Record
func (r *Record) Has(column string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.columns[column]; ok {
		return true
	}
	for k := range r.columns {
		if strings.EqualFold(k, column) {
			return true
		}
	}
	return false
}

// Keys returns all column names
func (r *Record) Keys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.columns))
	for k := range r.columns {
		keys = append(keys, k)
	}
	return keys
}

// Remove removes a column from the Record
func (r *Record) Remove(column string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.columns, column)
}

// Clear clears all columns
// Clear clears all columns
func (r *Record) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.columns = make(map[string]interface{})
}

// ToMap converts the Record to a map
// ToMap converts the Record to a map (returns a copy)
func (r *Record) ToMap() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	newMap := make(map[string]interface{}, len(r.columns))
	for k, v := range r.columns {
		newMap[k] = v
	}
	return newMap
}

// ToJson converts the Record to JSON string
func (r *Record) ToJson() string {
	if r == nil {
		return "{}"
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := json.Marshal(r.columns)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// FromJson parses JSON string into the Record
func (r *Record) FromJson(jsonStr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return json.Unmarshal([]byte(jsonStr), &r.columns)
}

// Str returns the column name in string format
func (r *Record) Str(column string) string {
	return r.GetString(column)
}

// Int returns the column value as int
func (r *Record) Int(column string) int {
	return r.GetInt(column)
}

// Int64 returns the column value as int64
func (r *Record) Int64(column string) int64 {
	return r.GetInt64(column)
}

// Float returns the column value as float64
func (r *Record) Float(column string) float64 {
	return r.GetFloat(column)
}

// Bool returns the column value as bool
func (r *Record) Bool(column string) bool {
	return r.GetBool(column)
}
