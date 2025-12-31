package main

import (
	"log"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/zzguang83325/dbkit"
)

func main() {
	// 1. 连接 SQL Server 数据库
	dsn := "sqlserver://sa:123456@192.168.10.44:1433?database=test"
	err := dbkit.OpenDatabaseWithDBName("sqlserver", dbkit.SQLServer, dsn, 25)
	if err != nil {
		log.Fatalf("SQL Server数据库连接失败: %v", err)
	}

	// 2. 确保表存在
	setupTable()

	// 3. 生成模型
	err = dbkit.Use("sqlserver").GenerateDbModel("demo", "../models", "Demo")
	if err != nil {
		log.Fatalf("生成模型失败: %v", err)
	}

	log.Println("SQL Server Demo 模型生成成功")
}

func setupTable() {
	sql := `
	IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'demo')
	CREATE TABLE demo (
		id INT IDENTITY(1,1) PRIMARY KEY,
		name NVARCHAR(100),
		age INT,
		salary DECIMAL(10, 2),
		is_active BIT,
		birthday DATE,
		created_at DATETIME DEFAULT GETDATE(),
		metadata NVARCHAR(MAX)
	)`
	_, err := dbkit.Use("sqlserver").Exec(sql)
	if err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
}
