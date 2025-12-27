package main

import (
	"dbkit"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// printLog 统一输出函数，同时在控制台打印并记录到日志系统
func printLog(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(msg)
	dbkit.LogInfo(msg)
}

func main() {
	printLog("=== DBKit 多数据库综合测试示例 ===")

	// 0. 初始化日志记录
	dbkit.SetDebugMode(true)
	logFilePath := filepath.Join(".", "log.log")
	dbkit.InitLoggerWithFile("debug", logFilePath)

	// 同时让标准库 log 也输出到该文件
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	printLog("0. 已启用调试模式，日志将同时输出到控制台和文件: %s", logFilePath)

	// 1. 注册五种数据库
	printLog("\n1. 正在连接数据库...")

	// MySQL
	dbkit.OpenDatabaseWithDBName("mysql", dbkit.MySQL, "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 25)

	// PostgreSQL
	dbkit.OpenDatabaseWithDBName("postgresql", dbkit.PostgreSQL, "user=test password=123456 host=192.168.10.220 port=5432 dbname=postgres sslmode=disable", 25)

	// Oracle
	dbkit.OpenDatabaseWithDBName("oracle", dbkit.Oracle, "oracle://test:123456@192.168.10.44:1521/orcl", 25)

	// SQL Server
	dbkit.OpenDatabaseWithDBName("sqlserver", dbkit.SQLServer, "sqlserver://sa:123456@192.168.10.44:1433?database=test", 25)

	// SQLite
	dbkit.OpenDatabaseWithDBName("sqlite", dbkit.SQLite3, "file:test_multi.db?cache=shared&mode=rwc", 10)

	// 显示已注册的数据库
	printLog("已注册的数据库列表: %v", dbkit.ListDatabases())
	printLog("当前默认数据库: %s", dbkit.GetCurrentDBName())

	defer dbkit.Close()

	// 2. 遍历所有已注册的数据库进行测试
	databases := []string{"mysql", "postgresql", "oracle", "sqlserver", "sqlite"}

	for _, dbName := range databases {
		runFullTest(dbName)
	}

	printLog("\n=== 所有数据库测试完成 ===")
}

// runFullTest 对指定的数据库执行完整的功能测试
func runFullTest(dbName string) {
	printLog("\n" + strings.Repeat("=", 50))
	printLog("开始测试数据库: %s", strings.ToUpper(dbName))
	printLog(strings.Repeat("=", 50))

	// 切换当前数据库上下文
	_ = dbkit.Use(dbName)

	// --- 准备工作：清理并创建测试表 ---
	//prepareTables(dbName)

	// --- 1. dbkit 自带的增删改查 (ActiveRecord 风格) ---
	testActiveRecordCRUD(dbName)

	// --- 2. 基于原生 SQL 语句的增删改查 (带多个参数，测试占位符) ---
	testRawSQLCRUD(dbName)

	// 3. 多表联合查询 (JOIN)
	printLog("\n[测试] 多表联合查询 (JOIN)...")
	testJoinQuery(dbName)

	// --- 4. 测试 dbkit 提供的所有其他函数 ---
	testOtherFunctions(dbName)

	// --- 5. 事务测试 ---
	testTransaction(dbName)
}

// prepareTables 根据不同数据库语法创建测试表
func prepareTables(dbName string) {
	printLog("\n[准备阶段] 创建测试表...")
	db := dbkit.Use(dbName)

	// 先删除旧表（注意顺序，先删从表）
	dropTable(db, "order_items", dbName)
	dropTable(db, "orders", dbName)
	dropTable(db, "products", dbName)

	var createProductsSQL string
	var createOrdersSQL string

	switch dbName {
	case "mysql":
		createProductsSQL = `CREATE TABLE products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			price DECIMAL(10,2)
		)`
		createOrdersSQL = `CREATE TABLE orders (
			id INT AUTO_INCREMENT PRIMARY KEY,
			product_id INT,
			customer_name VARCHAR(100),
			quantity INT,
			order_date DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "postgresql":
		createProductsSQL = `CREATE TABLE products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			price DECIMAL(10,2)
		)`
		createOrdersSQL = `CREATE TABLE orders (
			id SERIAL PRIMARY KEY,
			product_id INT,
			customer_name VARCHAR(100),
			quantity INT,
			order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	case "sqlite":
		createProductsSQL = `CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price REAL
		)`
		createOrdersSQL = `CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id INTEGER,
			customer_name TEXT,
			quantity INTEGER,
			order_date DATETIME DEFAULT CURRENT_TIMESTAMP
		)`
	case "oracle":
		// Oracle DDL: 使用 Sequence + Trigger 方案（兼容性最好）
		// 先创建序列
		db.Exec("CREATE SEQUENCE products_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE")
		db.Exec("CREATE SEQUENCE orders_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE")

		createProductsSQL = `CREATE TABLE products (
			id NUMBER PRIMARY KEY,
			name VARCHAR2(100) NOT NULL,
			price NUMBER(10, 2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
		createOrdersSQL = `CREATE TABLE orders (
			id NUMBER PRIMARY KEY,
			product_id NUMBER NOT NULL,
			customer_name VARCHAR2(100) NOT NULL,
			quantity NUMBER NOT NULL,
			order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	case "sqlserver":
		createProductsSQL = `CREATE TABLE products (
			id INT IDENTITY(1,1) PRIMARY KEY,
			name NVARCHAR(100) NOT NULL,
			price DECIMAL(10,2)
		)`
		createOrdersSQL = `CREATE TABLE orders (
			id INT IDENTITY(1,1) PRIMARY KEY,
			product_id INT,
			customer_name NVARCHAR(100),
			quantity INT,
			order_date DATETIME DEFAULT GETDATE()
		)`
	}

	if _, err := db.Exec(createProductsSQL); err != nil {
		log.Printf("[%s] 创建 products 表失败: %v", dbName, err)
	}
	if _, err := db.Exec(createOrdersSQL); err != nil {
		log.Printf("[%s] 创建 orders 表失败: %v", dbName, err)
	}

	// Oracle 特殊处理：创建触发器以实现自增 ID
	if dbName == "oracle" {
		db.Exec(`
			CREATE OR REPLACE TRIGGER products_bir
			BEFORE INSERT ON products
			FOR EACH ROW
			BEGIN
				IF :NEW.id IS NULL THEN
					SELECT products_seq.NEXTVAL INTO :NEW.id FROM dual;
				END IF;
			END;
		`)
		db.Exec(`
			CREATE OR REPLACE TRIGGER orders_bir
			BEFORE INSERT ON orders
			FOR EACH ROW
			BEGIN
				IF :NEW.id IS NULL THEN
					SELECT orders_seq.NEXTVAL INTO :NEW.id FROM dual;
				END IF;
			END;
		`)
	}

	printLog("✓ 测试表创建完成")
}

func dropTable(db *dbkit.DB, tableName string, dbName string) {
	var sql string
	if dbName == "sqlserver" {
		sql = fmt.Sprintf("IF OBJECT_ID('%s', 'U') IS NOT NULL DROP TABLE %s", tableName, tableName)
		db.Exec(sql)
	} else if dbName == "oracle" {
		// Oracle 不支持 IF EXISTS，使用 PL/SQL 块安全删除表和相关的序列
		db.Exec(fmt.Sprintf("BEGIN EXECUTE IMMEDIATE 'DROP TABLE %s PURGE'; EXCEPTION WHEN OTHERS THEN NULL; END;", tableName))
		db.Exec(fmt.Sprintf("BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE %s_seq'; EXCEPTION WHEN OTHERS THEN NULL; END;", tableName))
	} else {
		sql = fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
		db.Exec(sql)
	}
}

// testActiveRecordCRUD 测试 dbkit 内置的增删改查方法
func testActiveRecordCRUD(dbName string) {
	printLog("\n[测试] dbkit 内置 ActiveRecord 风格操作...")
	db := dbkit.Use(dbName)

	// 1. Save (新增)
	p1 := dbkit.NewRecord().Set("name", "笔记本电脑").Set("price", 5999.00)
	p2 := dbkit.NewRecord().Set("name", "智能手机").Set("price", 3999.00)

	id1, err := db.Insert("products", p1)
	if err != nil {
		log.Printf("  - 新增产品失败: %v", err)
	} else {
		printLog("  - 新增产品成功 (使用 Insert 方法)")
	}
	id2, err := db.Save("products", p2)
	if err != nil {
		log.Printf("  - 新增产品失败: %v", err)
	} else {
		printLog("  - 新增产品成功 (使用 Save 方法)")
	}
	printLog("  - 已插入产品: ID=%d, ID=%d", id1, id2)

	// 2. FindAll (查询所有)
	products, _ := db.FindAll("products")
	printLog("  - 当前产品总数: %d", len(products))

	// 3. Update (更新)
	p1.Set("price", 5499.00)
	rows, _ := db.Update("products", p1, "id = ?", id1)
	printLog("  - 更新产品价格, 影响行数: %d", rows)

	// 4. Delete (删除)
	rows, _ = db.Delete("products", "id = ?", id2)
	printLog("  - 删除 ID=%d 的产品, 影响行数: %d", id2, rows)
}

// testRawSQLCRUD 测试原生 SQL 语句操作，重点测试多参数和占位符转换
func testRawSQLCRUD(dbName string) {
	printLog("\n[测试] 原生 SQL 增删改查 (测试占位符转换)...")
	db := dbkit.Use(dbName)

	// 1. 原生 Insert (多参数)
	_, err := db.Exec("INSERT INTO products (name, price) VALUES (?, ?)", "平板电脑", 2999.50)
	if err != nil {
		log.Printf("  - 原生插入失败: %v", err)
	} else {
		printLog("  - 原生插入成功 (使用 ? 占位符)")
	}

	// 2. 原生 Query (多参数条件查询)
	sql := "SELECT * FROM products WHERE price > ? AND name LIKE ?"
	results, _ := db.Query(sql, 2000, "%电脑%")
	printLog("  - 原生查询结果 (price > 2000 且包含'电脑'): %d 条", len(results))
	for _, r := range results {
		printLog("    - ID: %d, 名称: %s, 价格: %.2f", r.GetInt("id"), r.GetString("name"), r.GetFloat("price"))
	}

	// 3. 原生 Update (多参数)
	_, err = db.Exec("UPDATE products SET price = ? WHERE name = ? AND price < ?", 2899.00, "平板电脑", 3000)
	printLog("  - 原生更新成功")

	// 4. 原生 QueryFirst
	row, _ := db.QueryFirst("SELECT * FROM products WHERE id = ?", 1)
	if row != nil {
		printLog("  - QueryFirst 结果: %s", row.GetString("name"))
	}
}

// testJoinQuery 测试多表联合查询
func testJoinQuery(dbName string) {
	printLog("\n[测试] 多表联合查询...")
	db := dbkit.Use(dbName)

	// 插入一些订单数据
	db.Exec("INSERT INTO orders (product_id, customer_name, quantity) VALUES (?, ?, ?)", 1, "张三", 2)
	db.Exec("INSERT INTO orders (product_id, customer_name, quantity) VALUES (?, ?, ?)", 1, "李四", 1)
	db.Exec("INSERT INTO orders (product_id, customer_name, quantity) VALUES (?, ?, ?)", 3, "王五", 5)

	// 复杂 Join 查询
	joinSQL := `
		SELECT o.id as order_id, o.customer_name, p.name as product_name, p.price, o.quantity, (p.price * o.quantity) as total_amount
		FROM orders o
		JOIN products p ON o.product_id = p.id
		WHERE o.quantity >= ?
		ORDER BY total_amount DESC
	`
	results, _ := db.Query(joinSQL, 1)
	printLog("  - 联合查询结果: %d 条", len(results))
	for _, r := range results {
		printLog("    - 订单#%d: 客户[%s] 购买了 [%s], 单价:%.2f, 数量:%d, 总额:%.2f",
			r.GetInt("order_id"), r.GetString("customer_name"), r.GetString("product_name"),
			r.GetFloat("price"), r.GetInt("quantity"), r.GetFloat("total_amount"))
	}
}

// testOtherFunctions 测试 dbkit 提供的其他常用函数
func testOtherFunctions(dbName string) {
	printLog("\n[测试] dbkit 其他功能函数...")
	db := dbkit.Use(dbName)

	// 1. Count
	count, _ := db.Count("products", "price > ?", 1000)
	printLog("  - Count (价格 > 1000): %d", count)

	// 2. Exists
	exists, _ := db.Exists("products", "name = ?", "笔记本电脑")
	printLog("  - Exists (是否存在'笔记本电脑'): %v", exists)

	// 3. Paginate (分页)
	// 3. Paginate (分页)
	pageResults, total, _ := db.Paginate(1, 2, "SELECT *", "products", "price > ?", "id", 0)
	printLog("  - Paginate: 第1页, 每页2条, 总记录数: %d, 当前页记录数: %d", total, len(pageResults))

	// 4. QueryMap (返回 Map 格式)
	mapResults, _ := db.QueryMap("SELECT id, name FROM products LIMIT 5")
	printLog("  - QueryMap 结果数量: %d", len(mapResults))

	// 5. BatchInsert (批量插入)
	printLog("  - 测试批量插入...")
	batchRecords := []*dbkit.Record{
		dbkit.NewRecord().Set("name", "批量产品A").Set("price", 100.1),
		dbkit.NewRecord().Set("name", "批量产品B").Set("price", 100.2),
		dbkit.NewRecord().Set("name", "批量产品C").Set("price", 100.3),
	}
	batchTotal, err := db.BatchInsertDefault("products", batchRecords)
	if err != nil {
		log.Printf("    - 批量插入失败: %v", err)
	} else {
		printLog("    - 成功批量插入 %d 条记录", batchTotal)
	}
}

// testTransaction 测试事务功能
func testTransaction(dbName string) {
	printLog("\n[测试] 事务功能...")
	db := dbkit.Use(dbName)

	// 1. 正常事务提交
	printLog("  - 1. 测试事务正常提交...")
	err := db.Transaction(func(tx *dbkit.Tx) error {
		// 在事务中插入
		p := dbkit.NewRecord().Set("name", "事务产品1").Set("price", 888.88)
		id, err := tx.Save("products", p)
		if err != nil {
			return err
		}
		printLog("    - 事务内插入产品成功, ID=%d", id)

		// 在事务中更新
		_, err = tx.Exec("UPDATE products SET price = ? WHERE id = ?", 999.99, id)
		return err
	})

	if err == nil {
		printLog("    ✓ 事务提交成功")
	} else {
		printLog("    ✗ 事务失败: %v", err)
	}

	// 2. 事务回滚测试
	printLog("  - 2. 测试事务异常回滚...")
	beforeCount, _ := db.Count("products", "name = ?", "回滚测试产品")

	err = db.Transaction(func(tx *dbkit.Tx) error {
		p := dbkit.NewRecord().Set("name", "回滚测试产品").Set("price", 123.45)
		tx.Save("products", p)

		printLog("    - 事务内已插入记录，即将手动触发错误以回滚...")
		return fmt.Errorf("故意触发的错误")
	})

	afterCount, _ := db.Count("products", "name = ?", "回滚测试产品")
	if beforeCount == afterCount {
		printLog("    ✓ 事务已成功回滚 (记录数未增加)")
	} else {
		printLog("    ✗ 事务回滚失败 (记录数增加了)")
	}
}
