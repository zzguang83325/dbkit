package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
)

// 快速入门示例 - 展示 DBKit SQL Template 连接 MySQL 数据库的核心功能
func main() {
	fmt.Println("========================================")
	fmt.Println("   DBKit SQL Template MySQL 快速入门")
	fmt.Println("========================================")

	// 步骤 1: 加载 SQL 配置文件
	fmt.Println("\n【步骤 1: 加载配置】")
	if err := initializeConfigs(); err != nil {
		log.Fatalf("❌ 初始化配置失败: %v", err)
	}

	dbkit.InitLogger("debug")
	// 步骤 2: 连接 MySQL 数据库
	fmt.Println("\n【步骤 2: 连接数据库】")
	if err := connectDatabase(); err != nil {
		log.Printf("❌ 数据库连接失败: %v", err)
		fmt.Println("💡 请确保 MySQL 数据库正在运行并修改连接参数")
		return
	}
	demonstrateInsert()
	// 步骤 3: 基础查询操作
	fmt.Println("\n【步骤 3: 基础查询】")
	demonstrateBasicQuery()
	fmt.Println("\n【步骤 4: 分页查询】")
	demonstratePaginate() //分页查询




	// 步骤 5: 更新操作
	fmt.Println("\n【步骤 5: 更新数据】")
	demonstrateUpdate()

	// 步骤 6: 动态查询
	fmt.Println("\n【步骤 6: 动态查询】")
	demonstrateDynamicQuery()

	// 步骤 7: 事务处理
	fmt.Println("\n【步骤 7: 事务处理】")
	demonstrateTransaction()

	fmt.Println("\n========================================")
	fmt.Println("   Sql模板 快速入门完成！")
	fmt.Println("========================================")
}

// 初始化配置
func initializeConfigs() error {
	// 加载用户服务配置
	if err := dbkit.LoadSqlConfig("./config/user_service.json"); err != nil {
		return fmt.Errorf("加载用户服务配置失败: %v", err)
	}
	fmt.Println("✅ 用户服务配置加载成功")

	// 加载订单服务配置
	if err := dbkit.LoadSqlConfig("./config/order_service.json"); err != nil {
		return fmt.Errorf("加载订单服务配置失败: %v", err)
	}
	fmt.Println("✅ 订单服务配置加载成功")

	// 加载通用配置
	if err := dbkit.LoadSqlConfig("./config/common.json"); err != nil {
		return fmt.Errorf("加载通用配置失败: %v", err)
	}
	fmt.Println("✅ 通用配置加载成功")

	return nil
}

// 连接数据库
func connectDatabase() error {
	// MySQL 连接字符串
	// 请根据实际情况修改以下连接参数
	dsn := "root:123456@tcp(localhost:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"

	fmt.Printf("正在连接 MySQL 数据库...\n")
	fmt.Printf("DSN: %s\n", dsn)

	// 使用 DBKit 的正确 API 连接数据库
	err := dbkit.OpenDatabase(dbkit.MySQL, dsn, 10)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	fmt.Println("✅ 数据库连接成功")
	return nil
}

// 基础查询演示
func demonstrateBasicQuery() {
	fmt.Println("--- 根据 ID 查询用户 ---")

	// 使用配置文件中的 SQL 模板查询单条记录
	record, err := dbkit.SqlTemplate("user_service.findById", 1).QueryFirst()
	if err != nil {
		log.Printf("❌ 查询失败: %v", err)
		return
	}

	if record != nil {
		fmt.Printf("✅ 查询成功: ID=%v, Name=%v, Email=%v\n",
			record.Get("id"), record.Get("name"), record.Get("email"))
	} else {
		fmt.Println("⚠️  未找到 ID=1 的用户")
	}

	fmt.Println("\n--- 根据邮箱查询用户 ---")
	record2, err := dbkit.SqlTemplate("user_service.findByEmail", "zhangsan@example.com").QueryFirst()
	if err != nil {
		log.Printf("❌ 查询失败: %v", err)
		return
	}

	if record2 != nil {
		fmt.Printf("✅ 查询成功: ID=%v, Name=%v, Email=%v\n",
			record2.Get("id"), record2.Get("name"), record2.Get("email"))
	} else {
		fmt.Println("⚠️  未找到该邮箱的用户")
	}
}

// 分页查询演示
func demonstratePaginate() {
	fmt.Println("\n--- SQL 模板分页查询演示 ---")

	// 基本分页查询
	fmt.Println("1. 基本分页查询（第1页，每页5条）")
	pageObj, err := dbkit.SqlTemplate("user_service.findUsers").Paginate(1, 5)
	if err != nil {
		log.Printf("❌ 分页查询失败: %v", err)
		return
	}

	if pageObj != nil {
		fmt.Printf("✅ 分页查询成功: 第%d页（共%d页），总条数: %d\n",
			pageObj.PageNumber, pageObj.TotalPage, pageObj.TotalRow)

		for i, record := range pageObj.List {
			fmt.Printf("   %d. ID=%v, Name=%v, Email=%v\n",
				i+1, record.Get("id"), record.Get("name"), record.Get("email"))
		}
	}

	// 带参数的分页查询
	fmt.Println("\n2. 带参数的分页查询（查询状态为1的用户，第2页）")
	params := map[string]interface{}{
		"status": 1,
	}
	pageObj2, err := dbkit.SqlTemplate("user_service.findUsers", params).Paginate(2, 3)
	if err != nil {
		log.Printf("❌ 带参数分页查询失败: %v", err)
		return
	}

	if pageObj2 != nil {
		fmt.Printf("✅ 带参数分页查询成功: 第%d页（共%d页），总条数: %d\n",
			pageObj2.PageNumber, pageObj2.TotalPage, pageObj2.TotalRow)

		for i, record := range pageObj2.List {
			fmt.Printf("   %d. ID=%v, Name=%v, Status=%v\n",
				i+1, record.Get("id"), record.Get("name"), record.Get("status"))
		}
	}

	// 带超时的分页查询
	fmt.Println("\n3. 带超时的分页查询（30秒超时）")
	pageObj3, err := dbkit.SqlTemplate("user_service.findUsers").
		Timeout(30*time.Second).
		Paginate(1, 10)
	if err != nil {
		log.Printf("❌ 超时分页查询失败: %v", err)
		return
	}

	if pageObj3 != nil {
		fmt.Printf("✅ 超时分页查询成功: 第%d页（共%d页），总条数: %d\n",
			pageObj3.PageNumber, pageObj3.TotalPage, pageObj3.TotalRow)
	}
}

// 插入操作演示
func demonstrateInsert() {
	fmt.Println("--- 插入新用户 ---")

	// 使用配置文件中的插入 SQL
	result, err := dbkit.SqlTemplate("user_service.insertUser",
		"张三", "zhangsan_new@example.com", 28, "北京", 1).Exec()

	if err != nil {
		log.Printf("❌ 插入失败: %v", err)
		return
	}

	fmt.Printf("✅ 插入成功: %+v\n", result)

	// 验证插入结果 - 查询最新插入的用户
	record, err := dbkit.SqlTemplate("user_service.findByEmail", "zhangsan_new@example.com").QueryFirst()
	if err == nil && record != nil {
		fmt.Printf("✅ 验证成功: ID=%v, Name=%v, Email=%v\n",
			record.Get("id"), record.Get("name"), record.Get("email"))
	}
}

// 更新操作演示
func demonstrateUpdate() {
	fmt.Println("--- 更新用户信息 ---")

	// 使用 Map 参数进行更新
	updateParams := map[string]interface{}{
		"name":  "李四2",
		"email": "lisi@example.com",
		"age":   30,
		"city":  "上海",
		"id":    2,
	}

	result, err := dbkit.SqlTemplate("user_service.updateUser", updateParams).Exec()
	if err != nil {
		log.Printf("❌ 更新失败: %v", err)
		return
	}

	fmt.Printf("✅ 更新成功: %+v\n", result)

	// 验证更新结果
	record, err := dbkit.SqlTemplate("user_service.findById", 2).QueryFirst()
	if err == nil && record != nil {
		fmt.Printf("✅ 验证更新: ID=%v, Name=%v, Email=%v, City=%v\n",
			record.Get("id"), record.Get("name"), record.Get("email"), record.Get("city"))
	}
}

// 动态查询演示
func demonstrateDynamicQuery() {
	fmt.Println("--- 动态条件查询 ---")

	// 测试不同的查询条件组合
	testCases := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "按状态查询",
			params: map[string]interface{}{"status": 1},
		},
		{
			name:   "按状态和姓名查询",
			params: map[string]interface{}{"status": 1, "name": "张"},
		},
		{
			name:   "按状态和年龄范围查询",
			params: map[string]interface{}{"status": 1, "ageMin": 25, "ageMax": 35},
		},
	}

	for i, tc := range testCases {
		fmt.Printf("\n--- 测试 %d: %s ---\n", i+1, tc.name)
		fmt.Printf("查询条件: %v\n", tc.params)

		records, err := dbkit.SqlTemplate("user_service.findUsers", tc.params).Query()
		if err != nil {
			log.Printf("❌ 查询失败: %v", err)
			continue
		}

		fmt.Printf("✅ 查询到 %d 条记录\n", len(records))
		for j, record := range records {
			if j < 3 { // 只显示前3条
				fmt.Printf("   %d. %v (%v) - %v岁, %v\n",
					record.Get("id"), record.Get("name"), record.Get("email"),
					record.Get("age"), record.Get("city"))
			}
		}
		if len(records) > 3 {
			fmt.Printf("   ... 还有 %d 条记录\n", len(records)-3)
		}
	}
}

// 事务处理演示
func demonstrateTransaction() {
	fmt.Println("--- 事务处理演示 ---")

	// 使用 DBKit 的事务处理
	err := dbkit.Transaction(func(tx *dbkit.Tx) error {
		fmt.Println("✅ 事务已开启")

		// 在事务中插入用户
		result1, err := tx.SqlTemplate("user_service.insertUser",
			"事务用户", "tx@example.com", 25, "深圳", 1).Exec()
		if err != nil {
			return fmt.Errorf("事务中插入用户失败: %v", err)
		}

		fmt.Printf("✅ 事务中插入用户成功: %+v\n", result1)

		// 在事务中创建订单（假设我们知道用户ID）
		result2, err := tx.SqlTemplate("order_service.createOrder",
			1, 299.99, "pending").Exec()
		if err != nil {
			return fmt.Errorf("事务中创建订单失败: %v", err)
		}

		fmt.Printf("✅ 事务中创建订单成功: %+v\n", result2)
		return nil
	})

	if err != nil {
		log.Printf("❌ 事务执行失败: %v", err)
		return
	}

	fmt.Println("✅ 事务提交成功")

	// 验证事务结果
	record, err := dbkit.SqlTemplate("user_service.findByEmail", "tx@example.com").QueryFirst()
	if err == nil && record != nil {
		fmt.Printf("✅ 验证用户: ID=%v, Name=%v, Email=%v\n",
			record.Get("id"), record.Get("name"), record.Get("email"))
	}
}
