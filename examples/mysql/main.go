package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/mysql/models"
)

// ============================================================================
// DBKit MySQL 综合测试示例
// ============================================================================
// 功能演示:
//   1. 数据库连接和配置
//   2. Record CRUD 操作（基础 API）
//   3. DbModel CRUD 操作（模型 API）
//   4. 链式查询（QueryBuilder）
//   5. 分页查询
//   6. 事务处理
//   7. 缓存机制
//   8. 批量操作
//   9. 工具函数
//
// 使用场景:
//   - 学习 DBKit 的各种 API 用法
//   - 理解 Record 和 DbModel 的区别
//   - 掌握链式查询的强大功能
//   - 了解事务和缓存的使用
//
// 前置条件:
//   - MySQL 数据库已启动
//   - 数据库连接信息正确
//   - 有相应的数据库和表权限
// ============================================================================

func main() {
	fmt.Println("========================================")
	fmt.Println("   DBKit MySQL 综合测试")
	fmt.Println("========================================")

	// 1. 数据库连接测试
	testDatabaseConnection()

	// 2. 初始化环境
	setupTable()
	prepareData()

	// 3. Record CRUD 测试
	testRecordCRUD()

	// 4. DbModel CRUD 测试
	testDbModelCRUD()

	// 5. 链式查询测试
	testChainQuery()

	// 6. 分页查询测试
	testPagination()

	// 7. 事务测试
	testTransaction()

	// 8. 缓存测试
	testCache()

	// 9. DbModel 缓存测试
	testDbModelCache()

	// 10. 批量操作测试
	testBatchOperations()

	// 11. 工具函数测试
	testUtilityFunctions()

	fmt.Println("\n========================================")
	fmt.Println("   所有测试完成!")
	fmt.Println("========================================")
}

// ============================================================================
// 1. 数据库连接测试
// ============================================================================
// 说明: 演示如何连接 MySQL 数据库并进行基本的连接测试
// 关键 API:
//   - OpenDatabaseWithDBName: 使用指定的数据库名称打开连接
//   - PingDB: 测试数据库连接是否正常
//   - SetDebugMode: 启用调试模式，输出 SQL 日志
func testDatabaseConnection() {
	fmt.Println("\n【1. 数据库连接测试】")

	// MySQL 连接字符串格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	// 参数说明:
	//   - user: 数据库用户名
	//   - password: 数据库密码
	//   - host: 数据库主机地址
	//   - port: 数据库端口（默认 3306）
	//   - dbname: 数据库名称
	//   - charset: 字符集（推荐 utf8mb4）
	//   - parseTime: 是否解析时间字段
	//   - loc: 时区设置
	dsn := "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"

	// 使用指定的数据库名称打开连接
	// 参数: 数据库别名, 数据库类型, 连接字符串, 最大连接数
	err := dbkit.OpenDatabaseWithDBName("mysql", dbkit.MySQL, dsn, 25)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	fmt.Println("✓ 数据库连接成功")

	// 测试数据库连接是否正常
	// 这会发送一个 PING 命令到数据库
	err = dbkit.PingDB("mysql")
	if err != nil {
		log.Fatalf("❌ Ping 失败: %v", err)
	}
	fmt.Println("✓ Ping 测试通过")

	// 启用调试模式
	// 启用后，所有 SQL 语句和执行时间都会被打印到日志
	// 用于开发和调试，生产环境应关闭
	dbkit.SetDebugMode(true)
	fmt.Println("✓ 调试模式已开启")
}

// ==================== 2. 初始化环境 ====================
func setupTable() {
	fmt.Println("\n【2. 初始化测试表】")

	// 删除旧表重新创建
	dbkit.Use("mysql").Exec("DROP TABLE IF EXISTS demo")

	sql := `
	CREATE TABLE demo (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100),
		age INT,
		salary DECIMAL(10, 2),
		is_active TINYINT(1) DEFAULT 1,
		birthday DATE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT
	)`
	_, err := dbkit.Use("mysql").Exec(sql)
	if err != nil {
		log.Fatalf("❌ 创建表失败: %v", err)
	}
	fmt.Println("✓ 表 'demo' 创建成功")
}

func prepareData() {
	fmt.Println("\n【准备测试数据】")

	records := make([]*dbkit.Record, 0, 50)
	for i := 1; i <= 50; i++ {
		record := dbkit.NewRecord().
			Set("name", fmt.Sprintf("User_%d", i)).
			Set("age", 18+rand.Intn(40)).
			Set("salary", 3000.0+rand.Float64()*7000.0).
			Set("is_active", i%2).
			Set("birthday", time.Now().AddDate(-20-rand.Intn(10), 0, 0)).
			Set("metadata", fmt.Sprintf(`{"tag": "tag_%d"}`, i))
		records = append(records, record)
	}

	affected, err := dbkit.Use("mysql").BatchInsert("demo", records, 50)
	if err != nil {
		log.Fatalf("❌ 批量插入失败: %v", err)
	}
	fmt.Printf("✓ 批量插入 %d 条数据\n", affected)
}

// ==================== 3. Record CRUD 测试 ====================
func testRecordCRUD() {
	fmt.Println("\n【3. Record CRUD 测试】")

	// Insert
	newRec := dbkit.NewRecord().
		Set("name", "TestRecord").
		Set("age", 30).
		Set("salary", 8000.50).
		Set("is_active", 1)
	id, err := dbkit.Use("mysql").Insert("demo", newRec)
	if err != nil {
		log.Printf("❌ Insert 失败: %v", err)
	} else {
		fmt.Printf("✓ Insert 成功, ID: %d\n", id)
	}

	// Query
	records, err := dbkit.Use("mysql").Query("SELECT * FROM demo WHERE id = ?", id)
	if err != nil || len(records) == 0 {
		log.Printf("❌ Query 失败: %v", err)
	} else {
		fmt.Printf("✓ Query 成功, Name: %s\n", records[0].GetString("name"))
	}

	// QueryFirst
	record, err := dbkit.Use("mysql").QueryFirst("SELECT * FROM demo WHERE id = ?", id)
	if err != nil || record == nil {
		log.Printf("❌ QueryFirst 失败: %v", err)
	} else {
		fmt.Printf("✓ QueryFirst 成功, Age: %d\n", record.GetInt("age"))
	}

	// Update
	updateRec := dbkit.NewRecord().Set("age", 35).Set("salary", 9500.00)
	affected, err := dbkit.Use("mysql").Update("demo", updateRec, "id = ?", id)
	if err != nil {
		log.Printf("❌ Update 失败: %v", err)
	} else {
		fmt.Printf("✓ Update 成功, 影响行数: %d\n", affected)
	}

	// Save (更新已存在记录)
	record.Set("name", "TestRecord_Updated")
	affected, err = dbkit.Use("mysql").Save("demo", record)
	if err != nil {
		log.Printf("❌ Save 失败: %v", err)
	} else {
		fmt.Printf("✓ Save 成功, 影响行数: %d\n", affected)
	}

	// Count
	count, err := dbkit.Use("mysql").Count("demo", "age > ?", 25)
	if err != nil {
		log.Printf("❌ Count 失败: %v", err)
	} else {
		fmt.Printf("✓ Count 成功, 数量: %d\n", count)
	}

	// Exists
	exists, err := dbkit.Use("mysql").Exists("demo", "id = ?", id)
	if err != nil {
		log.Printf("❌ Exists 失败: %v", err)
	} else {
		fmt.Printf("✓ Exists 成功, 存在: %v\n", exists)
	}

	// Delete
	affected, err = dbkit.Use("mysql").Delete("demo", "id = ?", id)
	if err != nil {
		log.Printf("❌ Delete 失败: %v", err)
	} else {
		fmt.Printf("✓ Delete 成功, 影响行数: %d\n", affected)
	}
}

// ==================== 4. DbModel CRUD 测试 ====================
func testDbModelCRUD() {
	fmt.Println("\n【4. DbModel CRUD 测试】")

	// Insert
	user := &models.Demo{
		Name:     "ModelUser",
		Age:      28,
		Salary:   7500.00,
		IsActive: 1,
		Birthday: time.Now().AddDate(-28, 0, 0),
		Metadata: `{"role": "admin"}`,
	}
	id, err := user.Insert()
	if err != nil {
		log.Printf("❌ DbModel Insert 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel Insert 成功, ID: %d\n", id)
		user.ID = id
	}

	// FindFirst
	found := &models.Demo{}
	found, err = found.FindFirst("id = ?", id)
	if err != nil {
		log.Printf("❌ DbModel FindFirst 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel FindFirst 成功, Name: %s\n", found.Name)
	}

	// Update
	found.Age = 30
	found.Salary = 8500.00
	affected, err := found.Update()
	if err != nil {
		log.Printf("❌ DbModel Update 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel Update 成功, 影响行数: %d\n", affected)
	}

	// Save
	found.Name = "ModelUser_Saved"
	affected, err = found.Save()
	if err != nil {
		log.Printf("❌ DbModel Save 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel Save 成功, 影响行数: %d\n", affected)
	}

	// Find (查询多条)
	model := &models.Demo{}
	results, err := model.Find("age >= ?", "id DESC", 20)
	if err != nil {
		log.Printf("❌ DbModel Find 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel Find 成功, 数量: %d\n", len(results))
	}

	// Paginate
	page, err := model.Paginate(1, 10, "select * from demo where age > ? order by id ASC", 20)
	if err != nil {
		log.Printf("❌ DbModel Paginate 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel Paginate 成功, 总数: %d, 当前页: %d\n", page.TotalRow, len(page.List))
	}

	// ToJson
	jsonStr := found.ToJson()
	fmt.Printf("✓ DbModel ToJson: %s\n", jsonStr[:min(80, len(jsonStr))]+"...")

	// Delete
	affected, err = found.Delete()
	if err != nil {
		log.Printf("❌ DbModel Delete 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel Delete 成功, 影响行数: %d\n", affected)
	}
}

// ==================== 5. 链式查询测试 ====================
func testChainQuery() {
	fmt.Println("\n【5. 链式查询测试】")

	// 基本链式查询
	records, err := dbkit.Use("mysql").Table("demo").
		Select("id, name, age, salary").
		Where("age >= ?", 25).
		Where("is_active = ?", 1).
		OrderBy("age DESC").
		Limit(5).
		Find()
	if err != nil {
		log.Printf("❌ 链式查询 Find 失败: %v", err)
	} else {
		fmt.Printf("✓ 链式查询 Find 成功, 数量: %d\n", len(records))
		for i, r := range records {
			if i < 3 {
				fmt.Printf("  - %s (Age: %d, Salary: %.2f)\n", r.GetString("name"), r.GetInt("age"), r.GetFloat("salary"))
			}
		}
	}

	// FindFirst
	record, err := dbkit.Use("mysql").Table("demo").
		Where("age > ?", 30).
		OrderBy("salary DESC").
		FindFirst()
	if err != nil {
		log.Printf("❌ 链式查询 FindFirst 失败: %v", err)
	} else if record != nil {
		fmt.Printf("✓ 链式查询 FindFirst 成功, Name: %s\n", record.GetString("name"))
	}

	// FindToDbModel
	var users []models.Demo
	err = dbkit.Use("mysql").Table("demo").
		Where("age > ?", 20).
		OrderBy("id ASC").
		Limit(5).
		FindToDbModel(&users)
	if err != nil {
		log.Printf("❌ 链式查询 FindToDbModel 失败: %v", err)
	} else {
		fmt.Printf("✓ 链式查询 FindToDbModel 成功, 数量: %d\n", len(users))
	}

	// Count
	count, err := dbkit.Use("mysql").Table("demo").
		Where("salary > ?", 5000).
		Count()
	if err != nil {
		log.Printf("❌ 链式查询 Count 失败: %v", err)
	} else {
		fmt.Printf("✓ 链式查询 Count 成功, 数量: %d\n", count)
	}

	// Delete (链式)
	// 先插入一条测试数据
	testRec := dbkit.NewRecord().Set("name", "ToDelete").Set("age", 99)
	dbkit.Use("mysql").Insert("demo", testRec)

	affected, err := dbkit.Use("mysql").Table("demo").
		Where("name = ?", "ToDelete").
		Delete()
	if err != nil {
		log.Printf("❌ 链式查询 Delete 失败: %v", err)
	} else {
		fmt.Printf("✓ 链式查询 Delete 成功, 影响行数: %d\n", affected)
	}
}

// ==================== 6. 分页查询测试 ====================
func testPagination() {
	fmt.Println("\n【6. 分页查询测试】")

	// 使用 Paginate 函数
	page, err := dbkit.Use("mysql").Paginate(1, 10, "select * from demo where age > ? order by id ASC", 20)
	if err != nil {
		log.Printf("❌ Paginate 失败: %v", err)
	} else {
		fmt.Printf("✓ Paginate 成功:\n")
		fmt.Printf("  - 当前页: %d\n", page.PageNumber)
		fmt.Printf("  - 每页大小: %d\n", page.PageSize)
		fmt.Printf("  - 总页数: %d\n", page.TotalPage)
		fmt.Printf("  - 总记录数: %d\n", page.TotalRow)
		fmt.Printf("  - 当前页数据量: %d\n", len(page.List))
	}

	// 链式分页
	page2, err := dbkit.Use("mysql").Table("demo").
		Select("id, name, age").
		Where("is_active = ?", 1).
		OrderBy("created_at DESC").
		Paginate(2, 5)
	if err != nil {
		log.Printf("❌ 链式 Paginate 失败: %v", err)
	} else {
		fmt.Printf("✓ 链式 Paginate 成功, 第%d页, 共%d条\n", page2.PageNumber, page2.TotalRow)
	}
}

// ==================== 7. 事务测试 ====================
func testTransaction() {
	fmt.Println("\n【7. 事务测试】")

	// 成功的事务
	err := dbkit.Use("mysql").Transaction(func(tx *dbkit.Tx) error {
		// 插入
		rec1 := dbkit.NewRecord().Set("name", "TxUser1").Set("age", 25)
		_, err := tx.Insert("demo", rec1)
		if err != nil {
			return err
		}

		// 更新
		rec2 := dbkit.NewRecord().Set("salary", 10000)
		_, err = tx.Update("demo", rec2, "name = ?", "TxUser1")
		return err
	})
	if err != nil {
		log.Printf("❌ 事务(成功) 失败: %v", err)
	} else {
		fmt.Println("✓ 事务(成功) 测试通过")
	}

	// 回滚的事务
	err = dbkit.Use("mysql").Transaction(func(tx *dbkit.Tx) error {
		rec := dbkit.NewRecord().Set("name", "TxUser_Rollback").Set("age", 30)
		tx.Insert("demo", rec)
		// 模拟错误，触发回滚
		return fmt.Errorf("模拟错误，触发回滚")
	})
	if err != nil {
		fmt.Printf("✓ 事务(回滚) 测试通过, 错误: %v\n", err)
	}

	// 验证回滚
	exists, _ := dbkit.Use("mysql").Exists("demo", "name = ?", "TxUser_Rollback")
	if !exists {
		fmt.Println("✓ 事务回滚验证通过, 数据未插入")
	}

	// 清理事务测试数据
	dbkit.Use("mysql").Delete("demo", "name LIKE ?", "TxUser%")
}

// ==================== 8. 缓存测试 ====================
func testCache() {
	fmt.Println("\n【8. 缓存测试】")

	// 创建缓存
	dbkit.CreateCache("test_cache", 5*time.Minute)
	fmt.Println("✓ 缓存 'test_cache' 创建成功")

	// 第一次查询 (从数据库)
	start := time.Now()
	records, err := dbkit.Use("mysql").Cache("test_cache", 60*time.Second).
		Query("SELECT * FROM demo WHERE age > ?", 25)
	if err != nil {
		log.Printf("❌ 缓存查询(1st) 失败: %v", err)
	} else {
		fmt.Printf("✓ 缓存查询(1st) 成功, 数量: %d, 耗时: %v\n", len(records), time.Since(start))
	}

	// 第二次查询 (从缓存)
	start = time.Now()
	records, err = dbkit.Use("mysql").Cache("test_cache", 60*time.Second).
		Query("SELECT * FROM demo WHERE age > ?", 25)
	if err != nil {
		log.Printf("❌ 缓存查询(2nd) 失败: %v", err)
	} else {
		fmt.Printf("✓ 缓存查询(2nd) 成功, 数量: %d, 耗时: %v (应更快)\n", len(records), time.Since(start))
	}

	// 手动缓存操作
	dbkit.CacheSet("manual_cache", "key1", "value1", 1*time.Minute)
	val, ok := dbkit.CacheGet("manual_cache", "key1")
	if ok {
		fmt.Printf("✓ 手动缓存 Get 成功, 值: %v\n", val)
	}

	dbkit.CacheDelete("manual_cache", "key1")
	_, ok = dbkit.CacheGet("manual_cache", "key1")
	if !ok {
		fmt.Println("✓ 手动缓存 Delete 成功")
	}

	// 缓存状态
	status := dbkit.CacheStatus()
	fmt.Printf("✓ 缓存状态: 类型=%v, 项数=%v\n", status["type"], status["total_items"])
}

// ==================== 9. DbModel 缓存测试 ====================
func testDbModelCache() {
	fmt.Println("\n【9. DbModel 缓存测试】")

	model := &models.Demo{}

	// 第一次查询 (从数据库)
	start := time.Now()
	results, err := model.Cache("model_cache", 60*time.Second).Find("age > ?", "id ASC", 20)
	if err != nil {
		log.Printf("❌ DbModel 缓存查询(1st) 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel 缓存查询(1st) 成功, 数量: %d, 耗时: %v\n", len(results), time.Since(start))
	}

	// 第二次查询 (从缓存)
	start = time.Now()
	results, err = model.Cache("model_cache", 60*time.Second).Find("age > ?", "id ASC", 20)
	if err != nil {
		log.Printf("❌ DbModel 缓存查询(2nd) 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel 缓存查询(2nd) 成功, 数量: %d, 耗时: %v (应更快)\n", len(results), time.Since(start))
	}

	// 分页缓存
	start = time.Now()
	page, err := model.Cache("page_cache", 60*time.Second).PaginateBuilder(1, 10, "age > ?", "id DESC", 20)
	if err != nil {
		log.Printf("❌ DbModel 分页缓存(1st) 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel 分页缓存(1st) 成功, 总数: %d, 耗时: %v\n", page.TotalRow, time.Since(start))
	}

	start = time.Now()
	page, err = model.Cache("page_cache", 60*time.Second).PaginateBuilder(1, 10, "age > ?", "id DESC", 20)
	if err != nil {
		log.Printf("❌ DbModel 分页缓存(2nd) 失败: %v", err)
	} else {
		fmt.Printf("✓ DbModel 分页缓存(2nd) 成功, 总数: %d, 耗时: %v (应更快)\n", page.TotalRow, time.Since(start))
	}
}

// ==================== 10. 批量操作测试 ====================
func testBatchOperations() {
	fmt.Println("\n【10. 批量操作测试】")

	// 批量插入
	records := make([]*dbkit.Record, 0, 20)
	for i := 1; i <= 20; i++ {
		rec := dbkit.NewRecord().
			Set("name", fmt.Sprintf("BatchUser_%d", i)).
			Set("age", 20+i).
			Set("salary", 5000.0+float64(i)*100)
		records = append(records, rec)
	}

	affected, err := dbkit.Use("mysql").BatchInsert("demo", records, 10)
	if err != nil {
		log.Printf("❌ BatchInsert 失败: %v", err)
	} else {
		fmt.Printf("✓ BatchInsert 成功, 影响行数: %d\n", affected)
	}

	// BatchInsertDefault
	records2 := make([]*dbkit.Record, 0, 5)
	for i := 1; i <= 5; i++ {
		rec := dbkit.NewRecord().
			Set("name", fmt.Sprintf("BatchDefault_%d", i)).
			Set("age", 30+i)
		records2 = append(records2, rec)
	}

	affected, err = dbkit.Use("mysql").BatchInsertDefault("demo", records2)
	if err != nil {
		log.Printf("❌ BatchInsertDefault 失败: %v", err)
	} else {
		fmt.Printf("✓ BatchInsertDefault 成功, 影响行数: %d\n", affected)
	}

	// 清理批量测试数据
	dbkit.Use("mysql").Delete("demo", "name LIKE ?", "Batch%")
}

// ==================== 11. 工具函数测试 ====================
func testUtilityFunctions() {
	fmt.Println("\n【11. 工具函数测试】")

	// Record 转换
	record := dbkit.NewRecord().
		Set("id", 1).
		Set("name", "TestUser").
		Set("age", 25).
		Set("salary", 8000.50)

	// ToJson
	jsonStr := record.ToJson()
	fmt.Printf("✓ Record.ToJson: %s\n", jsonStr)

	// ToMap
	m := record.ToMap()
	fmt.Printf("✓ Record.ToMap: %v\n", m)

	// 类型获取
	fmt.Printf("✓ Record.GetString: %s\n", record.GetString("name"))
	fmt.Printf("✓ Record.GetInt: %d\n", record.GetInt("age"))
	fmt.Printf("✓ Record.GetFloat: %.2f\n", record.GetFloat("salary"))

	// Has / Keys
	fmt.Printf("✓ Record.Has('name'): %v\n", record.Has("name"))
	fmt.Printf("✓ Record.Keys: %v\n", record.Keys())

	// FromJson
	newRecord := dbkit.NewRecord()
	newRecord.FromJson(`{"name": "JsonUser", "age": 30}`)
	fmt.Printf("✓ Record.FromJson: name=%s, age=%d\n", newRecord.GetString("name"), newRecord.GetInt("age"))

	// ToStruct
	type User struct {
		ID     int64   `column:"id"`
		Name   string  `column:"name"`
		Age    int     `column:"age"`
		Salary float64 `column:"salary"`
	}
	var user User
	record.ToStruct(&user)
	fmt.Printf("✓ Record.ToStruct: %+v\n", user)

	// SnakeToCamel
	camel := dbkit.SnakeToCamel("user_name_test")
	fmt.Printf("✓ SnakeToCamel: user_name_test -> %s\n", camel)

	// ValidateTableName
	err := dbkit.ValidateTableName("valid_table")
	if err == nil {
		fmt.Println("✓ ValidateTableName: 'valid_table' 验证通过")
	}

	err = dbkit.ValidateTableName("invalid;table")
	if err != nil {
		fmt.Printf("✓ ValidateTableName: 'invalid;table' 验证失败 (预期行为)\n")
	}

	// SupportedDrivers
	drivers := dbkit.SupportedDrivers()
	fmt.Printf("✓ SupportedDrivers: %v\n", drivers)

	// IsValidDriver
	fmt.Printf("✓ IsValidDriver(MySQL): %v\n", dbkit.IsValidDriver(dbkit.MySQL))
	fmt.Printf("✓ IsValidDriver('invalid'): %v\n", dbkit.IsValidDriver("invalid"))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
