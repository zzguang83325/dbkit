package main

import (
	"log"
	"strings"

	_ "github.com/sijms/go-ora/v2"
	"github.com/zzguang83325/dbkit"
)

func main() {
	// 1. 连接 Oracle 数据库
	dsn := "oracle://test:123456@192.168.10.44:1521/orcl"
	err := dbkit.OpenDatabaseWithDBName("oracle", dbkit.Oracle, dsn, 25)
	if err != nil {
		log.Fatalf("Oracle数据库连接失败: %v", err)
	}

	// 2. 确保表存在
	setupTable()

	// 3. 生成模型
	err = dbkit.Use("oracle").GenerateDbModel("DEMO", "../models", "Demo")
	if err != nil {
		log.Fatalf("生成模型失败: %v", err)
	}

	log.Println("Oracle Demo 模型生成成功")
}

func setupTable() {
	// 尝试直接创建表，忽略已存在的错误
	sql := `CREATE TABLE DEMO (
			ID NUMBER PRIMARY KEY,
			NAME VARCHAR2(100),
			AGE NUMBER,
			SALARY NUMBER(10,2),
			IS_ACTIVE NUMBER(1),
			BIRTHDAY DATE,
			CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			METADATA CLOB
		)`
	_, err := dbkit.Use("oracle").Exec(sql)
	if err != nil {
		// ORA-00955: name is already used by an existing object
		if !strings.Contains(err.Error(), "ORA-00955") {
			log.Printf("Oracle table setup warning: %v", err)
		}
	}
}
