package main

import (
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
)

func main() {
	// 1. 初始化数据库连接
	// 使用本地文件数据库，以便生成模型时能读取到表结构
	dbPath := "../comprehensive.db"
	err := dbkit.OpenDatabase(dbkit.SQLite3, dbPath, 10)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer dbkit.Close()

	// 2. 预先创建表结构 (如果不存在)
	setupTables()

	// 3. 执行生成任务
	// 针对 users 表生成模型，保存到 ../models 目录，结构体名为 User
	err = dbkit.GenerateDbModel("users", "../models/users.go", "User")
	if err != nil {
		log.Fatalf("生成 User 模型失败: %v", err)
	}

	// 针对 orders 表生成模型，保存到 ../models 目录，结构体名为 Order
	err = dbkit.GenerateDbModel("orders", "../models/orders.go", "Order")
	if err != nil {
		log.Fatalf("生成 Order 模型失败: %v", err)
	}

	log.Println("模型生成成功！代码已保存至 examples/comprehensive/models 目录")
}

func setupTables() {
	// 用户表
	userTable := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		age INTEGER,
		created_at DATETIME
	)`
	// 订单表
	orderTable := `CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		amount DECIMAL(10,2),
		status TEXT,
		created_at DATETIME
	)`
	dbkit.Exec(userTable)
	dbkit.Exec(orderTable)
}
