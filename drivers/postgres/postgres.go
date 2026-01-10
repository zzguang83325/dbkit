// Package postgres 提供PostgreSQL数据库驱动支持
// 使用 github.com/jackc/pgx/v5/stdlib 驱动（推荐，高性能）
package postgres

import (
	"database/sql"

	"github.com/jackc/pgx/v5/stdlib"
)

func init() {
	// 注册pgx驱动为"postgres"名称
	sql.Register("postgres", stdlib.GetDefaultDriver())
}

// 导入此包会自动注册PostgreSQL驱动
// 使用方式：
// import _ "github.com/zzguang83325/dbkit/drivers/postgres"
