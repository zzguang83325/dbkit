package main

import (
	"log"

	"github.com/zzguang83325/dbkit"
	_ "github.com/zzguang83325/dbkit/drivers/postgres"
)

func main() {
	// 1. 连接 PostgreSQL 数据库
	dsn := "user=postgres password=123456 host=127.0.0.1 port=5432 dbname=postgres sslmode=disable"
	err := dbkit.OpenDatabaseWithDBName("postgresql", dbkit.PostgreSQL, dsn, 25)
	if err != nil {
		log.Fatalf("PostgreSQL数据库连接失败: %v", err)
	}

	// 2. 确保表存在
	setupTable()

	// 3. 生成模型
	err = dbkit.Use("postgresql").GenerateDbModel("demo", "../models", "Demo")
	if err != nil {
		log.Fatalf("生成模型失败: %v", err)
	}

	log.Println("PostgreSQL Demo 模型生成成功")
}

func setupTable() {
	sql := `
	CREATE TABLE IF NOT EXISTS demo (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100),
		age INT,
		salary DECIMAL(10, 2),
		is_active BOOLEAN,
		birthday DATE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		metadata JSONB
	)`
	_, err := dbkit.Use("postgresql").Exec(sql)
	if err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
}
