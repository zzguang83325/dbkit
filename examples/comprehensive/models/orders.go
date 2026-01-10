package models

import (
	"time"

	"github.com/zzguang83325/dbkit"
)

// Order represents the orders table
type Order struct {
	dbkit.ModelCache
	ID        int64     `column:"id" json:"id"`
	UserID    int64     `column:"user_id" json:"user_id"`
	Amount    float64   `column:"amount" json:"amount"`
	Status    string    `column:"status" json:"status"`
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
	m.SetCache(name, ttl...)
	return m
}

// ToJson converts Order to a JSON string
func (m *Order) ToJson() string {
	return dbkit.ToJson(m)
}

// Save saves the Order record (insert or update)
func (m *Order) Save() (int64, error) {
	return dbkit.SaveDbModel(m)
}

// Insert inserts the Order record
func (m *Order) Insert() (int64, error) {
	return dbkit.InsertDbModel(m)
}

// Update updates the Order record based on its primary key
func (m *Order) Update() (int64, error) {
	return dbkit.UpdateDbModel(m)
}

// Delete deletes the Order record based on its primary key
func (m *Order) Delete() (int64, error) {
	return dbkit.DeleteDbModel(m)
}

// ForceDelete performs a physical delete, bypassing soft delete
func (m *Order) ForceDelete() (int64, error) {
	return dbkit.ForceDeleteModel(m)
}

// Restore restores a soft-deleted record
func (m *Order) Restore() (int64, error) {
	return dbkit.RestoreModel(m)
}

// FindFirst finds the first Order record based on conditions
func (m *Order) FindFirst(whereSql string, args ...interface{}) (*Order, error) {
	result := &Order{}
	return dbkit.FindFirstModel(result, m.GetCache(), whereSql, args...)
}

// Find finds Order records based on conditions
func (m *Order) Find(whereSql string, orderBySql string, args ...interface{}) ([]*Order, error) {
	return dbkit.FindModel[*Order](m, m.GetCache(), whereSql, orderBySql, args...)
}

// FindWithTrashed finds Order records including soft-deleted ones
func (m *Order) FindWithTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*Order, error) {
	return dbkit.FindModelWithTrashed[*Order](m, m.GetCache(), whereSql, orderBySql, args...)
}

// FindOnlyTrashed finds only soft-deleted Order records
func (m *Order) FindOnlyTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*Order, error) {
	return dbkit.FindModelOnlyTrashed[*Order](m, m.GetCache(), whereSql, orderBySql, args...)
}

// PaginateBuilder paginates Order records based on conditions (traditional method)
func (m *Order) PaginateBuilder(page int, pageSize int, whereSql string, orderBy string, args ...interface{}) (*dbkit.Page[*Order], error) {
	return dbkit.PaginateModel[*Order](m, m.GetCache(), page, pageSize, whereSql, orderBy, args...)
}

// Paginate paginates Order records using complete SQL statement (recommended)
// 使用完整SQL语句进行分页查询，自动解析SQL并根据数据库类型生成相应的分页语句
func (m *Order) Paginate(page int, pageSize int, fullSQL string, args ...interface{}) (*dbkit.Page[*Order], error) {
	return dbkit.PaginateModel_FullSql[*Order](m, m.GetCache(), page, pageSize, fullSQL, args...)
}
