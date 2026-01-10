// Package mysql 提供MySQL数据库驱动支持
// 使用 github.com/go-sql-driver/mysql 驱动
package mysql

import (
	_ "github.com/go-sql-driver/mysql" // MySQL驱动
)

// 导入此包会自动注册MySQL驱动
// 使用方式：
// import _ "github.com/zzguang83325/dbkit/drivers/mysql"
