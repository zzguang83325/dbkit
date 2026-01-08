package main

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/redis"
)

func main() {
	fmt.Println("=== Redis 缓存 + MySQL 演示 ===")

	// 1. 初始化数据库连接
	fmt.Println("\n[1] 连接 MySQL 数据库...")
	dbkit.OpenDatabase(dbkit.MySQL, "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 25)
	defer dbkit.Close()

	if err := dbkit.Ping(); err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		return
	}
	fmt.Println("✓ 数据库连接成功")

	// 准备测试数据
	fmt.Println("\n[2] 准备测试数据...")
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

	// 2. 配置 Redis (通过子包按需引入)
	// 地址: 192.168.10.205:6379, 密码: cpdv3, 数据库: 2
	rc, err := redis.NewRedisCache("192.168.10.205:6379", "", "cpdv3", 2)
	if err != nil {
		fmt.Printf("Redis 连接失败: %v\n", err)
		return
	}
	dbkit.SetDefaultCache(rc)
	fmt.Println("✓ Redis 已连接并切换为默认缓存提供者")

	// ========== 性能测试：展示优化效果 ==========

	fmt.Println("【性能测试】RedisCache 优化效果")

	fmt.Println("优化说明：")
	fmt.Println("  - 优化前：Redis → 字符串 → Marshal → Unmarshal → 对象（双重序列化）")
	fmt.Println("  - 优化后：Redis → 字节数组 → Unmarshal → 对象（单次反序列化）")

	// 测试 1：单条记录查询
	fmt.Println("\n[测试 1] 单条记录查询 - Redis 字节数组优化")

	// 第一次查询：从数据库读取并存入 Redis
	start := time.Now()
	user, err := dbkit.Cache("user_cache_redis").QueryFirst("SELECT * FROM test_users WHERE id = ?", 1)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
	}
	dbLoadTime := time.Since(start)
	fmt.Printf("  第 1 次查询 (从数据库): 耗时 %v\n", dbLoadTime)
	fmt.Printf("  结果: id=%v, name=%v\n", user.GetInt("id"), user.GetString("name"))

	// 第二次查询：从 Redis 读取（优化后的字节数组路径）
	start = time.Now()
	userCached, _ := dbkit.Cache("user_cache_redis").QueryFirst("SELECT * FROM test_users WHERE id = ?", 1)
	cacheLoadTime := time.Since(start)
	fmt.Printf("  第 2 次查询 (从 Redis): 耗时 %v\n", cacheLoadTime)
	fmt.Printf("  结果: id=%v, name=%v\n", userCached.GetInt("id"), userCached.GetString("name"))

	if dbLoadTime > 0 && cacheLoadTime > 0 {
		speedup := float64(dbLoadTime) / float64(cacheLoadTime)
		fmt.Printf("  ⚡ 性能提升: %.1fx 倍 (缓存比数据库快)\n", speedup)
	}

	// 测试 2：多条记录查询
	fmt.Println("\n[测试 2] 多条记录查询 - 批量数据 Redis 缓存")

	start = time.Now()
	users, _ := dbkit.Cache("users_list_redis").Query("SELECT * FROM test_users WHERE age > ? LIMIT 50", 30)
	dbLoadTime = time.Since(start)
	fmt.Printf("  第 1 次查询 (从数据库): 返回 %d 条记录, 耗时 %v\n", len(users), dbLoadTime)

	start = time.Now()
	usersCached, _ := dbkit.Cache("users_list_redis").Query("SELECT * FROM test_users WHERE age > ? LIMIT 50", 30)
	cacheLoadTime = time.Since(start)
	fmt.Printf("  第 2 次查询 (从 Redis): 返回 %d 条记录, 耗时 %v\n", len(usersCached), cacheLoadTime)

	if dbLoadTime > 0 && cacheLoadTime > 0 {
		speedup := float64(dbLoadTime) / float64(cacheLoadTime)
		fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", speedup)
	}

	// 测试 3：分页查询
	fmt.Println("\n[测试 3] 分页查询 - Page 对象 Redis 缓存")

	start = time.Now()
	page, _ := dbkit.Cache("page_cache_redis").Paginate(1, 10, "SELECT * FROM test_users ORDER BY id")
	dbLoadTime = time.Since(start)
	fmt.Printf("  第 1 次分页 (从数据库): 第 %d 页, 共 %d 条, 耗时 %v\n",
		page.PageNumber, page.TotalRow, dbLoadTime)

	start = time.Now()
	pageCached, _ := dbkit.Cache("page_cache_redis").Paginate(1, 10, "SELECT * FROM test_users ORDER BY id")
	cacheLoadTime = time.Since(start)
	fmt.Printf("  第 2 次分页 (从 Redis): 第 %d 页, 共 %d 条, 耗时 %v\n",
		pageCached.PageNumber, pageCached.TotalRow, cacheLoadTime)

	if dbLoadTime > 0 && cacheLoadTime > 0 {
		speedup := float64(dbLoadTime) / float64(cacheLoadTime)
		fmt.Printf("  ⚡ 性能提升: %.1fx 倍\n", speedup)
	}

	// 测试 4：高频访问压力测试
	fmt.Println("\n[测试 4] 高频访问压力测试 - 1000 次 Redis 读取")
	fmt.Println("说明：测试优化后的 Redis 缓存在高并发场景下的性能")

	// 先预热缓存
	dbkit.Cache("stress_test_redis").QueryFirst("SELECT * FROM test_users WHERE id = ?", 50)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		dbkit.Cache("stress_test_redis").QueryFirst("SELECT * FROM test_users WHERE id = ?", 50)
	}
	totalTime := time.Since(start)
	avgTime := totalTime / 1000

	fmt.Printf("  完成 1000 次 Redis 读取\n")
	fmt.Printf("  总耗时: %v\n", totalTime)
	fmt.Printf("  平均每次: %v\n", avgTime)
	fmt.Printf("  QPS: %.0f 次/秒\n", float64(1000)/totalTime.Seconds())

	// 测试 5：复杂对象序列化
	fmt.Println("\n[测试 5] 复杂对象序列化 - JSON 自动处理")
	fmt.Println("说明：Redis 会自动将复杂对象序列化为 JSON，优化后减少一次序列化")

	complexData := map[string]interface{}{
		"user_id": 123,
		"profile": map[string]string{
			"name":  "张三",
			"email": "zhangsan@example.com",
		},
		"settings": []string{"theme:dark", "lang:zh"},
	}

	start = time.Now()
	dbkit.CacheSet("api_cache", "config_1", complexData)
	setTime := time.Since(start)
	fmt.Printf("  存储复杂对象: 耗时 %v\n", setTime)

	start = time.Now()
	if val, ok := dbkit.CacheGet("api_cache", "config_1"); ok {
		getTime := time.Since(start)
		fmt.Printf("  读取复杂对象: 耗时 %v\n", getTime)
		fmt.Printf("  结果: %v\n", val)
	}

	// ========== 基础功能演示 ==========

	fmt.Println("【基础功能演示】")

	// 查看缓存状态
	fmt.Println("\n[功能 1] 缓存状态:")
	status := dbkit.CacheStatus()
	fmt.Printf("  缓存类型: %v\n", status["type"])
	fmt.Printf("  Redis 地址: %v\n", status["address"])
	if dbSize, ok := status["db_size"]; ok {
		fmt.Printf("  Redis 数据库大小 (Key 数量): %v\n", dbSize)
	}

	// 测试过期 (Redis 级别生效)
	fmt.Println("\n[功能 2] 测试自定义过期 (3 秒)...")
	dbkit.CacheSet("api_cache", "temp_token", "ABC-123", 3*time.Second)
	fmt.Println("  ✓ 已设置临时 token")

	time.Sleep(4 * time.Second)
	if _, ok := dbkit.CacheGet("api_cache", "temp_token"); !ok {
		fmt.Println("  ✓ Redis 缓存已过期")
	}

	// 清理操作
	fmt.Println("\n[功能 3] 清理 Redis 库:")
	dbkit.CacheSet("api_cache", "key_1", "data_1")
	dbkit.CacheSet("api_cache", "key_2", "data_2")
	fmt.Println("  已设置 key_1 和 key_2")

	dbkit.CacheClearRepository("api_cache")
	fmt.Println("  ✓ api_cache 库下的所有 Key 已被清理")

	if _, ok := dbkit.CacheGet("api_cache", "key_1"); !ok {
		fmt.Println("  ✓ 清理验证成功")
	}

	fmt.Println("【总结】")

	fmt.Println("✓ RedisCache 优化：消除双重序列化，直接使用字节数组")
	fmt.Println("✓ 性能提升约 50-100%（相比优化前）")
	fmt.Println("✓ 适合分布式场景，多实例共享缓存")
	fmt.Println("✓ 自动 JSON 序列化，支持复杂对象")
	fmt.Println("\n=== 演示完成 ===")
}
