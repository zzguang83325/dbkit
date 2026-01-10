// Package sqlserver 提供SQL Server数据库驱动支持
// 使用 github.com/denisenkom/go-mssqldb 驱动
package sqlserver

import (
	_ "github.com/denisenkom/go-mssqldb" // SQL Server驱动
)

// 导入此包会自动注册SQL Server驱动
// 使用方式：
// import _ "github.com/zzguang83325/dbkit/drivers/sqlserver"
