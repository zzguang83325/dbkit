package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/comprehensive/models"
)

func main() {
	// 1. 初始化数据库连接 - MySQL
	dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabase(dbkit.MySQL, dsn, 10)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer dbkit.Close()

	// 开启 Debug 模式输出 SQL
	dbkit.SetDebugMode(true)

	fmt.Println("\n" + repeat("=", 60))
	fmt.Println("DBKit 综合 API 测试示例")
	fmt.Println(repeat("=", 60))

	// 2. 初始化环境 (创建表)
	setupTables()

	// 3. 准备测试数据
	prepareData()

	// 运行所有测试
	testBasicCRUD()
	testChainQuery()
	testAdvancedWhere()
	testJoinQuery()
	testSubquery()
	testGroupByHaving()
	testTransaction()
	testPagination()
	testCache()
	testBatchOperations()
	testDbModel()

	fmt.Println("\n" + repeat("=", 60))
	fmt.Println("所有测试完成!")
	fmt.Println(repeat("=", 60))
}

// ==================== 基础 CRUD 测试 ====================
func testBasicCRUD() {
	fmt.Println("\n[测试 1: 基础 CRUD 操作]")

	// Insert
	user := dbkit.NewRecord().
		Set("username", "TestUser").
		Set("email", "test@example.com").
		Set("age", 25).
		Set("status", "active").
		Set("created_at", time.Now())
	id, err := dbkit.Insert("users", user)
	if err != nil {
		log.Printf("  Insert 失败: %v", err)
	} else {
		fmt.Printf("  ✓ Insert 成功, ID: %d\n", id)
	}

	// Query
	record, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		log.Printf("  QueryFirst 失败: %v", err)
	} else {
		fmt.Printf("  ✓ QueryFirst 成功: username=%s, age=%d\n", record.Str("username"), record.Int("age"))
	}

	// Update
	record.Set("age", 26)
	affected, err := dbkit.Update("users", record, "id = ?", id)
	if err != nil {
		log.Printf("  Update 失败: %v", err)
	} else {
		fmt.Printf("  ✓ Update 成功, 影响行数: %d\n", affected)
	}

	// Save (update existing)
	record.Set("age", 27)
	_, err = dbkit.Save("users", record)
	if err != nil {
		log.Printf("  Save 失败: %v", err)
	} else {
		fmt.Printf("  ✓ Save 成功 (更新)\n")
	}

	// Count
	count, err := dbkit.Count("users", "status = ?", "active")
	if err != nil {
		log.Printf("  Count 失败: %v", err)
	} else {
		fmt.Printf("  ✓ Count 成功: %d 条记录\n", count)
	}

	// Exists
	exists, err := dbkit.Exists("users", "username = ?", "TestUser")
	if err != nil {
		log.Printf("  Exists 失败: %v", err)
	} else {
		fmt.Printf("  ✓ Exists 成功: %v\n", exists)
	}
}

// ==================== 链式查询测试 ====================
func testChainQuery() {
	fmt.Println("\n[测试 2: 链式查询 (QueryBuilder)]")

	// 基本链式查询
	users, err := dbkit.Table("users").
		Select("id, username, age, status").
		Where("age > ?", 20).
		Where("status = ?", "active").
		OrderBy("age DESC").
		Limit(5).
		Find()
	if err != nil {
		log.Printf("  链式查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ 基本链式查询成功, 返回 %d 条记录\n", len(users))
		for i := range users {
			fmt.Printf("    - %s (age: %d)\n", users[i].Str("username"), users[i].Int("age"))
		}
	}

	// FindFirst
	user, err := dbkit.Table("users").
		Where("status = ?", "active").
		OrderBy("id DESC").
		FindFirst()
	if err != nil {
		log.Printf("  FindFirst 失败: %v", err)
	} else if user != nil {
		fmt.Printf("  ✓ FindFirst 成功: %s\n", user.Str("username"))
	}

	// Offset
	users2, err := dbkit.Table("users").
		OrderBy("id ASC").
		Limit(3).
		Offset(2).
		Find()
	if err != nil {
		log.Printf("  Offset 查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ Offset 查询成功, 返回 %d 条记录\n", len(users2))
	}
}

// ==================== 高级 WHERE 条件测试 ====================
func testAdvancedWhere() {
	fmt.Println("\n[测试 3: 高级 WHERE 条件]")

	// OrWhere
	users, err := dbkit.Table("users").
		Where("status = ?", "active").
		OrWhere("age > ?", 40).
		Find()
	if err != nil {
		log.Printf("  OrWhere 失败: %v", err)
	} else {
		fmt.Printf("  ✓ OrWhere 成功: %d 条记录\n", len(users))
	}

	// WhereGroup / OrWhereGroup
	users2, err := dbkit.Table("users").
		Where("status = ?", "active").
		OrWhereGroup(func(qb *dbkit.QueryBuilder) *dbkit.QueryBuilder {
			return qb.Where("age > ?", 30).Where("age < ?", 50)
		}).
		Find()
	if err != nil {
		log.Printf("  WhereGroup 失败: %v", err)
	} else {
		fmt.Printf("  ✓ OrWhereGroup 成功: %d 条记录\n", len(users2))
	}

	// WhereInValues
	users3, err := dbkit.Table("users").
		WhereInValues("age", []interface{}{25, 30, 35, 40}).
		Find()
	if err != nil {
		log.Printf("  WhereInValues 失败: %v", err)
	} else {
		fmt.Printf("  ✓ WhereInValues 成功: %d 条记录\n", len(users3))
	}

	// WhereNotInValues
	users4, err := dbkit.Table("users").
		WhereNotInValues("status", []interface{}{"banned", "deleted"}).
		Find()
	if err != nil {
		log.Printf("  WhereNotInValues 失败: %v", err)
	} else {
		fmt.Printf("  ✓ WhereNotInValues 成功: %d 条记录\n", len(users4))
	}

	// WhereBetween
	users5, err := dbkit.Table("users").
		WhereBetween("age", 25, 35).
		Find()
	if err != nil {
		log.Printf("  WhereBetween 失败: %v", err)
	} else {
		fmt.Printf("  ✓ WhereBetween 成功: %d 条记录\n", len(users5))
	}

	// WhereNotBetween
	users6, err := dbkit.Table("users").
		WhereNotBetween("age", 20, 25).
		Find()
	if err != nil {
		log.Printf("  WhereNotBetween 失败: %v", err)
	} else {
		fmt.Printf("  ✓ WhereNotBetween 成功: %d 条记录\n", len(users6))
	}

	// WhereNull / WhereNotNull
	users7, err := dbkit.Table("users").
		WhereNotNull("email").
		Find()
	if err != nil {
		log.Printf("  WhereNotNull 失败: %v", err)
	} else {
		fmt.Printf("  ✓ WhereNotNull 成功: %d 条记录\n", len(users7))
	}

	users8, err := dbkit.Table("users").
		WhereNull("deleted_at").
		Find()
	if err != nil {
		log.Printf("  WhereNull 失败: %v", err)
	} else {
		fmt.Printf("  ✓ WhereNull 成功: %d 条记录\n", len(users8))
	}
}

// ==================== JOIN 查询测试 ====================
func testJoinQuery() {
	fmt.Println("\n[测试 4: JOIN 查询]")

	// LEFT JOIN
	records, err := dbkit.Table("users").
		Select("users.username, orders.amount, orders.status as order_status").
		LeftJoin("orders", "users.id = orders.user_id").
		Where("orders.amount > ?", 100).
		OrderBy("orders.amount DESC").
		Limit(5).
		Find()
	if err != nil {
		log.Printf("  LEFT JOIN 失败: %v", err)
	} else {
		fmt.Printf("  ✓ LEFT JOIN 成功: %d 条记录\n", len(records))
		for i := range records {
			fmt.Printf("    - %s: ¥%.2f (%s)\n", records[i].Str("username"), records[i].Float("amount"), records[i].Str("order_status"))
		}
	}

	// INNER JOIN
	records2, err := dbkit.Table("orders").
		Select("orders.id, users.username, products.name as product_name, order_items.quantity").
		InnerJoin("users", "orders.user_id = users.id").
		InnerJoin("order_items", "orders.id = order_items.order_id").
		InnerJoin("products", "order_items.product_id = products.id").
		Where("orders.status = ?", "COMPLETED").
		Limit(5).
		Find()
	if err != nil {
		log.Printf("  多表 INNER JOIN 失败: %v", err)
	} else {
		fmt.Printf("  ✓ 多表 INNER JOIN 成功: %d 条记录\n", len(records2))
	}

	// RIGHT JOIN
	records3, err := dbkit.Table("users").
		Select("users.username, orders.amount").
		RightJoin("orders", "users.id = orders.user_id").
		Limit(5).
		Find()
	if err != nil {
		log.Printf("  RIGHT JOIN 失败: %v", err)
	} else {
		fmt.Printf("  ✓ RIGHT JOIN 成功: %d 条记录\n", len(records3))
	}

	// JOIN with parameters
	records4, err := dbkit.Table("users").
		Select("users.username, orders.amount").
		Join("orders", "users.id = orders.user_id AND orders.status = ?", "COMPLETED").
		Find()
	if err != nil {
		log.Printf("  带参数 JOIN 失败: %v", err)
	} else {
		fmt.Printf("  ✓ 带参数 JOIN 成功: %d 条记录\n", len(records4))
	}
}

// ==================== 子查询测试 ====================
func testSubquery() {
	fmt.Println("\n[测试 5: 子查询 (Subquery)]")

	// WhereIn with Subquery
	activeUsersSub := dbkit.NewSubquery().
		Table("orders").
		Select("DISTINCT user_id").
		Where("status = ?", "COMPLETED")

	users, err := dbkit.Table("users").
		Select("id, username").
		WhereIn("id", activeUsersSub).
		Find()
	if err != nil {
		log.Printf("  WhereIn 子查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ WhereIn 子查询成功: %d 条记录 (有已完成订单的用户)\n", len(users))
	}

	// WhereNotIn with Subquery
	bannedUsersSub := dbkit.NewSubquery().
		Table("users").
		Select("id").
		Where("status = ?", "banned")

	orders, err := dbkit.Table("orders").
		WhereNotIn("user_id", bannedUsersSub).
		Find()
	if err != nil {
		log.Printf("  WhereNotIn 子查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ WhereNotIn 子查询成功: %d 条记录\n", len(orders))
	}

	// FROM Subquery (TableSubquery)
	userTotalsSub := dbkit.NewSubquery().
		Table("orders").
		Select("user_id, SUM(amount) as total_spent, COUNT(*) as order_count")

	// Note: MySQL requires GROUP BY in subquery for aggregation
	records, err := dbkit.Query(`
		SELECT user_id, total_spent, order_count 
		FROM (SELECT user_id, SUM(amount) as total_spent, COUNT(*) as order_count FROM orders GROUP BY user_id) AS user_totals 
		WHERE total_spent > ?`, 200)
	if err != nil {
		log.Printf("  FROM 子查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ FROM 子查询成功: %d 条记录\n", len(records))
	}
	_ = userTotalsSub // 使用变量避免警告

	// SELECT Subquery
	orderCountSub := dbkit.NewSubquery().
		Table("orders").
		Select("COUNT(*)").
		Where("orders.user_id = users.id")

	users2, err := dbkit.Table("users").
		Select("users.id, users.username").
		SelectSubquery(orderCountSub, "order_count").
		Limit(5).
		Find()
	if err != nil {
		log.Printf("  SELECT 子查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ SELECT 子查询成功: %d 条记录\n", len(users2))
		for i := range users2 {
			fmt.Printf("    - %s: %d 个订单\n", users2[i].Str("username"), users2[i].Int("order_count"))
		}
	}
}

// ==================== GROUP BY / HAVING 测试 ====================
func testGroupByHaving() {
	fmt.Println("\n[测试 6: GROUP BY / HAVING]")

	// 基本 GroupBy
	stats, err := dbkit.Table("orders").
		Select("status, COUNT(*) as count, SUM(amount) as total_amount").
		GroupBy("status").
		Find()
	if err != nil {
		log.Printf("  GroupBy 失败: %v", err)
	} else {
		fmt.Printf("  ✓ GroupBy 成功:\n")
		for i := range stats {
			fmt.Printf("    - %s: %d 订单, 总金额 ¥%.2f\n", stats[i].Str("status"), stats[i].Int("count"), stats[i].Float("total_amount"))
		}
	}

	// GroupBy + Having
	userStats, err := dbkit.Table("orders").
		Select("user_id, COUNT(*) as order_count, SUM(amount) as total_spent").
		GroupBy("user_id").
		Having("COUNT(*) >= ?", 2).
		Find()
	if err != nil {
		log.Printf("  GroupBy + Having 失败: %v", err)
	} else {
		fmt.Printf("  ✓ GroupBy + Having 成功: %d 个用户有 2+ 订单\n", len(userStats))
	}

	// 多个 Having 条件
	userStats2, err := dbkit.Table("orders").
		Select("user_id, COUNT(*) as cnt, SUM(amount) as total").
		GroupBy("user_id").
		Having("COUNT(*) >= ?", 1).
		Having("SUM(amount) > ?", 100).
		Find()
	if err != nil {
		log.Printf("  多 Having 条件失败: %v", err)
	} else {
		fmt.Printf("  ✓ 多 Having 条件成功: %d 条记录\n", len(userStats2))
	}

	// GroupBy 多列
	stats2, err := dbkit.Table("orders").
		Select("user_id, status, COUNT(*) as count").
		GroupBy("user_id, status").
		OrderBy("user_id, status").
		Find()
	if err != nil {
		log.Printf("  多列 GroupBy 失败: %v", err)
	} else {
		fmt.Printf("  ✓ 多列 GroupBy 成功: %d 条记录\n", len(stats2))
	}
}

// ==================== 事务测试 ====================
func testTransaction() {
	fmt.Println("\n[测试 7: 事务处理]")

	// 成功的事务
	err := dbkit.Transaction(func(tx *dbkit.Tx) error {
		// 创建用户
		user := dbkit.NewRecord().
			Set("username", "TransUser").
			Set("email", "trans@example.com").
			Set("age", 30).
			Set("status", "active").
			Set("created_at", time.Now())
		uid, err := tx.Insert("users", user)
		if err != nil {
			return err
		}

		// 为用户创建订单
		order := dbkit.NewRecord().
			Set("user_id", uid).
			Set("amount", 999.99).
			Set("status", "COMPLETED").
			Set("created_at", time.Now())
		_, err = tx.Insert("orders", order)
		return err
	})
	if err != nil {
		log.Printf("  事务失败: %v", err)
	} else {
		fmt.Printf("  ✓ 事务成功: 用户和订单已创建\n")
	}

	// 事务中的链式查询
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		users, err := tx.Table("users").
			Where("status = ?", "active").
			Limit(3).
			Find()
		if err != nil {
			return err
		}
		fmt.Printf("  ✓ 事务中链式查询成功: %d 条记录\n", len(users))
		return nil
	})

	// 回滚测试 (故意失败)
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		user := dbkit.NewRecord().
			Set("username", "RollbackUser").
			Set("age", 25).
			Set("status", "active").
			Set("created_at", time.Now())
		_, err := tx.Insert("users", user)
		if err != nil {
			return err
		}
		// 故意返回错误触发回滚
		return fmt.Errorf("故意触发回滚")
	})
	if err != nil {
		fmt.Printf("  ✓ 事务回滚成功: %v\n", err)
	}
}

// ==================== 分页测试 ====================
func testPagination() {
	fmt.Println("\n[测试 8: 分页查询]")

	// 链式分页
	page1, err := dbkit.Table("orders").
		Where("status = ?", "COMPLETED").
		OrderBy("id DESC").
		Paginate(1, 5)
	if err != nil {
		log.Printf("  链式分页失败: %v", err)
	} else {
		fmt.Printf("  ✓ 链式分页成功:\n")
		fmt.Printf("    第 %d 页 / 共 %d 页, 总记录: %d\n", page1.PageNumber, page1.TotalPage, page1.TotalRow)
		for i := range page1.List {
			fmt.Printf("    - 订单 #%d: ¥%.2f\n", page1.List[i].GetInt("id"), page1.List[i].GetFloat("amount"))
		}
	}

	// 第二页
	page2, err := dbkit.Table("orders").
		Where("status = ?", "COMPLETED").
		OrderBy("id DESC").
		Paginate(2, 5)
	if err != nil {
		log.Printf("  第二页查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ 第二页查询成功: %d 条记录\n", len(page2.List))
	}

	// 原生分页
	page3, err := dbkit.Paginate(1, 10, "id, username, age", "users", "age > ?", "id ASC", 20)
	if err != nil {
		log.Printf("  原生分页失败: %v", err)
	} else {
		fmt.Printf("  ✓ 原生分页成功: 第 %d 页, 共 %d 条\n", page3.PageNumber, page3.TotalRow)
	}
}

// ==================== 缓存测试 ====================
func testCache() {
	fmt.Println("\n[测试 9: 缓存]")

	cacheName := "test_cache"

	// 第一次查询 (查数据库)
	start := time.Now()
	users1, err := dbkit.Table("users").
		Where("status = ?", "active").
		Cache(cacheName, 30*time.Second).
		Find()
	elapsed1 := time.Since(start)
	if err != nil {
		log.Printf("  第一次查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ 第一次查询 (数据库): %d 条, 耗时 %v\n", len(users1), elapsed1)
	}

	// 第二次查询 (应命中缓存)
	start = time.Now()
	users2, err := dbkit.Table("users").
		Where("status = ?", "active").
		Cache(cacheName, 30*time.Second).
		Find()
	elapsed2 := time.Since(start)
	if err != nil {
		log.Printf("  第二次查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ 第二次查询 (缓存): %d 条, 耗时 %v\n", len(users2), elapsed2)
	}

	// 手动缓存操作
	dbkit.CacheSet("manual_cache", "key1", "value1", 1*time.Minute)
	val, ok := dbkit.CacheGet("manual_cache", "key1")
	if ok {
		fmt.Printf("  ✓ 手动缓存 Get 成功: %v\n", val)
	}

	dbkit.CacheDelete("manual_cache", "key1")
	_, ok = dbkit.CacheGet("manual_cache", "key1")
	if !ok {
		fmt.Printf("  ✓ 手动缓存 Delete 成功\n")
	}

	// 缓存状态
	status := dbkit.CacheStatus()
	fmt.Printf("  ✓ 缓存状态: type=%v, items=%v\n", status["type"], status["total_items"])
}

// ==================== 批量操作测试 ====================
func testBatchOperations() {
	fmt.Println("\n[测试 10: 批量操作]")

	// 批量插入
	var records []*dbkit.Record
	for i := 0; i < 10; i++ {
		r := dbkit.NewRecord().
			Set("username", fmt.Sprintf("BatchUser%d", i)).
			Set("email", fmt.Sprintf("batch%d@example.com", i)).
			Set("age", 20+i).
			Set("status", "active").
			Set("created_at", time.Now())
		records = append(records, r)
	}
	affected, err := dbkit.BatchInsertDefault("users", records)
	if err != nil {
		log.Printf("  批量插入失败: %v", err)
	} else {
		fmt.Printf("  ✓ 批量插入成功: %d 条记录\n", affected)
	}

	// 批量更新
	var updateRecords []*dbkit.Record
	users, _ := dbkit.Table("users").
		Where("username LIKE ?", "BatchUser%").
		Limit(5).
		Find()
	for i := range users {
		r := dbkit.NewRecord().
			Set("id", users[i].GetInt64("id")).
			Set("age", users[i].GetInt("age")+10)
		updateRecords = append(updateRecords, r)
	}
	if len(updateRecords) > 0 {
		affected, err = dbkit.BatchUpdateDefault("users", updateRecords)
		if err != nil {
			log.Printf("  批量更新失败: %v", err)
		} else {
			fmt.Printf("  ✓ 批量更新成功: %d 条记录\n", affected)
		}
	}

	// 批量删除 (by IDs)
	ids := []interface{}{}
	delUsers, _ := dbkit.Table("users").
		Where("username LIKE ?", "BatchUser%").
		Limit(3).
		Find()
	for i := range delUsers {
		ids = append(ids, delUsers[i].GetInt64("id"))
	}
	if len(ids) > 0 {
		affected, err = dbkit.BatchDeleteByIdsDefault("users", ids)
		if err != nil {
			log.Printf("  批量删除失败: %v", err)
		} else {
			fmt.Printf("  ✓ 批量删除成功: %d 条记录\n", affected)
		}
	}
}

// ==================== 初始化表结构 ====================
func setupTables() {
	fmt.Println("\n[初始化] 创建测试表...")

	// 删除旧表
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

	fmt.Println("  ✓ 表创建完成")
}

// ==================== 准备测试数据 ====================
func prepareData() {
	fmt.Println("\n[初始化] 插入测试数据...")

	// 插入用户
	users := []struct {
		username string
		email    string
		age      int
		status   string
	}{
		{"Alice", "alice@example.com", 25, "active"},
		{"Bob", "bob@example.com", 30, "active"},
		{"Charlie", "charlie@example.com", 35, "active"},
		{"David", "david@example.com", 40, "active"},
		{"Eve", "eve@example.com", 28, "inactive"},
		{"Frank", "frank@example.com", 45, "active"},
	}

	for _, u := range users {
		record := dbkit.NewRecord().
			Set("username", u.username).
			Set("email", u.email).
			Set("age", u.age).
			Set("status", u.status).
			Set("created_at", time.Now())
		dbkit.Insert("users", record)
	}

	// 插入产品
	products := []struct {
		name  string
		price float64
		stock int
	}{
		{"iPhone 15", 7999.00, 100},
		{"MacBook Pro", 14999.00, 50},
		{"AirPods Pro", 1999.00, 200},
		{"iPad Air", 4799.00, 80},
	}

	for _, p := range products {
		record := dbkit.NewRecord().
			Set("name", p.name).
			Set("price", p.price).
			Set("stock", p.stock).
			Set("created_at", time.Now())
		dbkit.Insert("products", record)
	}

	// 插入订单和订单项
	orderData := []struct {
		userID int64
		amount float64
		status string
		items  []struct {
			productID int64
			quantity  int
			price     float64
		}
	}{
		{1, 9998.00, "COMPLETED", []struct {
			productID int64
			quantity  int
			price     float64
		}{{1, 1, 7999.00}, {3, 1, 1999.00}}},
		{1, 14999.00, "COMPLETED", []struct {
			productID int64
			quantity  int
			price     float64
		}{{2, 1, 14999.00}}},
		{2, 4799.00, "PENDING", []struct {
			productID int64
			quantity  int
			price     float64
		}{{4, 1, 4799.00}}},
		{2, 1999.00, "COMPLETED", []struct {
			productID int64
			quantity  int
			price     float64
		}{{3, 1, 1999.00}}},
		{3, 7999.00, "COMPLETED", []struct {
			productID int64
			quantity  int
			price     float64
		}{{1, 1, 7999.00}}},
		{4, 3998.00, "PENDING", []struct {
			productID int64
			quantity  int
			price     float64
		}{{3, 2, 1999.00}}},
	}

	for _, o := range orderData {
		orderRecord := dbkit.NewRecord().
			Set("user_id", o.userID).
			Set("amount", o.amount).
			Set("status", o.status).
			Set("created_at", time.Now())
		orderID, _ := dbkit.Insert("orders", orderRecord)

		for _, item := range o.items {
			itemRecord := dbkit.NewRecord().
				Set("order_id", orderID).
				Set("product_id", item.productID).
				Set("quantity", item.quantity).
				Set("price", item.price)
			dbkit.Insert("order_items", itemRecord)
		}
	}

	fmt.Println("  ✓ 测试数据插入完成")
}

// ==================== DbModel 测试 ====================
func testDbModel() {
	fmt.Println("\n[测试 11: DbModel 模型操作]")

	// 使用 Model 插入
	user := &models.User{
		Username:  "ModelUser",
		Email:     "model@example.com",
		Age:       32,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	id, err := user.Insert()
	if err != nil {
		log.Printf("  Model Insert 失败: %v", err)
	} else {
		user.ID = id
		fmt.Printf("  ✓ Model Insert 成功, ID: %d\n", id)
	}

	// 使用 Model 查询
	userModel := &models.User{}
	foundUser, err := userModel.FindFirst("username = ?", "ModelUser")
	if err != nil {
		log.Printf("  Model FindFirst 失败: %v", err)
	} else if foundUser != nil {
		fmt.Printf("  ✓ Model FindFirst 成功: %s (age: %d)\n", foundUser.Username, foundUser.Age)
	}

	// 使用 Model 更新
	if foundUser != nil {
		foundUser.Age = 33
		affected, err := foundUser.Update()
		if err != nil {
			log.Printf("  Model Update 失败: %v", err)
		} else {
			fmt.Printf("  ✓ Model Update 成功, 影响行数: %d\n", affected)
		}
	}

	// 使用 Model Find 查询多条
	users, err := userModel.Find("status = ?", "id DESC", "active")
	if err != nil {
		log.Printf("  Model Find 失败: %v", err)
	} else {
		fmt.Printf("  ✓ Model Find 成功: %d 条记录\n", len(users))
	}

	// 使用 Model 分页
	page, err := userModel.Paginate(1, 5, "status = ?", "id DESC", "active")
	if err != nil {
		log.Printf("  Model Paginate 失败: %v", err)
	} else {
		fmt.Printf("  ✓ Model Paginate 成功: 第 %d 页, 共 %d 条\n", page.PageNumber, page.TotalRow)
	}

	// 使用 Model 带缓存查询
	cachedUsers, err := userModel.Cache("user_cache", 30*time.Second).Find("age > ?", "id ASC", 25)
	if err != nil {
		log.Printf("  Model Cache Find 失败: %v", err)
	} else {
		fmt.Printf("  ✓ Model Cache Find 成功: %d 条记录\n", len(cachedUsers))
	}

	// 使用 Model Save (更新已存在记录)
	if foundUser != nil {
		foundUser.Age = 34
		_, err := foundUser.Save()
		if err != nil {
			log.Printf("  Model Save 失败: %v", err)
		} else {
			fmt.Printf("  ✓ Model Save 成功\n")
		}
	}

	// 使用 Model ToJson
	if foundUser != nil {
		json := foundUser.ToJson()
		fmt.Printf("  ✓ Model ToJson: %s\n", json[:min(len(json), 80)]+"...")
	}

	// Product Model 测试
	product := &models.Product{
		Name:      "Test Product",
		Price:     199.99,
		Stock:     50,
		CreatedAt: time.Now(),
	}
	pid, err := product.Insert()
	if err != nil {
		log.Printf("  Product Insert 失败: %v", err)
	} else {
		fmt.Printf("  ✓ Product Insert 成功, ID: %d\n", pid)
	}

	// Order Model 测试
	order := &models.Order{
		UserID:    user.ID,
		Amount:    199.99,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}
	oid, err := order.Insert()
	if err != nil {
		log.Printf("  Order Insert 失败: %v", err)
	} else {
		fmt.Printf("  ✓ Order Insert 成功, ID: %d\n", oid)
	}

	// OrderItem Model 测试
	orderItem := &models.OrderItem{
		OrderID:   oid,
		ProductID: pid,
		Quantity:  2,
		Price:     199.99,
	}
	iid, err := orderItem.Insert()
	if err != nil {
		log.Printf("  OrderItem Insert 失败: %v", err)
	} else {
		fmt.Printf("  ✓ OrderItem Insert 成功, ID: %d\n", iid)
	}

	// 使用 Model Delete
	if foundUser != nil {
		affected, err := foundUser.Delete()
		if err != nil {
			log.Printf("  Model Delete 失败: %v", err)
		} else {
			fmt.Printf("  ✓ Model Delete 成功, 影响行数: %d\n", affected)
		}
	}
}

// 辅助函数
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
