package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"dbkit"
)

func main() {
	fmt.Println("=== DBKit PostgreSQL 示例 ===")

	// 初始化日志记录
	logFilePath := filepath.Join(".", "log.log")
	dbkit.InitLoggerWithFile("debug", logFilePath)

	// 初始化数据库连接
	fmt.Println("1. 初始化PostgreSQL数据库...")
	// PostgreSQL 连接字符串格式: user=postgres password=xxx host=localhost port=5432 dbname=test sslmode=disable
	dbkit.OpenDatabase(dbkit.PostgreSQL, "user=test password=123456 host=192.168.10.220 port=5432 dbname=postgres sslmode=disable", 25)
	defer dbkit.Close()

	// 检查连接
	if err := dbkit.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("✓ PostgreSQL数据库连接成功")

	// 创建表
	fmt.Println("\n2. 创建示例表...")
	dbkit.Exec("DROP TABLE IF EXISTS orders")
	_, err := dbkit.Exec(`
		CREATE TABLE orders (
			id SERIAL PRIMARY KEY,
			customer_name VARCHAR(100) NOT NULL,
			product_name VARCHAR(200) NOT NULL,
			quantity INT DEFAULT 1,
			price DECIMAL(10, 2) DEFAULT 0,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("创建表失败（可能已存在）: %v", err)
	} else {
		fmt.Println("✓ 订单表创建成功")
	}

	// 插入订单数据
	fmt.Println("\n3. 插入订单数据...")

	// 插入单个订单
	order1 := dbkit.NewRecord()
	order1.Set("customer_name", 11111)
	order1.Set("product_name", "ThinkPad X1 Carbon")
	order1.Set("quantity", "1")
	order1.Set("price", "9999.00")
	order1.Set("status", "completed")

	id, err := dbkit.Save("orders", order1)
	if err != nil {
		log.Fatalf("插入订单失败: %v", err)
	}
	fmt.Printf("✓ 插入订单成功，ID: %d\n", id)

	// 批量插入订单
	fmt.Println("\n4. 批量插入订单...")

	order2 := dbkit.NewRecord()
	order2.Set("customer_name", "李四")
	order2.Set("product_name", "iPhone 15 Pro")
	order2.Set("quantity", 2)
	order2.Set("price", 7999.00)
	order2.Set("status", "pending")

	order3 := dbkit.NewRecord()
	order3.Set("customer_name", "王五")
	order3.Set("product_name", "MacBook Air M3")
	order3.Set("quantity", 1)
	order3.Set("price", 8999.00)
	order3.Set("status", "shipped")

	order4 := dbkit.NewRecord()
	order4.Set("customer_name", "赵六")
	order4.Set("product_name", "iPad Pro 12.9")
	order4.Set("quantity", 1)
	order4.Set("price", 7999.00)
	order4.Set("status", "completed")

	orders := []*dbkit.Record{order2, order3, order4}

	total, err := dbkit.BatchInsertDefault("orders", orders)
	if err != nil {
		log.Fatalf("批量插入失败: %v", err)
	}
	fmt.Printf("✓ 批量插入成功，插入%d个订单\n", total)

	// 查询所有订单
	fmt.Println("\n5. 查询所有订单:")
	allOrders, err := dbkit.Query("SELECT * FROM orders ORDER BY id")
	if err != nil {
		log.Printf("查询所有订单失败: %v", err)
	} else {
		for _, o := range allOrders {
			fmt.Printf("  - 订单#%d: %s 购买了 %s x%d，金额: ¥%.2f，状态: %s\n",
				o.GetInt("id"),
				o.GetString("customer_name"),
				o.GetString("product_name"),
				o.GetInt("quantity"),
				o.GetFloat("price"),
				o.GetString("status"))
		}
	}

	// 条件查询
	fmt.Println("\n6. 查询已完成的订单:")
	completedOrders, err := dbkit.Query("SELECT * FROM orders WHERE status = ?", "completed")
	if err != nil {
		log.Printf("查询已完成订单失败: %v", err)
	} else {
		for _, o := range completedOrders {
			fmt.Printf("  - %s: %s (¥%.2f)\n",
				o.GetString("customer_name"),
				o.GetString("product_name"),
				o.GetFloat("price"))
		}
	}

	// 查询单价大于8000的订单
	fmt.Println("\n7. 查询高价订单（单价 > 8000）:")
	expensiveOrders, err := dbkit.Query("SELECT * FROM orders WHERE price > ?", 8000)
	if err != nil {
		log.Printf("查询高价订单失败: %v", err)
	} else {
		for _, o := range expensiveOrders {
			fmt.Printf("  - %s 购买了 %s，金额: ¥%.2f\n",
				o.GetString("customer_name"),
				o.GetString("product_name"),
				o.GetFloat("price"))
		}
	}

	// 查询单个订单
	fmt.Println("\n8. 查询单个订单:")
	order, err := dbkit.QueryFirst("SELECT * FROM orders WHERE id = ?", 1)
	if err != nil {
		log.Printf("查询单个订单失败: %v", err)
	} else if order != nil {
		fmt.Printf("  订单详情: %s 购买了 %s x%d，状态: %s\n",
			order.GetString("customer_name"),
			order.GetString("product_name"),
			order.GetInt("quantity"),
			order.GetString("status"))
	}

	// 按ID查询
	fmt.Println("\n9. 按ID查询订单:")
	orderById, err := dbkit.QueryFirst("SELECT * FROM orders WHERE id = ?", 2)
	if err != nil {
		log.Printf("按ID查询订单失败: %v", err)
	} else if orderById != nil {
		fmt.Printf("  订单#2: %s\n", orderById.GetString("product_name"))
	}

	// 更新订单状态
	fmt.Println("\n10. 更新订单状态...")
	orderUpdate := dbkit.NewRecord()
	orderUpdate.Set("status", "delivered")
	affected, err := dbkit.Update("orders", orderUpdate, "id = ?", 2)
	if err != nil {
		log.Fatalf("更新失败: %v", err)
	}
	fmt.Printf("✓ 更新成功，影响行数: %d\n", affected)

	// 统计查询
	fmt.Println("\n11. 统计查询...")
	orderCount, _ := dbkit.Count("orders", "")
	fmt.Printf("  订单总数: %d\n", orderCount)

	completedCount, _ := dbkit.Count("orders", "status = ?", "completed")
	fmt.Printf("  已完成订单数: %d\n", completedCount)

	totalAmount, _ := dbkit.Count("orders", "price > ?", 8000)
	fmt.Printf("  高价订单数（>8000）: %d\n", totalAmount)

	// 分页查询
	fmt.Println("\n14. 分页查询...")
	page := 1
	perPage := 5
	ordersPage, totalOrders, err := dbkit.Paginate(page, perPage, "SELECT *", "orders", "", "id ASC")
	if err != nil {
		log.Printf("分页查询失败: %v", err)
	} else {
		fmt.Printf("  第%d页（每页%d条），总订单数: %d\n", page, perPage, totalOrders)
		for i, o := range ordersPage {
			fmt.Printf("    %d. %s - %s (¥%.2f)\n",
				i+1, o.GetString("customer_name"), o.GetString("product_name"), o.GetFloat("price"))
		}
	}

	// 检查订单是否存在
	fmt.Println("\n13. 检查订单是否存在...")
	exists := dbkit.Exists("orders", "id = ?", 1)
	fmt.Printf("  订单#1是否存在: %v\n", exists)

	exists = dbkit.Exists("orders", "id = ?", 999)
	fmt.Printf("  订单#999是否存在: %v\n", exists)

	// Record的各种获取方法
	fmt.Println("\n14. Record数据获取方法示例...")
	record, err := dbkit.QueryFirst("SELECT * FROM orders WHERE id = ?", 1)
	if err != nil {
		log.Printf("获取订单记录失败: %v", err)
	} else if record != nil {
		fmt.Printf("  客户姓名: %s\n", record.Str("customer_name"))
		fmt.Printf("  产品名称: %s\n", record.Str("product_name"))
		fmt.Printf("  数量: %d\n", record.Int("quantity"))
		fmt.Printf("  单价: %.2f\n", record.Float("price"))
		fmt.Printf("  总价: %.2f\n", record.Float("price")*float64(record.Int("quantity")))
		fmt.Printf("  状态: %s\n", record.Str("status"))
		fmt.Printf("  是否有状态字段: %v\n", record.Has("status"))
		fmt.Printf("  所有字段: %v\n", record.Keys())
		fmt.Printf("  JSON格式: %s\n", record.ToJson())
	}

	// 事务示例
	fmt.Println("\n15. 事务操作示例...")
	tx, err := dbkit.BeginTransaction()
	if err != nil {
		log.Fatalf("开启事务失败: %v", err)
	}

	// 在事务中插入新订单
	newOrder := dbkit.NewRecord()
	newOrder.Set("customer_name", "事务测试用户")
	newOrder.Set("product_name", "测试产品")
	newOrder.Set("quantity", 5)
	newOrder.Set("price", 100.00)
	newOrder.Set("status", "pending")

	txId, err := dbkit.SaveTx(tx, "orders", newOrder)
	if err != nil {
		tx.Rollback()
		log.Fatalf("事务插入失败: %v", err)
	}
	fmt.Printf("  事务中插入订单ID: %d\n", txId)

	// 在事务中更新订单
	orderTxUpdate := dbkit.NewRecord()
	orderTxUpdate.Set("status", "processing")
	_, err = dbkit.UpdateTx(tx, "orders", orderTxUpdate, "id = ?", txId)
	if err != nil {
		tx.Rollback()
		log.Fatalf("事务更新失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Fatalf("提交事务失败: %v", err)
	}
	fmt.Println("✓ 事务提交成功")

	// 验证事务更新
	fmt.Println("\n16. 验证事务更新结果...")
	verifyOrder, err := dbkit.QueryFirst("SELECT * FROM orders WHERE id = ?", txId)
	if err != nil {
		log.Printf("验证事务更新结果失败: %v", err)
	} else if verifyOrder != nil {
		fmt.Printf("  订单#%d 状态: %s\n", txId, verifyOrder.GetString("status"))
	}

	// 使用WithTransaction进行更简洁的事务处理
	fmt.Println("\n17. 使用WithTransaction进行事务处理...")
	err = dbkit.WithTransaction(func(tx *dbkit.Tx) error {
		// 插入订单
		orderA := dbkit.NewRecord()
		orderA.Set("customer_name", "批量事务用户1")
		orderA.Set("product_name", "产品A")
		orderA.Set("price", 500.00)
		orderA.Set("status", "pending")

		_, err := dbkit.SaveTx(tx, "orders", orderA)
		if err != nil {
			return err
		}

		// 插入另一个订单
		orderB := dbkit.NewRecord()
		orderB.Set("customer_name", "批量事务用户2")
		orderB.Set("product_name", "产品B")
		orderB.Set("price", 600.00)
		orderB.Set("status", "pending")

		_, err = dbkit.SaveTx(tx, "orders", orderB)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf("事务处理失败: %v", err)
	}
	fmt.Println("✓ WithTransaction事务处理成功")

	// 删除测试数据
	fmt.Println("\n18. 清理测试数据...")
	dbkit.Delete("orders", "customer_name = ?", "事务测试用户")
	dbkit.Delete("orders", "customer_name LIKE ?", "批量事务用户%")
	fmt.Println("✓ 测试数据清理完成")

	// 查询为Map
	fmt.Println("\n19. 查询结果转换为Map...")
	ordersMap, err := dbkit.QueryMap("SELECT * FROM orders WHERE id = ?", 1)
	if err != nil {
		log.Printf("查询Map失败: %v", err)
	} else if len(ordersMap) > 0 {
		for k, v := range ordersMap[0] {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}

	fmt.Println("\n=== PostgreSQL 示例完成 ===")

	// ========== 复杂 SQL 查询测试 ==========
	fmt.Println("\n========== 复杂 SQL 查询测试 ==========")

	// 创建额外的测试表以支持复杂查询
	fmt.Println("\n20. 创建复杂查询测试表...")
	_, err = dbkit.Exec(`
		CREATE TABLE IF NOT EXISTS customers (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(200),
			phone VARCHAR(20),
			city VARCHAR(50),
			join_date DATE DEFAULT CURRENT_DATE,
			total_orders INT DEFAULT 0,
			total_spent DECIMAL(12, 2) DEFAULT 0
		)
	`)
	if err != nil {
		log.Printf("创建客户表失败: %v", err)
	}

	_, err = dbkit.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			category VARCHAR(50),
			price DECIMAL(10, 2),
			stock INT DEFAULT 0,
			rating DECIMAL(3, 2) DEFAULT 0
		)
	`)
	if err != nil {
		log.Printf("创建产品表失败: %v", err)
	}

	_, err = dbkit.Exec(`
		CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id INT REFERENCES orders(id),
			product_id INT REFERENCES products(id),
			quantity INT DEFAULT 1,
			unit_price DECIMAL(10, 2),
			discount DECIMAL(5, 2) DEFAULT 0
		)
	`)
	if err != nil {
		log.Printf("创建订单明细表失败: %v", err)
	}

	fmt.Println("✓ 测试表创建成功")

	// 插入测试数据
	fmt.Println("\n21. 插入测试数据...")

	// 插入客户数据
	customers := []*dbkit.Record{
		func() *dbkit.Record {
			c := dbkit.NewRecord()
			c.Set("name", "客户A")
			c.Set("email", "customer_a@example.com")
			c.Set("phone", "13800138001")
			c.Set("city", "北京")
			c.Set("total_orders", 5)
			c.Set("total_spent", 15000.00)
			return c
		}(),
		func() *dbkit.Record {
			c := dbkit.NewRecord()
			c.Set("name", "客户B")
			c.Set("email", "customer_b@example.com")
			c.Set("phone", "13800138002")
			c.Set("city", "上海")
			c.Set("total_orders", 3)
			c.Set("total_spent", 8000.00)
			return c
		}(),
		func() *dbkit.Record {
			c := dbkit.NewRecord()
			c.Set("name", "客户C")
			c.Set("email", "customer_c@example.com")
			c.Set("phone", "13800138003")
			c.Set("city", "广州")
			c.Set("total_orders", 8)
			c.Set("total_spent", 25000.00)
			return c
		}(),
	}
	dbkit.BatchInsertDefault("customers", customers)

	// 插入产品数据
	products := []*dbkit.Record{
		func() *dbkit.Record {
			p := dbkit.NewRecord()
			p.Set("name", "笔记本电脑")
			p.Set("category", "电子产品")
			p.Set("price", 5999.00)
			p.Set("stock", 50)
			p.Set("rating", 4.5)
			return p
		}(),
		func() *dbkit.Record {
			p := dbkit.NewRecord()
			p.Set("name", "智能手机")
			p.Set("category", "电子产品")
			p.Set("price", 3999.00)
			p.Set("stock", 100)
			p.Set("rating", 4.8)
			return p
		}(),
		func() *dbkit.Record {
			p := dbkit.NewRecord()
			p.Set("name", "无线耳机")
			p.Set("category", "配件")
			p.Set("price", 299.00)
			p.Set("stock", 200)
			p.Set("rating", 4.2)
			return p
		}(),
	}
	dbkit.BatchInsertDefault("products", products)

	// 插入订单明细数据
	orderItems := []*dbkit.Record{
		func() *dbkit.Record {
			oi := dbkit.NewRecord()
			oi.Set("order_id", 1)
			oi.Set("product_id", 1)
			oi.Set("quantity", 1)
			oi.Set("unit_price", 9999.00)
			oi.Set("discount", 0)
			return oi
		}(),
		func() *dbkit.Record {
			oi := dbkit.NewRecord()
			oi.Set("order_id", 2)
			oi.Set("product_id", 2)
			oi.Set("quantity", 2)
			oi.Set("unit_price", 7999.00)
			oi.Set("discount", 500.00)
			return oi
		}(),
		func() *dbkit.Record {
			oi := dbkit.NewRecord()
			oi.Set("order_id", 3)
			oi.Set("product_id", 1)
			oi.Set("quantity", 1)
			oi.Set("unit_price", 8999.00)
			oi.Set("discount", 0)
			return oi
		}(),
	}
	dbkit.BatchInsertDefault("order_items", orderItems)

	fmt.Println("✓ 测试数据插入完成")

	// ========== WITH 语句测试 ==========
	fmt.Println("\n22. WITH 语句测试 - 简单 CTE...")
	// 使用 WITH 子句计算每个客户的订单总金额
	withResults, err := dbkit.Query(`
		WITH customer_totals AS (
			SELECT 
				customer_name,
				SUM(price * quantity) as total_spent
			FROM orders
			GROUP BY customer_name
		)
		SELECT * FROM customer_totals
		WHERE total_spent > 5000
		ORDER BY total_spent DESC
	`)
	if err != nil {
		log.Printf("WITH语句查询失败: %v", err)
	} else {
		for _, r := range withResults {
			fmt.Printf("  客户: %s, 总消费: ¥%.2f\n",
				r.GetString("customer_name"), r.GetFloat("total_spent"))
		}
	}

	fmt.Println("\n23. WITH 语句测试 - 多个 CTE 链接...")
	// 使用多个 CTE 进行复杂分析
	multiWithResults, err := dbkit.Query(`
	WITH 
	-- 第一个 CTE: 计算每个订单的总金额
	order_totals AS (
		SELECT 
			id as order_id,
			customer_name,
			SUM(price * quantity) as order_amount
		FROM orders
		GROUP BY id, customer_name
	),
	-- 第二个 CTE: 计算每个客户的平均订单金额
	customer_avg AS (
		SELECT 
			customer_name,
			AVG(order_amount) as avg_order_amount,
			COUNT(*) as order_count
		FROM order_totals
		GROUP BY customer_name
	)
	-- 最终查询: 结合两个 CTE 的结果
	SELECT 
		ot.customer_name,
		ot.order_id,
		ot.order_amount,
		ca.avg_order_amount,
		ca.order_count
	FROM order_totals ot
	JOIN customer_avg ca ON ot.customer_name = ca.customer_name
	WHERE ot.order_amount > ca.avg_order_amount
	ORDER BY ot.customer_name, ot.order_amount DESC
`)
	if err != nil {
		log.Printf("多个CTE查询失败: %v", err)
	} else {
		for _, r := range multiWithResults {
			fmt.Printf("  %s 订单#%d: ¥%.2f (平均: ¥%.2f, 订单数: %d)\n",
				r.GetString("customer_name"),
				r.GetInt("order_id"),
				r.GetFloat("order_amount"),
				r.GetFloat("avg_order_amount"),
				r.GetInt("order_count"))
		}
	}

	fmt.Println("\n24. WITH 语句测试 - 递归 CTE...")
	// 使用递归 CTE 生成数字序列
	recursiveResults, err := dbkit.Query(`
	WITH RECURSIVE numbers AS (
		-- 基础查询
		SELECT 1 as n
		UNION ALL
		-- 递归查询
		SELECT n + 1 FROM numbers WHERE n < 10
	)
	SELECT n FROM numbers
	ORDER BY n
`)
	if err != nil {
		log.Printf("递归CTE查询失败: %v", err)
	} else {
		fmt.Print("  数字序列: ")
		for _, r := range recursiveResults {
			fmt.Printf("%d ", r.GetInt("n"))
		}
		fmt.Println()
	}

	fmt.Println("\n25. WITH 语句测试 - CTE 与 JOIN 结合...")
	// 结合 CTE 和多表 JOIN
	joinWithResults, err := dbkit.Query(`
	WITH high_value_orders AS (
		SELECT 
			id,
			customer_name,
			product_name,
			price * quantity as total_value
		FROM orders
		WHERE price * quantity > 10000
	)
	SELECT 
		hvo.id,
		hvo.customer_name,
		hvo.product_name,
		hvo.total_value,
		c.city,
		c.total_spent as customer_lifetime_value
	FROM high_value_orders hvo
	JOIN customers c ON hvo.customer_name = c.name
	ORDER BY hvo.total_value DESC
`)
	if err != nil {
		log.Printf("CTE与JOIN结合查询失败: %v", err)
	} else {
		for _, r := range joinWithResults {
			fmt.Printf("  订单#%d: %s - %s, 订单金额: ¥%.2f, 城市: %s, 客户终身价值: ¥%.2f\n",
				r.GetInt("id"),
				r.GetString("customer_name"),
				r.GetString("product_name"),
				r.GetFloat("total_value"),
				r.GetString("city"),
				r.GetFloat("customer_lifetime_value"))
		}
	}

	// ========== 动态 SQL 测试 ==========
	fmt.Println("\n26. 动态 SQL 测试 - 动态构建 WHERE 子句...")
	// 根据条件动态构建查询
	buildDynamicQuery := func(minPrice float64, status string, customerName string) {
		sql := "SELECT * FROM orders WHERE 1=1"
		var args []interface{}

		if minPrice > 0 {
			sql += " AND price >= ?"
			args = append(args, minPrice)
		}

		if status != "" {
			sql += " AND status = ?"
			args = append(args, status)
		}

		if customerName != "" {
			sql += " AND customer_name LIKE ?"
			args = append(args, "%"+customerName+"%")
		}

		sql += " ORDER BY id"

		fmt.Printf("  动态 SQL: %s\n", sql)
		fmt.Printf("  参数: %v\n", args)

		results, err := dbkit.Query(sql, args...)
		if err != nil {
			log.Printf("动态查询失败: %v", err)
		} else {
			fmt.Printf("  查询结果数: %d\n", len(results))
			for _, r := range results {
				fmt.Printf("    - %s: %s (¥%.2f)\n",
					r.GetString("customer_name"),
					r.GetString("product_name"),
					r.GetFloat("price"))
			}
		}
	}

	buildDynamicQuery(5000, "completed", "")
	buildDynamicQuery(0, "", "张")
	buildDynamicQuery(8000, "pending", "")

	fmt.Println("\n27. 动态 SQL 测试 - 动态选择字段...")
	// 根据需求动态选择查询字段
	buildDynamicFields := func(includePrice bool, includeStatus bool) {
		fields := "id, customer_name, product_name"
		if includePrice {
			fields += ", price"
		}
		if includeStatus {
			fields += ", status"
		}

		sql := fmt.Sprintf("SELECT %s FROM orders ORDER BY id LIMIT 3", fields)
		fmt.Printf("  动态 SQL: %s\n", sql)

		results, err := dbkit.Query(sql)
		if err != nil {
			log.Printf("动态字段查询失败: %v", err)
		} else {
			for _, r := range results {
				fmt.Printf("    - 订单#%d: %s - %s",
					r.GetInt("id"),
					r.GetString("customer_name"),
					r.GetString("product_name"))
				if includePrice {
					fmt.Printf(" (¥%.2f)", r.GetFloat("price"))
				}
				if includeStatus {
					fmt.Printf(" [%s]", r.GetString("status"))
				}
				fmt.Println()
			}
		}
	}

	buildDynamicFields(true, true)
	buildDynamicFields(true, false)
	buildDynamicFields(false, true)

	fmt.Println("\n28. 动态 SQL 测试 - 动态排序...")
	// 根据参数动态排序
	buildDynamicOrder := func(orderBy string, desc bool) {
		sql := "SELECT * FROM orders"
		if desc {
			sql += " ORDER BY " + orderBy + " DESC"
		} else {
			sql += " ORDER BY " + orderBy + " ASC"
		}
		sql += " LIMIT 3"

		fmt.Printf("  动态 SQL: %s\n", sql)

		results, err := dbkit.Query(sql)
		if err != nil {
			log.Printf("动态排序查询失败: %v", err)
		} else {
			for _, r := range results {
				fmt.Printf("    - 订单#%d: %s - %s (¥%.2f)\n",
					r.GetInt("id"),
					r.GetString("customer_name"),
					r.GetString("product_name"),
					r.GetFloat("price"))
			}
		}
	}

	buildDynamicOrder("price", true)
	buildDynamicOrder("price", false)
	buildDynamicOrder("created_at", true)

	// ========== 动态参数测试 ==========
	fmt.Println("\n29. 动态参数测试 - 可变参数数量...")
	// 使用可变参数构建 IN 查询
	buildInQuery := func(statuses []string) {
		if len(statuses) == 0 {
			fmt.Println("  无状态参数")
			return
		}

		// 构建占位符
		placeholders := make([]string, len(statuses))
		for i := range placeholders {
			placeholders[i] = "?"
		}

		// 构建参数列表
		args := make([]interface{}, len(statuses))
		for i, status := range statuses {
			args[i] = status
		}

		sql := fmt.Sprintf("SELECT * FROM orders WHERE status IN (%s) ORDER BY id",
			fmt.Sprintf("%s", placeholders))

		// 修正占位符格式
		sql = fmt.Sprintf("SELECT * FROM orders WHERE status IN (%s) ORDER BY id",
			strings.Join(placeholders, ", "))

		fmt.Printf("  动态 SQL: %s\n", sql)
		fmt.Printf("  参数: %v\n", args)

		results, err := dbkit.Query(sql, args...)
		if err != nil {
			log.Printf("IN查询失败: %v", err)
		} else {
			fmt.Printf("  查询结果数: %d\n", len(results))
			for _, r := range results {
				fmt.Printf("    - %s: %s [%s]\n",
					r.GetString("customer_name"),
					r.GetString("product_name"),
					r.GetString("status"))
			}
		}
	}

	buildInQuery([]string{"completed", "delivered"})
	buildInQuery([]string{"pending"})
	buildInQuery([]string{})

	fmt.Println("\n30. 动态参数测试 - 条件参数...")
	// 根据条件添加参数
	buildConditionalParams := func(minPrice *float64, maxPrice *float64, statuses []string) {
		sql := "SELECT * FROM orders WHERE 1=1"
		var args []interface{}

		if minPrice != nil {
			sql += " AND price >= ?"
			args = append(args, *minPrice)
		}

		if maxPrice != nil {
			sql += " AND price <= ?"
			args = append(args, *maxPrice)
		}

		if len(statuses) > 0 {
			placeholders := make([]string, len(statuses))
			for i := range placeholders {
				placeholders[i] = "?"
			}
			sql += " AND status IN (" + strings.Join(placeholders, ", ") + ")"
			for _, status := range statuses {
				args = append(args, status)
			}
		}

		sql += " ORDER BY price DESC"

		fmt.Printf("  动态 SQL: %s\n", sql)
		fmt.Printf("  参数: %v\n", args)

		results, err := dbkit.Query(sql, args...)
		if err != nil {
			log.Printf("条件参数查询失败: %v", err)
		} else {
			fmt.Printf("  查询结果数: %d\n", len(results))
			for _, r := range results {
				fmt.Printf("    - %s: %s (¥%.2f) [%s]\n",
					r.GetString("customer_name"),
					r.GetString("product_name"),
					r.GetFloat("price"),
					r.GetString("status"))
			}
		}
	}

	minPrice := 5000.0
	maxPrice := 10000.0
	buildConditionalParams(&minPrice, &maxPrice, []string{"completed", "pending"})
	buildConditionalParams(nil, &maxPrice, []string{"completed"})
	buildConditionalParams(&minPrice, nil, []string{})

	fmt.Println("\n31. 动态参数测试 - 批量插入动态参数...")
	// 使用动态参数批量插入
	dynamicBatchInsert := func(orders []map[string]interface{}) error {
		if len(orders) == 0 {
			return nil
		}

		// 获取所有字段
		fields := make([]string, 0)
		for k := range orders[0] {
			fields = append(fields, k)
		}

		// 构建占位符
		placeholders := make([]string, len(fields))
		for i := range placeholders {
			placeholders[i] = "?"
		}

		// 构建批量插入 SQL
		valueGroups := make([]string, len(orders))
		for i := range valueGroups {
			valueGroups[i] = "(" + strings.Join(placeholders, ", ") + ")"
		}

		sql := fmt.Sprintf("INSERT INTO orders (%s) VALUES %s",
			strings.Join(fields, ", "),
			strings.Join(valueGroups, ", "))

		// 构建参数列表
		var args []interface{}
		for _, order := range orders {
			for _, field := range fields {
				args = append(args, order[field])
			}
		}

		fmt.Printf("  批量插入 SQL: %s\n", sql)
		fmt.Printf("  参数数量: %d\n", len(args))

		_, err := dbkit.Exec(sql, args...)
		return err
	}

	// 测试批量插入
	newOrders := []map[string]interface{}{
		{
			"customer_name": "批量客户1",
			"product_name":  "产品X",
			"quantity":      1,
			"price":         1000.00,
			"status":        "pending",
		},
		{
			"customer_name": "批量客户2",
			"product_name":  "产品Y",
			"quantity":      2,
			"price":         2000.00,
			"status":        "completed",
		},
	}

	err = dynamicBatchInsert(newOrders)
	if err != nil {
		log.Printf("批量插入失败: %v", err)
	} else {
		fmt.Println("  批量插入成功")
	}

	// ========== 复杂组合查询测试 ==========
	fmt.Println("\n32. 复杂组合查询 - WITH + 动态参数...")
	// 结合 WITH 子句和动态参数
	complexQuery := func(minTotal float64, city string) {
		sql := `
			WITH customer_order_stats AS (
				SELECT 
					c.name,
					c.city,
					COUNT(o.id) as order_count,
					SUM(o.price * o.quantity) as total_spent
				FROM customers c
				LEFT JOIN orders o ON c.name = o.customer_name
				GROUP BY c.id, c.name, c.city
			)
			SELECT * FROM customer_order_stats
			WHERE 1=1
		`

		var args []interface{}

		if minTotal > 0 {
			sql += " AND total_spent >= ?"
			args = append(args, minTotal)
		}

		if city != "" {
			sql += " AND city = ?"
			args = append(args, city)
		}

		sql += " ORDER BY total_spent DESC"

		fmt.Printf("  复杂查询 SQL: %s\n", sql)
		fmt.Printf("  参数: %v\n", args)

		results, err := dbkit.Query(sql, args...)
		if err != nil {
			log.Printf("复杂查询失败: %v", err)
		} else {
			for _, r := range results {
				fmt.Printf("    - %s (%s): 订单数=%d, 总消费=¥%.2f\n",
					r.GetString("name"),
					r.GetString("city"),
					r.GetInt("order_count"),
					r.GetFloat("total_spent"))
			}
		}
	}

	complexQuery(10000, "")
	complexQuery(0, "北京")

	// ========== 清理测试数据 ==========
	fmt.Println("\n33. 清理复杂查询测试数据...")
	dbkit.Exec("DROP TABLE IF EXISTS order_items")
	dbkit.Exec("DROP TABLE IF EXISTS products")
	dbkit.Exec("DROP TABLE IF EXISTS customers")
	dbkit.Delete("orders", "customer_name LIKE ?", "批量客户%")
	fmt.Println("✓ 测试数据清理完成")

	fmt.Println("\n========== 复杂 SQL 查询测试完成 ==========")

	// 调用合并后的示例函数
	selectExamples2()
	procedureExamples()
}

func selectExamples2() {
	fmt.Println("========== SELECT 语句动态参数使用方法 ==========\n")

	// 方法1: Query - 直接传递可变参数（最常用）
	fmt.Println("方法1: Query - 直接传递可变参数")
	results1, err := dbkit.Query(
		"SELECT * FROM orders WHERE price >= ? AND status = ?",
		5000.0,
		"completed",
	)
	if err != nil {
		log.Printf("直接传递可变参数查询失败: %v", err)
	} else {
		fmt.Printf("  查询结果数: %d\n", len(results1))
		if len(results1) > 0 {
			fmt.Printf("  首条记录: %s - %s\n\n",
				results1[0].GetString("customer_name"),
				results1[0].GetString("product_name"))
		}
	}

	// 方法2: Query - 使用 []interface{} 切片（动态构建参数）
	fmt.Println("方法2: Query - 使用 []interface{} 切片")
	var args2 []interface{}
	args2 = append(args2, 5000.0)
	args2 = append(args2, "completed")
	results2, err := dbkit.Query(
		"SELECT * FROM orders WHERE price >= ? AND status = ?",
		args2...,
	)
	if err != nil {
		log.Printf("使用[]interface{}切片查询失败: %v", err)
	} else {
		fmt.Printf("  查询结果数: %d\n\n", len(results2))
	}

	// 方法3: Query - 动态构建 IN 查询
	fmt.Println("方法3: Query - 动态构建 IN 查询")
	statuses := []string{"completed", "delivered", "pending"}
	placeholders := make([]string, len(statuses))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	sql3 := fmt.Sprintf(
		"SELECT * FROM orders WHERE status IN (%s) ORDER BY id",
		strings.Join(placeholders, ", "),
	)
	args3 := make([]interface{}, len(statuses))
	for i, status := range statuses {
		args3[i] = status
	}
	results3, err := dbkit.Query(sql3, args3...)
	if err != nil {
		log.Printf("动态构建IN查询失败: %v", err)
	} else {
		fmt.Printf("  SQL: %s\n", sql3)
		fmt.Printf("  参数: %v\n", args3)
		fmt.Printf("  查询结果数: %d\n\n", len(results3))
	}

	// 方法4: Query - 条件参数（根据条件动态添加）
	fmt.Println("方法4: Query - 条件参数（动态WHERE）")
	sql4 := "SELECT * FROM orders WHERE 1=1"
	var args4 []interface{}

	minPrice := 5000.0
	maxPrice := 10000.0
	customerName := "张三"

	sql4 += " AND price >= ?"
	args4 = append(args4, minPrice)

	sql4 += " AND price <= ?"
	args4 = append(args4, maxPrice)

	sql4 += " AND customer_name = ?"
	args4 = append(args4, customerName)

	sql4 += " ORDER BY price DESC"

	results4, err := dbkit.Query(sql4, args4...)
	if err != nil {
		log.Printf("条件参数查询失败: %v", err)
	} else {
		fmt.Printf("  SQL: %s\n", sql4)
		fmt.Printf("  参数: %v\n", args4)
		fmt.Printf("  查询结果数: %d\n\n", len(results4))
	}

	// 方法5: QueryFirst - 获取单条记录
	fmt.Println("方法5: QueryFirst - 获取单条记录")
	firstOrder, err := dbkit.QueryFirst(
		"SELECT * FROM orders WHERE customer_name = ? ORDER BY id LIMIT 1",
		"张三",
	)
	if err != nil {
		log.Printf("获取单条记录失败: %v", err)
	} else if firstOrder != nil {
		fmt.Printf("  首条订单: %s - %s (¥%.2f)\n\n",
			firstOrder.GetString("customer_name"),
			firstOrder.GetString("product_name"),
			firstOrder.GetFloat("price"))
	}

	// 方法6: QueryMap - 返回 map 格式
	fmt.Println("方法6: QueryMap - 返回 map 格式")
	results6, err := dbkit.QueryMap(
		"SELECT id, customer_name, product_name, price FROM orders WHERE price >= ? LIMIT 3",
		5000.0,
	)
	if err != nil {
		log.Printf("返回map格式查询失败: %v", err)
	} else {
		for i, r := range results6 {
			fmt.Printf("  订单%d: %v\n", i+1, r)
		}
		fmt.Println()
	}

	// 方法7: QueryFirst - 根据ID查询
	fmt.Println("方法7: QueryFirst - 根据ID查询")
	order, err := dbkit.QueryFirst("SELECT * FROM orders WHERE id = ?", 1)
	if err != nil {
		log.Printf("根据ID查询订单失败: %v", err)
	} else if order != nil {
		fmt.Printf("  ID=1的订单: %s - %s (¥%.2f)\n\n",
			order.GetString("customer_name"),
			order.GetString("product_name"),
			order.GetFloat("price"))
	}

	// 方法8: FindAll - 查询所有记录（简化方法）
	fmt.Println("方法8: FindAll - 查询所有记录")
	allOrders, _ := dbkit.FindAll("orders")
	fmt.Printf("  总订单数: %d\n\n", len(allOrders))

	// 方法9: Count - 统计记录数
	fmt.Println("方法9: Count - 统计记录数")
	count, err := dbkit.Count(
		"orders",
		"price >= ? AND status = ?",
		5000.0,
		"completed",
	)
	if err != nil {
		log.Printf("统计失败: %v", err)
	} else {
		fmt.Printf("  符合条件的记录数: %d\n\n", count)
	}

	// 方法10: Exists - 检查记录是否存在
	fmt.Println("方法10: Exists - 检查记录是否存在")
	exists := dbkit.Exists(
		"orders",
		"customer_name = ? AND status = ?",
		"张三",
		"completed",
	)
	fmt.Printf("  记录是否存在: %v\n\n", exists)

	// 方法11: Paginate - 分页查询
	fmt.Println("方法11: Paginate - 分页查询")
	orders, total, _ := dbkit.Paginate(
		1, 2, "SELECT *", "orders", "price >= ?", "id DESC", 5000.0,
	)
	fmt.Printf("  总数: %d, 当前页数: %d\n", total, len(orders))
	if len(orders) > 0 {
		fmt.Printf("  首条记录: %s - %s\n\n",
			orders[0].GetString("customer_name"),
			orders[0].GetString("product_name"))
	}

	// 方法12: 动态排序
	fmt.Println("方法12: 动态排序")
	orderBy := "price DESC"
	sql12 := fmt.Sprintf("SELECT * FROM orders WHERE price >= ? ORDER BY %s LIMIT 5", orderBy)
	results12, _ := dbkit.Query(sql12, 5000.0)
	fmt.Printf("  SQL: %s\n", sql12)
	fmt.Printf("  查询结果数: %d\n", len(results12))
	for _, r := range results12 {
		fmt.Printf("    - %s: %s (¥%.2f)\n",
			r.GetString("customer_name"),
			r.GetString("product_name"),
			r.GetFloat("price"))
	}
	fmt.Println()

	// 方法13: 动态选择字段
	fmt.Println("方法13: 动态选择字段")
	fields := []string{"id", "customer_name", "product_name", "price"}
	sql13 := fmt.Sprintf("SELECT %s FROM orders WHERE price >= ? LIMIT 3", strings.Join(fields, ", "))
	results13, _ := dbkit.Query(sql13, 5000.0)
	fmt.Printf("  SQL: %s\n", sql13)
	fmt.Printf("  查询结果数: %d\n", len(results13))
	for _, r := range results13 {
		fmt.Printf("    - ID=%d, %s: %s (¥%.2f)\n",
			r.GetInt("id"),
			r.GetString("customer_name"),
			r.GetString("product_name"),
			r.GetFloat("price"))
	}
	fmt.Println()

	// 方法14: 使用事务中的 Query
	fmt.Println("方法14: 使用事务中的 Query")
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		// 在事务中执行查询
		orders, _ := tx.Query(
			"SELECT * FROM orders WHERE price >= ? AND status = ?",
			5000.0,
			"completed",
		)
		fmt.Printf("  事务内查询结果数: %d\n", len(orders))

		// 在事务中执行 QueryFirst
		first, _ := tx.QueryFirst(
			"SELECT * FROM orders WHERE customer_name = ? ORDER BY id LIMIT 1",
			"张三",
		)
		if first != nil {
			fmt.Printf("  事务内首条订单: %s - %s\n",
				first.GetString("customer_name"),
				first.GetString("product_name"))
		}
		return nil
	})

	if err != nil {
		log.Printf("事务执行失败: %v", err)
	}
	fmt.Println()

	// 方法15: 复杂的动态条件组合
	fmt.Println("方法15: 复杂的动态条件组合")
	type QueryParams struct {
		MinPrice      *float64
		MaxPrice      *float64
		CustomerNames []string
		Statuses      []string
		OrderBy       string
		Limit         int
	}

	params := QueryParams{
		MinPrice:      func() *float64 { v := 5000.0; return &v }(),
		MaxPrice:      nil,
		CustomerNames: []string{"张三", "李四"},
		Statuses:      []string{"completed", "delivered"},
		OrderBy:       "price DESC",
		Limit:         10,
	}

	sql15 := "SELECT * FROM orders WHERE 1=1"
	var args15 []interface{}

	if params.MinPrice != nil {
		sql15 += " AND price >= ?"
		args15 = append(args15, *params.MinPrice)
	}

	if params.MaxPrice != nil {
		sql15 += " AND price <= ?"
		args15 = append(args15, *params.MaxPrice)
	}

	if len(params.CustomerNames) > 0 {
		placeholders := make([]string, len(params.CustomerNames))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		sql15 += " AND customer_name IN (" + strings.Join(placeholders, ", ") + ")"
		for _, name := range params.CustomerNames {
			args15 = append(args15, name)
		}
	}

	if len(params.Statuses) > 0 {
		placeholders := make([]string, len(params.Statuses))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		sql15 += " AND status IN (" + strings.Join(placeholders, ", ") + ")"
		for _, status := range params.Statuses {
			args15 = append(args15, status)
		}
	}

	sql15 += fmt.Sprintf(" ORDER BY %s", params.OrderBy)
	if params.Limit > 0 {
		sql15 += fmt.Sprintf(" LIMIT %d", params.Limit)
	}

	results15, _ := dbkit.Query(sql15, args15...)
	fmt.Printf("  SQL: %s\n", sql15)
	fmt.Printf("  参数: %v\n", args15)
	fmt.Printf("  查询结果数: %d\n", len(results15))

	fmt.Println("\n========== 所有 SELECT 方法演示完成 ==========")
}

func procedureExamples() {
	fmt.Println("========== 存储过程调用示例 ==========\n")

	// 检查连接
	if err := dbkit.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("✓ PostgreSQL数据库连接成功\n")

	// 1. 创建存储过程（PostgreSQL）
	fmt.Println("1. 创建存储过程...")

	// 创建一个简单的存储过程：根据客户名称查询订单
	_, err := dbkit.Exec(`
		CREATE OR REPLACE FUNCTION get_orders_by_customer(p_customer_name VARCHAR)
		RETURNS TABLE (
			id INT,
			customer_name VARCHAR,
			product_name VARCHAR,
			quantity INT,
			price DECIMAL,
			status VARCHAR
		) AS $$
		BEGIN
			RETURN QUERY
			SELECT o.id, o.customer_name, o.product_name, o.quantity, o.price, o.status
			FROM orders o
			WHERE o.customer_name = p_customer_name
			ORDER BY o.id;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		log.Printf("创建存储过程失败: %v", err)
	} else {
		fmt.Println("✓ 存储过程 get_orders_by_customer 创建成功")
	}

	// 创建一个带OUT参数的存储过程：计算订单总数和总金额
	_, err = dbkit.Exec(`
		CREATE OR REPLACE FUNCTION calculate_order_stats(
			p_status VARCHAR,
			OUT total_count INT,
			OUT total_amount DECIMAL
		) AS $$
		BEGIN
			SELECT COUNT(*), COALESCE(SUM(price * quantity), 0)
			INTO total_count, total_amount
			FROM orders
			WHERE status = p_status;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		log.Printf("创建存储过程失败: %v", err)
	} else {
		fmt.Println("✓ 存储过程 calculate_order_stats 创建成功")
	}

	// 创建一个带INOUT参数的存储过程：更新订单状态并返回旧状态
	_, err = dbkit.Exec(`
		CREATE OR REPLACE FUNCTION update_order_status(
			p_order_id INT,
			INOUT p_old_status VARCHAR,
			p_new_status VARCHAR
		) AS $$
		DECLARE
			v_old_status VARCHAR;
		BEGIN
			SELECT status INTO v_old_status FROM orders WHERE id = p_order_id;
			
			UPDATE orders SET status = p_new_status WHERE id = p_order_id;
			
			p_old_status := v_old_status;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		log.Printf("创建存储过程失败: %v", err)
	} else {
		fmt.Println("✓ 存储过程 update_order_status 创建成功")
	}

	// 2. 调用返回结果集的存储过程（使用 Query）
	fmt.Println("\n2. 调用返回结果集的存储过程...")

	results, _ := dbkit.Query("SELECT * FROM get_orders_by_customer(?)", "张三")
	fmt.Printf("  查询结果数: %d\n", len(results))
	for _, r := range results {
		fmt.Printf("    - 订单#%d: %s - %s (¥%.2f) [%s]\n",
			r.GetInt("id"),
			r.GetString("customer_name"),
			r.GetString("product_name"),
			r.GetFloat("price"),
			r.GetString("status"))
	}

	// 3. 调用返回单条记录的存储过程（使用 QueryFirst）
	fmt.Println("\n3. 调用返回单条记录的存储过程...")

	firstOrder, _ := dbkit.QueryFirst("SELECT * FROM get_orders_by_customer(?) LIMIT 1", "张三")
	if firstOrder != nil {
		fmt.Printf("  首条订单: %s - %s (¥%.2f)\n",
			firstOrder.GetString("customer_name"),
			firstOrder.GetString("product_name"),
			firstOrder.GetFloat("price"))
	}

	// 4. 调用带OUT参数的存储过程（使用 Query）
	fmt.Println("\n4. 调用带OUT参数的存储过程...")

	// PostgreSQL中，带OUT参数的函数可以直接用SELECT调用
	outResults, _ := dbkit.Query("SELECT * FROM calculate_order_stats(?)", "completed")
	if len(outResults) > 0 {
		fmt.Printf("  订单总数: %d\n", outResults[0].GetInt("total_count"))
		fmt.Printf("  总金额: ¥%.2f\n", outResults[0].GetFloat("total_amount"))
	}

	// 5. 调用带INOUT参数的存储过程
	fmt.Println("\n5. 调用带INOUT参数的存储过程...")

	// 先获取一个存在的订单ID
	testOrder, _ := dbkit.QueryFirst("SELECT id, status FROM orders WHERE customer_name = ? LIMIT 1", "张三")
	if testOrder != nil {
		orderId := testOrder.GetInt("id")
		oldStatus := testOrder.GetString("status")
		newStatus := "delivered"

		fmt.Printf("  更新前: 订单#%d, 状态: %s\n", orderId, oldStatus)

		// PostgreSQL中调用INOUT函数
		inoutResults, _ := dbkit.Query(
			"SELECT * FROM update_order_status(?, ?, ?)",
			orderId,
			oldStatus,
			newStatus,
		)

		if len(inoutResults) > 0 {
			fmt.Printf("  更新成功! 旧状态为: %s, 新状态已更新为: %s\n",
				inoutResults[0].GetString("p_old_status"),
				newStatus)
		}
	}

	// 6. 在事务中调用存储过程
	fmt.Println("\n6. 在事务中调用存储过程...")

	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		txResults, _ := tx.Query("SELECT * FROM get_orders_by_customer(?)", "李四")
		fmt.Printf("  事务内查询结果数: %d\n", len(txResults))

		// 在事务中执行存储过程
		_, err := tx.Exec("SELECT calculate_order_stats(?)", "pending")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("事务执行失败: %v", err)
	} else {
		fmt.Println("  ✓ 事务执行成功")
	}

	// 7. 使用存储过程进行批量操作
	fmt.Println("\n7. 使用存储过程进行批量操作...")

	// 创建一个批量更新状态的存储过程
	_, err = dbkit.Exec(`
		CREATE OR REPLACE FUNCTION batch_update_order_status(
			p_min_price DECIMAL,
			p_old_status VARCHAR,
			p_new_status VARCHAR
		) RETURNS INT AS $$
		DECLARE
			v_updated_count INT;
		BEGIN
			UPDATE orders
			SET status = p_new_status
			WHERE price >= p_min_price AND status = p_old_status;
			
			GET DIAGNOSTICS v_updated_count = ROW_COUNT;
			RETURN v_updated_count;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		log.Printf("创建存储过程失败: %v", err)
	} else {
		fmt.Println("✓ 存储过程 batch_update_order_status 创建成功")

		// 调用批量更新存储过程
		updateResults, _ := dbkit.Query("SELECT batch_update_order_status(?, ?, ?) AS updated_count",
			5000.0, "pending", "processing")
		if len(updateResults) > 0 {
			fmt.Printf("  批量更新影响行数: %d\n", updateResults[0].GetInt("updated_count"))
		}

		// 恢复状态
		dbkit.Exec("UPDATE orders SET status = 'pending' WHERE status = 'processing' AND price >= 5000")
	}

	// 8. 复杂的存储过程调用（带多个参数）
	fmt.Println("\n8. 复杂的存储过程调用...")

	// 创建一个复杂的存储过程：根据多个条件查询订单
	_, err = dbkit.Exec(`
		CREATE OR REPLACE FUNCTION search_orders(
			p_customer_name VARCHAR DEFAULT NULL,
			p_min_price DECIMAL DEFAULT NULL,
			p_max_price DECIMAL DEFAULT NULL,
			p_status VARCHAR DEFAULT NULL
		) RETURNS TABLE (
			id INT,
			customer_name VARCHAR,
			product_name VARCHAR,
			price DECIMAL,
			status VARCHAR
		) AS $$
		BEGIN
			RETURN QUERY
			SELECT o.id, o.customer_name, o.product_name, o.price, o.status
			FROM orders o
			WHERE (p_customer_name IS NULL OR o.customer_name = p_customer_name)
			  AND (p_min_price IS NULL OR o.price >= p_min_price)
			  AND (p_max_price IS NULL OR o.price <= p_max_price)
			  AND (p_status IS NULL OR o.status = p_status)
			ORDER BY o.id;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		log.Printf("创建存储过程失败: %v", err)
	} else {
		fmt.Println("✓ 存储过程 search_orders 创建成功")

		// 调用复杂存储过程（使用动态参数）
		customerName := "张三"
		minPrice := 5000.0
		status := "completed"

		complexResults, _ := dbkit.Query(
			"SELECT * FROM search_orders(?, ?, ?, ?)",
			customerName,
			minPrice,
			nil, // maxPrice 为 NULL
			status,
		)
		fmt.Printf("  查询结果数: %d\n", len(complexResults))
		for _, r := range complexResults {
			fmt.Printf("    - %s: %s (¥%.2f) [%s]\n",
				r.GetString("customer_name"),
				r.GetString("product_name"),
				r.GetFloat("price"),
				r.GetString("status"))
		}
	}

	// 9. 清理存储过程
	fmt.Println("\n9. 清理存储过程...")

	dbkit.Exec("DROP FUNCTION IF EXISTS get_orders_by_customer(VARCHAR)")
	dbkit.Exec("DROP FUNCTION IF EXISTS calculate_order_stats(VARCHAR)")
	dbkit.Exec("DROP FUNCTION IF EXISTS update_order_status(INT, VARCHAR, VARCHAR)")
	dbkit.Exec("DROP FUNCTION IF EXISTS batch_update_order_status(DECIMAL, VARCHAR, VARCHAR)")
	dbkit.Exec("DROP FUNCTION IF EXISTS search_orders(VARCHAR, DECIMAL, DECIMAL, VARCHAR)")

	fmt.Println("✓ 存储过程清理完成")

	fmt.Println("\n========== 存储过程调用示例完成 ==========")
}
