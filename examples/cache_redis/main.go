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
	if _, err := dbkit.Exec("CREATE TABLE test_users (id INT PRIMARY KEY, name VARCHAR(50))"); err != nil {
		fmt.Printf("创建表失败: %v\n", err)
	}
	if _, err := dbkit.Exec("INSERT INTO test_users (id, name) VALUES (1, 'Alice'), (2, 'Bob')"); err != nil {
		fmt.Printf("插入数据失败: %v\n", err)
	}

	// 2. 配置 Redis (通过子包按需引入)
	// 地址: 192.168.10.205:6379, 密码: cpdv3, 数据库: 2
	rc, err := redis.NewRedisCache("192.168.10.205:6379", "", "cpdv3", 2)
	if err != nil {
		fmt.Printf("Redis 连接失败: %v\n", err)
		return
	}
	dbkit.SetCache(rc)
	fmt.Println("Redis 已连接并切换为默认缓存提供者")

	// 3. 测试数据库查询缓存 (Redis)
	fmt.Println("\n[3] 测试数据库查询缓存 (Redis):")

	// 第一次查询：从数据库读取并存入 Redis
	start := time.Now()
	users, err := dbkit.Cache("user_cache_redis").Query("SELECT * FROM test_users WHERE id = ?", 1)
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
	}
	fmt.Printf("第一次查询 (从 DB): %v, 耗时: %v\n", users, time.Since(start))

	// 第二次查询：从 Redis 读取
	start = time.Now()
	usersCached, _ := dbkit.Cache("user_cache_redis").Query("SELECT * FROM test_users WHERE id = ?", 1)
	fmt.Printf("第二次查询 (从 Redis): %v, 耗时: %v\n", usersCached, time.Since(start))

	// 4. 基础的 CacheGet/Set 操作 (会自动序列化为 JSON)
	fmt.Println("\n[4] 基础操作:")
	dbkit.CacheSet("api_cache", "config_1", map[string]string{"theme": "dark", "lang": "zh"})

	if val, ok := dbkit.CacheGet("api_cache", "config_1"); ok {
		fmt.Printf("成功从 Redis 获取缓存: %v\n", val)
	}

	// 5. 查看缓存状态
	fmt.Println("\n[5] 缓存状态:")
	status := dbkit.CacheStatus()
	fmt.Printf("当前缓存类型: %v\n", status["type"])
	fmt.Printf("Redis 地址: %v\n", status["address"])
	if dbSize, ok := status["db_size"]; ok {
		fmt.Printf("Redis 数据库大小 (Key 数量): %v\n", dbSize)
	}

	// 5. 测试过期 (Redis 级别生效)
	fmt.Println("\n[5] 测试自定义过期 (3 秒)...")
	dbkit.CacheSet("api_cache", "temp_token", "ABC-123", 3*time.Second)

	time.Sleep(4 * time.Second)
	if _, ok := dbkit.CacheGet("api_cache", "temp_token"); !ok {
		fmt.Println("Redis 缓存已过期")
	}

	// 6. 清理操作
	fmt.Println("\n[6] 清理 Redis 库:")
	dbkit.CacheSet("api_cache", "key_1", "data_1")
	dbkit.CacheSet("api_cache", "key_2", "data_2")

	dbkit.CacheClear("api_cache")
	fmt.Println("api_cache 库下的所有 Key 已被清理 ")

	if _, ok := dbkit.CacheGet("api_cache", "key_1"); !ok {
		fmt.Println("清理验证成功")
	}

	fmt.Println("\n=== 演示完成 ===")
}
