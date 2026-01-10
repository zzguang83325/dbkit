package models

import (
	"time"

	"github.com/zzguang83325/dbkit"
)

// OrderItem represents the order_items table
type OrderItem struct {
	dbkit.ModelCache
	ID        int64   `column:"id" json:"id"`
	OrderID   int64   `column:"order_id" json:"order_id"`
	ProductID int64   `column:"product_id" json:"product_id"`
	Quantity  int64   `column:"quantity" json:"quantity"`
	Price     float64 `column:"price" json:"price"`
}

// TableName returns the table name for OrderItem struct
func (m *OrderItem) TableName() string {
	return "order_items"
}

// DatabaseName returns the database name for OrderItem struct
func (m *OrderItem) DatabaseName() string {
	return "default"
}

// Cache sets the cache name and TTL for the next query
func (m *OrderItem) Cache(name string, ttl ...time.Duration) *OrderItem {
	m.SetCache(name, ttl...)
	return m
}

// ToJson converts OrderItem to a JSON string
func (m *OrderItem) ToJson() string {
	return dbkit.ToJson(m)
}

// Save saves the OrderItem record (insert or update)
func (m *OrderItem) Save() (int64, error) {
	return dbkit.SaveDbModel(m)
}

// Insert inserts the OrderItem record
func (m *OrderItem) Insert() (int64, error) {
	return dbkit.InsertDbModel(m)
}

// Update updates the OrderItem record based on its primary key
func (m *OrderItem) Update() (int64, error) {
	return dbkit.UpdateDbModel(m)
}

// Delete deletes the OrderItem record based on its primary key
func (m *OrderItem) Delete() (int64, error) {
	return dbkit.DeleteDbModel(m)
}

// ForceDelete performs a physical delete, bypassing soft delete
func (m *OrderItem) ForceDelete() (int64, error) {
	return dbkit.ForceDeleteModel(m)
}

// Restore restores a soft-deleted record
func (m *OrderItem) Restore() (int64, error) {
	return dbkit.RestoreModel(m)
}

// FindFirst finds the first OrderItem record based on conditions
func (m *OrderItem) FindFirst(whereSql string, args ...interface{}) (*OrderItem, error) {
	result := &OrderItem{}
	return dbkit.FindFirstModel(result, m.GetCache(), whereSql, args...)
}

// Find finds OrderItem records based on conditions
func (m *OrderItem) Find(whereSql string, orderBySql string, args ...interface{}) ([]*OrderItem, error) {
	return dbkit.FindModel[*OrderItem](m, m.GetCache(), whereSql, orderBySql, args...)
}

// FindWithTrashed finds OrderItem records including soft-deleted ones
func (m *OrderItem) FindWithTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*OrderItem, error) {
	return dbkit.FindModelWithTrashed[*OrderItem](m, m.GetCache(), whereSql, orderBySql, args...)
}

// FindOnlyTrashed finds only soft-deleted OrderItem records
func (m *OrderItem) FindOnlyTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*OrderItem, error) {
	return dbkit.FindModelOnlyTrashed[*OrderItem](m, m.GetCache(), whereSql, orderBySql, args...)
}

// PaginateBuilder paginates OrderItem records based on conditions (traditional method)
func (m *OrderItem) PaginateBuilder(page int, pageSize int, whereSql string, orderBy string, args ...interface{}) (*dbkit.Page[*OrderItem], error) {
	return dbkit.PaginateModel[*OrderItem](m, m.GetCache(), page, pageSize, whereSql, orderBy, args...)
}

// Paginate paginates OrderItem records using complete SQL statement (recommended)
// 使用完整SQL语句进行分页查询，自动解析SQL并根据数据库类型生成相应的分页语句
func (m *OrderItem) Paginate(page int, pageSize int, fullSQL string, args ...interface{}) (*dbkit.Page[*OrderItem], error) {
	return dbkit.PaginateModel_FullSql[*OrderItem](m, m.GetCache(), page, pageSize, fullSQL, args...)
}
