// Package ansi 提供标准SQL驱动支持
// 用于适配其他类型的数据库或自定义数据库驱动
package ansi

import (
	"database/sql"
	"database/sql/driver"
)

// StandardDriver 标准SQL驱动接口
// 可以用来包装其他数据库驱动或实现自定义驱动
type StandardDriver struct {
	driver.Driver
}

// RegisterCustomDriver 注册自定义数据库驱动
// 参数：
//
//	driverName: 驱动名称（如 "custom", "other" 等）
//	d: 实现了 driver.Driver 接口的驱动实例
func RegisterCustomDriver(driverName string, d driver.Driver) {
	sql.Register(driverName, d)
}

// 导入此包可以使用标准SQL驱动功能
// 使用方式：
// import "github.com/zzguang83325/dbkit/drivers/ansi"
//
// 用途：
// 1. 适配其他类型的数据库
// 2. 注册自定义数据库驱动
// 3. 包装现有驱动以添加额外功能
//
// 示例：
// ansi.RegisterCustomDriver("mydb", myCustomDriver)
// dbkit.OpenDatabase(dbkit.DriverType("mydb"), dsn, 10)
