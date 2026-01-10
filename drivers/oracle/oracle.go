// Package oracle 提供Oracle数据库驱动支持
package oracle

import (
	_ "github.com/sijms/go-ora/v2" // Oracle驱动
)

// 导入此包会自动注册Oracle驱动
// 使用方式：
// import _ "github.com/zzguang83325/dbkit/drivers/oracle"
