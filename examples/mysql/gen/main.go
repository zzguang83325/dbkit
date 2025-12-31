package main

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
)

func main() {
	// 1. 连接 MySQL 数据库
	dsn := "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabaseWithDBName("mysql", dbkit.MySQL, dsn, 25)
	if err != nil {
		log.Fatalf("MySQL数据库连接失败: %v", err)
	}

	// 2. 确保表存在 (生成模型前需要表结构)
	setupTable()

	// 3. 生成模型
	// 参数: 表名, 输出路径(目录或文件), 结构体名称(可选)
	err = dbkit.Use("mysql").GenerateDbModel("demo", "../models", "Demo")
	if err != nil {
		log.Fatalf("生成模型失败: %v", err)
	}

	log.Println("MySQL Demo 模型生成成功")
}

func setupTable() {
	sql := `
	CREATE TABLE IF NOT EXISTS demo (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100),
		age INT,
		salary DECIMAL(10, 2),
		is_active BOOLEAN,
		birthday DATE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT
	)`
	_, err := dbkit.Use("mysql").Exec(sql)
	if err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
}
