package models

import (
	"time"
	"github.com/zzguang83325/dbkit"
)

// Article represents a blog post with optimistic locking and soft delete
type Article struct {
	dbkit.ModelCache
	ID        int64     `column:"id" json:"id"`
	AuthorID  int64     `column:"author_id" json:"author_id"`
	Title     string    `column:"title" json:"title"`
	Content   string    `column:"content" json:"content"`
	Status    string    `column:"status" json:"status"`
	Version   int64     `column:"version" json:"version"` // For Optimistic Lock
	DeletedAt time.Time `column:"deleted_at" json:"deleted_at"` // For Soft Delete
	CreatedAt time.Time `column:"created_at" json:"created_at"`
	UpdatedAt time.Time `column:"updated_at" json:"updated_at"`
}

func (m *Article) TableName() string {
	return "pro_articles"
}

func (m *Article) DatabaseName() string {
	return "default"
}

func (m *Article) Cache(name string, ttl ...time.Duration) *Article {
	m.SetCache(name, ttl...)
	return m
}

func (m *Article) Save() (int64, error) {
	return dbkit.SaveDbModel(m)
}

func (m *Article) Insert() (int64, error) {
	return dbkit.InsertDbModel(m)
}

func (m *Article) Update() (int64, error) {
	return dbkit.UpdateDbModel(m)
}

func (m *Article) Delete() (int64, error) {
	return dbkit.DeleteDbModel(m)
}

func (m *Article) ForceDelete() (int64, error) {
	return dbkit.ForceDeleteModel(m)
}

func (m *Article) Restore() (int64, error) {
	return dbkit.RestoreModel(m)
}

func (m *Article) FindFirst(whereSql string, args ...interface{}) (*Article, error) {
	result := &Article{}
	return dbkit.FindFirstModel(result, m.GetCache(), whereSql, args...)
}

func (m *Article) Find(whereSql string, orderBySql string, args ...interface{}) ([]*Article, error) {
	return dbkit.FindModel[*Article](m, m.GetCache(), whereSql, orderBySql, args...)
}

func (m *Article) FindWithTrashed(whereSql string, orderBySql string, args ...interface{}) ([]*Article, error) {
	return dbkit.FindModelWithTrashed[*Article](m, m.GetCache(), whereSql, orderBySql, args...)
}

func (m *Article) Paginate(page int, pageSize int, fullSQL string, args ...interface{}) (*dbkit.Page[*Article], error) {
	return dbkit.PaginateModel_FullSql[*Article](m, m.GetCache(), page, pageSize, fullSQL, args...)
}
