package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/postgresql/models"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("DBKit PostgreSQL 综合测试")
	fmt.Println("========================================")

	// 1. 连接 PostgreSQL 数据库
	dsn := "user=postgres password=123456 host=127.0.0.1 port=5432 dbname=postgres sslmode=disable"
	err := dbkit.OpenDatabaseWithDBName("postgresql", dbkit.PostgreSQL, dsn, 25)
	if err != nil {
		log.Fatalf("PostgreSQL数据库连接失败: %v", err)
	}
	defer dbkit.Close()
	dbkit.SetDebugMode(true)

	fmt.Println("✓ PostgreSQL 数据库连接成功")

	setupTable()
	prepareData()
	demoRecordOperations()
	demoDbModelOperations()
	demoChainOperations()
	demoBatchOperations() // 新增批量操作测试
	demoCacheOperations()
	demoTimeoutOperations() // 新增超时测试
	demoPoolMonitoring()    // 新增连接池监控测试

	// 清理
	cleanup()
}

func setupTable() {
	fmt.Println("\n【创建测试表】")
	sql := `
	CREATE TABLE IF NOT EXISTS demo (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100),
		age INT,
		salary DECIMAL(10, 2),
		is_active BOOLEAN,
		birthday DATE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		metadata JSONB
	)`
	_, err := dbkit.Use("postgresql").Exec(sql)
	if err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
	fmt.Println("✓ 表 'demo' 创建成功")
}

func prepareData() {
	fmt.Println("\n【准备测试数据】")
	// 清空表
	dbkit.Use("postgresql").Exec("TRUNCATE TABLE demo RESTART IDENTITY")

	fmt.Println("  插入 110 条测试数据...")
	records := make([]*dbkit.Record, 0, 110)
	for i := 1; i <= 110; i++ {
		record := dbkit.NewRecord().
			Set("name", fmt.Sprintf("PG_User_%d", i)).
			Set("age", 18+rand.Intn(40)).
			Set("salary", 3000.0+rand.Float64()*7000.0).
			Set("is_active", i%2 == 0).
			Set("birthday", time.Now().AddDate(-20-rand.Intn(10), 0, 0)).
			Set("metadata", fmt.Sprintf(`{"tag": "pg_%d"}`, i))
		records = append(records, record)
	}
	affected, err := dbkit.Use("postgresql").BatchInsert("demo", records, 100)
	if err != nil {
		log.Printf("批量插入失败: %v", err)
		return
	}
	fmt.Printf("✓ 批量插入完成，影响行数: %d\n", affected)
}

func demoRecordOperations() {
	fmt.Println("\n【Record 操作测试】")
	records, err := dbkit.Use("postgresql").Query("SELECT * FROM demo WHERE age > $1 LIMIT 5", 30)
	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}
	fmt.Printf("✓ Query 返回 %d 条记录\n", len(records))
	for _, r := range records {
		fmt.Printf("  - ID: %d, Name: %s, Age: %d\n", r.Int64("id"), r.Str("name"), r.Int("age"))
	}
}

func demoDbModelOperations() {
	fmt.Println("\n【DbModel CRUD 操作测试】")
	model := &models.Demo{}

	// 1. Insert
	newUser := &models.Demo{
		Name:      "New_PG_User",
		Age:       22,
		Salary:    7777.77,
		IsActive:  true,
		Birthday:  time.Now(),
		CreatedAt: time.Now(),
		Metadata:  `{"info": "PG Meta Data"}`,
	}
	id, err := newUser.Insert()
	if err != nil {
		log.Printf("Insert 失败: %v", err)
		return
	}
	fmt.Printf("✓ Insert: ID = %d\n", id)
	newUser.ID = id

	// 2. FindFirst
	foundUser, err := model.FindFirst("name = ?", "New_PG_User")
	if err != nil {
		log.Printf("FindFirst 失败: %v", err)
	} else if foundUser != nil {
		fmt.Printf("✓ FindFirst: 找到用户 %s, Age: %d\n", foundUser.Name, foundUser.Age)
	}

	// 3. Update
	foundUser.Age = 28
	foundUser.Salary = 8888.88
	affected, err := foundUser.Update()
	if err != nil {
		log.Printf("Update 失败: %v", err)
	} else {
		fmt.Printf("✓ Update: 影响 %d 行\n", affected)
	}

	// 4. Find
	results, err := model.Find("age >= ?", "id DESC", 20)
	if err != nil {
		log.Printf("Find 失败: %v", err)
	} else {
		fmt.Printf("✓ Find: 返回 %d 条结果\n", len(results))
	}

	// 5. Paginate
	page, err := model.Paginate(1, 10, "age > ?", "id ASC", 18)
	if err != nil {
		log.Printf("Paginate 失败: %v", err)
	} else {
		fmt.Printf("✓ Paginate: 总计 %d 行，当前页 %d 条\n", page.TotalRow, len(page.List))
	}

	// 6. Delete
	affected, err = foundUser.Delete()
	if err != nil {
		log.Printf("Delete 失败: %v", err)
	} else {
		fmt.Printf("✓ Delete: 影响 %d 行\n", affected)
	}
}

func demoChainOperations() {
	fmt.Println("\n【链式查询测试】")
	page, err := dbkit.Use("postgresql").Table("demo").Where("age > ?", 20).Paginate(1, 10)
	if err != nil {
		log.Printf("链式分页查询失败: %v", err)
		return
	}
	fmt.Printf("✓ 链式分页查询: 总计 %d 行\n", page.TotalRow)
}

func demoBatchOperations() {
	fmt.Println("\n【批量操作测试】")

	// 1. 批量插入
	fmt.Println("\n  [1] 批量插入测试")
	var insertRecords []*dbkit.Record
	for i := 1; i <= 20; i++ {
		record := dbkit.NewRecord().
			Set("name", fmt.Sprintf("Batch_User_%d", i)).
			Set("age", 25+i%10).
			Set("salary", 5000.0+float64(i)*100).
			Set("is_active", true).
			Set("birthday", time.Now().AddDate(-25, 0, 0)).
			Set("metadata", fmt.Sprintf(`{"batch": %d}`, i))
		insertRecords = append(insertRecords, record)
	}

	start := time.Now()
	affected, err := dbkit.Use("postgresql").BatchInsertDefault("demo", insertRecords)
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("  ✗ 批量插入失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 批量插入成功，影响行数: %d，耗时: %v\n", affected, elapsed)

	// 2. 批量更新
	fmt.Println("\n  [2] 批量更新测试")

	// 查询刚插入的记录
	records, err := dbkit.Use("postgresql").Query("SELECT id, name, age, salary FROM demo WHERE name LIKE 'Batch_User_%' ORDER BY id LIMIT 10")
	if err != nil {
		log.Printf("  ✗ 查询失败: %v", err)
		return
	}

	var updateRecords []*dbkit.Record
	for _, r := range records {
		record := dbkit.NewRecord().
			Set("id", r.Int64("id")).
			Set("name", r.Str("name")+"_updated").
			Set("age", r.Int("age")+100).
			Set("salary", 9999.99)
		updateRecords = append(updateRecords, record)
	}

	start = time.Now()
	affected, err = dbkit.Use("postgresql").BatchUpdateDefault("demo", updateRecords)
	elapsed = time.Since(start)
	if err != nil {
		log.Printf("  ✗ 批量更新失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 批量更新成功，影响行数: %d，耗时: %v\n", affected, elapsed)

	// 验证更新结果
	record, _ := dbkit.Use("postgresql").QueryFirst("SELECT * FROM demo WHERE name LIKE '%_updated' LIMIT 1")
	if record != nil {
		fmt.Printf("  验证: name=%s, age=%d, salary=%s\n",
			record.Str("name"), record.Int("age"), record.Str("salary"))
	}

	// 3. 批量删除（使用 Record）
	fmt.Println("\n  [3] 批量删除测试（使用 Record）")

	// 查询要删除的记录
	deleteRecords, err := dbkit.Use("postgresql").Query("SELECT id FROM demo WHERE name LIKE '%_updated' LIMIT 5")
	if err != nil {
		log.Printf("  ✗ 查询失败: %v", err)
		return
	}

	var toDelete []*dbkit.Record
	for _, r := range deleteRecords {
		toDelete = append(toDelete, dbkit.NewRecord().Set("id", r.Int64("id")))
	}

	start = time.Now()
	affected, err = dbkit.Use("postgresql").BatchDeleteDefault("demo", toDelete)
	elapsed = time.Since(start)
	if err != nil {
		log.Printf("  ✗ 批量删除失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 批量删除成功，影响行数: %d，耗时: %v\n", affected, elapsed)

	// 4. 批量删除（使用 ID 列表）
	fmt.Println("\n  [4] 批量删除测试（使用 ID 列表）")

	// 获取一些 ID
	idRecords, _ := dbkit.Use("postgresql").Query("SELECT id FROM demo WHERE name LIKE 'Batch_User_%' LIMIT 5")
	var ids []interface{}
	for _, r := range idRecords {
		ids = append(ids, r.Int64("id"))
	}

	if len(ids) > 0 {
		start = time.Now()
		affected, err = dbkit.Use("postgresql").BatchDeleteByIdsDefault("demo", ids)
		elapsed = time.Since(start)
		if err != nil {
			log.Printf("  ✗ 批量删除失败: %v", err)
		} else {
			fmt.Printf("  ✓ 批量删除成功，影响行数: %d，耗时: %v\n", affected, elapsed)
		}
	}

	// 5. 事务中的批量操作
	fmt.Println("\n  [5] 事务中的批量操作")

	err = dbkit.Use("postgresql").Transaction(func(tx *dbkit.Tx) error {
		// 事务内批量插入
		var txRecords []*dbkit.Record
		for i := 1; i <= 5; i++ {
			record := dbkit.NewRecord().
				Set("name", fmt.Sprintf("TX_User_%d", i)).
				Set("age", 30).
				Set("salary", 6000.0).
				Set("is_active", true).
				Set("birthday", time.Now().AddDate(-30, 0, 0)).
				Set("metadata", `{"tx": true}`)
			txRecords = append(txRecords, record)
		}

		affected, err := tx.BatchInsertDefault("demo", txRecords)
		if err != nil {
			return err
		}
		fmt.Printf("    事务内批量插入: %d 条\n", affected)

		// 事务内批量更新
		records, err := tx.Query("SELECT id, name, age FROM demo WHERE name LIKE 'TX_User_%'")
		if err != nil {
			return err
		}

		var updateRecords []*dbkit.Record
		for _, r := range records {
			record := dbkit.NewRecord().
				Set("id", r.Int64("id")).
				Set("name", r.Str("name")+"_tx_updated").
				Set("age", r.Int("age")+50)
			updateRecords = append(updateRecords, record)
		}

		affected, err = tx.BatchUpdateDefault("demo", updateRecords)
		if err != nil {
			return err
		}
		fmt.Printf("    事务内批量更新: %d 条\n", affected)

		return nil
	})

	if err != nil {
		log.Printf("  ✗ 事务失败: %v", err)
	} else {
		fmt.Println("  ✓ 事务提交成功")
	}

	// 统计
	count, _ := dbkit.Use("postgresql").Count("demo", "")
	fmt.Printf("\n  当前表中总记录数: %d\n", count)
}

func demoCacheOperations() {
	fmt.Println("\n【缓存操作测试】")
	var results []models.Demo

	// 第一次查询 - 应该查询数据库并缓存
	start := time.Now()
	err := dbkit.Use("postgresql").Cache("pg_demo_cache", 60*time.Second).Table("demo").Where("age > ?", 35).FindToDbModel(&results)
	if err != nil {
		log.Printf("缓存查询(第1次)失败: %v", err)
	} else {
		fmt.Printf("✓ 缓存查询(第1次): %d 条结果，耗时 %v\n", len(results), time.Since(start))
	}

	// 第二次查询 - 应该命中缓存
	start = time.Now()
	err = dbkit.Use("postgresql").Cache("pg_demo_cache", 60*time.Second).Table("demo").Where("age > ?", 35).FindToDbModel(&results)
	if err != nil {
		log.Printf("缓存查询(第2次)失败: %v", err)
	} else {
		fmt.Printf("✓ 缓存查询(第2次): %d 条结果，耗时 %v (来自缓存)\n", len(results), time.Since(start))
	}
}

func demoTimeoutOperations() {
	fmt.Println("\n【超时控制测试】")

	// 正常查询（设置超时）
	start := time.Now()
	records, err := dbkit.Use("postgresql").Timeout(5 * time.Second).Query("SELECT * FROM demo LIMIT 10")
	elapsed := time.Since(start)
	if err != nil {
		log.Printf("✗ 超时查询失败: %v", err)
	} else {
		fmt.Printf("✓ 超时查询成功，返回 %d 条记录，耗时: %v\n", len(records), elapsed)
	}

	// 链式查询超时
	start = time.Now()
	records, err = dbkit.Use("postgresql").Table("demo").
		Where("age > ?", 20).
		Timeout(3 * time.Second).
		Find()
	elapsed = time.Since(start)
	if err != nil {
		log.Printf("✗ 链式超时查询失败: %v", err)
	} else {
		fmt.Printf("✓ 链式超时查询成功，返回 %d 条记录，耗时: %v\n", len(records), elapsed)
	}

	// 测试超时触发（PostgreSQL 使用 pg_sleep）
	fmt.Println("\n  测试超时触发（pg_sleep 2秒，超时设置1秒）")
	start = time.Now()
	_, err = dbkit.Use("postgresql").Timeout(1 * time.Second).Query("SELECT pg_sleep(2)")
	elapsed = time.Since(start)
	if err != nil {
		fmt.Printf("✓ 预期的超时/取消，耗时: %v\n", elapsed)
		fmt.Printf("  错误信息: %v\n", err)
	} else {
		fmt.Printf("✗ 预期超时但查询成功了，耗时: %v\n", elapsed)
	}
}

func demoPoolMonitoring() {
	fmt.Println("\n【连接池监控测试】")

	// 1. 获取单个数据库的连接池统计
	fmt.Println("\n  [1] 获取 PostgreSQL 连接池统计")
	stats := dbkit.GetPoolStatsDB("postgresql")
	if stats != nil {
		fmt.Printf("  %s\n", stats.String())
		fmt.Printf("    - 数据库: %s\n", stats.DBName)
		fmt.Printf("    - 驱动: %s\n", stats.Driver)
		fmt.Printf("    - 最大连接数: %d\n", stats.MaxOpenConnections)
		fmt.Printf("    - 当前打开: %d\n", stats.OpenConnections)
		fmt.Printf("    - 使用中: %d\n", stats.InUse)
		fmt.Printf("    - 空闲: %d\n", stats.Idle)
		fmt.Printf("    - 等待次数: %d\n", stats.WaitCount)
		fmt.Printf("    - 等待时长: %v\n", stats.WaitDuration)
	}

	// 2. 获取所有数据库的连接池统计
	fmt.Println("\n  [2] 获取所有数据库连接池统计")
	allStats := dbkit.AllPoolStats()
	for name, s := range allStats {
		fmt.Printf("    %s: Open=%d (InUse=%d, Idle=%d)\n",
			name, s.OpenConnections, s.InUse, s.Idle)
	}

	// 3. 转换为 Map（便于 JSON）
	fmt.Println("\n  [3] 转换为 Map")
	if stats != nil {
		statsMap := stats.ToMap()
		fmt.Printf("    max_open_connections: %v\n", statsMap["max_open_connections"])
		fmt.Printf("    open_connections: %v\n", statsMap["open_connections"])
	}

	// 4. Prometheus 指标
	fmt.Println("\n  [4] Prometheus 指标（部分）")
	if stats != nil {
		metrics := stats.PrometheusMetrics()
		// 只打印前几行
		lines := strings.Split(metrics, "\n")
		for i, line := range lines {
			if i < 8 && line != "" {
				fmt.Printf("    %s\n", line)
			}
		}
		fmt.Println("    ...")
	}

	// 5. 模拟并发查询后查看连接池变化
	fmt.Println("\n  [5] 并发查询后的连接池状态")
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			dbkit.Use("postgresql").Query("SELECT pg_sleep(0.1)")
		}(i)
	}

	// 查询进行中时查看状态
	time.Sleep(50 * time.Millisecond)
	stats = dbkit.GetPoolStatsDB("postgresql")
	if stats != nil {
		fmt.Printf("    并发查询中: Open=%d, InUse=%d, Idle=%d\n",
			stats.OpenConnections, stats.InUse, stats.Idle)
	}

	wg.Wait()

	// 查询完成后查看状态
	stats = dbkit.GetPoolStatsDB("postgresql")
	if stats != nil {
		fmt.Printf("    查询完成后: Open=%d, InUse=%d, Idle=%d\n",
			stats.OpenConnections, stats.InUse, stats.Idle)
	}
}

func cleanup() {
	fmt.Println("\n【清理测试数据】")
	// 可选：删除测试表
	// dbkit.Use("postgresql").Exec("DROP TABLE IF EXISTS demo")
	fmt.Println("✓ 测试完成（保留测试表供查看）")

	fmt.Println("\n========================================")
	fmt.Println("PostgreSQL 综合测试完成")
	fmt.Println("========================================")
}
