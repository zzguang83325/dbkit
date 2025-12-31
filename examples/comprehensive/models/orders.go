package models

import (
	"time"
	"github.com/zzguang83325/dbkit"
)

// Order represents the orders table
type Order struct {
	cacheName string
	cacheTTL  time.Duration
	ID int64 `column:"id" json:"id"`
	UserID int64 `column:"user_id" json:"user_id"`
	Amount float64 `column:"amount" json:"amount"`
	Status string `column:"status" json:"status"`
	CreatedAt time.Time `column:"created_at" json:"created_at"`
}

// TableName returns the table name for Order struct
func (m *Order) TableName() string {
	return "orders"
}

// DatabaseName returns the database name for Order struct
func (m *Order) DatabaseName() string {
	return "default"
}

// Cache sets the cache name and TTL for the next query
func (m *Order) Cache(name string, ttl ...time.Duration) *Order {
	m.cacheName = name
	if len(ttl) > 0 {
		m.cacheTTL = ttl[0]
	} else {
		m.cacheTTL = -1
	}
	return m
}

// ToJson converts Order to a JSON string
func (m *Order) ToJson() string {
	return dbkit.ToJson(m)
}

// Save saves the Order record (insert or update)
func (m *Order) Save() (int64, error) {
	return dbkit.Use(m.DatabaseName()).SaveDbModel(m)
}

// Insert inserts the Order record
func (m *Order) Insert() (int64, error) {
	return dbkit.Use(m.DatabaseName()).InsertDbModel(m)
}

// Update updates the Order record based on its primary key
func (m *Order) Update() (int64, error) {
	return dbkit.Use(m.DatabaseName()).UpdateDbModel(m)
}

// Delete deletes the Order record based on its primary key
func (m *Order) Delete() (int64, error) {
	return dbkit.Use(m.DatabaseName()).DeleteDbModel(m)
}

// FindFirst finds the first Order record based on conditions
func (m *Order) FindFirst(whereSql string, args ...interface{}) (*Order, error) {
	result := &Order{}
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

// Find finds Order records based on conditions
func (m *Order) Find(whereSql string, orderBySql string, args ...interface{}) ([]*Order, error) {
	var results []*Order
	db := dbkit.Use(m.DatabaseName())
	if m.cacheName != "" {
		db = db.Cache(m.cacheName, m.cacheTTL)
	}
	err := db.Table(m.TableName()).Where(whereSql, args...).OrderBy(orderBySql).FindToDbModel(&results)
	return results, err
}

// Paginate paginates Order records based on conditions
func (m *Order) Paginate(page int, pageSize int, whereSql string, orderBy string, args ...interface{}) (*dbkit.Page[*Order], error) {
	db := dbkit.Use(m.DatabaseName())
	if m.cacheName != "" {
		db = db.Cache(m.cacheName, m.cacheTTL)
	}
	recordsPage, err := db.Table(m.TableName()).Where(whereSql, args...).OrderBy(orderBy).Paginate(page, pageSize)
	if err != nil {
		return nil, err
	}
	return dbkit.RecordPageToDbModelPage[*Order](recordsPage)
}
