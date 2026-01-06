package models

import (
	"time"

	"github.com/zzguang83325/dbkit"
)

// User 用户模型
type User struct {
	dbkit.ModelCache
	ID        int64     `column:"id" json:"id"`
	Name      string    `column:"name" json:"name"`
	Email     string    `column:"email" json:"email"`
	Age       int64     `column:"age" json:"age"`
	Status    string    `column:"status" json:"status"`
	CreatedAt time.Time `column:"created_at" json:"created_at"`
}

func (u *User) TableName() string {
	return "pagination_demo_users"
}

func (u *User) DatabaseName() string {
	return "mysql"
}

// Cache 设置缓存
func (u *User) Cache(name string, ttl ...time.Duration) *User {
	u.SetCache(name, ttl...)
	return u
}

// PaginateBuilder 传统分页方法（构建式）
func (u *User) PaginateBuilder(page int, pageSize int, whereSql string, orderBy string, args ...interface{}) (*dbkit.Page[*User], error) {
	return dbkit.PaginateModel[*User](u, u.GetCache(), page, pageSize, whereSql, orderBy, args...)
}

// Paginate 使用完整SQL进行分页查询（推荐方法）
func (u *User) Paginate(page int, pageSize int, querySQL string, args ...interface{}) (*dbkit.Page[*User], error) {
	db := dbkit.Use(u.DatabaseName())
	if cache := u.GetCache(); cache != nil && cache.CacheName != "" {
		db = db.Cache(cache.CacheName, cache.CacheTTL)
	}
	recordsPage, err := db.Paginate(page, pageSize, querySQL, args...)
	if err != nil {
		return nil, err
	}
	return dbkit.RecordPageToDbModelPage[*User](recordsPage)
}
