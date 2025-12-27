package main

import (
	"fmt"
	"log"
	"os"

	"dbkit"
)

func main() {
	fmt.Println("=== DBKit SQLite 示例 ===")

	// 删除旧的测试数据库
	os.Remove("test.db")

	// 初始化数据库连接（SQLite）
	fmt.Println("1. 初始化SQLite数据库...")
	dbkit.OpenDatabase(dbkit.SQLite3, "file:test.db?cache=shared&mode=rwc", 10)
	defer func() {
		dbkit.Close()
		os.Remove("test.db")
	}()

	// 检查连接
	if err := dbkit.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("✓ SQLite数据库连接成功")

	// 创建表
	fmt.Println("\n2. 创建示例表...")
	_, err := dbkit.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price REAL DEFAULT 0,
			stock INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
	fmt.Println("✓ 产品表创建成功")

	// 插入产品
	fmt.Println("\n3. 插入产品数据...")

	product1 := dbkit.NewRecord()
	product1.Set("name", "iPhone 15")
	product1.Set("price", 6999.00)
	product1.Set("stock", 100)

	product2 := dbkit.NewRecord()
	product2.Set("name", "MacBook Pro")
	product2.Set("price", 14999.00)
	product2.Set("stock", 50)

	product3 := dbkit.NewRecord()
	product3.Set("name", "iPad Air")
	product3.Set("price", 4799.00)
	product3.Set("stock", 200)

	products := []*dbkit.Record{product1, product2, product3}

	total, err := dbkit.BatchInsertDefault("products", products)
	if err != nil {
		log.Fatalf("批量插入失败: %v", err)
	}
	fmt.Printf("✓ 插入%d个产品\n", total)

	// 查询所有产品
	fmt.Println("\n4. 查询所有产品:")
	allProducts, err := dbkit.Query("SELECT * FROM products")
	if err != nil {
		log.Printf("查询所有产品失败: %v", err)
	} else {
		for _, p := range allProducts {
			fmt.Printf("  - %s: ¥%.2f (库存: %d)\n",
				p.GetString("name"),
				p.GetFloat("price"),
				p.GetInt("stock"))
		}
	}

	// 查询价格大于5000的产品
	fmt.Println("\n5. 查询高价产品（价格 > 5000）:")
	expensiveProducts, err := dbkit.Query("SELECT * FROM products WHERE price > ?", 5000)
	if err != nil {
		log.Printf("查询高价产品失败: %v", err)
	} else {
		for _, p := range expensiveProducts {
			fmt.Printf("  - %s: ¥%.2f\n", p.GetString("name"), p.GetFloat("price"))
		}
	}

	// 更新库存
	fmt.Println("\n6. 更新产品库存...")
	record := dbkit.NewRecord()
	record.Set("stock", 99)
	affected, err := dbkit.Update("products", record, "name = ?", "iPhone 15")
	if err != nil {
		log.Fatalf("更新失败: %v", err)
	}
	fmt.Printf("✓ 更新成功，影响行数: %d\n", affected)

	// 验证更新
	fmt.Println("\n7. 验证更新结果:")
	updatedProduct, err := dbkit.QueryFirst("SELECT * FROM products WHERE name = ?", "iPhone 15")
	if err != nil {
		log.Printf("验证更新结果失败: %v", err)
	} else if updatedProduct != nil {
		fmt.Printf("  iPhone 15 库存: %d\n", updatedProduct.GetInt("stock"))
	}

	// 统计查询
	fmt.Println("\n8. 统计查询...")
	totalCount, err := dbkit.Count("products", "")
	if err != nil {
		log.Printf("统计产品总数失败: %v", err)
	} else {
		fmt.Printf("  产品总数: %d\n", totalCount)
	}

	totalStock, err := dbkit.Count("products", "stock > ?", 100)
	if err != nil {
		log.Printf("统计库存大于100的产品数失败: %v", err)
	} else {
		fmt.Printf("  库存大于100的产品数: %d\n", totalStock)
	}

	// 检查产品是否存在
	fmt.Println("\n9. 检查产品是否存在...")
	exists := dbkit.Exists("products", "name = ?", "iPhone 15")
	fmt.Printf("  iPhone 15是否存在: %v\n", exists)

	exists = dbkit.Exists("products", "name = ?", "不存在的产品")
	fmt.Printf("  不存在的产品是否存在: %v\n", exists)

	// 使用事务
	fmt.Println("\n10. 事务操作示例...")
	tx, err := dbkit.BeginTransaction()
	if err != nil {
		log.Fatalf("开启事务失败: %v", err)
	}

	// 批量更新库存（使用新变量避免类型冲突）
	queriedProducts, err := dbkit.Query("SELECT * FROM products")
	if err != nil {
		tx.Rollback()
		log.Fatalf("查询产品失败: %v", err)
	}
	for _, p := range queriedProducts {
		stock := p.GetInt("stock")
		stock--
		_, err = dbkit.ExecTx(tx, "UPDATE products SET stock = ? WHERE id = ?", stock, p.Get("id"))
		if err != nil {
			tx.Rollback()
			log.Fatalf("事务更新失败: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("提交事务失败: %v", err)
	}
	fmt.Println("✓ 事务更新成功（所有产品库存-1）")

	// 显示更新后的库存
	fmt.Println("\n11. 更新后的产品列表:")
	products = []*dbkit.Record{}
	queriedProducts, err = dbkit.Query("SELECT * FROM products")
	if err != nil {
		log.Printf("查询更新后产品列表失败: %v", err)
	} else {
		for _, p := range queriedProducts {
			rec := p
			products = append(products, &rec)
			fmt.Printf("  - %s: 库存 %d\n", p.GetString("name"), p.GetInt("stock"))
		}
	}

	// 删除产品
	fmt.Println("\n12. 删除产品...")
	deleted, err := dbkit.Delete("products", "name = ?", "iPad Air")
	if err != nil {
		log.Fatalf("删除失败: %v", err)
	}
	fmt.Printf("✓ 删除iPad Air，影响行数: %d\n", deleted)

	// Record的高级用法
	fmt.Println("\n13. Record高级用法示例...")
	product, err := dbkit.QueryFirst("SELECT * FROM products WHERE name = ?", "MacBook Pro")
	if err != nil {
		log.Printf("查询产品失败: %v", err)
	} else if product != nil {
		fmt.Printf("  产品名称: %s\n", product.Str("name"))
		fmt.Printf("  产品价格: %.2f\n", product.Float("price"))
		fmt.Printf("  库存数量: %d\n", product.Int("stock"))
		fmt.Printf("  是否有货: %v\n", product.Bool("stock") && product.GetInt("stock") > 0)
		fmt.Printf("  所有字段: %v\n", product.Keys())
		fmt.Printf("  JSON格式: %s\n", product.ToJson())
	}

	// 分页查询
	fmt.Println("\n14. 分页查询...")
	page1, totalProducts, err := dbkit.Paginate(1, 1, "SELECT *", "products", "", "id ASC")
	if err != nil {
		log.Printf("分页查询失败: %v", err)
	} else {
		fmt.Printf("  第1页，每页1条，总数: %d\n", totalProducts)
		for _, p := range page1 {
			fmt.Printf("    - %s\n", p.GetString("name"))
		}
	}

	fmt.Println("\n=== SQLite 示例完成 ===")
}
