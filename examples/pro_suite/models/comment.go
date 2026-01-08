package models

import (
	"time"
	"github.com/zzguang83325/dbkit"
)

// Comment represents a comment on an article
type Comment struct {
	dbkit.ModelCache
	ID        int64     `column:"id" json:"id"`
	ArticleID int64     `column:"article_id" json:"article_id"`
	UserID    int64     `column:"user_id" json:"user_id"`
	ParentID  int64     `column:"parent_id" json:"parent_id"` // For nested comments
	Content   string    `column:"content" json:"content"`
	CreatedAt time.Time `column:"created_at" json:"created_at"`
}

func (m *Comment) TableName() string {
	return "pro_comments"
}

func (m *Comment) DatabaseName() string {
	return "default"
}

func (m *Comment) Cache(name string, ttl ...time.Duration) *Comment {
	m.SetCache(name, ttl...)
	return m
}

func (m *Comment) Insert() (int64, error) {
	return dbkit.InsertDbModel(m)
}

func (m *Comment) Find(whereSql string, orderBySql string, args ...interface{}) ([]*Comment, error) {
	return dbkit.FindModel[*Comment](m, m.GetCache(), whereSql, orderBySql, args...)
}
