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

	// 步骤 8: 缓存功能测试
	fmt.Println("\n【步骤 8: 缓存功能测试】")
	demonstrateCacheFeatures()

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


// 缓存功能演示
func demonstrateCacheFeatures() {
	fmt.Println("--- SQL 模板缓存功能演示 ---")

	// 初始化本地缓存
	fmt.Println("\n1. 初始化本地缓存")
	dbkit.InitLocalCache(10 * time.Minute)
	fmt.Println("✅ 本地缓存已初始化")

	// 测试 1: 基本查询 + 本地缓存
	fmt.Println("\n2. 基本查询 + 本地缓存")
	
	// 第一次查询 - 从数据库读取
	start := time.Now()
	record1, err := dbkit.SqlTemplate("user_service.findById", 1).
		LocalCache("user_cache", 5*time.Minute).
		QueryFirst()
	time1 := time.Since(start)
	
	if err != nil {
		log.Printf("❌ 第 1 次查询失败: %v", err)
	} else if record1 != nil {
		fmt.Printf("✅ 第 1 次查询 (从数据库): ID=%v, Name=%v, 耗时 %v\n",
			record1.Get("id"), record1.Get("name"), time1)
	}
	
	// 第二次查询 - 从本地缓存读取
	start = time.Now()
	record2, err := dbkit.SqlTemplate("user_service.findById", 1).
		LocalCache("user_cache", 5*time.Minute).
		QueryFirst()
	time2 := time.Since(start)
	
	if err != nil {
		log.Printf("❌ 第 2 次查询失败: %v", err)
	} else if record2 != nil {
		fmt.Printf("✅ 第 2 次查询 (从本地缓存): ID=%v, Name=%v, 耗时 %v\n",
			record2.Get("id"), record2.Get("name"), time2)
		if time2 < time1 {
			fmt.Printf("⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))
		}
	}

	// 测试 2: 分页查询 + 本地缓存
	fmt.Println("\n3. 分页查询 + 本地缓存")
	
	// 第一次分页查询 - 从数据库读取
	start = time.Now()
	page1, err := dbkit.SqlTemplate("user_service.findUsers").
		LocalCache("users_page", 5*time.Minute).
		Paginate(1, 5)
	time1 = time.Since(start)
	
	if err != nil {
		log.Printf("❌ 第 1 次分页查询失败: %v", err)
	} else if page1 != nil {
		fmt.Printf("✅ 第 1 次分页查询 (从数据库): 第%d页, 共%d条记录, 耗时 %v\n",
			page1.PageNumber, page1.TotalRow, time1)
	}
	
	// 第二次分页查询 - 从本地缓存读取
	start = time.Now()
	page2, err := dbkit.SqlTemplate("user_service.findUsers").
		LocalCache("users_page", 5*time.Minute).
		Paginate(1, 5)
	time2 = time.Since(start)
	
	if err != nil {
		log.Printf("❌ 第 2 次分页查询失败: %v", err)
	} else if page2 != nil {
		fmt.Printf("✅ 第 2 次分页查询 (从本地缓存): 第%d页, 共%d条记录, 耗时 %v\n",
			page2.PageNumber, page2.TotalRow, time2)
		if time2 < time1 {
			fmt.Printf("⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))
		}
	}

	// 测试 3: 带参数的分页查询 + 本地缓存
	fmt.Println("\n4. 带参数的分页查询 + 本地缓存")
	
	params := map[string]interface{}{
		"status": 1,
	}
	
	// 第一次查询
	start = time.Now()
	page3, err := dbkit.SqlTemplate("user_service.findUsers", params).
		LocalCache("users_status_page", 5*time.Minute).
		Paginate(1, 3)
	time1 = time.Since(start)
	
	if err != nil {
		log.Printf("❌ 第 1 次带参数分页查询失败: %v", err)
	} else if page3 != nil {
		fmt.Printf("✅ 第 1 次带参数分页查询 (从数据库): 状态=1, 共%d条记录, 耗时 %v\n",
			page3.TotalRow, time1)
	}
	
	// 第二次查询
	start = time.Now()
	page4, err := dbkit.SqlTemplate("user_service.findUsers", params).
		LocalCache("users_status_page", 5*time.Minute).
		Paginate(1, 3)
	time2 = time.Since(start)
	
	if err != nil {
		log.Printf("❌ 第 2 次带参数分页查询失败: %v", err)
	} else if page4 != nil {
		fmt.Printf("✅ 第 2 次带参数分页查询 (从本地缓存): 状态=1, 共%d条记录, 耗时 %v\n",
			page4.TotalRow, time2)
		if time2 < time1 {
			fmt.Printf("⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))
		}
	}

	// 测试 4: DB 实例的 SQL 模板缓存
	fmt.Println("\n5. DB 实例的 SQL 模板缓存")
	
	db, _ := dbkit.UseWithError("default")
	
	// 方式 1: db.LocalCache().SqlTemplate()
	start = time.Now()
	page5, err := db.LocalCache("db_template_page", 5*time.Minute).
		SqlTemplate("user_service.findUsers").
		Paginate(1, 5)
	time1 = time.Since(start)
	
	if err != nil {
		log.Printf("❌ 方式 1 查询失败: %v", err)
	} else if page5 != nil {
		fmt.Printf("✅ 方式 1 (db.LocalCache().SqlTemplate()): 共%d条记录, 耗时 %v\n",
			page5.TotalRow, time1)
	}
	
	// 方式 2: db.SqlTemplate().LocalCache()
	start = time.Now()
	page6, err := db.SqlTemplate("user_service.findUsers").
		LocalCache("db_template_page2", 5*time.Minute).
		Paginate(1, 5)
	time1 = time.Since(start)
	
	if err != nil {
		log.Printf("❌ 方式 2 查询失败: %v", err)
	} else if page6 != nil {
		fmt.Printf("✅ 方式 2 (db.SqlTemplate().LocalCache()): 共%d条记录, 耗时 %v\n",
			page6.TotalRow, time1)
	}

	// 测试 5: 事务中的 SQL 模板缓存
	fmt.Println("\n6. 事务中的 SQL 模板缓存")
	
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		// 在事务中使用缓存
		page, err := tx.LocalCache("tx_template_page", 5*time.Minute).
			SqlTemplate("user_service.findUsers").
			Paginate(1, 5)
		
		if err != nil {
			return err
		}
		
		fmt.Printf("✅ 事务中分页查询成功: 共%d条记录\n", page.TotalRow)
		return nil
	})
	
	if err != nil {
		log.Printf("❌ 事务失败: %v", err)
	}

	// 测试 6: 不同页码的缓存
	fmt.Println("\n7. 不同页码的缓存")
	
	// 查询第 1 页
	page7, _ := dbkit.SqlTemplate("user_service.findUsers").
		LocalCache("multi_page", 5*time.Minute).
		Paginate(1, 3)
	if page7 != nil {
		fmt.Printf("✅ 第 1 页: 共%d条记录\n", page7.TotalRow)
	}
	
	// 查询第 2 页
	page8, _ := dbkit.SqlTemplate("user_service.findUsers").
		LocalCache("multi_page", 5*time.Minute).
		Paginate(2, 3)
	if page8 != nil {
		fmt.Printf("✅ 第 2 页: 共%d条记录\n", page8.TotalRow)
	}
	
	fmt.Println("说明: 每个页码都会单独缓存")

	// 测试 7: Query 方法的缓存
	fmt.Println("\n8. Query 方法的缓存")
	
	// 第一次查询
	start = time.Now()
	records1, err := dbkit.SqlTemplate("user_service.findUsers").
		LocalCache("users_list", 5*time.Minute).
		Query()
	time1 = time.Since(start)
	
	if err != nil {
		log.Printf("❌ 第 1 次 Query 失败: %v", err)
	} else {
		fmt.Printf("✅ 第 1 次 Query (从数据库): 查询到%d条记录, 耗时 %v\n",
			len(records1), time1)
	}
	
	// 第二次查询
	start = time.Now()
	records2, err := dbkit.SqlTemplate("user_service.findUsers").
		LocalCache("users_list", 5*time.Minute).
		Query()
	time2 = time.Since(start)
	
	if err != nil {
		log.Printf("❌ 第 2 次 Query 失败: %v", err)
	} else {
		fmt.Printf("✅ 第 2 次 Query (从本地缓存): 查询到%d条记录, 耗时 %v\n",
			len(records2), time2)
		if time2 < time1 {
			fmt.Printf("⚡ 性能提升: %.1fx 倍\n", float64(time1)/float64(time2))
		}
	}

	// 总结
	fmt.Println("\n【缓存功能总结】")
	fmt.Println("✅ SQL 模板支持本地缓存")
	fmt.Println("✅ QueryFirst、Query、Paginate 都支持缓存")
	fmt.Println("✅ 支持 DB 实例和事务中的缓存")
	fmt.Println("✅ 不同页码会单独缓存")
	fmt.Println("✅ 缓存显著提升查询性能")
}
