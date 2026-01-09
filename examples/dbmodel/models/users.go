package models

import (
	"time"

	"github.com/zzguang83325/dbkit"
)

// User represents the users table
type User struct {
	dbkit.ModelCache `column:"-"`
	ID               int64      `column:"id" json:"id"`                 // 用户ID
	Username         string     `column:"username" json:"username"`     // 用户名
	Email            string     `column:"email" json:"email"`           // 邮箱
	Password         string     `column:"password" json:"password"`     // 密码
	Balance          *float64   `column:"balance" json:"balance"`       // 账户余额
	Status           *int64     `column:"status" json:"status"`         // 状态: 0-禁用, 1-启用
	CreatedAt        *time.Time `column:"created_at" json:"created_at"` // 创建时间
	UpdatedAt        *time.Time `column:"updated_at" json:"updated_at"` // 更新时间
	DeletedAt        *time.Time `column:"deleted_at" json:"deleted_at"` // 删除时间(软删除)
	Version          int64      `column:"version" json:"version"`       // 版本号(乐观锁)
}

// TableName returns the table name for User struct
func (m *User) TableName() string {
	return "users"
}

// DatabaseName returns the database name for User struct
func (m *User) DatabaseName() string {
	return "mysql"
}

// Cache sets the cache name and TTL for the next query
func (m *User) Cache(cacheRepositoryName string, ttl ...time.Duration) *User {
	m.SetCache(cacheRepositoryName, ttl...)
	return m
}

// WithCountCache 设置分页计数缓存时间，支持链式调用
func (m *User) WithCountCache(ttl time.Duration) *User {
	m.ModelCache.WithCountCache(ttl)
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
// Uses complete SQL statement for pagination query, automatically parses SQL and generates corresponding pagination statements based on database type
func (m *User) Paginate(page int, pageSize int, fullSQL string, args ...interface{}) (*dbkit.Page[*User], error) {
	return dbkit.PaginateModel_FullSql[*User](m, m.GetCache(), page, pageSize, fullSQL, args...)
}
