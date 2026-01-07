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
	if _, err := dbkit.Exec("CREATE TABLE test_users (id INT PRIMARY KEY, name VARCHAR(50))"); err != nil {
		fmt.Printf("创建表失败: %v\n", err)
	}
	if _, err := dbkit.Exec("INSERT INTO test_users (id, name) VALUES (1, 'Alice'), (2, 'Bob')"); err != nil {
		fmt.Printf("插入数据失败: %v\n", err)
	}

	// 2. 设置默认过期时间为 5 秒
	dbkit.SetDefaultTtl(5 * time.Second)

	// 3. 测试数据库查询缓存
	fmt.Println("\n[2] 测试数据库查询缓存:")

	// 第一次查询：从数据库读取并存入缓存
	start := time.Now()
	// 使用 FindAll 或 Query 配合 Cache 链式调用
	users, _ := dbkit.Cache("user_cache").Query("SELECT * FROM test_users WHERE id = ?", 1)
	fmt.Printf("第一次查询 (从 DB): %v, 耗时: %v\n", users, time.Since(start))

	// 第二次查询：从缓存读取
	start = time.Now()
	usersCached, _ := dbkit.Cache("user_cache").Query("SELECT * FROM test_users WHERE id = ?", 1)
	fmt.Printf("第二次查询 (从 Cache): %v, 耗时: %v\n", usersCached, time.Since(start))

	// 分页查询
	fmt.Printf("\n[3] 分页查询:\n")
	page, _ := dbkit.Cache("user_cache").Paginate(1, 10, "select * from test_users")
	fmt.Printf("分页查询结果: %v\n", page.ToJson())

	page2, _ := dbkit.Cache("user_cache").Paginate(1, 10, "select * from test_users")
	fmt.Printf("分页查询结果2: %v\n", page2.ToJson())

	// 4. 基础的 CacheGet/Set 操作
	fmt.Println("\n[3] 基础操作:")
	dbkit.CacheSet("user_store", "user_1", "张三")
	if val, ok := dbkit.CacheGet("user_store", "user_1"); ok {
		fmt.Printf("成功获取缓存: %v\n", val)
	}

	// 5. 查看缓存状态
	fmt.Println("\n[4] 缓存状态:")
	status := dbkit.CacheStatus()
	fmt.Printf("当前缓存类型: %v\n", status["type"])
	fmt.Printf("存储库数量: %v\n", status["store_count"])
	fmt.Printf("总缓存项数量: %v\n", status["total_items"])
	fmt.Printf("预估内存占用: %v (%v bytes)\n", status["estimated_memory_human"], status["estimated_memory_bytes"])
	fmt.Printf("清理间隔: %v\n", status["cleanup_interval"])

	// 3. 测试过期
	fmt.Println("\n[2] 测试过期 (等待 6 秒)...")
	time.Sleep(6 * time.Second)
	if _, ok := dbkit.CacheGet("user_store", "user_1"); !ok {
		fmt.Println("缓存已按预期过期")
	}

	// 4. 为特定库预设 TTL
	fmt.Println("\n[3] 为特定库预设 TTL (2 秒):")
	dbkit.CreateCache("short_lived", 2*time.Second)
	dbkit.CacheSet("short_lived", "temp_key", "瞬时数据")

	val, _ := dbkit.CacheGet("short_lived", "temp_key")
	fmt.Printf("立即获取: %v\n", val)

	time.Sleep(3 * time.Second)
	if _, ok := dbkit.CacheGet("short_lived", "temp_key"); !ok {
		fmt.Println("预设 TTL 数据已过期")
	}

	// 5. 手动删除和清理
	fmt.Println("\n[4] 手动删除和清理:")
	dbkit.CacheSet("user_store", "user_2", "李四")
	dbkit.CacheDelete("user_store", "user_2")
	if _, ok := dbkit.CacheGet("user_store", "user_2"); !ok {
		fmt.Println("user_2 已手动删除")
	}

	dbkit.CacheSet("user_store", "user_3", "王五")
	dbkit.CacheClear("user_store")
	if _, ok := dbkit.CacheGet("user_store", "user_3"); !ok {
		fmt.Println("user_store 已全部清空")
	}

	fmt.Println("\n=== 演示完成 ===")
}
