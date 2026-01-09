package models

import (
	"time"

	"github.com/zzguang83325/dbkit"
)

// OrderItem represents the order_items table
type OrderItem struct {
	dbkit.ModelCache `column:"-"`
	ID               int64      `column:"id" json:"id"`                     // 订单明细ID
	OrderID          int64      `column:"order_id" json:"order_id"`         // 订单ID
	ProductID        int64      `column:"product_id" json:"product_id"`     // 商品ID
	ProductName      string     `column:"product_name" json:"product_name"` // 商品名称(快照)
	Price            float64    `column:"price" json:"price"`               // 单价(快照)
	Quantity         int64      `column:"quantity" json:"quantity"`         // 数量
	Subtotal         float64    `column:"subtotal" json:"subtotal"`         // 小计
	CreatedAt        *time.Time `column:"created_at" json:"created_at"`     // 创建时间
}

// TableName returns the table name for OrderItem struct
func (m *OrderItem) TableName() string {
	return "order_items"
}

// DatabaseName returns the database name for OrderItem struct
func (m *OrderItem) DatabaseName() string {
	return "mysql"
}

// Cache sets the cache name and TTL for the next query
func (m *OrderItem) Cache(cacheRepositoryName string, ttl ...time.Duration) *OrderItem {
	m.SetCache(cacheRepositoryName, ttl...)
	return m
}

// WithCountCache 设置分页计数缓存时间，支持链式调用
func (m *OrderItem) WithCountCache(ttl time.Duration) *OrderItem {
	m.ModelCache.WithCountCache(ttl)
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
// Uses complete SQL statement for pagination query, automatically parses SQL and generates corresponding pagination statements based on database type
func (m *OrderItem) Paginate(page int, pageSize int, fullSQL string, args ...interface{}) (*dbkit.Page[*OrderItem], error) {
	return dbkit.PaginateModel_FullSql[*OrderItem](m, m.GetCache(), page, pageSize, fullSQL, args...)
}
