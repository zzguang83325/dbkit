package main

import (
	"fmt"
	"log"
	"path/filepath"

	"dbkit"
)

func main() {
	fmt.Println("=== DBKit MySQL 示例 ===")

	// 初始化日志记录器（debug模式）
	logFilePath := filepath.Join(".", "log.log")
	dbkit.InitLoggerWithFile("debug", logFilePath)
	// 初始化数据库连接
	fmt.Println("1. 初始化数据库连接...")
	dbkit.OpenDatabase(dbkit.MySQL, "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 25)
	defer dbkit.Close()

	// 检查连接
	if err := dbkit.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("✓ 数据库连接成功")

	// 创建表
	fmt.Println("\n2. 创建示例表...")
	// 先删除旧表
	dbkit.Exec("DROP TABLE IF EXISTS order_items")
	dbkit.Exec("DROP TABLE IF EXISTS orders")
	dbkit.Exec("DROP TABLE IF EXISTS products")
	dbkit.Exec("DROP TABLE IF EXISTS users")

	_, err := dbkit.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(200) NOT NULL,
			age INT DEFAULT 0,
			salary DECIMAL(10,2) DEFAULT 0,
			department VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("创建表失败: %v", err)
	} else {
		fmt.Println("✓ users 表创建成功")
	}

	// 创建订单表
	_, err = dbkit.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			order_date DATE NOT NULL,
			total_amount DECIMAL(10,2) NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		log.Printf("创建订单表失败（可能已存在）: %v", err)
	} else {
		fmt.Println("✓ orders 表创建成功")
	}

	// 创建产品表
	_, err = dbkit.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			category VARCHAR(50),
			price DECIMAL(10,2) NOT NULL,
			stock INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("创建产品表失败（可能已存在）: %v", err)
	} else {
		fmt.Println("✓ products 表创建成功")
	}

	// 创建订单明细表
	_, err = dbkit.Exec(`
		CREATE TABLE IF NOT EXISTS order_items (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			order_id BIGINT NOT NULL,
			product_id BIGINT NOT NULL,
			quantity INT NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id),
			FOREIGN KEY (product_id) REFERENCES products(id)
		)
	`)
	if err != nil {
		log.Printf("创建订单明细表失败（可能已存在）: %v", err)
	} else {
		fmt.Println("✓ order_items 表创建成功")
	}

	// 插入数据
	fmt.Println("\n3. 插入数据...")
	user := dbkit.NewRecord()
	user.Set("name", "张三")
	user.Set("email", "zhangsan@example.com")
	user.Set("age", 25)

	id, err := dbkit.Save("users", user)
	if err != nil {
		log.Fatalf("插入失败: %v", err)
	}
	fmt.Printf("✓ 插入成功，ID: %d\n", id)

	// 批量插入
	fmt.Println("\n4. 批量插入数据...")
	user1 := dbkit.NewRecord()
	user1.Set("name", "李四")
	user1.Set("email", "lisi@example.com")
	user1.Set("age", 30)

	user2 := dbkit.NewRecord()
	user2.Set("name", "王五")
	user2.Set("email", "wangwu@example.com")
	user2.Set("age", 28)

	user3 := dbkit.NewRecord()
	user3.Set("name", "赵六")
	user3.Set("email", "zhaoliu@example.com")
	user3.Set("age", 35)

	users := []*dbkit.Record{user1, user2, user3}

	total, err := dbkit.BatchInsertDefault("users", users)
	if err != nil {
		log.Fatalf("批量插入失败: %v", err)
	}
	fmt.Printf("✓ 批量插入成功，影响行数: %d\n", total)

	// 查询数据
	fmt.Println("\n5. 查询数据...")

	// 查询所有用户
	fmt.Println("\n  5.1 查询所有用户:")
	allUsers, err := dbkit.Query("SELECT * FROM users")
	if err != nil {
		log.Printf("查询所有用户失败: %v", err)
	} else {
		for i, u := range allUsers {
			fmt.Printf("    用户%d: %s, %s, 年龄: %d\n",
				i+1,
				u.GetString("name"),
				u.GetString("email"),
				u.GetInt("age"))
		}
	}

	// 条件查询
	fmt.Println("\n  5.2 查询年龄大于25的用户:")
	adultUsers, err := dbkit.Query("SELECT * FROM users WHERE age > ?", 25)
	if err != nil {
		log.Printf("查询年龄大于25的用户失败: %v", err)
	} else {
		for _, u := range adultUsers {
			fmt.Printf("    - %s, 年龄: %d\n", u.GetString("name"), u.GetInt("age"))
		}
	}

	// 查询单条记录
	fmt.Println("\n  5.3 查询单个用户:")
	firstUser, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)
	if err != nil {
		log.Printf("查询单个用户失败: %v", err)
	} else if firstUser != nil {
		fmt.Printf("    第一个用户: %s, 邮箱: %s\n",
			firstUser.GetString("name"),
			firstUser.GetString("email"))
	}

	// 按ID查询
	fmt.Println("\n  5.4 按ID查询:")
	userById, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)
	if err != nil {
		log.Printf("按ID查询用户失败: %v", err)
	} else if userById != nil {
		fmt.Printf("    ID=1的用户: %s\n", userById.GetString("name"))
	}

	// 更新数据
	fmt.Println("\n6. 更新数据...")
	userToUpdate := dbkit.NewRecord()
	userToUpdate.Set("name", "张三三")
	userToUpdate.Set("age", 26)

	affected, err := dbkit.Update("users", userToUpdate, "id = ?", 1)
	if err != nil {
		log.Fatalf("更新失败: %v", err)
	}
	fmt.Printf("✓ 更新成功，影响行数: %d\n", affected)

	// 统计查询
	fmt.Println("\n7. 统计查询...")
	count, _ := dbkit.Count("users", "")
	fmt.Printf("  用户总数: %d\n", count)

	countAdult, _ := dbkit.Count("users", "age > ?", 25)
	fmt.Printf("  年龄大于25的用户数: %d\n", countAdult)

	// 分页查询
	fmt.Println("\n8. 分页查询...")
	page := 1
	perPage := 2
	usersPage, totalUsers, err := dbkit.Paginate(page, perPage, "SELECT *", "users", "", "id ASC")
	if err != nil {
		log.Printf("分页查询失败: %v", err)
	} else {
		fmt.Printf("  第%d页（每页%d条），总用户数: %d\n", page, perPage, totalUsers)
		for i, u := range usersPage {
			fmt.Printf("    %d. %s (ID: %d)\n", i+1, u.GetString("name"), u.GetInt("id"))
		}
	}

	// 删除数据
	fmt.Println("\n9. 删除数据...")
	deleted, err := dbkit.Delete("users", "id = ?", 100)
	if err != nil {
		log.Fatalf("删除失败: %v", err)
	}
	fmt.Printf("✓ 删除成功，影响行数: %d\n", deleted)

	// 检查记录是否存在
	fmt.Println("\n10. 检查记录是否存在...")
	exists := dbkit.Exists("users", "id = ?", 1)
	fmt.Printf("  ID=1的用户是否存在: %v\n", exists)

	// Record的各种获取方法
	fmt.Println("\n11. Record数据获取方法...")
	record, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)
	if err != nil {
		log.Printf("查询单条记录失败: %v", err)
	} else if record != nil {
		fmt.Printf("  姓名: %s\n", record.Str("name"))
		fmt.Printf("  年龄: %d\n", record.Int("age"))
		fmt.Printf("  年龄(Int64): %d\n", record.Int64("age"))
		fmt.Printf("  年龄(Float): %.2f\n", record.Float("age"))
		fmt.Printf("  有邮箱字段: %v\n", record.Has("email"))
		fmt.Printf("  所有字段: %v\n", record.Keys())
		fmt.Printf("  转换为JSON: %s\n", record.ToJson())
	}

	// 查询为Map
	fmt.Println("\n12. 查询结果转换为Map...")
	usersMap, err := dbkit.QueryMap("SELECT * FROM users WHERE id = ?", 1)
	if err != nil {
		log.Printf("查询Map结果失败: %v", err)
	} else if len(usersMap) > 0 {
		for k, v := range usersMap[0] {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}

	// 事务示例
	fmt.Println("\n13. 事务操作...")
	tx, err := dbkit.BeginTransaction()
	if err != nil {
		log.Fatalf("开启事务失败: %v", err)
	}

	// 在事务中插入
	userTx := dbkit.NewRecord()
	userTx.Set("name", "事务用户")
	userTx.Set("email", "tx@example.com")
	userTx.Set("age", 40)

	txId, err := dbkit.SaveTx(tx, "users", userTx)
	if err != nil {
		tx.Rollback()
		log.Fatalf("事务插入失败: %v", err)
	}
	fmt.Printf("  事务中插入用户ID: %d\n", txId)

	// 提交事务
	if err := tx.Commit(); err != nil {
		log.Fatalf("提交事务失败: %v", err)
	}
	fmt.Println("✓ 事务提交成功")

	// 清理测试数据
	fmt.Println("\n14. 清理测试数据...")
	dbkit.Delete("users", "name = ?", "事务用户")
	dbkit.Delete("users", "name = ?", "张三三")
	fmt.Println("✓ 测试数据清理完成")

	// ====== 复杂 SQL 查询测试 ======
	fmt.Println("\n========== 复杂 SQL 查询测试 ==========")

	// 15. 多表连接查询 (JOIN)
	fmt.Println("\n15. 多表连接查询 (JOIN)...")
	// 先插入一些测试数据
	testUsers := []*dbkit.Record{
		func() *dbkit.Record {
			r := dbkit.NewRecord()
			r.Set("name", "用户A")
			r.Set("email", "usera@test.com")
			r.Set("age", 28)
			r.Set("salary", 8000.00)
			r.Set("department", "技术部")
			return r
		}(),
		func() *dbkit.Record {
			r := dbkit.NewRecord()
			r.Set("name", "用户B")
			r.Set("email", "userb@test.com")
			r.Set("age", 32)
			r.Set("salary", 12000.00)
			r.Set("department", "市场部")
			return r
		}(),
		func() *dbkit.Record {
			r := dbkit.NewRecord()
			r.Set("name", "用户C")
			r.Set("email", "userc@test.com")
			r.Set("age", 35)
			r.Set("salary", 15000.00)
			r.Set("department", "技术部")
			return r
		}(),
	}
	c, err := dbkit.BatchInsertDefault("users", testUsers)
	if err != nil {
		log.Fatalf("批量插入用户失败: %v", err)
	}
	fmt.Printf("✓ 批量插入 %d 条用户记录\n", c)
	// 插入产品
	testProducts := []*dbkit.Record{
		func() *dbkit.Record {
			r := dbkit.NewRecord()
			r.Set("name", "笔记本电脑")
			r.Set("category", "电子产品")
			r.Set("price", 5999.00)
			r.Set("stock", 100)
			return r
		}(),
		func() *dbkit.Record {
			r := dbkit.NewRecord()
			r.Set("name", "无线鼠标")
			r.Set("category", "配件")
			r.Set("price", 99.00)
			r.Set("stock", 500)
			return r
		}(),
		func() *dbkit.Record {
			r := dbkit.NewRecord()
			r.Set("name", "机械键盘")
			r.Set("category", "配件")
			r.Set("price", 299.00)
			r.Set("stock", 200)
			return r
		}(),
	}
	dbkit.BatchInsertDefault("products", testProducts)

	// 插入订单
	_, err = dbkit.Exec("INSERT INTO orders (user_id, order_date, total_amount, status) VALUES (?, CURDATE(), ?, 'completed')", 1, 5999.00)
	if err != nil {
		log.Println("插入订单失败: %v", err)
	}
	_, err = dbkit.Exec("INSERT INTO orders (user_id, order_date, total_amount, status) VALUES (?, CURDATE(), ?, 'completed')", 2, 99.00)
	if err != nil {
		log.Println("插入订单失败: %v", err)
	}
	_, err = dbkit.Exec("INSERT INTO orders (user_id, order_date, total_amount, status) VALUES (?, CURDATE(), ?, 'pending')", 3, 299.00)
	if err != nil {
		log.Println("插入订单失败: %v", err)
	}

	// INNER JOIN 查询
	fmt.Println("\n  15.1 INNER JOIN - 查询有订单的用户:")
	joinResults, err := dbkit.Query(`
		SELECT u.id, u.name, u.email, o.id as order_id, o.total_amount, o.status
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		ORDER BY u.id, o.id
	`)
	if err != nil {
		log.Printf("INNER JOIN 查询失败: %v", err)
	} else {
		for _, row := range joinResults {
			fmt.Printf("    用户: %s (ID:%d) - 订单ID: %d, 金额: %.2f, 状态: %s\n",
				row.GetString("name"),
				row.GetInt("id"),
				row.GetInt("order_id"),
				row.Float("total_amount"),
				row.GetString("status"))
		}
	}

	// LEFT JOIN 查询
	fmt.Println("\n  15.2 LEFT JOIN - 查询所有用户及其订单（包括没有订单的）:")
	leftJoinResults, err := dbkit.Query(`
		SELECT u.id, u.name, u.email, o.id as order_id, o.total_amount
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		ORDER BY u.id
	`)
	if err != nil {
		log.Printf("LEFT JOIN 查询失败: %v", err)
	} else {
		for _, row := range leftJoinResults {
			if row.GetInt("order_id") == 0 {
				fmt.Printf("    用户: %s (ID:%d) - 无订单\n", row.GetString("name"), row.GetInt("id"))
			} else {
				fmt.Printf("    用户: %s (ID:%d) - 订单ID: %d, 金额: %.2f\n",
					row.GetString("name"),
					row.GetInt("id"),
					row.GetInt("order_id"),
					row.Float("total_amount"))
			}
		}
	}

	// 16. 聚合函数和 GROUP BY
	fmt.Println("\n16. 聚合函数和 GROUP BY...")

	// 按部门统计
	fmt.Println("\n  16.1 按部门统计员工数量和平均薪资:")
	deptStats, err := dbkit.Query(`
		SELECT 
			department,
			COUNT(*) as employee_count,
			AVG(salary) as avg_salary,
			MAX(salary) as max_salary,
			MIN(salary) as min_salary,
			SUM(salary) as total_salary
		FROM users
		WHERE department IS NOT NULL
		GROUP BY department
		ORDER BY avg_salary DESC
	`)
	if err != nil {
		log.Printf("按部门统计查询失败: %v", err)
	} else {
		for _, row := range deptStats {
			fmt.Printf("    部门: %s | 员工数: %d | 平均薪资: %.2f | 最高: %.2f | 最低: %.2f | 总和: %.2f\n",
				row.GetString("department"),
				row.GetInt("employee_count"),
				row.Float("avg_salary"),
				row.Float("max_salary"),
				row.Float("min_salary"),
				row.Float("total_salary"))
		}
	}

	// 按年龄分组统计
	fmt.Println("\n  16.2 按年龄段统计用户数量:")
	ageGroups, err := dbkit.Query(`
		SELECT 
			CASE 
				WHEN age < 25 THEN '25岁以下'
				WHEN age BETWEEN 25 AND 30 THEN '25-30岁'
				WHEN age BETWEEN 31 AND 35 THEN '31-35岁'
				ELSE '35岁以上'
			END as age_group,
			COUNT(*) as user_count
		FROM users
		GROUP BY age_group
		ORDER BY age_group
	`)
	if err != nil {
		log.Printf("按年龄分组统计查询失败: %v", err)
	} else {
		for _, row := range ageGroups {
			fmt.Printf("    %s: %d 人\n", row.GetString("age_group"), row.GetInt("user_count"))
		}
	}

	// 17. 子查询
	fmt.Println("\n17. 子查询...")

	// 使用子查询查找薪资高于平均薪资的员工
	fmt.Println("\n  17.1 查找薪资高于平均薪资的员工:")
	aboveAvgSalary, err := dbkit.Query(`
		SELECT id, name, email, salary, department
		FROM users
		WHERE salary > (SELECT AVG(salary) FROM users WHERE salary > 0)
		ORDER BY salary DESC
	`)
	if err != nil {
		log.Printf("查找薪资高于平均薪资的员工查询失败: %v", err)
	} else {
		for _, row := range aboveAvgSalary {
			fmt.Printf("    %s (部门: %s) - 薪资: %.2f\n",
				row.GetString("name"),
				row.GetString("department"),
				row.Float("salary"))
		}
	}

	// 使用子查询查找没有订单的用户
	fmt.Println("\n  17.2 查找没有订单的用户:")
	usersWithoutOrders, err := dbkit.Query(`
		SELECT id, name, email
		FROM users
		WHERE id NOT IN (SELECT DISTINCT user_id FROM orders)
		ORDER BY id
	`)
	if err != nil {
		log.Printf("查找没有订单的用户查询失败: %v", err)
	} else {
		for _, row := range usersWithoutOrders {
			fmt.Printf("    %s (ID: %d, 邮箱: %s)\n",
				row.GetString("name"),
				row.GetInt("id"),
				row.GetString("email"))
		}
	}

	// 18. HAVING 子句
	fmt.Println("\n18. HAVING 子句...")

	// 查找订单总额超过 5000 的用户
	fmt.Println("\n  18.1 查找订单总额超过 5000 的用户:")
	highValueUsers, err := dbkit.Query(`
		SELECT 
			u.id, u.name, u.email,
			COUNT(o.id) as order_count,
			SUM(o.total_amount) as total_spent
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		GROUP BY u.id, u.name, u.email
		HAVING SUM(o.total_amount) > 5000
		ORDER BY total_spent DESC
	`)
	if err != nil {
		log.Printf("查找高价值用户查询失败: %v", err)
	} else {
		for _, row := range highValueUsers {
			fmt.Printf("    %s - 订单数: %d, 总消费: %.2f\n",
				row.GetString("name"),
				row.GetInt("order_count"),
				row.Float("total_spent"))
		}
	}

	// 19. 窗口函数 (Window Functions)
	fmt.Println("\n19. 窗口函数 (Window Functions)...")

	// 使用 ROW_NUMBER() 为用户排名
	fmt.Println("\n  19.1 按薪资为用户排名:")
	rankedUsers, err := dbkit.Query(`
		SELECT 
			id, name, email, salary, department,
			ROW_NUMBER() OVER (ORDER BY salary DESC) as salary_rank,
			DENSE_RANK() OVER (PARTITION BY department ORDER BY salary DESC) as dept_rank,
			LAG(salary) OVER (ORDER BY salary DESC) as prev_salary,
			LEAD(salary) OVER (ORDER BY salary DESC) as next_salary
		FROM users
		WHERE salary > 0
		ORDER BY salary DESC
		LIMIT 5
	`)
	if err != nil {
		log.Printf("用户薪资排名查询失败: %v", err)
	} else {
		for _, row := range rankedUsers {
			fmt.Printf("    %s (部门: %s) - 薪资: %.2f | 全局排名: %d | 部门排名: %d\n",
				row.GetString("name"),
				row.GetString("department"),
				row.Float("salary"),
				row.GetInt("salary_rank"),
				row.GetInt("dept_rank"))
		}
	}

	// 20. UNION 和 UNION ALL
	fmt.Println("\n20. UNION 和 UNION ALL...")

	// 合并查询
	fmt.Println("\n  20.1 查询所有用户和产品名称:")
	allItems, err := dbkit.Query(`
		SELECT id, name, 'user' as item_type, email as extra_info
		FROM users
		WHERE id <= 3
		UNION ALL
		SELECT id, name, 'product' as item_type, category as extra_info
		FROM products
		WHERE id <= 3
		ORDER BY item_type, id
	`)
	if err != nil {
		log.Printf("合并查询失败: %v", err)
	} else {
		for _, row := range allItems {
			fmt.Printf("    %s (ID: %d, 类型: %s, 额外信息: %s)\n",
				row.GetString("name"),
				row.GetInt("id"),
				row.GetString("item_type"),
				row.GetString("extra_info"))
		}
	}

	// 21. CASE 表达式
	fmt.Println("\n21. CASE 表达式...")

	// 使用 CASE 进行条件判断
	fmt.Println("\n  21.1 根据薪资评定等级:")
	salaryGrades, err := dbkit.Query(`
		SELECT 
			id, name, salary, department,
			CASE 
				WHEN salary >= 15000 THEN '高级'
				WHEN salary >= 10000 THEN '中级'
				WHEN salary >= 5000 THEN '初级'
				ELSE '实习'
			END as salary_grade,
			CASE 
				WHEN age < 30 THEN '青年'
				WHEN age < 40 THEN '中年'
				ELSE '资深'
			END as age_group
		FROM users
		WHERE salary > 0
		ORDER BY salary DESC
	`)
	if err != nil {
		log.Printf("薪资等级查询失败: %v", err)
	} else {
		for _, row := range salaryGrades {
			fmt.Printf("    %s - 薪资等级: %s, 年龄段: %s\n",
				row.GetString("name"),
				row.GetString("salary_grade"),
				row.GetString("age_group"))
		}
	}

	// 22. 复杂的 WHERE 条件
	fmt.Println("\n22. 复杂的 WHERE 条件...")

	// 使用多个条件组合
	fmt.Println("\n  22.1 复杂条件查询:")
	complexQuery, err := dbkit.Query(`
		SELECT id, name, email, age, salary, department
		FROM users
		WHERE (age BETWEEN 25 AND 35)
		  AND (salary > 5000 OR department = '技术部')
		  AND (name LIKE '张%' OR name LIKE '李%')
		ORDER BY salary DESC
	`)
	if err != nil {
		log.Printf("复杂条件查询失败: %v", err)
	} else {
		for _, row := range complexQuery {
			fmt.Printf("    %s (年龄: %d, 薪资: %.2f, 部门: %s)\n",
				row.GetString("name"),
				row.GetInt("age"),
				row.Float("salary"),
				row.GetString("department"))
		}
	}

	// 23. DISTINCT 去重
	fmt.Println("\n23. DISTINCT 去重...")

	// 获取不同的部门
	fmt.Println("\n  23.1 获取所有不同的部门:")
	departments, err := dbkit.Query(`
		SELECT DISTINCT department
		FROM users
		WHERE department IS NOT NULL
		ORDER BY department
	`)
	if err != nil {
		log.Printf("获取不同部门失败: %v", err)
	} else {
		for _, row := range departments {
			fmt.Printf("    部门: %s\n", row.GetString("department"))
		}
	}

	// 24. LIMIT 和 OFFSET 分页
	fmt.Println("\n24. LIMIT 和 OFFSET 分页...")

	// 分页查询用户
	fmt.Println("\n  24.1 分页查询用户（第2页，每页3条）:")
	pagedUsers, err := dbkit.Query(`
		SELECT id, name, email, age, salary
		FROM users
		WHERE salary > 0
		ORDER BY salary DESC
		LIMIT 3 OFFSET 3
	`)
	if err != nil {
		log.Printf("分页查询用户失败: %v", err)
	} else {
		for _, row := range pagedUsers {
			fmt.Printf("    %s (ID: %d, 薪资: %.2f)\n",
				row.GetString("name"),
				row.GetInt("id"),
				row.Float("salary"))
		}
	}

	// 25. 日期和时间函数
	fmt.Println("\n25. 日期和时间函数...")

	// 使用日期函数
	fmt.Println("\n  25.1 日期相关查询:")
	dateQuery, err := dbkit.Query(`
		SELECT 
			id,
			name,
			created_at,
			DATEDIFF(NOW(), created_at) as days_since_creation,
			YEAR(created_at) as year,
			MONTH(created_at) as month,
			DAY(created_at) as day
		FROM users
	WHERE id <= 5
	ORDER BY created_at DESC
`)
	if err != nil {
		log.Printf("日期相关查询失败: %v", err)
	} else {
		for _, row := range dateQuery {
			fmt.Printf("    %s - 创建于: %s, 已创建 %d 天\n",
				row.GetString("name"),
				row.GetString("created_at"),
				row.GetInt("days_since_creation"))
		}
	}

	// 26. 字符串函数
	fmt.Println("\n26. 字符串函数...")

	// 使用字符串函数
	fmt.Println("\n  26.1 字符串操作:")
	stringQuery, err := dbkit.Query(`
		SELECT 
			id,
			name,
			email,
			UPPER(name) as name_upper,
			LOWER(email) as email_lower,
			CONCAT(name, ' <', email, '>') as full_contact,
			SUBSTRING(email, 1, LOCATE('@', email) - 1) as email_prefix,
			LENGTH(name) as name_length
		FROM users
		WHERE id <= 5
		ORDER BY id
	`)
	if err != nil {
		log.Printf("字符串操作查询失败: %v", err)
	} else {
		for _, row := range stringQuery {
			fmt.Printf("    %s - 联系方式: %s\n",
				row.GetString("name"),
				row.GetString("full_contact"))
		}
	}

	// 27. 数学函数
	fmt.Println("\n27. 数学函数...")

	// 使用数学函数
	fmt.Println("\n  27.1 数学计算:")
	mathQuery, err := dbkit.Query(`
		SELECT 
			id,
			name,
			salary,
			ROUND(salary, -2) as salary_rounded_hundreds,
			CEIL(salary / 1000) * 1000 as salary_ceil_thousand,
			FLOOR(salary / 1000) * 1000 as salary_floor_thousand,
			MOD(age, 10) as age_mod_10
		FROM users
		WHERE salary > 0 AND id <= 5
		ORDER BY salary DESC
	`)
	if err != nil {
		log.Printf("数学计算查询失败: %v", err)
	} else {
		for _, row := range mathQuery {
			fmt.Printf("    %s - 原薪资: %.2f, 百位取整: %.2f\n",
				row.GetString("name"),
				row.Float("salary"),
				row.Float("salary_rounded_hundreds"))
		}
	}

	// 28. EXISTS 和 NOT EXISTS
	fmt.Println("\n28. EXISTS 和 NOT EXISTS...")

	// 使用 EXISTS
	fmt.Println("\n  28.1 使用 EXISTS 查找有订单的用户:")
	existsQuery, err := dbkit.Query(`
		SELECT id, name, email
		FROM users u
		WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id)
		ORDER BY id
	`)
	if err != nil {
		log.Printf("EXISTS查询失败: %v", err)
	} else {
		for _, row := range existsQuery {
			fmt.Printf("    %s (ID: %d)\n", row.GetString("name"), row.GetInt("id"))
		}
	}

	// 29. 多表连接（三个表）
	fmt.Println("\n29. 多表连接（三个表）...")

	// 连接三个表
	fmt.Println("\n  29.1 用户-订单-产品关联查询:")
	threeTableJoin, err := dbkit.Query(`
		SELECT 
			u.name as user_name,
			o.id as order_id,
			o.total_amount,
			p.name as product_name,
			p.category,
			p.price as product_price
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		LEFT JOIN order_items oi ON o.id = oi.order_id
		LEFT JOIN products p ON oi.product_id = p.id
		WHERE u.id <= 3
	ORDER BY u.id, o.id
`)
	if err != nil {
		log.Printf("三表连接查询失败: %v", err)
	} else {
		for _, row := range threeTableJoin {
			if row.GetString("product_name") != "" {
				fmt.Printf("    %s - 订单: %d (%.2f) - 产品: %s (%s, %.2f)\n",
					row.GetString("user_name"),
					row.GetInt("order_id"),
					row.Float("total_amount"),
					row.GetString("product_name"),
					row.GetString("category"),
					row.Float("product_price"))
			}
		}
	}

	// 30. 自连接 (Self Join)
	fmt.Println("\n30. 自连接 (Self Join)...")

	// 查找薪资比自己高的同事
	fmt.Println("\n  30.1 查找薪资比自己高的同事:")
	selfJoinQuery, err := dbkit.Query(`
		SELECT 
			u1.name as employee_name,
			u1.salary as employee_salary,
			u2.name as higher_paid_colleague,
			u2.salary as colleague_salary
		FROM users u1
		INNER JOIN users u2 ON u1.department = u2.department 
			AND u2.salary > u1.salary
			AND u1.salary > 0
			AND u2.salary > 0
		WHERE u1.id <= 3
		ORDER BY u1.name, u2.salary DESC
	`)
	if err != nil {
		log.Printf("自连接查询失败: %v", err)
	} else {
		for _, row := range selfJoinQuery {
			fmt.Printf("    %s (%.2f) < %s (%.2f)\n",
				row.GetString("employee_name"),
				row.Float("employee_salary"),
				row.GetString("higher_paid_colleague"),
				row.Float("colleague_salary"))
		}
	}

	fmt.Println("\n========== 复杂 SQL 查询测试完成 ==========")
	fmt.Println("\n=== MySQL 示例完成 ===")
}
