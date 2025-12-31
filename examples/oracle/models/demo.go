package models

import (
	"time"
	"github.com/zzguang83325/dbkit"
)

// Demo represents the DEMO table
type Demo struct {
	cacheName string
	cacheTTL  time.Duration
	ID float64 `column:"ID" json:"id"`
	Name string `column:"NAME" json:"name"`
	Age float64 `column:"AGE" json:"age"`
	Salary float64 `column:"SALARY" json:"salary"`
	IsActive float64 `column:"IS_ACTIVE" json:"is_active"`
	Birthday time.Time `column:"BIRTHDAY" json:"birthday"`
	CreatedAt time.Time `column:"CREATED_AT" json:"created_at"`
	Metadata interface{} `column:"METADATA" json:"metadata"`
}

// TableName returns the table name for Demo struct
func (m *Demo) TableName() string {
	return "DEMO"
}

// DatabaseName returns the database name for Demo struct
func (m *Demo) DatabaseName() string {
	return "oracle"
}

// Cache sets the cache name and TTL for the next query
func (m *Demo) Cache(name string, ttl ...time.Duration) *Demo {
	m.cacheName = name
	if len(ttl) > 0 {
		m.cacheTTL = ttl[0]
	} else {
		m.cacheTTL = -1
	}
	return m
}

// ToJson converts Demo to a JSON string
func (m *Demo) ToJson() string {
	return dbkit.ToJson(m)
}

// Save saves the Demo record (insert or update)
func (m *Demo) Save() (int64, error) {
	return dbkit.Use(m.DatabaseName()).SaveDbModel(m)
}

// Insert inserts the Demo record
func (m *Demo) Insert() (int64, error) {
	return dbkit.Use(m.DatabaseName()).InsertDbModel(m)
}

// Update updates the Demo record based on its primary key
func (m *Demo) Update() (int64, error) {
	return dbkit.Use(m.DatabaseName()).UpdateDbModel(m)
}

// Delete deletes the Demo record based on its primary key
func (m *Demo) Delete() (int64, error) {
	return dbkit.Use(m.DatabaseName()).DeleteDbModel(m)
}

// FindFirst finds the first Demo record based on conditions
func (m *Demo) FindFirst(whereSql string, args ...interface{}) (*Demo, error) {
	result := &Demo{}
	db := dbkit.Use(m.DatabaseName())
	if m.cacheName != "" {
		db = db.Cache(m.cacheName, m.cacheTTL)
	}
	err := db.Table(m.TableName()).Where(whereSql, args...).FindFirstToDbModel(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Find finds Demo records based on conditions
func (m *Demo) Find(whereSql string, orderBySql string, args ...interface{}) ([]*Demo, error) {
	var results []*Demo
	db := dbkit.Use(m.DatabaseName())
	if m.cacheName != "" {
		db = db.Cache(m.cacheName, m.cacheTTL)
	}
	err := db.Table(m.TableName()).Where(whereSql, args...).OrderBy(orderBySql).FindToDbModel(&results)
	return results, err
}

// Paginate paginates Demo records based on conditions
func (m *Demo) Paginate(page int, pageSize int, whereSql string, orderBy string, args ...interface{}) (*dbkit.Page[*Demo], error) {
	db := dbkit.Use(m.DatabaseName())
	if m.cacheName != "" {
		db = db.Cache(m.cacheName, m.cacheTTL)
	}
	recordsPage, err := db.Table(m.TableName()).Where(whereSql, args...).OrderBy(orderBy).Paginate(page, pageSize)
	if err != nil {
		return nil, err
	}
	return dbkit.RecordPageToDbModelPage[*Demo](recordsPage)
}
