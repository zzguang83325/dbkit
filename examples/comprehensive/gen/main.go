package main

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
)

func main() {
	// 1. 初始化数据库连接 - MySQL
	dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabase(dbkit.MySQL, dsn, 10)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer dbkit.Close()

	// 2. 预先创建表结构 (如果不存在)
	setupTables()

	// 3. 执行生成任务
	// 针对 users 表生成模型
	err = dbkit.GenerateDbModel("users", "../models/users.go", "User")
	if err != nil {
		log.Fatalf("生成 User 模型失败: %v", err)
	}
	log.Println("✓ User 模型生成成功")

	// 针对 orders 表生成模型
	err = dbkit.GenerateDbModel("orders", "../models/orders.go", "Order")
	if err != nil {
		log.Fatalf("生成 Order 模型失败: %v", err)
	}
	log.Println("✓ Order 模型生成成功")

	// 针对 products 表生成模型
	err = dbkit.GenerateDbModel("products", "../models/products.go", "Product")
	if err != nil {
		log.Fatalf("生成 Product 模型失败: %v", err)
	}
	log.Println("✓ Product 模型生成成功")

	// 针对 order_items 表生成模型
	err = dbkit.GenerateDbModel("order_items", "../models/order_items.go", "OrderItem")
	if err != nil {
		log.Fatalf("生成 OrderItem 模型失败: %v", err)
	}
	log.Println("✓ OrderItem 模型生成成功")

	log.Println("\n所有模型生成成功！代码已保存至 examples/comprehensive/models 目录")
}

func setupTables() {
	// 删除旧表 (按依赖顺序)
	dbkit.Exec("DROP TABLE IF EXISTS order_items")
	dbkit.Exec("DROP TABLE IF EXISTS orders")
	dbkit.Exec("DROP TABLE IF EXISTS products")
	dbkit.Exec("DROP TABLE IF EXISTS users")

	// 用户表
	_, err := dbkit.Exec(`CREATE TABLE users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(100) NOT NULL,
		email VARCHAR(100),
		age INT DEFAULT 0,
		status VARCHAR(20) DEFAULT 'active',
		deleted_at DATETIME NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Printf("创建 users 表失败: %v", err)
	}

	// 产品表
	_, err = dbkit.Exec(`CREATE TABLE products (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		price DECIMAL(10,2) DEFAULT 0,
		stock INT DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Printf("创建 products 表失败: %v", err)
	}

	// 订单表
	_, err = dbkit.Exec(`CREATE TABLE orders (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		amount DECIMAL(10,2) DEFAULT 0,
		status VARCHAR(20) DEFAULT 'PENDING',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Printf("创建 orders 表失败: %v", err)
	}

	// 订单项表
	_, err = dbkit.Exec(`CREATE TABLE order_items (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		order_id BIGINT NOT NULL,
		product_id BIGINT NOT NULL,
		quantity INT DEFAULT 1,
		price DECIMAL(10,2) DEFAULT 0
	)`)
	if err != nil {
		log.Printf("创建 order_items 表失败: %v", err)
	}

	log.Println("✓ 表结构创建完成")
}
