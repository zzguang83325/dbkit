//go:build ignore
// +build ignore

package main

import "fmt"

// 数据库配置示例
// 复制此文件为 config.go 并修改连接参数

const (
	// MySQL 连接配置
	MySQLHost     = "localhost"
	MySQLPort     = "3306"
	MySQLUser     = "root"
	MySQLPassword = "your_password_here"
	MySQLDatabase = "test_db"
)

// GetMySQLDSN 获取 MySQL 连接字符串
func GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		MySQLUser, MySQLPassword, MySQLHost, MySQLPort, MySQLDatabase)
}

// 使用方法：
// 1. 复制此文件为 config.go
// 2. 修改上面的数据库连接参数
// 3. 在 main.go 中使用 GetMySQLDSN() 替换硬编码的连接字符串
