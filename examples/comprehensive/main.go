package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/comprehensive/models"
)

// main is the entry point of the comprehensive DBKit API test example.
// 主函数：DBKit 综合 API 测试示例的入口点
func main() {
	// 1. Initialize database connection - MySQL
	// 1. 初始化数据库连接 - MySQL
	dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabase(dbkit.MySQL, dsn, 10)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer dbkit.Close()

	// Enable Debug mode to output SQL statements
	// 开启 Debug 模式输出 SQL
	dbkit.SetDebugMode(true)

	fmt.Println("\n" + repeat("=", 60))
	fmt.Println("DBKit Comprehensive API Test Example")
	fmt.Println("DBKit 综合 API 测试示例")
	fmt.Println(repeat("=", 60))

	// 2. Initialize environment (create tables)
	// 2. 初始化环境 (创建表)
	setupTables()

	// 3. Prepare test data
	// 3. 准备测试数据
	prepareData()

	// Run all tests
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
	fmt.Println("All tests completed!")
	fmt.Println("所有测试完成!")
	fmt.Println(repeat("=", 60))
}

// ==================== Basic CRUD Test ====================
// ==================== 基础 CRUD 操作测试 ====================
func testBasicCRUD() {
	fmt.Println("\n[Test 1: Basic CRUD Operations]")
	fmt.Println("\n[测试 1: 基础 CRUD 操作]")

	// Insert - Create a new record
	// Insert - 创建新记录
	user := dbkit.NewRecord().
		Set("username", "TestUser").
		Set("email", "test@example.com").
		Set("age", 25).
		Set("status", "active").
		Set("created_at", time.Now())
	id, err := dbkit.Insert("users", user)
	if err != nil {
		log.Printf("  Insert failed: %v", err)
	} else {
		fmt.Printf("  ✓ Insert successful, ID: %d\n", id)
	}

	// Query - Retrieve a single record
	// Query - 查询单条记录
	record, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		log.Printf("  QueryFirst failed: %v", err)
	} else {
		fmt.Printf("  ✓ QueryFirst successful: username=%s, age=%d\n", record.Str("username"), record.Int("age"))
	}

	// Update - Modify an existing record
	// Update - 修改现有记录
	record.Set("age", 26)
	affected, err := dbkit.Update("users", record, "id = ?", id)
	if err != nil {
		log.Printf("  Update failed: %v", err)
	} else {
		fmt.Printf("  ✓ Update successful, affected rows: %d\n", affected)
	}

	// Save - Smart insert or update (update if exists, insert if not)
	// Save - 智能插入或更新（存在则更新，不存在则插入）
	record.Set("age", 27)
	_, err = dbkit.Save("users", record)
	if err != nil {
		log.Printf("  Save failed: %v", err)
	} else {
		fmt.Printf("  ✓ Save successful (update)\n")
	}

	// Count - Count records matching condition
	// Count - 统计符合条件的记录
	count, err := dbkit.Count("users", "status = ?", "active")
	if err != nil {
		log.Printf("  Count failed: %v", err)
	} else {
		fmt.Printf("  ✓ Count successful: %d records\n", count)
	}

	// Exists - Check if record exists
	// Exists - 检查记录是否存在
	exists, err := dbkit.Exists("users", "username = ?", "TestUser")
	if err != nil {
		log.Printf("  Exists failed: %v", err)
	} else {
		fmt.Printf("  ✓ Exists successful: %v\n", exists)
	}
}

// ==================== Chain Query Test ====================
// ==================== 链式查询测试 ====================
func testChainQuery() {
	fmt.Println("\n[Test 2: Chain Query (QueryBuilder)]")
	fmt.Println("\n[测试 2: 链式查询 (QueryBuilder)]")

	// Basic chain query with multiple conditions
	// 基本链式查询，包含多个条件
	users, err := dbkit.Table("users").
		Select("id, username, age, status").
		Where("age > ?", 20).
		Where("status = ?", "active").
		OrderBy("age DESC").
		Limit(5).
		Find()
	if err != nil {
		log.Printf("  Chain query failed: %v", err)
	} else {
		fmt.Printf("  ✓ Basic chain query successful, returned %d records\n", len(users))
		for i := range users {
			fmt.Printf("    - %s (age: %d)\n", users[i].Str("username"), users[i].Int("age"))
		}
	}

	// FindFirst - Get the first matching record
	// FindFirst - 获取第一条匹配的记录
	user, err := dbkit.Table("users").
		Where("status = ?", "active").
		OrderBy("id DESC").
		FindFirst()
	if err != nil {
		log.Printf("  FindFirst failed: %v", err)
	} else if user != nil {
		fmt.Printf("  ✓ FindFirst successful: %s\n", user.Str("username"))
	}

	// Offset - Skip records and get next batch
	// Offset - 跳过记录并获取下一批
	users2, err := dbkit.Table("users").
		OrderBy("id ASC").
		Limit(3).
		Offset(2).
		Find()
	if err != nil {
		log.Printf("  Offset query failed: %v", err)
	} else {
		fmt.Printf("  ✓ Offset query successful, returned %d records\n", len(users2))
	}
}

// ==================== Advanced WHERE Conditions Test ====================
// ==================== 高级 WHERE 条件测试 ====================
func testAdvancedWhere() {
	fmt.Println("\n[Test 3: Advanced WHERE Conditions]")
	fmt.Println("\n[测试 3: 高级 WHERE 条件]")

	// OrWhere - OR condition
	// OrWhere - OR 条件
	users, err := dbkit.Table("users").
		Where("status = ?", "active").
		OrWhere("age > ?", 40).
		Find()
	if err != nil {
		log.Printf("  OrWhere failed: %v", err)
	} else {
		fmt.Printf("  ✓ OrWhere successful: %d records\n", len(users))
	}

	// WhereGroup / OrWhereGroup - Group conditions with parentheses
	// WhereGroup / OrWhereGroup - 用括号分组条件
	users2, err := dbkit.Table("users").
		Where("status = ?", "active").
		OrWhereGroup(func(qb *dbkit.QueryBuilder) *dbkit.QueryBuilder {
			return qb.Where("age > ?", 30).Where("age < ?", 50)
		}).
		Find()
	if err != nil {
		log.Printf("  WhereGroup failed: %v", err)
	} else {
		fmt.Printf("  ✓ OrWhereGroup successful: %d records\n", len(users2))
	}

	// WhereInValues - IN query with value list
	// WhereInValues - 值列表 IN 查询
	users3, err := dbkit.Table("users").
		WhereInValues("age", []interface{}{25, 30, 35, 40}).
		Find()
	if err != nil {
		log.Printf("  WhereInValues failed: %v", err)
	} else {
		fmt.Printf("  ✓ WhereInValues successful: %d records\n", len(users3))
	}

	// WhereNotInValues - NOT IN query
	// WhereNotInValues - 值列表 NOT IN 查询
	users4, err := dbkit.Table("users").
		WhereNotInValues("status", []interface{}{"banned", "deleted"}).
		Find()
	if err != nil {
		log.Printf("  WhereNotInValues failed: %v", err)
	} else {
		fmt.Printf("  ✓ WhereNotInValues successful: %d records\n", len(users4))
	}

	// WhereBetween - Range query
	// WhereBetween - 范围查询
	users5, err := dbkit.Table("users").
		WhereBetween("age", 25, 35).
		Find()
	if err != nil {
		log.Printf("  WhereBetween failed: %v", err)
	} else {
		fmt.Printf("  ✓ WhereBetween successful: %d records\n", len(users5))
	}

	// WhereNotBetween - Exclude range
	// WhereNotBetween - 排除范围
	users6, err := dbkit.Table("users").
		WhereNotBetween("age", 20, 25).
		Find()
	if err != nil {
		log.Printf("  WhereNotBetween failed: %v", err)
	} else {
		fmt.Printf("  ✓ WhereNotBetween successful: %d records\n", len(users6))
	}

	// WhereNotNull - IS NOT NULL check
	// WhereNotNull - IS NOT NULL 检查
	users7, err := dbkit.Table("users").
		WhereNotNull("email").
		Find()
	if err != nil {
		log.Printf("  WhereNotNull failed: %v", err)
	} else {
		fmt.Printf("  ✓ WhereNotNull successful: %d records\n", len(users7))
	}

	// WhereNull - IS NULL check
	// WhereNull - IS NULL 检查
	users8, err := dbkit.Table("users").
		WhereNull("deleted_at").
		Find()
	if err != nil {
		log.Printf("  WhereNull failed: %v", err)
	} else {
		fmt.Printf("  ✓ WhereNull successful: %d records\n", len(users8))
	}
}

// ==================== JOIN Query Test ====================
// ==================== JOIN 查询测试 ====================
func testJoinQuery() {
	fmt.Println("\n[Test 4: JOIN Query]")
	fmt.Println("\n[测试 4: JOIN 查询]")

	// LEFT JOIN - Include all records from left table
	// LEFT JOIN - 包含左表的所有记录
	records, err := dbkit.Table("users").
		Select("users.username, orders.amount, orders.status as order_status").
		LeftJoin("orders", "users.id = orders.user_id").
		Where("orders.amount > ?", 100).
		OrderBy("orders.amount DESC").
		Limit(5).
		Find()
	if err != nil {
		log.Printf("  LEFT JOIN failed: %v", err)
	} else {
		fmt.Printf("  ✓ LEFT JOIN successful: %d records\n", len(records))
		for i := range records {
			fmt.Printf("    - %s: ¥%.2f (%s)\n", records[i].Str("username"), records[i].Float("amount"), records[i].Str("order_status"))
		}
	}

	// INNER JOIN - Multiple table join
	// INNER JOIN - 多表连接
	records2, err := dbkit.Table("orders").
		Select("orders.id, users.username, products.name as product_name, order_items.quantity").
		InnerJoin("users", "orders.user_id = users.id").
		InnerJoin("order_items", "orders.id = order_items.order_id").
		InnerJoin("products", "order_items.product_id = products.id").
		Where("orders.status = ?", "COMPLETED").
		Limit(5).
		Find()
	if err != nil {
		log.Printf("  Multiple INNER JOIN failed: %v", err)
	} else {
		fmt.Printf("  ✓ Multiple INNER JOIN successful: %d records\n", len(records2))
	}

	// RIGHT JOIN - Include all records from right table
	// RIGHT JOIN - 包含右表的所有记录
	records3, err := dbkit.Table("users").
		Select("users.username, orders.amount").
		RightJoin("orders", "users.id = orders.user_id").
		Limit(5).
		Find()
	if err != nil {
		log.Printf("  RIGHT JOIN failed: %v", err)
	} else {
		fmt.Printf("  ✓ RIGHT JOIN successful: %d records\n", len(records3))
	}

	// JOIN with parameters - Add conditions in JOIN clause
	// JOIN with parameters - 在 JOIN 子句中添加条件
	records4, err := dbkit.Table("users").
		Select("users.username, orders.amount").
		Join("orders", "users.id = orders.user_id AND orders.status = ?", "COMPLETED").
		Find()
	if err != nil {
		log.Printf("  JOIN with parameters failed: %v", err)
	} else {
		fmt.Printf("  ✓ JOIN with parameters successful: %d records\n", len(records4))
	}
}

// ==================== Subquery Test ====================
// ==================== 子查询测试 ====================
func testSubquery() {
	fmt.Println("\n[Test 5: Subquery]")
	fmt.Println("\n[测试 5: 子查询 (Subquery)]")

	// WhereIn with Subquery - Find users with completed orders
	// WhereIn with Subquery - 查找有已完成订单的用户
	activeUsersSub := dbkit.NewSubquery().
		Table("orders").
		Select("DISTINCT user_id").
		Where("status = ?", "COMPLETED")

	users, err := dbkit.Table("users").
		Select("id, username").
		WhereIn("id", activeUsersSub).
		Find()
	if err != nil {
		log.Printf("  WhereIn subquery failed: %v", err)
	} else {
		fmt.Printf("  ✓ WhereIn subquery successful: %d records (users with completed orders)\n", len(users))
	}

	// WhereNotIn with Subquery - Find orders from non-banned users
	// WhereNotIn with Subquery - 查找非禁用用户的订单
	bannedUsersSub := dbkit.NewSubquery().
		Table("users").
		Select("id").
		Where("status = ?", "banned")

	orders, err := dbkit.Table("orders").
		WhereNotIn("user_id", bannedUsersSub).
		Find()
	if err != nil {
		log.Printf("  WhereNotIn subquery failed: %v", err)
	} else {
		fmt.Printf("  ✓ WhereNotIn subquery successful: %d records\n", len(orders))
	}

	// FROM Subquery - Query from derived table
	// FROM Subquery - 从派生表查询
	userTotalsSub := dbkit.NewSubquery().
		Table("orders").
		Select("user_id, SUM(amount) as total_spent, COUNT(*) as order_count")

	// Note: MySQL requires GROUP BY in subquery for aggregation
	// 注意：MySQL 在子查询中需要 GROUP BY 来进行聚合
	records, err := dbkit.Query(`
		SELECT user_id, total_spent, order_count 
		FROM (SELECT user_id, SUM(amount) as total_spent, COUNT(*) as order_count FROM orders GROUP BY user_id) AS user_totals 
		WHERE total_spent > ?`, 200)
	if err != nil {
		log.Printf("  FROM subquery failed: %v", err)
	} else {
		fmt.Printf("  ✓ FROM subquery successful: %d records\n", len(records))
	}
	_ = userTotalsSub // Use variable to avoid warning / 使用变量避免警告

	// SELECT Subquery - Add subquery result as column
	// SELECT Subquery - 将子查询结果作为列添加
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
		log.Printf("  SELECT subquery failed: %v", err)
	} else {
		fmt.Printf("  ✓ SELECT subquery successful: %d records\n", len(users2))
		for i := range users2 {
			fmt.Printf("    - %s: %d orders\n", users2[i].Str("username"), users2[i].Int("order_count"))
		}
	}
}

// ==================== GROUP BY / HAVING Test ====================
// ==================== GROUP BY / HAVING 测试 ====================
func testGroupByHaving() {
	fmt.Println("\n[Test 6: GROUP BY / HAVING]")
	fmt.Println("\n[测试 6: GROUP BY / HAVING]")

	// Basic GroupBy - Group records and aggregate
	// 基本 GroupBy - 分组记录并聚合
	stats, err := dbkit.Table("orders").
		Select("status, COUNT(*) as count, SUM(amount) as total_amount").
		GroupBy("status").
		Find()
	if err != nil {
		log.Printf("  GroupBy failed: %v", err)
	} else {
		fmt.Printf("  ✓ GroupBy successful:\n")
		for i := range stats {
			fmt.Printf("    - %s: %d orders, total ¥%.2f\n", stats[i].Str("status"), stats[i].Int("count"), stats[i].Float("total_amount"))
		}
	}

	// GroupBy + Having - Filter groups by aggregate condition
	// GroupBy + Having - 按聚合条件过滤分组
	userStats, err := dbkit.Table("orders").
		Select("user_id, COUNT(*) as order_count, SUM(amount) as total_spent").
		GroupBy("user_id").
		Having("COUNT(*) >= ?", 2).
		Find()
	if err != nil {
		log.Printf("  GroupBy + Having failed: %v", err)
	} else {
		fmt.Printf("  ✓ GroupBy + Having successful: %d users with 2+ orders\n", len(userStats))
	}

	// Multiple Having conditions - Apply multiple filters
	// 多个 Having 条件 - 应用多个过滤器
	userStats2, err := dbkit.Table("orders").
		Select("user_id, COUNT(*) as cnt, SUM(amount) as total").
		GroupBy("user_id").
		Having("COUNT(*) >= ?", 1).
		Having("SUM(amount) > ?", 100).
		Find()
	if err != nil {
		log.Printf("  Multiple Having conditions failed: %v", err)
	} else {
		fmt.Printf("  ✓ Multiple Having conditions successful: %d records\n", len(userStats2))
	}

	// GroupBy multiple columns - Group by multiple fields
	// GroupBy 多列 - 按多个字段分组
	stats2, err := dbkit.Table("orders").
		Select("user_id, status, COUNT(*) as count").
		GroupBy("user_id, status").
		OrderBy("user_id, status").
		Find()
	if err != nil {
		log.Printf("  Multiple column GroupBy failed: %v", err)
	} else {
		fmt.Printf("  ✓ Multiple column GroupBy successful: %d records\n", len(stats2))
	}
}

// ==================== Transaction Test ====================
// ==================== 事务测试 ====================
func testTransaction() {
	fmt.Println("\n[Test 7: Transaction Handling]")
	fmt.Println("\n[测试 7: 事务处理]")

	// Successful transaction - Auto commit on success
	// 成功的事务 - 成功时自动提交
	err := dbkit.Transaction(func(tx *dbkit.Tx) error {
		// Create user
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

		// Create order for user
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
		log.Printf("  Transaction failed: %v", err)
	} else {
		fmt.Printf("  ✓ Transaction successful: user and order created\n")
	}

	// Chain query in transaction
	// 事务中的链式查询
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		users, err := tx.Table("users").
			Where("status = ?", "active").
			Limit(3).
			Find()
		if err != nil {
			return err
		}
		fmt.Printf("  ✓ Chain query in transaction successful: %d records\n", len(users))
		return nil
	})

	// Rollback test - Intentionally fail to trigger rollback
	// 回滚测试 - 故意失败以触发回滚
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
		// Intentionally return error to trigger rollback
		// 故意返回错误触发回滚
		return fmt.Errorf("intentionally trigger rollback")
	})
	if err != nil {
		fmt.Printf("  ✓ Transaction rollback successful: %v\n", err)
	}
}

// ==================== Pagination Test ====================
// ==================== 分页测试 ====================
func testPagination() {
	fmt.Println("\n[Test 8: Pagination Query]")
	fmt.Println("\n[测试 8: 分页查询]")

	// Chain pagination - Paginate with chain query
	// 链式分页 - 使用链式查询进行分页
	page1, err := dbkit.Table("orders").
		Where("status = ?", "COMPLETED").
		OrderBy("id DESC").
		Paginate(1, 5)
	if err != nil {
		log.Printf("  Chain pagination failed: %v", err)
	} else {
		fmt.Printf("  ✓ Chain pagination successful:\n")
		fmt.Printf("    Page %d / Total %d pages, Total records: %d\n", page1.PageNumber, page1.TotalPage, page1.TotalRow)
		for i := range page1.List {
			fmt.Printf("    - Order #%d: ¥%.2f\n", page1.List[i].GetInt("id"), page1.List[i].GetFloat("amount"))
		}
	}

	// Second page - Get next page of results
	// 第二页 - 获取下一页结果
	page2, err := dbkit.Table("orders").
		Where("status = ?", "COMPLETED").
		OrderBy("id DESC").
		Paginate(2, 5)
	if err != nil {
		log.Printf("  Second page query failed: %v", err)
	} else {
		fmt.Printf("  ✓ Second page query successful: %d records\n", len(page2.List))
	}

	// Native pagination - Use raw SQL pagination
	// 原生分页 - 使用原始 SQL 分页
	page3, err := dbkit.Paginate(1, 10, "id, username, age", "users", "age > ?", "id ASC", 20)
	if err != nil {
		log.Printf("  Native pagination failed: %v", err)
	} else {
		fmt.Printf("  ✓ Native pagination successful: Page %d, Total %d records\n", page3.PageNumber, page3.TotalRow)
	}
}

// ==================== Cache Test ====================
// ==================== 缓存测试 ====================
func testCache() {
	fmt.Println("\n[Test 9: Caching]")
	fmt.Println("\n[测试 9: 缓存]")

	cacheName := "test_cache"

	// First query (from database)
	// 第一次查询 (查数据库)
	start := time.Now()
	users1, err := dbkit.Table("users").
		Where("status = ?", "active").
		Cache(cacheName, 30*time.Second).
		Find()
	elapsed1 := time.Since(start)
	if err != nil {
		log.Printf("  First query failed: %v", err)
	} else {
		fmt.Printf("  ✓ First query (database): %d records, elapsed %v\n", len(users1), elapsed1)
	}

	// Second query (should hit cache)
	// 第二次查询 (应命中缓存)
	start = time.Now()
	users2, err := dbkit.Table("users").
		Where("status = ?", "active").
		Cache(cacheName, 30*time.Second).
		Find()
	elapsed2 := time.Since(start)
	if err != nil {
		log.Printf("  Second query failed: %v", err)
	} else {
		fmt.Printf("  ✓ Second query (cache): %d records, elapsed %v\n", len(users2), elapsed2)
	}

	// Manual cache operations
	// 手动缓存操作
	dbkit.CacheSet("manual_cache", "key1", "value1", 1*time.Minute)
	val, ok := dbkit.CacheGet("manual_cache", "key1")
	if ok {
		fmt.Printf("  ✓ Manual cache Get successful: %v\n", val)
	}

	dbkit.CacheDelete("manual_cache", "key1")
	_, ok = dbkit.CacheGet("manual_cache", "key1")
	if !ok {
		fmt.Printf("  ✓ Manual cache Delete successful\n")
	}

	// Cache status
	// 缓存状态
	status := dbkit.CacheStatus()
	fmt.Printf("  ✓ Cache status: type=%v, items=%v\n", status["type"], status["total_items"])
}

// ==================== Batch Operations Test ====================
// ==================== 批量操作测试 ====================
func testBatchOperations() {
	fmt.Println("\n[Test 10: Batch Operations]")
	fmt.Println("\n[测试 10: 批量操作]")

	// Batch insert - Insert multiple records at once
	// 批量插入 - 一次插入多条记录
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
		log.Printf("  Batch insert failed: %v", err)
	} else {
		fmt.Printf("  ✓ Batch insert successful: %d records\n", affected)
	}

	// Batch update - Update multiple records at once
	// 批量更新 - 一次更新多条记录
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
			log.Printf("  Batch update failed: %v", err)
		} else {
			fmt.Printf("  ✓ Batch update successful: %d records\n", affected)
		}
	}

	// Batch delete by IDs - Delete multiple records by ID list
	// 批量删除 (by IDs) - 按 ID 列表删除多条记录
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
			log.Printf("  Batch delete failed: %v", err)
		} else {
			fmt.Printf("  ✓ Batch delete successful: %d records\n", affected)
		}
	}
}

// ==================== Initialize Table Structure ====================
// ==================== 初始化表结构 ====================
func setupTables() {
	fmt.Println("\n[Initialize] Creating test tables...")
	fmt.Println("\n[初始化] 创建测试表...")

	// Drop old tables
	// 删除旧表
	dbkit.Exec("DROP TABLE IF EXISTS order_items")
	dbkit.Exec("DROP TABLE IF EXISTS orders")
	dbkit.Exec("DROP TABLE IF EXISTS products")
	dbkit.Exec("DROP TABLE IF EXISTS users")

	// Users table
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
		log.Printf("Failed to create users table: %v", err)
	}

	// Products table
	// 产品表
	_, err = dbkit.Exec(`CREATE TABLE products (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		price DECIMAL(10,2) DEFAULT 0,
		stock INT DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Printf("Failed to create products table: %v", err)
	}

	// Orders table
	// 订单表
	_, err = dbkit.Exec(`CREATE TABLE orders (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		amount DECIMAL(10,2) DEFAULT 0,
		status VARCHAR(20) DEFAULT 'PENDING',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Printf("Failed to create orders table: %v", err)
	}

	// Order items table
	// 订单项表
	_, err = dbkit.Exec(`CREATE TABLE order_items (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		order_id BIGINT NOT NULL,
		product_id BIGINT NOT NULL,
		quantity INT DEFAULT 1,
		price DECIMAL(10,2) DEFAULT 0
	)`)
	if err != nil {
		log.Printf("Failed to create order_items table: %v", err)
	}

	fmt.Println("  ✓ Tables created successfully")
}

// ==================== Prepare Test Data ====================
// ==================== 准备测试数据 ====================
func prepareData() {
	fmt.Println("\n[Initialize] Inserting test data...")
	fmt.Println("\n[初始化] 插入测试数据...")

	// Insert users
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

	// Insert products
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

	// Insert orders and order items
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

	fmt.Println("  ✓ Test data inserted successfully")
}

// ==================== Auto Timestamps Test ====================
// ==================== 自动时间戳测试 ====================
func testAutoTimestamps() {
	fmt.Println("\n[Test 11: Auto Timestamps]")
	fmt.Println("\n[测试 11: 自动时间戳 (Auto Timestamps)]")

	// Enable timestamp auto-update
	// 启用时间戳自动更新
	dbkit.EnableTimestamps()
	fmt.Println("  ✓ Timestamp auto-update enabled")
	fmt.Println("  ✓ 已启用时间戳自动更新")

	// Create table with timestamp fields
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
		log.Printf("  Failed to create articles table: %v", err)
		return
	}

	// Configure auto timestamps (using default field names)
	// 配置自动时间戳（使用默认字段名）
	dbkit.ConfigTimestamps("articles")
	fmt.Println("  ✓ Auto timestamps configured (created_at, updated_at)")
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

// ==================== Soft Delete Test ====================
// ==================== 软删除测试 ====================
func testSoftDelete() {
	fmt.Println("\n[Test 12: Soft Delete]")
	fmt.Println("\n[测试 12: 软删除 (Soft Delete)]")

	// Enable soft delete functionality
	// 启用软删除功能
	dbkit.EnableSoftDelete()
	fmt.Println("  ✓ Soft delete enabled")
	fmt.Println("  ✓ 已启用软删除功能")

	// Create table with soft delete field
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
		log.Printf("  Failed to create documents table: %v", err)
		return
	}

	// Configure soft delete
	// 配置软删除
	dbkit.ConfigSoftDelete("documents", "deleted_at")
	fmt.Println("  ✓ Soft delete configured (deleted_at)")
	fmt.Println("  ✓ 已配置软删除 (deleted_at)")

	// Insert test data
	// 插入测试数据
	for i := 1; i <= 5; i++ {
		doc := dbkit.NewRecord().
			Set("title", fmt.Sprintf("Document %d", i)).
			Set("content", fmt.Sprintf("Content of document %d", i)).
			Set("status", "published")
		dbkit.Insert("documents", doc)
	}
	fmt.Println("  ✓ Inserted 5 documents")
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

// ==================== Optimistic Lock Test ====================
// ==================== 乐观锁测试 ====================
func testOptimisticLock() {
	fmt.Println("\n[Test 13: Optimistic Lock]")
	fmt.Println("\n[测试 13: 乐观锁 (Optimistic Lock)]")

	// Enable optimistic lock functionality
	// 启用乐观锁功能
	dbkit.EnableOptimisticLock()
	fmt.Println("  ✓ Optimistic lock enabled")
	fmt.Println("  ✓ 已启用乐观锁功能")

	// Create table with version field
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
		log.Printf("  Failed to create inventory table: %v", err)
		return
	}

	// Configure optimistic lock (using default field name version)
	// 配置乐观锁（使用默认字段名 version）
	dbkit.ConfigOptimisticLock("inventory")
	fmt.Println("  ✓ Optimistic lock configured (version)")
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

// ==================== DbModel Test ====================
// ==================== DbModel 测试 ====================
func testDbModel() {
	fmt.Println("\n[Test 14: DbModel Operations]")
	fmt.Println("\n[测试 14: DbModel 模型操作]")

	// Insert using Model
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
		log.Printf("  Model Insert failed: %v", err)
	} else {
		user.ID = id
		fmt.Printf("  ✓ Model Insert successful, ID: %d\n", id)
	}

	// Query using Model
	// 使用 Model 查询
	userModel := &models.User{}
	foundUser, err := userModel.FindFirst("username = ?", "ModelUser")
	if err != nil {
		log.Printf("  Model FindFirst failed: %v", err)
	} else if foundUser != nil {
		fmt.Printf("  ✓ Model FindFirst successful: %s (age: %d)\n", foundUser.Username, foundUser.Age)
	}

	// Update using Model
	// 使用 Model 更新
	if foundUser != nil {
		foundUser.Age = 33
		affected, err := foundUser.Update()
		if err != nil {
			log.Printf("  Model Update failed: %v", err)
		} else {
			fmt.Printf("  ✓ Model Update successful, affected rows: %d\n", affected)
		}
	}

	// Find multiple records using Model
	// 使用 Model Find 查询多条
	users, err := userModel.Find("status = ?", "id DESC", "active")
	if err != nil {
		log.Printf("  Model Find failed: %v", err)
	} else {
		fmt.Printf("  ✓ Model Find successful: %d records\n", len(users))
	}

	// Pagination using Model
	// 使用 Model 分页
	page, err := userModel.Paginate(1, 5, "status = ?", "id DESC", "active")
	if err != nil {
		log.Printf("  Model Paginate failed: %v", err)
	} else {
		fmt.Printf("  ✓ Model Paginate successful: Page %d, Total %d records\n", page.PageNumber, page.TotalRow)
	}

	// Cache query using Model
	// 使用 Model 带缓存查询
	cachedUsers, err := userModel.Cache("user_cache", 30*time.Second).Find("age > ?", "id ASC", 25)
	if err != nil {
		log.Printf("  Model Cache Find failed: %v", err)
	} else {
		fmt.Printf("  ✓ Model Cache Find successful: %d records\n", len(cachedUsers))
	}

	// Save using Model (update existing record)
	// 使用 Model Save (更新已存在记录)
	if foundUser != nil {
		foundUser.Age = 34
		_, err := foundUser.Save()
		if err != nil {
			log.Printf("  Model Save failed: %v", err)
		} else {
			fmt.Printf("  ✓ Model Save successful\n")
		}
	}

	// Convert Model to JSON
	// 使用 Model ToJson
	if foundUser != nil {
		json := foundUser.ToJson()
		fmt.Printf("  ✓ Model ToJson: %s\n", json[:min(len(json), 80)]+"...")
	}

	// Product Model test
	// Product Model 测试
	product := &models.Product{
		Name:      "Test Product",
		Price:     199.99,
		Stock:     50,
		CreatedAt: time.Now(),
	}
	pid, err := product.Insert()
	if err != nil {
		log.Printf("  Product Insert failed: %v", err)
	} else {
		fmt.Printf("  ✓ Product Insert successful, ID: %d\n", pid)
	}

	// Order Model test
	// Order Model 测试
	order := &models.Order{
		UserID:    user.ID,
		Amount:    199.99,
		Status:    "PENDING",
		CreatedAt: time.Now(),
	}
	oid, err := order.Insert()
	if err != nil {
		log.Printf("  Order Insert failed: %v", err)
	} else {
		fmt.Printf("  ✓ Order Insert successful, ID: %d\n", oid)
	}

	// OrderItem Model test
	// OrderItem Model 测试
	orderItem := &models.OrderItem{
		OrderID:   oid,
		ProductID: pid,
		Quantity:  2,
		Price:     199.99,
	}
	iid, err := orderItem.Insert()
	if err != nil {
		log.Printf("  OrderItem Insert failed: %v", err)
	} else {
		fmt.Printf("  ✓ OrderItem Insert successful, ID: %d\n", iid)
	}

	// Delete using Model
	// 使用 Model Delete
	if foundUser != nil {
		affected, err := foundUser.Delete()
		if err != nil {
			log.Printf("  Model Delete failed: %v", err)
		} else {
			fmt.Printf("  ✓ Model Delete successful, affected rows: %d\n", affected)
		}
	}
}

// Helper function - Repeat string n times
// 辅助函数 - 重复字符串 n 次
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// Helper function - Return minimum of two integers
// 辅助函数 - 返回两个整数的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
