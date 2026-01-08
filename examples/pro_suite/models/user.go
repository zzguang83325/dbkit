package models

import (
	"time"
	"github.com/zzguang83325/dbkit"
)

// User represents a user with advanced fields
type User struct {
	dbkit.ModelCache
	ID        int64     `column:"id" json:"id"`
	Username  string    `column:"username" json:"username"`
	Email     string    `column:"email" json:"email"`
	Role      string    `column:"role" json:"role"`
	Settings  string    `column:"settings" json:"settings"` // JSON field
	Credits   float64   `column:"credits" json:"credits"`
	CreatedAt time.Time `column:"created_at" json:"created_at"`
	UpdatedAt time.Time `column:"updated_at" json:"updated_at"`
}

func (m *User) TableName() string {
	return "pro_users"
}

func (m *User) DatabaseName() string {
	return "default"
}

func (m *User) Cache(name string, ttl ...time.Duration) *User {
	m.SetCache(name, ttl...)
	return m
}

func (m *User) Save() (int64, error) {
	return dbkit.SaveDbModel(m)
}

func (m *User) Insert() (int64, error) {
	return dbkit.InsertDbModel(m)
}

func (m *User) Update() (int64, error) {
	return dbkit.UpdateDbModel(m)
}

func (m *User) Delete() (int64, error) {
	return dbkit.DeleteDbModel(m)
}

func (m *User) FindFirst(whereSql string, args ...interface{}) (*User, error) {
	result := &User{}
	return dbkit.FindFirstModel(result, m.GetCache(), whereSql, args...)
}

func (m *User) Find(whereSql string, orderBySql string, args ...interface{}) ([]*User, error) {
	return dbkit.FindModel[*User](m, m.GetCache(), whereSql, orderBySql, args...)
}
