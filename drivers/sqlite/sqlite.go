// Package sqlite 提供SQLite数据库驱动支持
// 使用 github.com/mattn/go-sqlite3 驱动
package sqlite

import (
	_ "github.com/mattn/go-sqlite3" // SQLite3驱动
)

// 导入此包会自动注册SQLite驱动
// 使用方式：
// import _ "github.com/zzguang83325/dbkit/drivers/sqlite"
