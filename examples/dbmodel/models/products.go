package models

import (
	"time"

	"github.com/zzguang83325/dbkit"
)

// Product represents the products table
type Product struct {
	dbkit.ModelCache `column:"-"`
	ID               int64      `column:"id" json:"id"`                   // 商品ID
	Name             string     `column:"name" json:"name"`               // 商品名称
	Description      *string    `column:"description" json:"description"` // 商品描述
	Price            float64    `column:"price" json:"price"`             // 价格
	Stock            *int64     `column:"stock" json:"stock"`             // 库存数量
	Category         *string    `column:"category" json:"category"`       // 分类
	IsActive         *int64     `column:"is_active" json:"is_active"`     // 是否上架
	CreatedAt        *time.Time `column:"created_at" json:"created_at"`   // 创建时间
	UpdatedAt        *time.Time `column:"updated_at" json:"updated_at"`   // 更新时间
	DeletedAt        *time.Time `column:"deleted_at" json:"deleted_at"`   // 删除时间
}

// TableName returns the table name for Product struct
func (m *Product) TableName() string {
	return "products"
}

// DatabaseName returns the database name for Product struct
func (m *Product) DatabaseName() string {
	return "mysql"
}

// Cache sets the cache name and TTL for the next query
func (m *Product) Cache(cacheRepositoryName string, ttl ...time.Duration) *Product {
	m.SetCache(cacheRepositoryName, ttl...)
	return m
}

// WithCountCache 设置分页计数缓存时间，支持链式调用
func (m *Product) WithCountCache(ttl time.Duration) *Product {
	m.ModelCache.WithCountCache(ttl)
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
// Uses complete SQL statement for pagination query, automatically parses SQL and generates corresponding pagination statements based on database type
func (m *Product) Paginate(page int, pageSize int, fullSQL string, args ...interface{}) (*dbkit.Page[*Product], error) {
	return dbkit.PaginateModel_FullSql[*Product](m, m.GetCache(), page, pageSize, fullSQL, args...)
}
