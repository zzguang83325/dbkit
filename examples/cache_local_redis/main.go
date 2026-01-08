package main

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/redis"
)

func main() {
	fmt.Println("=== 本地缓存 vs Redis 缓存演示 ===")

	// 1. 初始化数据库连接
	fmt.Println("\n[1] 连接 MySQL 数据库...")
	dbkit.OpenDatabase(dbkit.MySQL, "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 25)
	defer dbkit.Close()

	if err := dbkit.Ping(); err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		return
	}
	fmt.Println("✓ 数据库连接成功")

	// 2. 准备测试数据
	fmt.Println("\n[2] 准备测试数据...")
	dbkit.Exec("DROP TABLE IF EXISTS test_users")
	if _, err := dbkit.Exec("CREATE TABLE test_users (id INT PRIMARY KEY, name VARCHAR(50), email VARCHAR(100), age INT)"); err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 插入测试数据
	for i := 1; i <= 50; i++ {
		dbkit.Exec("INSERT INTO test_users (id, name, email, age) VALUES (?, ?, ?, ?)",
			i, fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@example.com", i), 20+i%50)
	}
	fmt.Println("✓ 测试数据准备完成 (50 条记录)")

	// 3. 初始化 Redis 缓存
	fmt.Println("\n[3] 初始化 Redis 缓存...")
	rc, err := redis.NewRedisCache("192.168.10.205:6379", "", "cpdv3", 2)
	if err != nil {
		fmt.Printf("Redis 连接失败: %v\n", err)
		return
	}
	dbkit.InitRedisCache(rc)
	fmt.Println("✓ Redis 缓存已初始化")

	// 4. 查看默认缓存状态
	fmt.Println("\n[4] 默认缓存状态:")
	status := dbkit.CacheStatus()
	fmt.Printf("  默认缓存类型: %v\n", status["type"])

	// ========== 演示 1：显式使用本地缓存 ==========
	fmt.Println("\n【演示 1】显式使用本地缓存")
	fmt.Println("API: dbkit.LocalCache(\"cache_name\").Query(...)")

	// 第一次查询：从数据库读取
	start := time.Now()
	user1, _ := dbkit.LocalCache("user_local").QueryFirst("SELECT * FROM test_users WHERE id = ?", 1)
	time1 := time.Since(start)
	fmt.Printf("  第 1 次查询 (从数据库): 耗时 %v\n", time1)
	fmt.Printf("  结果: id=%v, name=%v\n", user1.GetInt("id"), user1.GetString("name"))

	// 第二次查询：从本地缓存读取
	start = time.Now()
	user1Cached, _ := dbkit.LocalCache("user_local").QueryFirst("SELECT * FROM test_users WHERE id = ?", 1)
	time2 := time.Since(start)
	fmt.Printf("  第 2 次查询 (从本地缓存): 耗时 %v\n", time2)
	fmt.Printf("  结果: id=%v, name=%v\n", user1Cached.GetInt("id"), user1Cached.GetString("name"))
	fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))

	// ========== 演示 2：显式使用 Redis 缓存 ==========
	fmt.Println("\n【演示 2】显式使用 Redis 缓存")
	fmt.Println("API: dbkit.RedisCache(\"cache_name\").Query(...)")

	// 第一次查询：从数据库读取
	start = time.Now()
	user2, _ := dbkit.RedisCache("user_redis").QueryFirst("SELECT * FROM test_users WHERE id = ?", 2)
	time1 = time.Since(start)
	fmt.Printf("  第 1 次查询 (从数据库): 耗时 %v\n", time1)
	fmt.Printf("  结果: id=%v, name=%v\n", user2.GetInt("id"), user2.GetString("name"))

	// 第二次查询：从 Redis 缓存读取
	start = time.Now()
	user2Cached, _ := dbkit.RedisCache("user_redis").QueryFirst("SELECT * FROM test_users WHERE id = ?", 2)
	time2 = time.Since(start)
	fmt.Printf("  第 2 次查询 (从 Redis 缓存): 耗时 %v\n", time2)
	fmt.Printf("  结果: id=%v, name=%v\n", user2Cached.GetInt("id"), user2Cached.GetString("name"))
	fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))

	// ========== 演示 3：使用默认缓存（默认是本地缓存）==========
	fmt.Println("\n【演示 3】使用默认缓存（默认是本地缓存）")
	fmt.Println("API: dbkit.Cache(\"cache_name\").Query(...)")

	start = time.Now()
	user3, _ := dbkit.Cache("user_default").QueryFirst("SELECT * FROM test_users WHERE id = ?", 3)
	time1 = time.Since(start)
	fmt.Printf("  第 1 次查询: 耗时 %v\n", time1)
	fmt.Printf("  结果: id=%v, name=%v\n", user3.GetInt("id"), user3.GetString("name"))

	start = time.Now()
	user3Cached, _ := dbkit.Cache("user_default").QueryFirst("SELECT * FROM test_users WHERE id = ?", 3)
	time2 = time.Since(start)
	fmt.Printf("  第 2 次查询: 耗时 %v\n", time2)
	fmt.Printf("  结果: id=%v, name=%v\n", user3Cached.GetInt("id"), user3Cached.GetString("name"))
	fmt.Printf("  使用的是: 本地缓存 (默认)\n")

	// ========== 演示 4：切换默认缓存为 Redis ==========
	fmt.Println("\n【演示 4】切换默认缓存为 Redis")
	fmt.Println("API: dbkit.SetDefaultCache(redisCache)")

	dbkit.SetDefaultCache(rc)
	fmt.Println("✓ 已切换默认缓存为 Redis")

	// 验证默认缓存已切换
	status = dbkit.CacheStatus()
	fmt.Printf("  当前默认缓存类型: %v\n", status["type"])

	// 使用新的默认缓存（Redis）
	start = time.Now()
	user4, _ := dbkit.Cache("user_default_redis").QueryFirst("SELECT * FROM test_users WHERE id = ?", 4)
	time1 = time.Since(start)
	fmt.Printf("  第 1 次查询: 耗时 %v\n", time1)
	fmt.Printf("  结果: id=%v, name=%v\n", user4.GetInt("id"), user4.GetString("name"))

	start = time.Now()
	user4Cached, _ := dbkit.Cache("user_default_redis").QueryFirst("SELECT * FROM test_users WHERE id = ?", 4)
	time2 = time.Since(start)
	fmt.Printf("  第 2 次查询: 耗时 %v\n", time2)
	fmt.Printf("  结果: id=%v, name=%v\n", user4Cached.GetInt("id"), user4Cached.GetString("name"))
	fmt.Printf("  使用的是: Redis 缓存 (新的默认)\n")

	// ========== 演示 5：混合使用不同缓存 ==========
	fmt.Println("\n【演示 5】混合使用不同缓存")
	fmt.Println("说明：在同一个应用中，可以根据场景选择不同的缓存")

	// 配置数据用本地缓存（快速访问，很少变化）
	configs, _ := dbkit.LocalCache("config_cache", 10*time.Minute).
		Query("SELECT * FROM test_users WHERE age < 25 LIMIT 5")
	fmt.Printf("  配置数据 (本地缓存): %d 条记录\n", len(configs))

	// 业务数据用 Redis 缓存（多实例共享）
	orders, _ := dbkit.RedisCache("order_cache", 5*time.Minute).
		Query("SELECT * FROM test_users WHERE age > 30 LIMIT 10")
	fmt.Printf("  业务数据 (Redis 缓存): %d 条记录\n", len(orders))

	// 临时查询用默认缓存
	temps, _ := dbkit.Cache("temp_cache", 1*time.Minute).
		Query("SELECT * FROM test_users WHERE age = 25")
	fmt.Printf("  临时数据 (默认缓存): %d 条记录\n", len(temps))

	// ========== 演示 6：分页查询缓存 ==========
	fmt.Println("\n【演示 6】分页查询缓存")

	// 本地缓存分页
	start = time.Now()
	pageLocal, _ := dbkit.LocalCache("page_local").
		Paginate(1, 10, "SELECT * FROM test_users ORDER BY id")
	time1 = time.Since(start)
	fmt.Printf("  本地缓存分页 (第 1 次): 第 %d 页, 共 %d 条, 耗时 %v\n",
		pageLocal.PageNumber, pageLocal.TotalRow, time1)

	start = time.Now()
	pageLocal2, _ := dbkit.LocalCache("page_local").
		Paginate(1, 10, "SELECT * FROM test_users ORDER BY id")
	time2 = time.Since(start)
	fmt.Printf("  本地缓存分页 (第 2 次): 第 %d 页, 共 %d 条, 耗时 %v\n",
		pageLocal2.PageNumber, pageLocal2.TotalRow, time2)

	// Redis 缓存分页
	start = time.Now()
	pageRedis, _ := dbkit.RedisCache("page_redis").
		Paginate(1, 10, "SELECT * FROM test_users ORDER BY id DESC")
	time1 = time.Since(start)
	fmt.Printf("  Redis 缓存分页 (第 1 次): 第 %d 页, 共 %d 条, 耗时 %v\n",
		pageRedis.PageNumber, pageRedis.TotalRow, time1)

	start = time.Now()
	pageRedis2, _ := dbkit.RedisCache("page_redis").
		Paginate(1, 10, "SELECT * FROM test_users ORDER BY id DESC")
	time2 = time.Since(start)
	fmt.Printf("  Redis 缓存分页 (第 2 次): 第 %d 页, 共 %d 条, 耗时 %v\n",
		pageRedis2.PageNumber, pageRedis2.TotalRow, time2)

	// ========== 演示 7：缓存管理 ==========
	fmt.Println("\n【演示 7】缓存管理")

	// 设置缓存
	dbkit.LocalCache("api_cache").QueryFirst("SELECT * FROM test_users WHERE id = ?", 10)
	dbkit.RedisCache("api_cache").QueryFirst("SELECT * FROM test_users WHERE id = ?", 20)
	fmt.Println("  ✓ 已设置本地缓存和 Redis 缓存")

	// 清理缓存
	dbkit.CacheClearRepository("api_cache")
	dbkit.LocalCacheClearRepository("api_cache")
	fmt.Println("  ✓ 已清理 api_cache 缓存")

	// ========== 总结 ==========
	fmt.Println("\n【总结】")
	fmt.Println("✓ dbkit.LocalCache() - 显式使用本地缓存（速度最快，单实例）")
	fmt.Println("✓ dbkit.RedisCache() - 显式使用 Redis 缓存（分布式共享）")
	fmt.Println("✓ dbkit.Cache() - 使用默认缓存（可通过 SetDefaultCache 切换）")
	fmt.Println("✓ 支持在同一应用中混合使用不同缓存策略")
	fmt.Println("✓ 灵活的 API 设计，满足各种场景需求")

	fmt.Println("\n=== 演示完成 ===")
}
