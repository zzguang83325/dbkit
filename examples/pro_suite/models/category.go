package models

import (
	"github.com/zzguang83325/dbkit"
)

// Category represents an article category
type Category struct {
	dbkit.ModelCache
	ID   int64  `column:"id" json:"id"`
	Name string `column:"name" json:"name"`
}

func (m *Category) TableName() string {
	return "pro_categories"
}

func (m *Category) DatabaseName() string {
	return "default"
}

func (m *Category) Insert() (int64, error) {
	return dbkit.InsertDbModel(m)
}

func (m *Category) Find(whereSql string, orderBySql string, args ...interface{}) ([]*Category, error) {
	return dbkit.FindModel[*Category](m, m.GetCache(), whereSql, orderBySql, args...)
}
