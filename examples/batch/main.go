package main

import (
	"fmt"
	"time"

	"github.com/zzguang83325/dbkit"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("DBKit 批量操作测试")
	fmt.Println("========================================")

	// 1. 初始化数据库
	fmt.Println("\n【1. 数据库初始化】")
	err := dbkit.OpenDatabase(dbkit.MySQL,
		"root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local", 10)
	if err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		fmt.Println("\n请修改 DSN 配置后重试")
		return
	}
	defer dbkit.Close()

	if err := dbkit.Ping(); err != nil {
		fmt.Printf("数据库 Ping 失败: %v\n", err)
		return
	}
	fmt.Println("✓ 数据库连接成功")

	dbkit.SetDebugMode(true)

	// 2. 创建测试表
	fmt.Println("\n【2. 创建测试表】")
	_, _ = dbkit.Exec(`DROP TABLE IF EXISTS batch_test`)
	_, err = dbkit.Exec(`CREATE TABLE batch_test (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100),
		age INT,
		status VARCHAR(20),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}
	fmt.Println("✓ 测试表创建成功")

	// 3. 批量插入测试
	fmt.Println("\n【3. 批量插入测试】")
	var insertRecords []*dbkit.Record
	for i := 1; i <= 20; i++ {
		record := dbkit.NewRecord().
			Set("name", fmt.Sprintf("user_%d", i)).
			Set("age", 20+i%10).
			Set("status", "active")
		insertRecords = append(insertRecords, record)
	}

	start := time.Now()
	affected, err := dbkit.BatchInsertDefault("batch_test", insertRecords)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("✗ 批量插入失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 批量插入成功，影响行数: %d，耗时: %v\n", affected, elapsed)

	// 查询验证
	count, _ := dbkit.Count("batch_test", "")
	fmt.Printf("  当前表中记录数: %d\n", count)

	// 4. 批量更新测试
	fmt.Println("\n【4. 批量更新测试】")

	// 先查询出要更新的记录
	records, err := dbkit.Query("SELECT id, name, age, status FROM batch_test WHERE id <= 10")
	if err != nil {
		fmt.Printf("✗ 查询失败: %v\n", err)
		return
	}

	// 修改记录
	var updateRecords []*dbkit.Record
	for _, r := range records {
		record := dbkit.NewRecord().
			Set("id", r.Int64("id")).
			Set("name", r.Str("name")+"_updated").
			Set("age", r.Int("age")+100).
			Set("status", "updated")
		updateRecords = append(updateRecords, record)
	}

	start = time.Now()
	affected, err = dbkit.BatchUpdateDefault("batch_test", updateRecords)
	elapsed = time.Since(start)
	if err != nil {
		fmt.Printf("✗ 批量更新失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 批量更新成功，影响行数: %d，耗时: %v\n", affected, elapsed)

	// 验证更新结果
	record, _ := dbkit.QueryFirst("SELECT * FROM batch_test WHERE id = 1")
	if record != nil {
		fmt.Printf("  验证 ID=1: name=%s, age=%d, status=%s\n",
			record.Str("name"), record.Int("age"), record.Str("status"))
	}

	// 5. 批量删除测试（使用 Record）
	fmt.Println("\n【5. 批量删除测试（使用 Record）】")

	// 准备要删除的记录（只需要主键）
	var deleteRecords []*dbkit.Record
	for i := 1; i <= 5; i++ {
		record := dbkit.NewRecord().Set("id", i)
		deleteRecords = append(deleteRecords, record)
	}

	start = time.Now()
	affected, err = dbkit.BatchDeleteDefault("batch_test", deleteRecords)
	elapsed = time.Since(start)
	if err != nil {
		fmt.Printf("✗ 批量删除失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 批量删除成功，影响行数: %d，耗时: %v\n", affected, elapsed)

	count, _ = dbkit.Count("batch_test", "")
	fmt.Printf("  删除后记录数: %d\n", count)

	// 6. 批量删除测试（使用 ID 列表）
	fmt.Println("\n【6. 批量删除测试（使用 ID 列表）】")

	ids := []interface{}{6, 7, 8, 9, 10}
	start = time.Now()
	affected, err = dbkit.BatchDeleteByIdsDefault("batch_test", ids)
	elapsed = time.Since(start)
	if err != nil {
		fmt.Printf("✗ 批量删除失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 批量删除成功，影响行数: %d，耗时: %v\n", affected, elapsed)

	count, _ = dbkit.Count("batch_test", "")
	fmt.Printf("  删除后记录数: %d\n", count)

	// 7. 事务中的批量操作
	fmt.Println("\n【7. 事务中的批量操作】")

	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		// 事务内批量插入
		var txRecords []*dbkit.Record
		for i := 100; i <= 105; i++ {
			record := dbkit.NewRecord().
				Set("name", fmt.Sprintf("tx_user_%d", i)).
				Set("age", 30).
				Set("status", "tx_active")
			txRecords = append(txRecords, record)
		}

		affected, err := tx.BatchInsertDefault("batch_test", txRecords)
		if err != nil {
			return err
		}
		fmt.Printf("  事务内批量插入: %d 条\n", affected)

		// 事务内批量更新
		records, err := tx.Query("SELECT id, name, age, status FROM batch_test WHERE status = 'tx_active'")
		if err != nil {
			return err
		}

		var updateRecords []*dbkit.Record
		for _, r := range records {
			record := dbkit.NewRecord().
				Set("id", r.Int64("id")).
				Set("name", r.Str("name")).
				Set("age", r.Int("age")+10).
				Set("status", "tx_updated")
			updateRecords = append(updateRecords, record)
		}

		affected, err = tx.BatchUpdateDefault("batch_test", updateRecords)
		if err != nil {
			return err
		}
		fmt.Printf("  事务内批量更新: %d 条\n", affected)

		return nil
	})

	if err != nil {
		fmt.Printf("✗ 事务失败: %v\n", err)
	} else {
		fmt.Println("✓ 事务提交成功")
	}

	// 8. 使用 Use() 的批量操作
	fmt.Println("\n【8. Use() 批量操作】")

	var useRecords []*dbkit.Record
	for i := 200; i <= 203; i++ {
		record := dbkit.NewRecord().
			Set("name", fmt.Sprintf("use_user_%d", i)).
			Set("age", 25).
			Set("status", "use_active")
		useRecords = append(useRecords, record)
	}

	affected, err = dbkit.Use("default").BatchInsertDefault("batch_test", useRecords)
	if err != nil {
		fmt.Printf("✗ Use() 批量插入失败: %v\n", err)
	} else {
		fmt.Printf("✓ Use() 批量插入成功: %d 条\n", affected)
	}

	// 9. 查看最终结果
	fmt.Println("\n【9. 最终数据统计】")
	count, _ = dbkit.Count("batch_test", "")
	fmt.Printf("  总记录数: %d\n", count)

	// 按状态统计
	statusRecords, _ := dbkit.Query("SELECT status, COUNT(*) as cnt FROM batch_test GROUP BY status")
	fmt.Println("  按状态统计:")
	for _, r := range statusRecords {
		fmt.Printf("    - %s: %d 条\n", r.Str("status"), r.Int64("cnt"))
	}

	// 10. 清理
	fmt.Println("\n【10. 清理测试数据】")
	_, err = dbkit.Exec("DROP TABLE IF EXISTS batch_test")
	if err != nil {
		fmt.Printf("✗ 清理失败: %v\n", err)
	} else {
		fmt.Println("✓ 测试表已删除")
	}

	fmt.Println("\n========================================")
	fmt.Println("批量操作测试完成")
	fmt.Println("========================================")
}
