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
// columns 保留原始大小写用于生成 SQL，lowerKeyMap 用于大小写不敏感的快速查找
type Record struct {
	columns     map[string]interface{} // 原始键名 -> 值
	lowerKeyMap map[string]string      // 小写键名 -> 原始键名（用于快速查找）
	mu          sync.RWMutex
}

// NewRecord creates a new empty Record
func NewRecord() *Record {
	return &Record{
		columns:     make(map[string]interface{}),
		lowerKeyMap: make(map[string]string),
	}
}

// Set sets a column value in the Record with case-insensitive support for existing columns
// 保留原始大小写用于 SQL 生成，同时维护小写映射用于快速查找
func (r *Record) Set(column string, value interface{}) *Record {
	r.mu.Lock()
	defer r.mu.Unlock()

	lowerKey := strings.ToLower(column)

	// 如果已存在相同小写键名的字段，更新原有字段
	if existingKey, exists := r.lowerKeyMap[lowerKey]; exists {
		r.columns[existingKey] = value
		return r
	}

	// 新字段：保存原始大小写和映射关系
	r.columns[column] = value
	r.lowerKeyMap[lowerKey] = column
	return r
}

// getValue gets a column value from the Record with case-insensitive support
// 通过小写映射快速查找，O(1) 复杂度
func (r *Record) getValue(column string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lowerKey := strings.ToLower(column)
	if actualKey, exists := r.lowerKeyMap[lowerKey]; exists {
		return r.columns[actualKey]
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
// Has checks if a column exists in the Record
func (r *Record) Has(column string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lowerKey := strings.ToLower(column)
	_, exists := r.lowerKeyMap[lowerKey]
	return exists
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

// Remove removes a column from the Record with case-insensitive support
func (r *Record) Remove(column string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	lowerKey := strings.ToLower(column)
	if actualKey, exists := r.lowerKeyMap[lowerKey]; exists {
		delete(r.columns, actualKey)
		delete(r.lowerKeyMap, lowerKey)
	}
}

// Clear clears all columns
func (r *Record) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.columns = make(map[string]interface{})
	r.lowerKeyMap = make(map[string]string)
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
	data, err := r.MarshalJSON()
	if err != nil {
		return "{}"
	}
	return string(data)
}

// MarshalJSON implements the json.Marshaler interface
func (r *Record) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("{}"), nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return json.Marshal(r.columns)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (r *Record) UnmarshalJSON(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.columns == nil {
		r.columns = make(map[string]interface{})
	}
	if r.lowerKeyMap == nil {
		r.lowerKeyMap = make(map[string]string)
	}

	// 清空现有数据
	r.columns = make(map[string]interface{})
	r.lowerKeyMap = make(map[string]string)

	// 反序列化
	if err := json.Unmarshal(data, &r.columns); err != nil {
		return err
	}

	// 重建小写映射
	for k := range r.columns {
		r.lowerKeyMap[strings.ToLower(k)] = k
	}

	return nil
}

// FromJson parses JSON string into the Record
func (r *Record) FromJson(jsonStr string) error {
	return r.UnmarshalJSON([]byte(jsonStr))
}

// ToStruct converts the Record to a struct
func (r *Record) ToStruct(dest interface{}) error {
	return ToStruct(r, dest)
}

// FromStruct populates the Record from a struct
func (r *Record) FromStruct(src interface{}) error {
	return FromStruct(src, r)
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
