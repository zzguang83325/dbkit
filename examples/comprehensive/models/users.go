package models

import (
	"time"
	"github.com/zzguang83325/dbkit"
)

// User represents the users table
type User struct {
	cacheName string
	cacheTTL  time.Duration
	ID int64 `column:"id" json:"id"`
	Username string `column:"username" json:"username"`
	Age int64 `column:"age" json:"age"`
	CreatedAt time.Time `column:"created_at" json:"created_at"`
}

// TableName returns the table name for User struct
func (m *User) TableName() string {
	return "users"
}

// DatabaseName returns the database name for User struct
func (m *User) DatabaseName() string {
	return "default"
}

// Cache sets the cache name and TTL for the next query
func (m *User) Cache(name string, ttl ...time.Duration) *User {
	m.cacheName = name
	if len(ttl) > 0 {
		m.cacheTTL = ttl[0]
	} else {
		m.cacheTTL = -1
	}
	return m
}

// ToJson converts User to a JSON string
func (m *User) ToJson() string {
	return dbkit.ToJson(m)
}

// Save saves the User record (insert or update)
func (m *User) Save() (int64, error) {
	return dbkit.Use(m.DatabaseName()).SaveDbModel(m)
}

// Insert inserts the User record
func (m *User) Insert() (int64, error) {
	return dbkit.Use(m.DatabaseName()).InsertDbModel(m)
}

// Update updates the User record based on its primary key
func (m *User) Update() (int64, error) {
	return dbkit.Use(m.DatabaseName()).UpdateDbModel(m)
}

// Delete deletes the User record based on its primary key
func (m *User) Delete() (int64, error) {
	return dbkit.Use(m.DatabaseName()).DeleteDbModel(m)
}

// FindFirst finds the first User record based on conditions
func (m *User) FindFirst(whereSql string, args ...interface{}) (*User, error) {
	result := &User{}
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

// Find finds User records based on conditions
func (m *User) Find(whereSql string, orderBySql string, args ...interface{}) ([]*User, error) {
	var results []*User
	db := dbkit.Use(m.DatabaseName())
	if m.cacheName != "" {
		db = db.Cache(m.cacheName, m.cacheTTL)
	}
	err := db.Table(m.TableName()).Where(whereSql, args...).OrderBy(orderBySql).FindToDbModel(&results)
	return results, err
}

// Paginate paginates User records based on conditions
func (m *User) Paginate(page int, pageSize int, whereSql string, orderBy string, args ...interface{}) (*dbkit.Page[*User], error) {
	db := dbkit.Use(m.DatabaseName())
	if m.cacheName != "" {
		db = db.Cache(m.cacheName, m.cacheTTL)
	}
	recordsPage, err := db.Table(m.TableName()).Where(whereSql, args...).OrderBy(orderBy).Paginate(page, pageSize)
	if err != nil {
		return nil, err
	}
	return dbkit.RecordPageToDbModelPage[*User](recordsPage)
}
