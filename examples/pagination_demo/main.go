package main

import (
	"fmt"
	"log"
	"pagination_demo/models"
	"time"

	"github.com/zzguang83325/dbkit"
	_ "github.com/zzguang83325/dbkit/drivers/mysql"
)

func main() {
	fmt.Println("🚀 分页函数测试示例 - MySQL 数据库")
	fmt.Println("=====================================")

	dbkit.InitLogger("debug")
	// 1. 连接 MySQL 数据库
	// 注意：请根据你的实际 MySQL 配置修改连接字符串
	dsn := "root:123456@tcp(localhost:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabaseWithDBName("mysql", dbkit.MySQL, dsn, 10)
	if err != nil {
		log.Printf("⚠️  MySQL 数据库连接失败: %v", err)
		log.Println("💡 请确保 MySQL 服务正在运行，并修改 main.go 中的数据库连接字符串")
		log.Println("💡 连接字符串格式: username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local")
		return
	}
	fmt.Println("✅ MySQL 数据库连接成功")

	// 2. 创建测试表和数据
	if err := setupTestData(); err != nil {
		log.Fatalf("❌ 设置测试数据失败: %v", err)
	}
	fmt.Println("✅ 测试数据准备完成")

	// 3. 演示各种分页功能
	fmt.Println("\n📊 开始分页功能演示...")

	// 演示1: 基本的 Paginate 用法
	demonstrateBasicPaginate()

	// 演示2: 传统 Paginate 方法对比
	demonstrateTraditionalPaginate()

	// 演示3: 复杂查询分页
	demonstrateComplexQuery()

	// 演示4: 带缓存的分页
	demonstrateCachedPagination()

	// 演示5: 全局分页函数
	demonstrateGlobalPaginate()

	// 演示6: JOIN 查询分页
	demonstrateJoinQueries()

	// 演示7: 子查询分页
	demonstrateSubqueries()

	// 演示8: 窗口函数和高级聚合
	demonstrateWindowFunctions()

	// 演示9: 复杂的多表关联查询
	demonstrateComplexJoins()

	fmt.Println("\n🎉 所有分页功能演示完成！")
}

// setupTestData 创建测试表和插入测试数据
func setupTestData() error {
	db := dbkit.Use("mysql")

	// 创建用户表
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS pagination_demo_users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(150) NOT NULL,
		age INT NOT NULL,
		status VARCHAR(20) DEFAULT 'active',
		department_id BIGINT,
		salary DECIMAL(10,2) DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_age (age),
		INDEX idx_status (status),
		INDEX idx_department_id (department_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

	// 创建部门表
	createDepartmentsTableSQL := `
	CREATE TABLE IF NOT EXISTS pagination_demo_departments (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		budget DECIMAL(12,2) DEFAULT 0,
		manager_id BIGINT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_manager_id (manager_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

	// 创建订单表
	createOrdersTableSQL := `
	CREATE TABLE IF NOT EXISTS pagination_demo_orders (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		amount DECIMAL(10,2) NOT NULL,
		status VARCHAR(20) DEFAULT 'pending',
		order_date DATE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_user_id (user_id),
		INDEX idx_status (status),
		INDEX idx_order_date (order_date)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

	// 创建产品表
	createProductsTableSQL := `
	CREATE TABLE IF NOT EXISTS pagination_demo_products (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(200) NOT NULL,
		category VARCHAR(50) NOT NULL,
		price DECIMAL(8,2) NOT NULL,
		stock INT DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_category (category),
		INDEX idx_price (price)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

	// 创建订单项表
	createOrderItemsTableSQL := `
	CREATE TABLE IF NOT EXISTS pagination_demo_order_items (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		order_id BIGINT NOT NULL,
		product_id BIGINT NOT NULL,
		quantity INT NOT NULL,
		unit_price DECIMAL(8,2) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_order_id (order_id),
		INDEX idx_product_id (product_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

	// 先删除现有表（按依赖关系顺序）
	dropTables := []string{
		"pagination_demo_order_items",
		"pagination_demo_orders",
		"pagination_demo_products",
		"pagination_demo_users",
		"pagination_demo_departments",
	}

	for _, table := range dropTables {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		if err != nil {
			return fmt.Errorf("删除表 %s 失败: %v", table, err)
		}
	}

	// 执行表创建
	tables := []string{
		createUsersTableSQL,
		createDepartmentsTableSQL,
		createOrdersTableSQL,
		createProductsTableSQL,
		createOrderItemsTableSQL,
	}

	for _, sql := range tables {
		_, err := db.Exec(sql)
		if err != nil {
			return fmt.Errorf("创建表失败: %v", err)
		}
	}

	// 清空现有数据
	clearTables := []string{
		"pagination_demo_order_items",
		"pagination_demo_orders",
		"pagination_demo_products",
		"pagination_demo_users",
		"pagination_demo_departments",
	}

	for _, table := range clearTables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("清空表 %s 失败: %v", table, err)
		}
	}

	// 插入部门数据
	departments := []struct {
		name   string
		budget float64
	}{
		{"技术部", 500000.00},
		{"销售部", 300000.00},
		{"市场部", 200000.00},
		{"人事部", 150000.00},
		{"财务部", 180000.00},
	}

	for _, dept := range departments {
		_, err := db.Exec("INSERT INTO pagination_demo_departments (name, budget) VALUES (?, ?)",
			dept.name, dept.budget)
		if err != nil {
			return fmt.Errorf("插入部门数据失败: %v", err)
		}
	}

	// 插入用户数据
	statuses := []string{"active", "inactive", "pending"}
	for i := 1; i <= 50; i++ {
		status := statuses[i%3]
		age := 20 + (i % 40)                   // 年龄在 20-59 之间
		departmentId := (i % 5) + 1            // 部门ID 1-5
		salary := float64(3000 + (i*100)%8000) // 薪资 3000-11000

		_, err := db.Exec("INSERT INTO pagination_demo_users (name, email, age, status, department_id, salary) VALUES (?, ?, ?, ?, ?, ?)",
			fmt.Sprintf("用户%d", i),
			fmt.Sprintf("user%d@example.com", i),
			age,
			status,
			departmentId,
			salary)
		if err != nil {
			return fmt.Errorf("插入用户数据失败: %v", err)
		}
	}

	// 插入产品数据
	categories := []string{"电子产品", "服装", "食品", "图书", "家居"}
	for i := 1; i <= 30; i++ {
		category := categories[i%5]
		price := float64(10 + (i*5)%500) // 价格 10-500
		stock := 10 + (i*3)%100          // 库存 10-100

		_, err := db.Exec("INSERT INTO pagination_demo_products (name, category, price, stock) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("产品%d", i),
			category,
			price,
			stock)
		if err != nil {
			return fmt.Errorf("插入产品数据失败: %v", err)
		}
	}

	// 插入订单数据
	orderStatuses := []string{"pending", "completed", "cancelled"}
	for i := 1; i <= 100; i++ {
		userId := (i % 50) + 1 // 用户ID 1-50
		status := orderStatuses[i%3]
		amount := float64(50 + (i*10)%1000)                            // 订单金额 50-1000
		orderDate := fmt.Sprintf("2024-%02d-%02d", (i%12)+1, (i%28)+1) // 2024年的随机日期

		_, err := db.Exec("INSERT INTO pagination_demo_orders (user_id, amount, status, order_date) VALUES (?, ?, ?, ?)",
			userId, amount, status, orderDate)
		if err != nil {
			return fmt.Errorf("插入订单数据失败: %v", err)
		}
	}

	// 插入订单项数据
	for i := 1; i <= 200; i++ {
		orderId := (i % 100) + 1             // 订单ID 1-100
		productId := (i % 30) + 1            // 产品ID 1-30
		quantity := (i % 5) + 1              // 数量 1-5
		unitPrice := float64(10 + (i*3)%200) // 单价 10-200

		_, err := db.Exec("INSERT INTO pagination_demo_order_items (order_id, product_id, quantity, unit_price) VALUES (?, ?, ?, ?)",
			orderId, productId, quantity, unitPrice)
		if err != nil {
			return fmt.Errorf("插入订单项数据失败: %v", err)
		}
	}

	return nil
}

// demonstrateBasicPaginate 演示基本的 Paginate 用法
func demonstrateBasicPaginate() {
	fmt.Println("\n1️⃣ 基本 Paginate 用法演示")
	fmt.Println("--------------------------------")

	user := &models.User{}
	querySQL := "SELECT id, name, email, age, status, created_at FROM pagination_demo_users WHERE age > ? ORDER BY age ASC"

	page, err := user.Paginate(1, 10, querySQL, 25)
	if err != nil {
		log.Printf("❌ Paginate 查询失败: %v", err)
		return
	}

	fmt.Printf("📄 查询条件: 年龄 > 25，按年龄升序排列\n")
	fmt.Printf("📊 分页信息: 第 %d 页，每页 %d 条，共 %d 条记录，共 %d 页\n",
		page.PageNumber, page.PageSize, page.TotalRow, page.TotalPage)

	fmt.Println("📋 查询结果:")
	for i, u := range page.List {
		if i >= 5 { // 只显示前5条
			fmt.Printf("   ... 还有 %d 条记录\n", len(page.List)-5)
			break
		}
		fmt.Printf("   ID: %d, 姓名: %s, 年龄: %d, 状态: %s\n",
			u.ID, u.Name, u.Age, u.Status)
	}
}

// demonstrateTraditionalPaginate 演示传统 Paginate 方法
func demonstrateTraditionalPaginate() {
	fmt.Println("\n2️⃣ 传统 Paginate 方法对比")
	fmt.Println("--------------------------------")

	user := &models.User{}
	page, err := user.PaginateBuilder(1, 10, "age > ?", "age ASC", 25)
	if err != nil {
		log.Printf("❌ 传统 Paginate 查询失败: %v", err)
		return
	}

	fmt.Printf("📄 查询条件: 年龄 > 25，按年龄升序排列 (传统方法)\n")
	fmt.Printf("📊 分页信息: 第 %d 页，每页 %d 条，共 %d 条记录，共 %d 页\n",
		page.PageNumber, page.PageSize, page.TotalRow, page.TotalPage)

	fmt.Printf("📋 结果数量: %d 条记录\n", len(page.List))
	fmt.Println("💡 传统方法需要分别指定 WHERE 和 ORDER BY 子句")
	fmt.Println("💡 现在推荐使用 Paginate(page, pageSize, querySQL, args...) 方法")
}

// demonstrateComplexQuery 演示复杂查询分页
func demonstrateComplexQuery() {
	fmt.Println("\n3️⃣ 复杂查询分页演示")
	fmt.Println("--------------------------------")

	// 复杂的聚合查询
	querySQL := "SELECT status, COUNT(*) as user_count, AVG(age) as avg_age, MIN(age) as min_age, MAX(age) as max_age FROM pagination_demo_users WHERE age BETWEEN ? AND ? GROUP BY status HAVING COUNT(*) > ? ORDER BY user_count DESC"

	// 注意：这里我们使用 Record 类型，因为这是聚合查询
	recordPage, err := dbkit.Use("mysql").Paginate(1, 10, querySQL, 20, 50, 5)
	if err != nil {
		log.Printf("❌ 复杂查询失败: %v", err)
		return
	}

	fmt.Printf("📄 查询: 按状态分组统计，年龄在 20-50 之间，用户数 > 5\n")
	fmt.Printf("📊 分页信息: 第 %d 页，每页 %d 条，共 %d 条记录\n",
		recordPage.PageNumber, recordPage.PageSize, recordPage.TotalRow)

	fmt.Println("📋 统计结果:")
	for i := range recordPage.List {
		record := &recordPage.List[i]
		fmt.Printf("   状态: %s, 用户数: %d, 平均年龄: %.1f, 年龄范围: %d-%d\n",
			record.GetString("status"),
			record.GetInt64("user_count"),
			record.GetFloat("avg_age"),
			record.GetInt64("min_age"),
			record.GetInt64("max_age"))
	}
	fmt.Println("💡 Paginate 支持复杂的聚合查询和分组")
}

// demonstrateCachedPagination 演示带缓存的分页
func demonstrateCachedPagination() {
	fmt.Println("\n4️⃣ 带缓存的分页演示")
	fmt.Println("--------------------------------")

	user := &models.User{}
	querySQL := "SELECT id, name, email, age, status FROM pagination_demo_users WHERE status = ? ORDER BY created_at DESC"

	// 第一次查询（会缓存结果）
	start := time.Now()
	page1, err := user.Cache("user_active_list", 30*time.Second).Paginate(1, 8, querySQL, "active")
	if err != nil {
		log.Printf("❌ 缓存查询失败: %v", err)
		return
	}
	duration1 := time.Since(start)

	// 第二次查询（从缓存获取）
	start = time.Now()
	page2, err := user.Cache("user_active_list", 30*time.Second).Paginate(1, 8, querySQL, "active")
	if err != nil {
		log.Printf("❌ 缓存查询失败: %v", err)
		return
	}
	duration2 := time.Since(start)

	fmt.Printf("📄 查询条件: 状态为 'active' 的用户\n")
	fmt.Printf("📊 第一次查询耗时: %v (数据库查询)\n", duration1)
	fmt.Printf("📊 第二次查询耗时: %v (缓存获取)\n", duration2)
	fmt.Printf("📊 缓存加速比: %.2fx\n", float64(duration1.Nanoseconds())/float64(duration2.Nanoseconds()))
	fmt.Printf("📋 查询结果: %d 条记录\n", len(page1.List))

	// 显示部分结果
	fmt.Println("📋 活跃用户列表 (前3条):")
	for i, u := range page2.List {
		if i >= 3 {
			break
		}
		fmt.Printf("   %s (%s) - 年龄: %d\n", u.Name, u.Email, u.Age)
	}
	fmt.Println("💡 缓存可以显著提升重复查询的性能")
}

// demonstrateGlobalPaginate 演示全局分页函数
func demonstrateGlobalPaginate() {
	fmt.Println("\n5️⃣ 全局分页函数演示")
	fmt.Println("--------------------------------")

	// 使用全局 Paginate 函数
	querySQL := "SELECT id, name, email, age, status FROM pagination_demo_users WHERE age BETWEEN ? AND ? ORDER BY age DESC"

	page, err := dbkit.Paginate(1, 12, querySQL, 30, 45)
	if err != nil {
		log.Printf("❌ 全局 Paginate 查询失败: %v", err)
		return
	}

	fmt.Printf("📄 查询条件: 年龄在 30-45 之间，按年龄降序排列\n")
	fmt.Printf("📊 分页信息: 第 %d 页，每页 %d 条，共 %d 条记录\n",
		page.PageNumber, page.PageSize, page.TotalRow)

	fmt.Println("📋 中年用户列表:")
	for i := range page.List {
		record := &page.List[i]
		if i >= 6 { // 只显示前6条
			fmt.Printf("   ... 还有 %d 条记录\n", len(page.List)-6)
			break
		}
		fmt.Printf("   ID: %d, 姓名: %s, 年龄: %d, 邮箱: %s\n",
			record.GetInt64("id"),
			record.GetString("name"),
			record.GetInt64("age"),
			record.GetString("email"))
	}
	fmt.Println("💡 全局函数返回 Record 类型，适合动态查询")
}

// demonstrateJoinQueries 演示 JOIN 查询分页
func demonstrateJoinQueries() {
	fmt.Println("\n6️⃣ JOIN 查询分页演示")
	fmt.Println("--------------------------------")

	// 演示1: INNER JOIN - 用户和部门信息
	fmt.Println("📋 INNER JOIN 示例 - 用户部门信息:")
	querySQL1 := "SELECT u.id, u.name, u.age, u.salary, d.name as department_name, d.budget FROM pagination_demo_users u INNER JOIN pagination_demo_departments d ON u.department_id = d.id WHERE u.salary > ? ORDER BY u.salary DESC"

	page1, err := dbkit.Use("mysql").Paginate(1, 8, querySQL1, 5000)
	if err != nil {
		log.Printf("❌ INNER JOIN 查询失败: %v", err)
	} else {
		fmt.Printf("📊 高薪员工信息: 第 %d 页，共 %d 条记录\n", page1.PageNumber, page1.TotalRow)
		for i := range page1.List {
			record := &page1.List[i]
			if i >= 3 {
				fmt.Printf("   ... 还有 %d 条记录\n", len(page1.List)-3)
				break
			}
			fmt.Printf("   %s (薪资: %.0f) - %s部门 (预算: %.0f)\n",
				record.GetString("name"),
				record.GetFloat("salary"),
				record.GetString("department_name"),
				record.GetFloat("budget"))
		}
	}

	// 演示2: LEFT JOIN - 用户订单统计
	fmt.Println("\n📋 LEFT JOIN 示例 - 用户订单统计:")
	querySQL2 := "SELECT u.id, u.name, u.email, COUNT(o.id) as order_count, COALESCE(SUM(o.amount), 0) as total_amount FROM pagination_demo_users u LEFT JOIN pagination_demo_orders o ON u.id = o.user_id WHERE u.status = ? GROUP BY u.id, u.name, u.email HAVING COUNT(o.id) >= ? ORDER BY total_amount DESC"

	page2, err := dbkit.Use("mysql").Paginate(1, 10, querySQL2, "active", 1)
	if err != nil {
		log.Printf("❌ LEFT JOIN 查询失败: %v", err)
	} else {
		fmt.Printf("📊 活跃用户订单统计: 第 %d 页，共 %d 条记录\n", page2.PageNumber, page2.TotalRow)
		for i := range page2.List {
			record := &page2.List[i]
			if i >= 4 {
				fmt.Printf("   ... 还有 %d 条记录\n", len(page2.List)-4)
				break
			}
			fmt.Printf("   %s: %d个订单，总金额: %.2f\n",
				record.GetString("name"),
				record.GetInt64("order_count"),
				record.GetFloat("total_amount"))
		}
	}

	fmt.Println("💡 Paginate 完美支持各种 JOIN 查询")
}

// demonstrateSubqueries 演示子查询分页
func demonstrateSubqueries() {
	fmt.Println("\n7️⃣ 子查询分页演示")
	fmt.Println("--------------------------------")

	// 演示1: 标量子查询
	fmt.Println("📋 标量子查询示例 - 高于平均薪资的员工:")
	querySQL1 := "SELECT u.id, u.name, u.salary, d.name as department_name, (u.salary - (SELECT AVG(salary) FROM pagination_demo_users)) as salary_diff FROM pagination_demo_users u INNER JOIN pagination_demo_departments d ON u.department_id = d.id WHERE u.salary > (SELECT AVG(salary) FROM pagination_demo_users) ORDER BY salary_diff DESC"

	page1, err := dbkit.Use("mysql").Paginate(1, 8, querySQL1)
	if err != nil {
		log.Printf("❌ 标量子查询失败: %v", err)
	} else {
		fmt.Printf("📊 高于平均薪资员工: 第 %d 页，共 %d 条记录\n", page1.PageNumber, page1.TotalRow)
		for i := range page1.List {
			record := &page1.List[i]
			if i >= 3 {
				fmt.Printf("   ... 还有 %d 条记录\n", len(page1.List)-3)
				break
			}
			fmt.Printf("   %s (薪资: %.0f, 超出平均: +%.0f) - %s\n",
				record.GetString("name"),
				record.GetFloat("salary"),
				record.GetFloat("salary_diff"),
				record.GetString("department_name"))
		}
	}

	// 演示2: EXISTS 子查询
	fmt.Println("\n📋 EXISTS 子查询示例 - 有订单的用户:")
	querySQL2 := "SELECT u.id, u.name, u.email, u.age, d.name as department_name FROM pagination_demo_users u INNER JOIN pagination_demo_departments d ON u.department_id = d.id WHERE EXISTS (SELECT 1 FROM pagination_demo_orders o WHERE o.user_id = u.id AND o.status = 'completed') ORDER BY u.age DESC"

	page2, err := dbkit.Use("mysql").Paginate(1, 10, querySQL2)
	if err != nil {
		log.Printf("❌ EXISTS 子查询失败: %v", err)
	} else {
		fmt.Printf("📊 有完成订单的用户: 第 %d 页，共 %d 条记录\n", page2.PageNumber, page2.TotalRow)
		for i := range page2.List {
			record := &page2.List[i]
			if i >= 4 {
				fmt.Printf("   ... 还有 %d 条记录\n", len(page2.List)-4)
				break
			}
			fmt.Printf("   %s (年龄: %d) - %s\n",
				record.GetString("name"),
				record.GetInt64("age"),
				record.GetString("department_name"))
		}
	}

	// 演示3: IN 子查询
	fmt.Println("\n📋 IN 子查询示例 - 购买过电子产品的用户:")
	querySQL3 := "SELECT DISTINCT u.id, u.name, u.email, u.age FROM pagination_demo_users u WHERE u.id IN (SELECT DISTINCT o.user_id FROM pagination_demo_orders o INNER JOIN pagination_demo_order_items oi ON o.id = oi.order_id INNER JOIN pagination_demo_products p ON oi.product_id = p.id WHERE p.category = '电子产品') ORDER BY u.age ASC"

	page3, err := dbkit.Use("mysql").Paginate(1, 12, querySQL3)
	if err != nil {
		log.Printf("❌ IN 子查询失败: %v", err)
	} else {
		fmt.Printf("📊 购买过电子产品的用户: 第 %d 页，共 %d 条记录\n", page3.PageNumber, page3.TotalRow)
		for i, record := range page3.List {
			if i >= 5 {
				fmt.Printf("   ... 还有 %d 条记录\n", len(page3.List)-5)
				break
			}
			fmt.Printf("   %s (年龄: %d) - %s\n",
				record.GetString("name"),
				record.GetInt64("age"),
				record.GetString("email"))
		}
	}

	fmt.Println("💡 Paginate 支持各种复杂的子查询结构")
}

// demonstrateWindowFunctions 演示窗口函数和高级聚合
func demonstrateWindowFunctions() {
	fmt.Println("\n8️⃣ 窗口函数和高级聚合演示")
	fmt.Println("--------------------------------")

	// 演示1: ROW_NUMBER 窗口函数
	fmt.Println("📋 ROW_NUMBER 示例 - 部门内薪资排名:")
	querySQL1 := "SELECT u.id, u.name, u.salary, d.name as department_name, ROW_NUMBER() OVER (PARTITION BY d.id ORDER BY u.salary DESC) as salary_rank, RANK() OVER (PARTITION BY d.id ORDER BY u.salary DESC) as salary_rank_with_ties FROM pagination_demo_users u INNER JOIN pagination_demo_departments d ON u.department_id = d.id WHERE u.status = 'active' ORDER BY d.name, u.salary DESC"

	page1, err := dbkit.Use("mysql").Paginate(1, 15, querySQL1)
	if err != nil {
		log.Printf("❌ 窗口函数查询失败: %v", err)
	} else {
		fmt.Printf("📊 部门薪资排名: 第 %d 页，共 %d 条记录\n", page1.PageNumber, page1.TotalRow)
		currentDept := ""
		count := 0
		for i := range page1.List {
			record := &page1.List[i]
			dept := record.GetString("department_name")
			if dept != currentDept {
				if currentDept != "" {
					fmt.Println()
				}
				fmt.Printf("   【%s】:\n", dept)
				currentDept = dept
				count = 0
			}
			count++
			if count <= 3 {
				fmt.Printf("     第%d名: %s (薪资: %.0f)\n",
					record.GetInt64("salary_rank"),
					record.GetString("name"),
					record.GetFloat("salary"))
			} else if count == 4 {
				fmt.Printf("     ...\n")
			}
		}
	}

	// 演示2: 聚合窗口函数
	fmt.Println("\n📋 聚合窗口函数示例 - 累计订单金额:")
	querySQL2 := "SELECT o.id, o.user_id, o.amount, o.order_date, SUM(o.amount) OVER (PARTITION BY o.user_id ORDER BY o.order_date) as cumulative_amount, AVG(o.amount) OVER (PARTITION BY o.user_id ORDER BY o.order_date ROWS BETWEEN 2 PRECEDING AND CURRENT ROW) as moving_avg FROM pagination_demo_orders o WHERE o.status = 'completed' ORDER BY o.user_id, o.order_date"

	page2, err := dbkit.Use("mysql").Paginate(1, 12, querySQL2)
	if err != nil {
		log.Printf("❌ 聚合窗口函数查询失败: %v", err)
	} else {
		fmt.Printf("📊 用户累计订单金额: 第 %d 页，共 %d 条记录\n", page2.PageNumber, page2.TotalRow)
		currentUser := int64(0)
		count := 0
		for i := range page2.List {
			record := &page2.List[i]
			userId := record.GetInt64("user_id")
			if userId != currentUser {
				if currentUser != 0 {
					fmt.Println()
				}
				fmt.Printf("   【用户%d】:\n", userId)
				currentUser = userId
				count = 0
			}
			count++
			if count <= 2 {
				fmt.Printf("     订单%d: %.2f (累计: %.2f, 移动平均: %.2f)\n",
					record.GetInt64("id"),
					record.GetFloat("amount"),
					record.GetFloat("cumulative_amount"),
					record.GetFloat("moving_avg"))
			} else if count == 3 {
				fmt.Printf("     ...\n")
			}
		}
	}

	fmt.Println("\n💡 Paginate 支持 MySQL 8.0+ 的窗口函数")
}

// demonstrateComplexJoins 演示复杂的多表关联查询
func demonstrateComplexJoins() {
	fmt.Println("\n9️⃣ 复杂多表关联查询演示")
	fmt.Println("--------------------------------")

	// 演示1: 五表关联查询
	fmt.Println("📋 五表关联示例 - 完整的订单详情:")
	querySQL1 := "SELECT u.name as user_name, u.email, d.name as department_name, o.id as order_id, o.amount as order_amount, o.order_date, p.name as product_name, p.category, p.price, oi.quantity, oi.unit_price, (oi.quantity * oi.unit_price) as item_total FROM pagination_demo_users u INNER JOIN pagination_demo_departments d ON u.department_id = d.id INNER JOIN pagination_demo_orders o ON u.id = o.user_id INNER JOIN pagination_demo_order_items oi ON o.id = oi.order_id INNER JOIN pagination_demo_products p ON oi.product_id = p.id WHERE o.status = 'completed' AND o.order_date >= '2024-06-01' ORDER BY o.order_date DESC, o.id, oi.id"

	page1, err := dbkit.Use("mysql").Paginate(1, 20, querySQL1)
	if err != nil {
		log.Printf("❌ 五表关联查询失败: %v", err)
	} else {
		fmt.Printf("📊 完整订单详情: 第 %d 页，共 %d 条记录\n", page1.PageNumber, page1.TotalRow)
		currentOrder := int64(0)
		count := 0
		for i := range page1.List {
			record := &page1.List[i]
			orderId := record.GetInt64("order_id")
			if orderId != currentOrder {
				if currentOrder != 0 {
					fmt.Println()
				}
				fmt.Printf("   【订单%d】%s - %s (%s部门)\n",
					orderId,
					record.GetString("user_name"),
					record.GetString("email"),
					record.GetString("department_name"))
				fmt.Printf("     订单日期: %s, 总金额: %.2f\n",
					record.GetString("order_date"),
					record.GetFloat("order_amount"))
				currentOrder = orderId
				count = 0
			}
			count++
			if count <= 3 {
				fmt.Printf("     - %s (%s) x%d = %.2f\n",
					record.GetString("product_name"),
					record.GetString("category"),
					record.GetInt64("quantity"),
					record.GetFloat("item_total"))
			} else if count == 4 {
				fmt.Printf("     - ...\n")
			}
		}
	}

	// 演示2: 复杂的聚合统计查询
	fmt.Println("\n📋 复杂聚合统计示例 - 部门销售业绩:")
	querySQL2 := "SELECT d.name as department_name, d.budget, COUNT(DISTINCT u.id) as employee_count, COUNT(DISTINCT o.id) as order_count, COUNT(DISTINCT oi.id) as item_count, SUM(o.amount) as total_revenue, AVG(o.amount) as avg_order_amount, SUM(oi.quantity * oi.unit_price) as calculated_revenue, (SUM(o.amount) / d.budget * 100) as revenue_budget_ratio FROM pagination_demo_departments d LEFT JOIN pagination_demo_users u ON d.id = u.department_id LEFT JOIN pagination_demo_orders o ON u.id = o.user_id AND o.status = 'completed' LEFT JOIN pagination_demo_order_items oi ON o.id = oi.order_id GROUP BY d.id, d.name, d.budget HAVING COUNT(DISTINCT o.id) > 0 ORDER BY total_revenue DESC"

	page2, err := dbkit.Use("mysql").Paginate(1, 10, querySQL2)
	if err != nil {
		log.Printf("❌ 复杂聚合查询失败: %v", err)
	} else {
		fmt.Printf("📊 部门销售业绩: 第 %d 页，共 %d 条记录\n", page2.PageNumber, page2.TotalRow)
		for i, record := range page2.List {
			fmt.Printf("   %d. %s部门:\n", i+1, record.GetString("department_name"))
			fmt.Printf("      员工数: %d, 订单数: %d, 商品项: %d\n",
				record.GetInt64("employee_count"),
				record.GetInt64("order_count"),
				record.GetInt64("item_count"))
			fmt.Printf("      总收入: %.2f, 平均订单: %.2f\n",
				record.GetFloat("total_revenue"),
				record.GetFloat("avg_order_amount"))
			fmt.Printf("      预算完成率: %.1f%%\n",
				record.GetFloat("revenue_budget_ratio"))
			fmt.Println()
		}
	}

	// 演示3: 复杂的条件查询（替代UNION）
	fmt.Println("📋 复杂条件查询示例 - 高价值客户:")
	querySQL3 := "SELECT u.name as customer_name, u.email, SUM(o.amount) as total_spent, COUNT(o.id) as order_count, AVG(o.amount) as avg_order_amount FROM pagination_demo_users u INNER JOIN pagination_demo_orders o ON u.id = o.user_id WHERE o.status = 'completed' GROUP BY u.id, u.name, u.email HAVING SUM(o.amount) > 500 ORDER BY total_spent DESC"

	page3, err := dbkit.Use("mysql").Paginate(1, 15, querySQL3)
	if err != nil {
		log.Printf("❌ UNION 查询失败: %v", err)
	} else {
		fmt.Printf("📊 高价值客户: 第 %d 页，共 %d 条记录\n", page3.PageNumber, page3.TotalRow)
		for i := range page3.List {
			record := &page3.List[i]
			fmt.Printf("   🏆 %s (%s) - 总消费: %.2f (平均: %.2f/单, %d个订单)\n",
				record.GetString("customer_name"),
				record.GetString("email"),
				record.GetFloat("total_spent"),
				record.GetFloat("avg_order_amount"),
				record.GetInt64("order_count"))
		}
	}

	fmt.Println("\n💡 Paginate 支持最复杂的 SQL 查询结构")
}
