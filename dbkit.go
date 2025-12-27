// Package dbkit provides a database ActiveRecord-style library for Go
// inspired by JFinal's ActiveRecord pattern
package dbkit

// Re-export all public types and functions
import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)
