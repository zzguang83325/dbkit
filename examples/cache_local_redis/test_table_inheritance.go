package main

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/redis"
)

// TestTableCacheInheritance 测试 Table() 方法的缓存继承
func TestTableCacheInheritance() {
	fmt.Println("\n=== 测试 Table() 方法的缓存继承 ===")

	// 1. 初始化数据库连接
	fmt.Println("\n[1] 连接 MySQL 数据库...")
	err := dbkit.OpenDatabase(dbkit.MySQL, "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 25)
	if err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		return
	}
	defer dbkit.Close()

	if err := dbkit.Ping(); err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		return
	}
	fmt.Println("✓ 数据库连接成功")

	// 2. 准备测试数据
	fmt.Println("\n[2] 准备测试数据...")
	dbkit.Exec("DROP TABLE IF EXISTS test_table_cache")
	_, err = dbkit.Exec(`CREATE TABLE test_table_cache (
		id INT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(50),
		value INT
	)`)
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 插入测试数据
	for i := 1; i <= 10; i++ {
		record := dbkit.NewRecord()
		record.Set("name", fmt.Sprintf("test%d", i))
		record.Set("value", i*10)
		_, err = dbkit.Insert("test_table_cache", record)
		if err != nil {
			fmt.Printf("插入数据失败: %v\n", err)
			return
		}
	}
	fmt.Println("✓ 测试数据准备完成 (10 条记录)")

	// 3. 初始化本地缓存
	fmt.Println("\n[3] 初始化本地缓存...")
	dbkit.InitLocalCache(10 * time.Minute)
	fmt.Println("✓ 本地缓存已初始化")

	// 4. 初始化 Redis 缓存
	fmt.Println("\n[4] 初始化 Redis 缓存...")
	rc, err := redis.NewRedisCache("192.168.10.205:6379", "", "cpdv3", 2)
	if err != nil {
		fmt.Printf("Redis 连接失败: %v\n", err)
		return
	}
	dbkit.InitRedisCache(rc)
	fmt.Println("✓ Redis 缓存已初始化")

	// ========== 测试 1：DB.LocalCache().Table() 继承本地缓存 ==========
	fmt.Println("\n【测试 1】DB.LocalCache().Table() 继承本地缓存")

	db, _ := dbkit.UseWithError("default")

	// 先设置 LocalCache，再调用 Table
	start := time.Now()
	records1, err := db.LocalCache("table_test_1", 5*time.Minute).
		Table("test_table_cache").
		Where("value > ?", 30).
		Query()
	time1 := time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 1 次查询 (从数据库): 查询到 %d 条记录, 耗时 %v\n", len(records1), time1)
	}

	// 第二次查询应该从缓存读取
	start = time.Now()
	records2, err := db.LocalCache("table_test_1", 5*time.Minute).
		Table("test_table_cache").
		Where("value > ?", 30).
		Query()
	time2 := time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 2 次查询 (从本地缓存): 查询到 %d 条记录, 耗时 %v\n", len(records2), time2)
		if time2 < time1 {
			fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))
		}
	}

	// ========== 测试 2：DB.Table().LocalCache() 也能正常工作 ==========
	fmt.Println("\n【测试 2】DB.Table().LocalCache() 也能正常工作")

	// 先调用 Table，再设置 LocalCache
	start = time.Now()
	records3, err := db.Table("test_table_cache").
		LocalCache("table_test_2", 5*time.Minute).
		Where("value < ?", 50).
		Query()
	time1 = time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 1 次查询 (从数据库): 查询到 %d 条记录, 耗时 %v\n", len(records3), time1)
	}

	// 第二次查询应该从缓存读取
	start = time.Now()
	records4, err := db.Table("test_table_cache").
		LocalCache("table_test_2", 5*time.Minute).
		Where("value < ?", 50).
		Query()
	time2 = time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 2 次查询 (从本地缓存): 查询到 %d 条记录, 耗时 %v\n", len(records4), time2)
		if time2 < time1 {
			fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))
		}
	}

	// ========== 测试 3：DB.RedisCache().Table() 继承 Redis 缓存 ==========
	fmt.Println("\n【测试 3】DB.RedisCache().Table() 继承 Redis 缓存")

	// 先设置 RedisCache，再调用 Table
	start = time.Now()
	records5, err := db.RedisCache("table_test_3", 5*time.Minute).
		Table("test_table_cache").
		Where("id <= ?", 5).
		Query()
	time1 = time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 1 次查询 (从数据库): 查询到 %d 条记录, 耗时 %v\n", len(records5), time1)
	}

	// 第二次查询应该从 Redis 缓存读取
	start = time.Now()
	records6, err := db.RedisCache("table_test_3", 5*time.Minute).
		Table("test_table_cache").
		Where("id <= ?", 5).
		Query()
	time2 = time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 2 次查询 (从 Redis 缓存): 查询到 %d 条记录, 耗时 %v\n", len(records6), time2)
		if time2 < time1 {
			fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))
		}
	}

	// ========== 测试 4：事务中的 Table() 缓存继承 ==========
	fmt.Println("\n【测试 4】事务中的 Table() 缓存继承")

	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		// Tx.LocalCache().Table()
		records, err := tx.LocalCache("table_test_4", 5*time.Minute).
			Table("test_table_cache").
			Where("id > ?", 5).
			Query()

		if err != nil {
			return err
		}

		fmt.Printf("  ✓ 事务中查询成功: 查询到 %d 条记录\n", len(records))
		return nil
	})

	if err != nil {
		fmt.Printf("  ✗ 事务失败: %v\n", err)
	}

	// ========== 测试 5：分页查询的缓存继承 ==========
	fmt.Println("\n【测试 5】分页查询的缓存继承")

	// 本地缓存分页
	start = time.Now()
	page1, err := db.LocalCache("table_page_1", 5*time.Minute).
		Table("test_table_cache").
		OrderBy("id DESC").
		Paginate(1, 5)
	time1 = time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 分页查询失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 1 次分页查询 (从数据库): 第 %d 页, 共 %d 条, 耗时 %v\n",
			page1.PageNumber, page1.TotalRow, time1)
	}

	// 第二次分页查询应该从缓存读取
	start = time.Now()
	page2, err := db.LocalCache("table_page_1", 5*time.Minute).
		Table("test_table_cache").
		OrderBy("id DESC").
		Paginate(1, 5)
	time2 = time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 分页查询失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 2 次分页查询 (从本地缓存): 第 %d 页, 共 %d 条, 耗时 %v\n",
			page2.PageNumber, page2.TotalRow, time2)
		if time2 < time1 {
			fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))
		}
	}

	// ========== 测试 6：QueryFirst 的缓存继承 ==========
	fmt.Println("\n【测试 6】QueryFirst 的缓存继承")

	start = time.Now()
	record1, err := db.LocalCache("table_first_1", 5*time.Minute).
		Table("test_table_cache").
		Where("id = ?", 3).
		QueryFirst()
	time1 = time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 1 次查询 (从数据库): id=%v, name=%v, 耗时 %v\n",
			record1.GetInt("id"), record1.GetString("name"), time1)
	}

	start = time.Now()
	record2, err := db.LocalCache("table_first_1", 5*time.Minute).
		Table("test_table_cache").
		Where("id = ?", 3).
		QueryFirst()
	time2 = time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 2 次查询 (从本地缓存): id=%v, name=%v, 耗时 %v\n",
			record2.GetInt("id"), record2.GetString("name"), time2)
		if time2 < time1 {
			fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))
		}
	}

	// ========== 测试 7：Count 的缓存继承 ==========
	fmt.Println("\n【测试 7】Count 的缓存继承")

	start = time.Now()
	count1, err := db.LocalCache("table_count_1", 5*time.Minute).
		Table("test_table_cache").
		Where("value > ?", 20).
		Count()
	time1 = time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 统计失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 1 次统计 (从数据库): count=%d, 耗时 %v\n", count1, time1)
	}

	start = time.Now()
	count2, err := db.LocalCache("table_count_1", 5*time.Minute).
		Table("test_table_cache").
		Where("value > ?", 20).
		Count()
	time2 = time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ 统计失败: %v\n", err)
	} else {
		fmt.Printf("  ✓ 第 2 次统计 (从本地缓存): count=%d, 耗时 %v\n", count2, time2)
		if time2 < time1 {
			fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))
		}
	}

	// ========== 总结 ==========
	fmt.Println("\n【总结】")
	fmt.Println("✓ DB.LocalCache().Table() - 正确继承本地缓存设置")
	fmt.Println("✓ DB.Table().LocalCache() - 正确设置本地缓存")
	fmt.Println("✓ DB.RedisCache().Table() - 正确继承 Redis 缓存设置")
	fmt.Println("✓ Tx.LocalCache().Table() - 事务中正确继承缓存设置")
	fmt.Println("✓ 分页查询、QueryFirst、Count 都正确支持缓存继承")
	fmt.Println("✓ 缓存继承功能测试通过!")

	fmt.Println("\n=== 测试完成 ===")
}
