package dbkit

import (
	"fmt"
	"strings"
)

// Subquery represents a subquery that can be used in various SQL contexts
type Subquery struct {
	table     string
	selectSql string
	whereSql  []string
	whereArgs []interface{}
	orderBy   string
	limit     int
}

// NewSubquery creates a new Subquery builder
func NewSubquery() *Subquery {
	return &Subquery{
		selectSql: "*",
	}
}

// Table sets the table for the subquery
func (s *Subquery) Table(name string) *Subquery {
	s.table = name
	return s
}

// Select sets the columns to select
func (s *Subquery) Select(columns string) *Subquery {
	s.selectSql = columns
	return s
}

// Where adds a where condition
func (s *Subquery) Where(condition string, args ...interface{}) *Subquery {
	s.whereSql = append(s.whereSql, condition)
	s.whereArgs = append(s.whereArgs, args...)
	return s
}

// OrderBy sets the order by clause
func (s *Subquery) OrderBy(orderBy string) *Subquery {
	s.orderBy = orderBy
	return s
}

// Limit sets the limit
func (s *Subquery) Limit(limit int) *Subquery {
	s.limit = limit
	return s
}

// ToSQL returns the SQL string and arguments
func (s *Subquery) ToSQL() (string, []interface{}) {
	if s.table == "" {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("SELECT ")
	sb.WriteString(s.selectSql)
	sb.WriteString(" FROM ")
	sb.WriteString(s.table)

	if len(s.whereSql) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(s.whereSql, " AND "))
	}

	if s.orderBy != "" {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(s.orderBy)
	}

	if s.limit > 0 {
		fmt.Fprintf(&sb, " LIMIT %d", s.limit)
	}

	return sb.String(), s.whereArgs
}
