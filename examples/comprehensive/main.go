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
	testAutoTimestamps()
	testSoftDelete()
	testOptimisticLock()
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

// ==================== 自动时间戳测试 ====================
func testAutoTimestamps() {
	fmt.Println("\n[测试 11: 自动时间戳 (Auto Timestamps)]")

	// 启用时间戳检查
	dbkit.EnableTimestampCheck()
	fmt.Println("  ✓ 已启用时间戳自动更新")

	// 创建带时间戳的表
	dbkit.Exec("DROP TABLE IF EXISTS articles")
	_, err := dbkit.Exec(`CREATE TABLE articles (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(200) NOT NULL,
		content TEXT,
		author VARCHAR(100),
		created_at DATETIME NULL,
		updated_at DATETIME NULL
	)`)
	if err != nil {
		log.Printf("  创建 articles 表失败: %v", err)
		return
	}

	// 配置自动时间戳（使用默认字段名）
	dbkit.ConfigTimestamps("articles")
	fmt.Println("  ✓ 已配置自动时间戳 (created_at, updated_at)")

	// 测试 1: 插入数据（created_at 自动填充）
	article := dbkit.NewRecord().
		Set("title", "DBKit 入门教程").
		Set("content", "这是一篇关于 DBKit 的教程...").
		Set("author", "张三")
	// 注意：不设置 created_at，让它自动填充
	articleID, err := dbkit.Insert("articles", article)
	if err != nil {
		log.Printf("  插入文章失败: %v", err)
	} else {
		fmt.Printf("  ✓ 插入文章成功, ID: %d (created_at 自动填充)\n", articleID)
	}

	// 查询验证 created_at
	record, _ := dbkit.QueryFirst("SELECT * FROM articles WHERE id = ?", articleID)
	if record != nil {
		fmt.Printf("    - created_at: %v\n", record.Get("created_at"))
		fmt.Printf("    - updated_at: %v\n", record.Get("updated_at"))
	}

	// 等待1秒，让时间戳有明显差异
	time.Sleep(1 * time.Second)

	// 测试 2: 更新数据（updated_at 自动填充）
	updateRecord := dbkit.NewRecord().
		Set("content", "这是更新后的内容...")
	// 注意：不设置 updated_at，让它自动填充
	affected, err := dbkit.Update("articles", updateRecord, "id = ?", articleID)
	if err != nil {
		log.Printf("  更新文章失败: %v", err)
	} else {
		fmt.Printf("  ✓ 更新文章成功, 影响行数: %d (updated_at 自动填充)\n", affected)
	}

	// 查询验证 updated_at
	record2, _ := dbkit.QueryFirst("SELECT * FROM articles WHERE id = ?", articleID)
	if record2 != nil {
		fmt.Printf("    - created_at: %v (未变)\n", record2.Get("created_at"))
		fmt.Printf("    - updated_at: %v (已更新)\n", record2.Get("updated_at"))
	}

	// 测试 3: 手动指定 created_at（不会被覆盖）
	customTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	article2 := dbkit.NewRecord().
		Set("title", "历史文章").
		Set("content", "这是一篇历史文章").
		Set("author", "李四").
		Set("created_at", customTime) // 手动指定
	articleID2, err := dbkit.Insert("articles", article2)
	if err != nil {
		log.Printf("  插入历史文章失败: %v", err)
	} else {
		fmt.Printf("  ✓ 插入历史文章成功, ID: %d\n", articleID2)
	}

	record3, _ := dbkit.QueryFirst("SELECT * FROM articles WHERE id = ?", articleID2)
	if record3 != nil {
		fmt.Printf("    - created_at: %v (保持为 2020-01-01)\n", record3.Get("created_at"))
	}

	// 测试 4: 临时禁用自动时间戳
	time.Sleep(1 * time.Second)
	updateRecord2 := dbkit.NewRecord().
		Set("author", "王五")
	affected2, err := dbkit.Table("articles").
		Where("id = ?", articleID).
		WithoutTimestamps().
		Update(updateRecord2)
	if err != nil {
		log.Printf("  禁用时间戳更新失败: %v", err)
	} else {
		fmt.Printf("  ✓ 禁用时间戳更新成功, 影响行数: %d\n", affected2)
	}

	record4, _ := dbkit.QueryFirst("SELECT * FROM articles WHERE id = ?", articleID)
	if record4 != nil {
		fmt.Printf("    - updated_at: %v (未变化)\n", record4.Get("updated_at"))
	}

	// 测试 5: 使用自定义字段名
	dbkit.Exec("DROP TABLE IF EXISTS posts")
	_, err = dbkit.Exec(`CREATE TABLE posts (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(200),
		create_time DATETIME NULL,
		modify_time DATETIME NULL
	)`)
	if err == nil {
		dbkit.ConfigTimestampsWithFields("posts", "create_time", "modify_time")
		fmt.Println("  ✓ 配置自定义时间戳字段 (create_time, modify_time)")

		post := dbkit.NewRecord().Set("title", "测试帖子")
		postID, _ := dbkit.Insert("posts", post)
		postRecord, _ := dbkit.QueryFirst("SELECT * FROM posts WHERE id = ?", postID)
		if postRecord != nil {
			fmt.Printf("    - create_time: %v\n", postRecord.Get("create_time"))
		}
	}

	// 测试 6: 仅配置 created_at
	dbkit.Exec("DROP TABLE IF EXISTS logs")
	_, err = dbkit.Exec(`CREATE TABLE logs (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		message TEXT,
		log_time DATETIME NULL
	)`)
	if err == nil {
		dbkit.ConfigCreatedAt("logs", "log_time")
		fmt.Println("  ✓ 仅配置 created_at (log_time)")

		logRecord := dbkit.NewRecord().Set("message", "系统启动")
		logID, _ := dbkit.Insert("logs", logRecord)
		log, _ := dbkit.QueryFirst("SELECT * FROM logs WHERE id = ?", logID)
		if log != nil {
			fmt.Printf("    - log_time: %v\n", log.Get("log_time"))
		}
	}
}

// ==================== 软删除测试 ====================
func testSoftDelete() {
	fmt.Println("\n[测试 12: 软删除 (Soft Delete)]")

	// 启用软删除检查
	dbkit.EnableSoftDeleteCheck()
	fmt.Println("  ✓ 已启用软删除检查")

	// 创建带软删除字段的表
	dbkit.Exec("DROP TABLE IF EXISTS documents")
	_, err := dbkit.Exec(`CREATE TABLE documents (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(200) NOT NULL,
		content TEXT,
		status VARCHAR(20) DEFAULT 'draft',
		deleted_at DATETIME NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Printf("  创建 documents 表失败: %v", err)
		return
	}

	// 配置软删除
	dbkit.ConfigSoftDelete("documents", "deleted_at")
	fmt.Println("  ✓ 已配置软删除 (deleted_at)")

	// 插入测试数据
	for i := 1; i <= 5; i++ {
		doc := dbkit.NewRecord().
			Set("title", fmt.Sprintf("文档 %d", i)).
			Set("content", fmt.Sprintf("这是文档 %d 的内容", i)).
			Set("status", "published")
		dbkit.Insert("documents", doc)
	}
	fmt.Println("  ✓ 插入 5 条文档")

	// 测试 1: 软删除（自动更新 deleted_at）
	affected, err := dbkit.Delete("documents", "id = ?", 1)
	if err != nil {
		log.Printf("  软删除失败: %v", err)
	} else {
		fmt.Printf("  ✓ 软删除成功, 影响行数: %d\n", affected)
	}

	// 验证软删除
	doc1, _ := dbkit.QueryFirst("SELECT * FROM documents WHERE id = ?", 1)
	if doc1 != nil {
		fmt.Printf("    - deleted_at: %v (已标记删除)\n", doc1.Get("deleted_at"))
	}

	// 测试 2: 普通查询（自动过滤已删除记录）
	docs, err := dbkit.Table("documents").Find()
	if err != nil {
		log.Printf("  查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ 普通查询: %d 条记录 (自动过滤已删除)\n", len(docs))
	}

	// 测试 3: 查询包含已删除记录
	allDocs, err := dbkit.Table("documents").WithTrashed().Find()
	if err != nil {
		log.Printf("  WithTrashed 查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ WithTrashed 查询: %d 条记录 (包含已删除)\n", len(allDocs))
	}

	// 测试 4: 只查询已删除记录
	deletedDocs, err := dbkit.Table("documents").OnlyTrashed().Find()
	if err != nil {
		log.Printf("  OnlyTrashed 查询失败: %v", err)
	} else {
		fmt.Printf("  ✓ OnlyTrashed 查询: %d 条记录 (仅已删除)\n", len(deletedDocs))
	}

	// 测试 5: 恢复已删除记录
	affected2, err := dbkit.Restore("documents", "id = ?", 1)
	if err != nil {
		log.Printf("  恢复失败: %v", err)
	} else {
		fmt.Printf("  ✓ 恢复成功, 影响行数: %d\n", affected2)
	}

	// 验证恢复
	doc1After, _ := dbkit.QueryFirst("SELECT * FROM documents WHERE id = ?", 1)
	if doc1After != nil {
		fmt.Printf("    - deleted_at: %v (已恢复)\n", doc1After.Get("deleted_at"))
	}

	// 测试 6: 物理删除（真正删除数据）
	affected3, err := dbkit.ForceDelete("documents", "id = ?", 2)
	if err != nil {
		log.Printf("  物理删除失败: %v", err)
	} else {
		fmt.Printf("  ✓ 物理删除成功, 影响行数: %d\n", affected3)
	}

	// 验证物理删除
	doc2, _ := dbkit.QueryFirst("SELECT * FROM documents WHERE id = ?", 2)
	if doc2 == nil {
		fmt.Println("    - 记录已被物理删除")
	}

	// 测试 7: 链式调用软删除
	affected4, err := dbkit.Table("documents").Where("id = ?", 3).Delete()
	if err != nil {
		log.Printf("  链式软删除失败: %v", err)
	} else {
		fmt.Printf("  ✓ 链式软删除成功, 影响行数: %d\n", affected4)
	}

	// 测试 8: 链式调用恢复
	affected5, err := dbkit.Table("documents").Where("id = ?", 3).Restore()
	if err != nil {
		log.Printf("  链式恢复失败: %v", err)
	} else {
		fmt.Printf("  ✓ 链式恢复成功, 影响行数: %d\n", affected5)
	}

	// 测试 9: 链式调用物理删除
	affected6, err := dbkit.Table("documents").Where("id = ?", 4).ForceDelete()
	if err != nil {
		log.Printf("  链式物理删除失败: %v", err)
	} else {
		fmt.Printf("  ✓ 链式物理删除成功, 影响行数: %d\n", affected6)
	}

	// 测试 10: 统计（自动过滤已删除）
	count1, _ := dbkit.Table("documents").Count()
	fmt.Printf("  ✓ Count (过滤已删除): %d 条\n", count1)

	// 测试 11: 统计（包含已删除）
	count2, _ := dbkit.Table("documents").WithTrashed().Count()
	fmt.Printf("  ✓ Count (包含已删除): %d 条\n", count2)

	// 测试 12: 使用布尔类型软删除
	dbkit.Exec("DROP TABLE IF EXISTS tasks")
	_, err = dbkit.Exec(`CREATE TABLE tasks (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(200),
		is_deleted TINYINT(1) DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err == nil {
		dbkit.ConfigSoftDeleteWithType("tasks", "is_deleted", dbkit.SoftDeleteBool)
		fmt.Println("  ✓ 配置布尔类型软删除 (is_deleted)")

		task := dbkit.NewRecord().Set("title", "测试任务")
		taskID, _ := dbkit.Insert("tasks", task)
		dbkit.Delete("tasks", "id = ?", taskID)
		taskRecord, _ := dbkit.QueryFirst("SELECT * FROM tasks WHERE id = ?", taskID)
		if taskRecord != nil {
			fmt.Printf("    - is_deleted: %v\n", taskRecord.Get("is_deleted"))
		}
	}
}

// ==================== 乐观锁测试 ====================
func testOptimisticLock() {
	fmt.Println("\n[测试 13: 乐观锁 (Optimistic Lock)]")

	// 启用乐观锁检查
	dbkit.EnableOptimisticLockCheck()
	fmt.Println("  ✓ 已启用乐观锁检查")

	// 创建带版本字段的表
	dbkit.Exec("DROP TABLE IF EXISTS inventory")
	_, err := dbkit.Exec(`CREATE TABLE inventory (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		product_name VARCHAR(200) NOT NULL,
		stock INT DEFAULT 0,
		price DECIMAL(10,2) DEFAULT 0,
		version BIGINT DEFAULT 0,
		updated_at DATETIME NULL
	)`)
	if err != nil {
		log.Printf("  创建 inventory 表失败: %v", err)
		return
	}

	// 配置乐观锁（使用默认字段名 version）
	dbkit.ConfigOptimisticLock("inventory")
	fmt.Println("  ✓ 已配置乐观锁 (version)")

	// 测试 1: 插入数据（version 自动初始化为 1）
	product := dbkit.NewRecord().
		Set("product_name", "iPhone 15 Pro").
		Set("stock", 100).
		Set("price", 7999.00)
	// 注意：不设置 version，让它自动初始化
	productID, err := dbkit.Insert("inventory", product)
	if err != nil {
		log.Printf("  插入商品失败: %v", err)
	} else {
		fmt.Printf("  ✓ 插入商品成功, ID: %d (version 自动初始化为 1)\n", productID)
	}

	// 查询验证 version
	record, _ := dbkit.QueryFirst("SELECT * FROM inventory WHERE id = ?", productID)
	if record != nil {
		fmt.Printf("    - version: %v\n", record.Get("version"))
		fmt.Printf("    - stock: %v\n", record.Get("stock"))
	}

	// 测试 2: 正常更新（带正确版本号）
	currentVersion := record.GetInt64("version")
	updateRecord := dbkit.NewRecord().
		Set("version", currentVersion). // 设置当前版本
		Set("stock", 95)                // 减少库存
	affected, err := dbkit.Update("inventory", updateRecord, "id = ?", productID)
	if err != nil {
		log.Printf("  更新失败: %v", err)
	} else {
		fmt.Printf("  ✓ 更新成功, 影响行数: %d (version 自动递增为 %d)\n", affected, currentVersion+1)
	}

	// 查询验证 version 已递增
	record2, _ := dbkit.QueryFirst("SELECT * FROM inventory WHERE id = ?", productID)
	if record2 != nil {
		fmt.Printf("    - version: %v (已递增)\n", record2.Get("version"))
		fmt.Printf("    - stock: %v (已更新)\n", record2.Get("stock"))
	}

	// 测试 3: 并发冲突检测（使用过期版本）
	staleVersion := int64(1) // 过期的版本号
	staleRecord := dbkit.NewRecord().
		Set("version", staleVersion). // 使用过期版本
		Set("stock", 90)
	affected2, err := dbkit.Update("inventory", staleRecord, "id = ?", productID)
	if err != nil {
		if err == dbkit.ErrVersionMismatch {
			fmt.Printf("  ✓ 检测到版本冲突: %v\n", err)
		} else {
			log.Printf("  更新失败: %v", err)
		}
	} else {
		fmt.Printf("  ⚠ 更新成功但不应该成功, 影响行数: %d\n", affected2)
	}

	// 测试 4: 正确处理并发 - 先读取最新版本
	latestRecord, _ := dbkit.Table("inventory").Where("id = ?", productID).FindFirst()
	if latestRecord != nil {
		currentVer := latestRecord.GetInt64("version")
		fmt.Printf("  ✓ 读取最新版本: %d\n", currentVer)

		updateRecord2 := dbkit.NewRecord().
			Set("version", currentVer).
			Set("stock", 90)
		affected3, err := dbkit.Update("inventory", updateRecord2, "id = ?", productID)
		if err != nil {
			log.Printf("  更新失败: %v", err)
		} else {
			fmt.Printf("  ✓ 使用最新版本更新成功, 影响行数: %d\n", affected3)
		}
	}

	// 测试 5: 不带版本字段更新（跳过版本检查）
	noVersionRecord := dbkit.NewRecord().
		Set("price", 7899.00) // 只更新价格，不设置 version
	affected4, err := dbkit.Update("inventory", noVersionRecord, "id = ?", productID)
	if err != nil {
		log.Printf("  无版本更新失败: %v", err)
	} else {
		fmt.Printf("  ✓ 无版本字段更新成功, 影响行数: %d (跳过版本检查)\n", affected4)
	}

	// 测试 6: 事务中使用乐观锁
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		rec, err := tx.Table("inventory").Where("id = ?", productID).FindFirst()
		if err != nil {
			return err
		}

		currentVer := rec.GetInt64("version")
		updateRec := dbkit.NewRecord().
			Set("version", currentVer).
			Set("stock", 85)
		_, err = tx.Update("inventory", updateRec, "id = ?", productID)
		return err
	})
	if err != nil {
		log.Printf("  事务中乐观锁失败: %v", err)
	} else {
		fmt.Println("  ✓ 事务中乐观锁更新成功")
	}

	// 测试 7: 模拟并发场景
	fmt.Println("\n  模拟并发场景:")
	// 用户A读取数据
	userARecord, _ := dbkit.Table("inventory").Where("id = ?", productID).FindFirst()
	userAVersion := userARecord.GetInt64("version")
	fmt.Printf("    - 用户A读取: version=%d, stock=%d\n", userAVersion, userARecord.GetInt("stock"))

	// 用户B读取数据
	userBRecord, _ := dbkit.Table("inventory").Where("id = ?", productID).FindFirst()
	userBVersion := userBRecord.GetInt64("version")
	fmt.Printf("    - 用户B读取: version=%d, stock=%d\n", userBVersion, userBRecord.GetInt("stock"))

	// 用户A先更新
	userAUpdate := dbkit.NewRecord().
		Set("version", userAVersion).
		Set("stock", userARecord.GetInt("stock")-5)
	_, err = dbkit.Update("inventory", userAUpdate, "id = ?", productID)
	if err != nil {
		fmt.Printf("    - 用户A更新失败: %v\n", err)
	} else {
		fmt.Println("    - 用户A更新成功")
	}

	// 用户B尝试更新（使用过期版本）
	userBUpdate := dbkit.NewRecord().
		Set("version", userBVersion). // 此时版本已过期
		Set("stock", userBRecord.GetInt("stock")-3)
	_, err = dbkit.Update("inventory", userBUpdate, "id = ?", productID)
	if err != nil {
		if err == dbkit.ErrVersionMismatch {
			fmt.Printf("    - 用户B更新失败: %v (版本冲突)\n", err)
			fmt.Println("    - 用户B需要重新读取最新数据")
		} else {
			fmt.Printf("    - 用户B更新失败: %v\n", err)
		}
	} else {
		fmt.Println("    - 用户B更新成功")
	}

	// 测试 8: 使用自定义版本字段名
	dbkit.Exec("DROP TABLE IF EXISTS accounts")
	_, err = dbkit.Exec(`CREATE TABLE accounts (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(100),
		balance DECIMAL(10,2) DEFAULT 0,
		revision BIGINT DEFAULT 0
	)`)
	if err == nil {
		dbkit.ConfigOptimisticLockWithField("accounts", "revision")
		fmt.Println("\n  ✓ 配置自定义版本字段 (revision)")

		account := dbkit.NewRecord().
			Set("username", "testuser").
			Set("balance", 1000.00)
		accID, _ := dbkit.Insert("accounts", account)
		accRecord, _ := dbkit.QueryFirst("SELECT * FROM accounts WHERE id = ?", accID)
		if accRecord != nil {
			fmt.Printf("    - revision: %v (自动初始化)\n", accRecord.Get("revision"))
		}
	}
}

// ==================== DbModel 测试 ====================
func testDbModel() {
	fmt.Println("\n[测试 14: DbModel 模型操作]")

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
