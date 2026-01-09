package main

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
)

// DbModel 代码生成器示例
func main() {
	fmt.Println("\n===========================================")
	fmt.Println("   DBKit DbModel 代码生成器")
	fmt.Println("===========================================")

	// 1. 连接数据库
	fmt.Println("\n[Step 1] 连接 MySQL 数据库...")
	dsn := "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabaseWithDBName("mysql", dbkit.MySQL, dsn, 10)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	defer dbkit.Close()
	fmt.Println("✓ 连接成功")

	// 2. 创建示例表并插入初始数据
	fmt.Println("\n[Step 2] 创建示例表并插入数据...")
	createTables()
	insertSampleData()

	// 3. 生成 DbModel
	fmt.Println("\n[Step 3] 生成 DbModel 代码...")

	// 定义要生成的表和对应的结构体名称
	tables := map[string]string{
		"users":       "User",
		"products":    "Product",
		"orders":      "Order",
		"order_items": "OrderItem",
	}

	// 生成到 models 目录
	outputDir := "../models"

	for tableName, structName := range tables {
		fmt.Printf("  生成 %s -> %s...\n", tableName, structName)
		err := dbkit.GenerateDbModel(tableName, outputDir, structName)
		if err != nil {
			log.Printf("  ⚠️ 生成 %s 失败: %v", tableName, err)
		} else {
			fmt.Printf("  ✓ 生成成功: %s/%s.go\n", outputDir, tableName)
		}
	}

	fmt.Println("\n===========================================")
	fmt.Println("   代码生成完成!")
	fmt.Println("===========================================")
	fmt.Println("\n生成的文件位于: examples2/dbmodel/models/")
	fmt.Println("现在可以运行主程序: cd .. && go run main.go")
}

// createTables 创建示例表
func createTables() {
	// 强制重建表以确保 Schema 正确
	// dbkit.Exec("DROP TABLE IF EXISTS order_items")
	// dbkit.Exec("DROP TABLE IF EXISTS orders")
	// dbkit.Exec("DROP TABLE IF EXISTS products")
	// dbkit.Exec("DROP TABLE IF EXISTS users")

	// 用户表
	usersTable := `
	CREATE TABLE users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID',
		username VARCHAR(100) NOT NULL UNIQUE COMMENT '用户名',
		email VARCHAR(255) NOT NULL COMMENT '邮箱',
		password VARCHAR(255) NOT NULL COMMENT '密码',
		balance DECIMAL(15, 2) DEFAULT 0.00 COMMENT '账户余额',
		status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP not null COMMENT '创建时间',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
		deleted_at DATETIME NULL COMMENT '删除时间(软删除)',
		version INT DEFAULT 1 not null COMMENT '版本号(乐观锁)',
		INDEX idx_username (username),
		INDEX idx_email (email),
		INDEX idx_deleted_at (deleted_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表'
	`

	// 商品表
	productsTable := `
	CREATE TABLE IF NOT EXISTS products (
		id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '商品ID',
		name VARCHAR(200) NOT NULL COMMENT '商品名称',
		description TEXT COMMENT '商品描述',
		price DECIMAL(10, 2) NOT NULL COMMENT '价格',
		stock INT DEFAULT 0 COMMENT '库存数量',
		category VARCHAR(50) COMMENT '分类',
		is_active TINYINT(1) DEFAULT 1 COMMENT '是否上架',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
		deleted_at DATETIME NULL COMMENT '删除时间',
		INDEX idx_category (category),
		INDEX idx_is_active (is_active)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品表'
	`

	// 订单表
	ordersTable := `
	CREATE TABLE IF NOT EXISTS orders (
		id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '订单ID',
		user_id BIGINT NOT NULL COMMENT '用户ID',
		order_no VARCHAR(50) NOT NULL UNIQUE COMMENT '订单号',
		total_amount DECIMAL(15, 2) NOT NULL COMMENT '订单总金额',
		status TINYINT DEFAULT 0 COMMENT '订单状态: 0-待支付, 1-已支付, 2-已发货, 3-已完成, 4-已取消',
		payment_method VARCHAR(20) COMMENT '支付方式',
		shipping_address TEXT COMMENT '收货地址',
		remark TEXT COMMENT '备注',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
		paid_at DATETIME NULL COMMENT '支付时间',
		shipped_at DATETIME NULL COMMENT '发货时间',
		completed_at DATETIME NULL COMMENT '完成时间',
		INDEX idx_user_id (user_id),
		INDEX idx_order_no (order_no),
		INDEX idx_status (status),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表'
	`

	// 订单明细表
	orderItemsTable := `
	CREATE TABLE IF NOT EXISTS order_items (
		id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '订单明细ID',
		order_id BIGINT NOT NULL COMMENT '订单ID',
		product_id BIGINT NOT NULL COMMENT '商品ID',
		product_name VARCHAR(200) NOT NULL COMMENT '商品名称(快照)',
		price DECIMAL(10, 2) NOT NULL COMMENT '单价(快照)',
		quantity INT NOT NULL COMMENT '数量',
		subtotal DECIMAL(15, 2) NOT NULL COMMENT '小计',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
		INDEX idx_order_id (order_id),
		INDEX idx_product_id (product_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细表'
	`

	// 执行建表语句
	tables := map[string]string{
		"users":       usersTable,
		"products":    productsTable,
		"orders":      ordersTable,
		"order_items": orderItemsTable,
	}

	for name, sql := range tables {
		_, err := dbkit.Exec(sql)
		if err != nil {
			// log.Printf("⚠️ 创建表 %s 失败: %v", name, err)
		} else {
			fmt.Printf("  ✓ 创建表: %s\n", name)
		}
	}

	// 插入示例数据
	insertSampleData()
}

// insertSampleData 插入示例数据
func insertSampleData() {
	fmt.Println("  插入示例数据...")

	// 插入用户
	users := []*dbkit.Record{
		dbkit.NewRecord().
			Set("username", "alice").
			Set("email", "alice@example.com").
			Set("password", "hashed_password_1").
			Set("balance", 1000.00),
		dbkit.NewRecord().
			Set("username", "bob").
			Set("email", "bob@example.com").
			Set("password", "hashed_password_2").
			Set("balance", 500.00),
		dbkit.NewRecord().
			Set("username", "charlie").
			Set("email", "charlie@example.com").
			Set("password", "hashed_password_3").
			Set("balance", 2000.00),
	}
	dbkit.BatchInsert("users", users, 10)

	// 插入商品
	products := []*dbkit.Record{
		dbkit.NewRecord().
			Set("name", "iPhone 15 Pro").
			Set("description", "最新款苹果手机").
			Set("price", 7999.00).
			Set("stock", 50).
			Set("category", "电子产品"),
		dbkit.NewRecord().
			Set("name", "MacBook Pro").
			Set("description", "专业笔记本电脑").
			Set("price", 12999.00).
			Set("stock", 30).
			Set("category", "电子产品"),
		dbkit.NewRecord().
			Set("name", "AirPods Pro").
			Set("description", "无线降噪耳机").
			Set("price", 1999.00).
			Set("stock", 100).
			Set("category", "配件"),
	}
	dbkit.BatchInsert("products", products, 10)

	// 插入订单
	orders := []*dbkit.Record{
		dbkit.NewRecord().
			Set("user_id", 1).
			Set("order_no", "ORD20260109001").
			Set("total_amount", 9998.00).
			Set("status", 1).
			Set("payment_method", "支付宝").
			Set("shipping_address", "北京市朝阳区xxx街道xxx号"),
		dbkit.NewRecord().
			Set("user_id", 2).
			Set("order_no", "ORD20260109002").
			Set("total_amount", 1999.00).
			Set("status", 0).
			Set("payment_method", "微信").
			Set("shipping_address", "上海市浦东新区xxx路xxx号"),
	}
	dbkit.BatchInsert("orders", orders, 10)

	// 插入订单明细
	orderItems := []*dbkit.Record{
		dbkit.NewRecord().
			Set("order_id", 1).
			Set("product_id", 1).
			Set("product_name", "iPhone 15 Pro").
			Set("price", 7999.00).
			Set("quantity", 1).
			Set("subtotal", 7999.00),
		dbkit.NewRecord().
			Set("order_id", 1).
			Set("product_id", 3).
			Set("product_name", "AirPods Pro").
			Set("price", 1999.00).
			Set("quantity", 1).
			Set("subtotal", 1999.00),
		dbkit.NewRecord().
			Set("order_id", 2).
			Set("product_id", 3).
			Set("product_name", "AirPods Pro").
			Set("price", 1999.00).
			Set("quantity", 1).
			Set("subtotal", 1999.00),
	}
	dbkit.BatchInsert("order_items", orderItems, 10)

	fmt.Println("  ✓ 示例数据插入完成")
}
