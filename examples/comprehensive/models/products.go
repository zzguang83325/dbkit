package models

import (
	"time"

	"github.com/zzguang83325/dbkit"
)

// Product represents the products table
type Product struct {
	dbkit.ModelCache
	ID        int64     `column:"id" json:"id"`
	Name      string    `column:"name" json:"name"`
	Price     float64   `column:"price" json:"price"`
	Stock     int64     `column:"stock" json:"stock"`
	CreatedAt time.Time `column:"created_at" json:"created_at"`
}

// TableName returns the table name for Product struct
func (m *Product) TableName() string {
	return "products"
}

// DatabaseName returns the database name for Product struct
func (m *Product) DatabaseName() string {
	return "default"
}

// Cache sets the cache name and TTL for the next query
func (m *Product) Cache(name string, ttl ...time.Duration) *Product {
	m.SetCache(name, ttl...)
	return m
}

// ToJson converts Product to a JSON string
func (m *Product) ToJson() string {
	return dbkit.ToJson(m)
}

// Save saves the Product record (insert or update)
func (m *Product) Save() (int64, error) {
	return dbkit.SaveDbModel(m)
}

// Insert inserts the Product record
func (m *Product) Insert() (int64, error) {
	return dbkit.InsertDbModel(m)
}

// Update updates the Product record based on its primary key
func (m *Product) Update() (int64, error) {
	return dbkit.UpdateDbModel(m)
}

// Delete deletes the Product record based on its primary key
func (m *Product) Delete() (int64, error) {
	return dbkit.DeleteDbModel(m)
}

// ForceDelete performs a physical delete, bypassing soft delete
func (m *Product) ForceDelete() (int64, error) {
	return dbkit.ForceDeleteModel(m)
}

// Restore restores a soft-deleted record
func (m *Product) Restore() (int64, error) {
	return dbkit.RestoreModel(m)
}

// FindFirst finds the first Product record based on conditions
func (m *Product) FindFirst(whereSql string, args ...interface{}) (*Product, error) {
	result := &Product{}
	return dbkit.FindFirstModel(result, m.GetCache(), whereSql, args...)
}

// Find finds Product records based on conditions
func (m *Product) Find(whereSql string, orderBySql string, args ...interface{}) ([]*Product, error) {
	return dbkit.FindModel[*Product](m, m.GetCache(), whereSql, orderBySql, args...)
}

// FindWithTrashed finds Product records including soft-deleted ones
func (m *Product) FindWithTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*Product, error) {
	return dbkit.FindModelWithTrashed[*Product](m, m.GetCache(), whereSql, orderBySql, args...)
}

// FindOnlyTrashed finds only soft-deleted Product records
func (m *Product) FindOnlyTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*Product, error) {
	return dbkit.FindModelOnlyTrashed[*Product](m, m.GetCache(), whereSql, orderBySql, args...)
}

// PaginateBuilder paginates Product records based on conditions (traditional method)
func (m *Product) PaginateBuilder(page int, pageSize int, whereSql string, orderBy string, args ...interface{}) (*dbkit.Page[*Product], error) {
	return dbkit.PaginateModel[*Product](m, m.GetCache(), page, pageSize, whereSql, orderBy, args...)
}

// Paginate paginates Product records using complete SQL statement (recommended)
// 使用完整SQL语句进行分页查询，自动解析SQL并根据数据库类型生成相应的分页语句
func (m *Product) Paginate(page int, pageSize int, fullSQL string, args ...interface{}) (*dbkit.Page[*Product], error) {
	return dbkit.PaginateModel_FullSql[*Product](m, m.GetCache(), page, pageSize, fullSQL, args...)
}
