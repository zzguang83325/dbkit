package models

import (
	"github.com/zzguang83325/dbkit"
)

// ArticleCategory represents the join table between articles and categories
type ArticleCategory struct {
	dbkit.ModelCache
	ArticleID  int64 `column:"article_id" json:"article_id"`
	CategoryID int64 `column:"category_id" json:"category_id"`
}

func (m *ArticleCategory) TableName() string {
	return "pro_article_categories"
}

func (m *ArticleCategory) DatabaseName() string {
	return "default"
}

func (m *ArticleCategory) Insert() (int64, error) {
	return dbkit.InsertDbModel(m)
}
