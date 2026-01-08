package main

import (
	"fmt"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
)

func main() {
	fmt.Println("=== 本机缓存 (LocalCache) + MySQL 演示 ===")

	// 1. 初始化数据库连接
	fmt.Println("\n[1] 连接 MySQL 数据库...")
	logFilePath := filepath.Join(".", "log.log")
	dbkit.InitLoggerWithFile("debug", logFilePath)
	dbkit.OpenDatabase(dbkit.MySQL, "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 25)
	defer dbkit.Close()

	if err := dbkit.Ping(); err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		return
	}
	fmt.Println("✓ 数据库连接成功")

	// 准备测试数据
	dbkit.Exec("DROP TABLE IF EXISTS test_users")
	if _, err := dbkit.Exec("CREATE TABLE test_users (id INT PRIMARY KEY, name VARCHAR(50), email VARCHAR(100), age INT)"); err != nil {
		fmt.Printf("创建表失败: %v\n", err)
	}

	// 插入更多测试数据
	for i := 1; i <= 100; i++ {
		dbkit.Exec("INSERT INTO test_users (id, name, email, age) VALUES (?, ?, ?, ?)",
			i, fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@example.com", i), 20+i%50)
	}
	fmt.Println("✓ 测试数据准备完成 (100 条记录)")

	// 2. 设置默认过期时间为 5 秒
	dbkit.SetDefaultTtl(5 * time.Second)

	// ========== 性能测试：展示优化效果 ==========

	fmt.Println("【性能测试】LocalCache 类型断言优化效果")

	// 测试 1：单条记录查询（类型断言路径）
	fmt.Println("\n[测试 1] 单条记录查询 - 类型断言优化")
	fmt.Println("说明：LocalCache 直接存储 Go 对象，读取时使用类型断言（零序列化开销）")

	// 第一次查询：从数据库读取
	start := time.Now()
	user, _ := dbkit.Cache("user_cache").QueryFirst("SELECT * FROM test_users WHERE id = ?", 1)
	dbLoadTime := time.Since(start)
	fmt.Printf("  第 1 次查询 (从数据库): 耗时 %v\n", dbLoadTime)

	fmt.Printf("  结果: id=%v, name=%v\n", user.GetInt("id"), user.GetString("name"))

	// 第二次查询：从缓存读取（类型断言路径，极快）
	start = time.Now()
	userCached, _ := dbkit.Cache("user_cache").QueryFirst("SELECT * FROM test_users WHERE id = ?", 1)
	cacheLoadTime := time.Since(start)
	fmt.Printf("  第 2 次查询 (从缓存): 耗时 %v\n", cacheLoadTime)
	fmt.Printf("  结果: id=%v, name=%v\n", userCached.GetInt("id"), userCached.GetString("name"))

	if dbLoadTime > 0 && cacheLoadTime > 0 {
		speedup := float64(dbLoadTime) / float64(cacheLoadTime)
		fmt.Printf("  ⚡ 性能提升: %.1fx 倍 (缓存比数据库快)\n", speedup)
	}

	// 测试 2：多条记录查询
	fmt.Println("\n[测试 2] 多条记录查询 - 批量数据缓存")
	fmt.Println("说明：查询 50 条记录，测试批量数据的缓存性能")

	start = time.Now()
	users, _ := dbkit.Cache("users_list").Query("SELECT * FROM test_users WHERE age > ? LIMIT 50", 30)
	dbLoadTime = time.Since(start)
	fmt.Printf("  第 1 次查询 (从数据库): 返回 %d 条记录, 耗时 %v\n", len(users), dbLoadTime)

	start = time.Now()
	usersCached, _ := dbkit.Cache("users_list").Query("SELECT * FROM test_users WHERE age > ? LIMIT 50", 30)
	cacheLoadTime = time.Since(start)
	fmt.Printf("  第 2 次查询 (从缓存): 返回 %d 条记录, 耗时 %v\n", len(usersCached), cacheLoadTime)

	if dbLoadTime > 0 && cacheLoadTime > 0 {
		speedup := float64(dbLoadTime) / float64(cacheLoadTime)
		fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", speedup)
	}

	// 测试 3：分页查询
	fmt.Println("\n[测试 3] 分页查询 - Page 对象缓存")
	fmt.Println("说明：分页查询会缓存 Page 对象，包含数据和分页信息")

	start = time.Now()
	page, _ := dbkit.Cache("page_cache").Paginate(1, 10, "SELECT * FROM test_users ORDER BY id")
	dbLoadTime = time.Since(start)
	fmt.Printf("  第 1 次分页 (从数据库): 第 %d 页, 共 %d 条, 耗时 %v\n",
		page.PageNumber, page.TotalRow, dbLoadTime)

	start = time.Now()
	pageCached, _ := dbkit.Cache("page_cache").Paginate(1, 10, "SELECT * FROM test_users ORDER BY id")
	cacheLoadTime = time.Since(start)
	fmt.Printf("  第 2 次分页 (从缓存): 第 %d 页, 共 %d 条, 耗时 %v\n",
		pageCached.PageNumber, pageCached.TotalRow, cacheLoadTime)

	if dbLoadTime > 0 && cacheLoadTime > 0 {
		speedup := float64(dbLoadTime) / float64(cacheLoadTime)
		fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", speedup)
	}

	// 测试 4：高频访问压力测试
	fmt.Println("\n[测试 4] 高频访问压力测试 - 1000 次缓存读取")
	fmt.Println("说明：模拟高并发场景，测试缓存的稳定性和性能")

	// 先预热缓存
	dbkit.Cache("stress_test").QueryFirst("SELECT * FROM test_users WHERE id = ?", 50)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		dbkit.Cache("stress_test").QueryFirst("SELECT * FROM test_users WHERE id = ?", 50)
	}
	totalTime := time.Since(start)
	avgTime := totalTime / 1000

	fmt.Printf("  完成 1000 次缓存读取\n")
	fmt.Printf("  总耗时: %v\n", totalTime)
	fmt.Printf("  平均每次: %v\n", avgTime)
	fmt.Printf("  QPS: %.0f 次/秒\n", float64(1000)/totalTime.Seconds())

	// ========== 基础功能演示 ==========

	fmt.Println("【基础功能演示】")

	// 3. 基础的 CacheGet/Set 操作
	fmt.Println("\n[功能 1] 基础缓存操作:")
	dbkit.CacheSet("user_store", "user_1", "张三")
	if val, ok := dbkit.CacheGet("user_store", "user_1"); ok {
		fmt.Printf("  ✓ 成功获取缓存: %v\n", val)
	}

	// 4. 查看缓存状态
	fmt.Println("\n[功能 2] 缓存状态:")
	status := dbkit.CacheStatus()
	fmt.Printf("  缓存类型: %v\n", status["type"])
	fmt.Printf("  存储库数量: %v\n", status["store_count"])
	fmt.Printf("  总缓存项数量: %v\n", status["total_items"])
	fmt.Printf("  预估内存占用: %v (%v bytes)\n", status["estimated_memory_human"], status["estimated_memory_bytes"])
	fmt.Printf("  清理间隔: %v\n", status["cleanup_interval"])

	// 5. 测试过期
	fmt.Println("\n[功能 3] 测试过期 (等待 6 秒)...")
	time.Sleep(6 * time.Second)
	if _, ok := dbkit.CacheGet("user_store", "user_1"); !ok {
		fmt.Println("  ✓ 缓存已按预期过期")
	}

	// 6. 为特定库预设 TTL
	fmt.Println("\n[功能 4] 为特定库预设 TTL (2 秒):")
	dbkit.CreateCache("short_lived", 2*time.Second)
	dbkit.CacheSet("short_lived", "temp_key", "瞬时数据")

	val, _ := dbkit.CacheGet("short_lived", "temp_key")
	fmt.Printf("  立即获取: %v\n", val)

	time.Sleep(3 * time.Second)
	if _, ok := dbkit.CacheGet("short_lived", "temp_key"); !ok {
		fmt.Println("  ✓ 预设 TTL 数据已过期")
	}

	// 7. 手动删除和清理
	fmt.Println("\n[功能 5] 手动删除和清理:")
	dbkit.CacheSet("user_store", "user_2", "李四")
	dbkit.CacheDelete("user_store", "user_2")
	if _, ok := dbkit.CacheGet("user_store", "user_2"); !ok {
		fmt.Println("  ✓ user_2 已手动删除")
	}

	dbkit.CacheSet("user_store", "user_3", "王五")
	dbkit.CacheClearRepository("user_store")
	if _, ok := dbkit.CacheGet("user_store", "user_3"); !ok {
		fmt.Println("  ✓ user_store 已全部清空")
	}
	fmt.Println("【总结】")
	fmt.Println("✓ LocalCache 使用类型断言优化，避免 JSON 序列化")
	fmt.Println("✓ 缓存命中时性能提升 10-100 倍（相比数据库查询）")
	fmt.Println("✓ 零序列化开销，内存占用低，适合高频访问场景")
	fmt.Println("\n=== 演示完成 ===")
}
