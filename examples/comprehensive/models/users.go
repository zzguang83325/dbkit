package models

import (
	"time"

	"github.com/zzguang83325/dbkit"
)

// User represents the users table
type User struct {
	dbkit.ModelCache
	ID        int64     `column:"id" json:"id"`
	Username  string    `column:"username" json:"username"`
	Email     string    `column:"email" json:"email"`
	Age       int64     `column:"age" json:"age"`
	Status    string    `column:"status" json:"status"`
	DeletedAt time.Time `column:"deleted_at" json:"deleted_at"`
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
	m.SetCache(name, ttl...)
	return m
}

// ToJson converts User to a JSON string
func (m *User) ToJson() string {
	return dbkit.ToJson(m)
}

// Save saves the User record (insert or update)
func (m *User) Save() (int64, error) {
	return dbkit.SaveDbModel(m)
}

// Insert inserts the User record
func (m *User) Insert() (int64, error) {
	return dbkit.InsertDbModel(m)
}

// Update updates the User record based on its primary key
func (m *User) Update() (int64, error) {
	return dbkit.UpdateDbModel(m)
}

// Delete deletes the User record based on its primary key
func (m *User) Delete() (int64, error) {
	return dbkit.DeleteDbModel(m)
}

// ForceDelete performs a physical delete, bypassing soft delete
func (m *User) ForceDelete() (int64, error) {
	return dbkit.ForceDeleteModel(m)
}

// Restore restores a soft-deleted record
func (m *User) Restore() (int64, error) {
	return dbkit.RestoreModel(m)
}

// FindFirst finds the first User record based on conditions
func (m *User) FindFirst(whereSql string, args ...interface{}) (*User, error) {
	result := &User{}
	return dbkit.FindFirstModel(result, m.GetCache(), whereSql, args...)
}

// Find finds User records based on conditions
func (m *User) Find(whereSql string, orderBySql string, args ...interface{}) ([]*User, error) {
	return dbkit.FindModel[*User](m, m.GetCache(), whereSql, orderBySql, args...)
}

// FindWithTrashed finds User records including soft-deleted ones
func (m *User) FindWithTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*User, error) {
	return dbkit.FindModelWithTrashed[*User](m, m.GetCache(), whereSql, orderBySql, args...)
}

// FindOnlyTrashed finds only soft-deleted User records
func (m *User) FindOnlyTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*User, error) {
	return dbkit.FindModelOnlyTrashed[*User](m, m.GetCache(), whereSql, orderBySql, args...)
}

// PaginateBuilder paginates User records based on conditions (traditional method)
func (m *User) PaginateBuilder(page int, pageSize int, whereSql string, orderBy string, args ...interface{}) (*dbkit.Page[*User], error) {
	return dbkit.PaginateModel[*User](m, m.GetCache(), page, pageSize, whereSql, orderBy, args...)
}

// Paginate paginates User records using complete SQL statement (recommended)
// 使用完整SQL语句进行分页查询，自动解析SQL并根据数据库类型生成相应的分页语句
func (m *User) Paginate(page int, pageSize int, fullSQL string, args ...interface{}) (*dbkit.Page[*User], error) {
	return dbkit.PaginateModel_FullSql[*User](m, m.GetCache(), page, pageSize, fullSQL, args...)
}
