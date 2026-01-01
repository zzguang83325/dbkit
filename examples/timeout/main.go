package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("DBKit 查询超时控制测试")
	fmt.Println("========================================")

	// 1. 使用带超时配置的方式初始化数据库
	fmt.Println("\n【1. 数据库初始化（带全局超时配置）】")
	config := &dbkit.Config{
		Driver:          dbkit.MySQL,
		DSN:             "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
		MaxOpen:         10,
		MaxIdle:         5,
		ConnMaxLifetime: time.Hour,
		QueryTimeout:    30 * time.Second, // 全局默认超时30秒
	}

	err := dbkit.OpenDatabaseWithConfig(config)
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
	fmt.Println("✓ 数据库连接成功，全局超时设置为 30 秒")

	// 开启调试模式查看SQL
	dbkit.SetDebugMode(true)

	// 2. 创建测试表
	fmt.Println("\n【2. 创建测试表】")
	_, err = dbkit.Exec(`DROP TABLE IF EXISTS timeout_test`)
	if err != nil {
		fmt.Printf("删除表失败: %v\n", err)
	}

	_, err = dbkit.Exec(`CREATE TABLE timeout_test (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100),
		value INT
	)`)
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}
	fmt.Println("✓ 测试表创建成功")

	// 插入测试数据
	for i := 1; i <= 5; i++ {
		record := dbkit.NewRecord().
			Set("name", fmt.Sprintf("test_%d", i)).
			Set("value", i*100)
		dbkit.Insert("timeout_test", record)
	}
	fmt.Println("✓ 测试数据插入成功")

	// 3. 测试正常查询（不超时）
	fmt.Println("\n【3. 正常查询测试（设置5秒超时）】")
	start := time.Now()
	records, err := dbkit.Timeout(5 * time.Second).Query("SELECT * FROM timeout_test")
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("✓ 查询成功，返回 %d 条记录，耗时: %v\n", len(records), elapsed)
	}

	// 4. 测试 Use().Timeout() 链式调用
	fmt.Println("\n【4. Use().Timeout() 链式调用测试】")
	start = time.Now()
	records, err = dbkit.Use("default").Timeout(3*time.Second).Query("SELECT * FROM timeout_test WHERE value > ?", 200)
	elapsed = time.Since(start)
	if err != nil {
		fmt.Printf("✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("✓ 查询成功，返回 %d 条记录，耗时: %v\n", len(records), elapsed)
	}

	// 5. 测试 Table().Timeout() 链式查询
	fmt.Println("\n【5. Table().Timeout() 链式查询测试】")
	start = time.Now()
	records, err = dbkit.Table("timeout_test").
		Where("value > ?", 100).
		OrderBy("id DESC").
		Timeout(5 * time.Second).
		Find()
	elapsed = time.Since(start)
	if err != nil {
		fmt.Printf("✗ 查询失败: %v\n", err)
	} else {
		fmt.Printf("✓ 链式查询成功，返回 %d 条记录，耗时: %v\n", len(records), elapsed)
		for _, r := range records {
			fmt.Printf("  - ID: %d, Name: %s, Value: %d\n",
				r.Int64("id"), r.Str("name"), r.Int("value"))
		}
	}

	// 6. 测试 QueryFirst 超时
	fmt.Println("\n【6. QueryFirst 超时测试】")
	start = time.Now()
	record, err := dbkit.Timeout(3 * time.Second).QueryFirst("SELECT * FROM timeout_test ORDER BY id LIMIT 1")
	elapsed = time.Since(start)
	if err != nil {
		fmt.Printf("✗ 查询失败: %v\n", err)
	} else if record != nil {
		fmt.Printf("✓ QueryFirst 成功，ID: %d, Name: %s，耗时: %v\n",
			record.Int64("id"), record.Str("name"), elapsed)
	}

	// 7. 测试 Exec 超时
	fmt.Println("\n【7. Exec 超时测试】")
	start = time.Now()
	_, err = dbkit.Timeout(3 * time.Second).Exec("UPDATE timeout_test SET value = value + 1 WHERE id = 1")
	elapsed = time.Since(start)
	if err != nil {
		fmt.Printf("✗ 更新失败: %v\n", err)
	} else {
		fmt.Printf("✓ Exec 成功，耗时: %v\n", elapsed)
	}

	// 8. 测试事务中的超时
	fmt.Println("\n【8. 事务中的超时测试】")
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		start := time.Now()
		_, err := tx.Timeout(5 * time.Second).Query("SELECT * FROM timeout_test")
		elapsed := time.Since(start)
		if err != nil {
			return err
		}
		fmt.Printf("✓ 事务内查询成功，耗时: %v\n", elapsed)

		// 事务内更新
		_, err = tx.Timeout(3 * time.Second).Exec("UPDATE timeout_test SET value = value + 10 WHERE id = 2")
		if err != nil {
			return err
		}
		fmt.Println("✓ 事务内更新成功")

		return nil
	})
	if err != nil {
		fmt.Printf("✗ 事务失败: %v\n", err)
	} else {
		fmt.Println("✓ 事务提交成功")
	}

	// 9. 测试超时触发（使用 SLEEP 函数模拟慢查询）
	fmt.Println("\n【9. 超时触发测试（SLEEP 2秒，超时设置1秒）】")
	fmt.Println("  注意：此测试预期会超时")
	start = time.Now()
	_, err = dbkit.Timeout(1 * time.Second).Query("SELECT SLEEP(2)")
	elapsed = time.Since(start)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("✓ 预期的超时错误，耗时: %v\n", elapsed)
			fmt.Println("  错误类型: context.DeadlineExceeded")
		} else {
			fmt.Printf("✓ 查询被取消，错误: %v，耗时: %v\n", err, elapsed)
		}
	} else {
		fmt.Printf("✗ 预期超时但查询成功了，耗时: %v\n", elapsed)
	}

	// 10. 测试较长超时（SLEEP 1秒，超时设置3秒）
	fmt.Println("\n【10. 长超时测试（SLEEP 1秒，超时设置3秒）】")
	start = time.Now()
	_, err = dbkit.Timeout(3 * time.Second).Query("SELECT SLEEP(1)")
	elapsed = time.Since(start)
	if err != nil {
		fmt.Printf("✗ 查询失败: %v，耗时: %v\n", err, elapsed)
	} else {
		fmt.Printf("✓ 查询成功（未超时），耗时: %v\n", elapsed)
	}

	// 11. 测试 FindFirst 链式查询超时
	fmt.Println("\n【11. FindFirst 链式查询超时测试】")
	start = time.Now()
	record, err = dbkit.Table("timeout_test").
		Where("id = ?", 1).
		Timeout(2 * time.Second).
		FindFirst()
	elapsed = time.Since(start)
	if err != nil {
		fmt.Printf("✗ 查询失败: %v\n", err)
	} else if record != nil {
		fmt.Printf("✓ FindFirst 成功，ID: %d, Value: %d，耗时: %v\n",
			record.Int64("id"), record.Int("value"), elapsed)
	}

	// 清理测试表
	fmt.Println("\n【12. 清理测试数据】")
	_, err = dbkit.Exec("DROP TABLE IF EXISTS timeout_test")
	if err != nil {
		fmt.Printf("✗ 清理失败: %v\n", err)
	} else {
		fmt.Println("✓ 测试表已删除")
	}

	fmt.Println("\n========================================")
	fmt.Println("查询超时控制测试完成")
	fmt.Println("========================================")
}
