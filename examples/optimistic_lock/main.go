package main

import (
	"errors"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
)

// ============================================================================
// DBKit 乐观锁功能演示
// ============================================================================
// 功能说明:
//   - 乐观锁：通过版本号检测并发修改
//   - 自动版本递增：每次更新时版本号自动加 1
//   - 冲突检测：检测到版本不匹配时返回错误
//   - 自定义版本字段：支持自定义版本字段名称
//
// 使用场景:
//   - 防止并发更新冲突
//   - 库存管理（防止超卖）
//   - 订单状态更新
//   - 数据一致性保证
//
// 工作原理:
//   - 每条记录有一个版本号字段（默认为 version）
//   - 更新时必须指定当前版本号
//   - 如果版本号不匹配，更新失败
//   - 成功更新后版本号自动加 1
// ============================================================================

func main() {
	// 1. 初始化数据库连接
	// 使用 SQLite 内存数据库
	err := dbkit.OpenDatabase(dbkit.SQLite3, ":memory:", 10)
	if err != nil {
		log.Fatal(err)
	}

	// 2. 创建测试表，包含版本字段
	_, err = dbkit.Exec(`
		CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price REAL,
			stock INTEGER,
			version INTEGER DEFAULT 1  -- 版本号字段
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 启用乐观锁功能（全局）
	// 必须在配置表之前调用
	dbkit.EnableOptimisticLock()

	// 4. 为 products 表配置乐观锁
	// 使用默认的 version 字段名
	dbkit.ConfigOptimisticLock("products")

	fmt.Println("=== Optimistic Lock Demo ===")
	fmt.Println()

	// ========================================================================
	// 示例 1: 插入记录时自动初始化版本号
	// ========================================================================
	// 说明: 新插入的记录版本号自动设置为 1
	fmt.Println("1. Inserting a new product (version will be auto-initialized to 1)...")
	record := dbkit.NewRecord()
	record.Set("name", "Laptop")
	record.Set("price", 999.99)
	record.Set("stock", 100)
	id, _ := dbkit.Insert("products", record)
	fmt.Printf("   Inserted product with ID: %d\n", id)
	printProduct(id)

	// ========================================================================
	// 示例 2: 使用正确的版本号更新
	// ========================================================================
	// 说明: 更新时指定当前版本号，更新成功后版本号自动加 1
	// 用途: 正常的并发安全更新
	fmt.Println("\n2. Updating product with correct version...")
	updateRecord := dbkit.NewRecord()
	updateRecord.Set("version", int64(1)) // 当前版本号
	updateRecord.Set("price", 899.99)
	updateRecord.Set("stock", 95)
	rows, err := dbkit.Update("products", updateRecord, "id = ?", id)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Updated %d row(s)\n", rows)
	}
	printProduct(id)

	// ========================================================================
	// 示例 3: 使用过期的版本号更新（模拟并发冲突）
	// ========================================================================
	// 说明: 如果版本号不匹配，更新会失败并返回 ErrVersionMismatch 错误
	// 场景: 两个请求同时读取数据，第一个更新成功版本号变为 2，
	//      第二个仍使用版本号 1 更新就会失败
	fmt.Println("\n3. Simulating concurrent update with stale version (version=1, but current is 2)...")
	staleRecord := dbkit.NewRecord()
	staleRecord.Set("version", int64(1)) // 过期的版本号！
	staleRecord.Set("price", 799.99)
	rows, err = dbkit.Update("products", staleRecord, "id = ?", id)
	if errors.Is(err, dbkit.ErrVersionMismatch) {
		fmt.Println("   ✓ Detected version mismatch! Concurrent modification prevented.")
		fmt.Printf("   Error: %v\n", err)
	} else if err != nil {
		fmt.Printf("   Unexpected error: %v\n", err)
	}
	printProduct(id) // 价格应该仍为 899.99

	// ========================================================================
	// 示例 4: 正确处理并发更新的方式
	// ========================================================================
	// 说明: 先读取最新版本号，再进行更新
	// 用途: 应用程序中处理并发冲突的标准做法
	fmt.Println("\n4. Correct way: Read latest version, then update...")
	latestRecord, _ := dbkit.Table("products").Where("id = ?", id).FindFirst()
	if latestRecord != nil {
		currentVersion := latestRecord.GetInt("version")
		fmt.Printf("   Current version: %d\n", currentVersion)

		updateRecord2 := dbkit.NewRecord()
		updateRecord2.Set("version", currentVersion)
		updateRecord2.Set("price", 799.99)
		rows, err = dbkit.Update("products", updateRecord2, "id = ?", id)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
		} else {
			fmt.Printf("   Updated %d row(s)\n", rows)
		}
	}
	printProduct(id)

	// ========================================================================
	// 示例 5: 不包含版本字段的更新（跳过版本检查）
	// ========================================================================
	// 说明: 如果更新记录中不包含 version 字段，则不进行版本检查
	// 用途: 某些不需要并发控制的更新操作
	fmt.Println("\n5. Update without version field (no version check)...")
	noVersionRecord := dbkit.NewRecord()
	noVersionRecord.Set("stock", 90)
	rows, err = dbkit.Update("products", noVersionRecord, "id = ?", id)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Updated %d row(s) (no version check)\n", rows)
	}
	printProduct(id) // 注意：版本号不变，因为更新中没有包含 version 字段

	// ========================================================================
	// 示例 6: 自定义版本字段名称
	// ========================================================================
	// 说明: 如果表使用不同的版本字段名，可以通过 ConfigOptimisticLockWithField 配置
	// 用途: 兼容不同的数据库设计规范
	fmt.Println("\n6. Demo with custom version field name...")
	_, err = dbkit.Exec(`
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer TEXT NOT NULL,
			total REAL,
			revision INTEGER DEFAULT 1  -- 自定义版本字段名
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// 使用自定义版本字段名配置乐观锁
	// 参数: 表名, 版本字段名
	dbkit.ConfigOptimisticLockWithField("orders", "revision")

	orderRecord := dbkit.NewRecord()
	orderRecord.Set("customer", "John Doe")
	orderRecord.Set("total", 150.00)
	orderId, _ := dbkit.Insert("orders", orderRecord)
	fmt.Printf("   Inserted order with ID: %d\n", orderId)
	printOrder(orderId)

	// 更新订单
	orderUpdate := dbkit.NewRecord()
	orderUpdate.Set("revision", int64(1))
	orderUpdate.Set("total", 175.00)
	rows, _ = dbkit.Update("orders", orderUpdate, "id = ?", orderId)
	fmt.Printf("   Updated %d row(s)\n", rows)
	printOrder(orderId)

	// ========================================================================
	// 示例 7: 使用 UpdateRecord 进行乐观锁更新
	// ========================================================================
	// 说明: UpdateRecord 会自动从记录中提取主键和版本号
	// 优点: 更简洁的 API，自动处理版本号
	fmt.Println("\n7. Using UpdateRecord with optimistic lock...")
	product, _ := dbkit.Table("products").Where("id = ?", id).FindFirst()
	if product != nil {
		product.Set("name", "Gaming Laptop")
		// version 字段已经在查询时获取
		rows, err = dbkit.Use("default").UpdateRecord("products", product)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
		} else {
			fmt.Printf("   Updated %d row(s)\n", rows)
		}
	}
	printProduct(id)

	// ========================================================================
	// 示例 8: 事务中的乐观锁
	// ========================================================================
	// 说明: 乐观锁也可以在事务中使用
	// 用途: 确保多个操作的原子性和一致性
	fmt.Println("\n8. Transaction with optimistic lock...")
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		// 在事务中读取最新数据
		rec, err := tx.Table("products").Where("id = ?", id).FindFirst()
		if err != nil {
			return err
		}

		currentVersion := rec.GetInt("version")
		fmt.Printf("   In transaction: current version = %d\n", currentVersion)

		// 在事务中进行版本检查的更新
		updateRec := dbkit.NewRecord()
		updateRec.Set("version", currentVersion)
		updateRec.Set("stock", 80)
		_, err = tx.Update("products", updateRec, "id = ?", id)
		return err
	})
	if err != nil {
		fmt.Printf("   Transaction error: %v\n", err)
	} else {
		fmt.Println("   Transaction committed successfully")
	}
	printProduct(id)

	fmt.Println("\n=== Demo Complete ===")
}

// ============================================================================
// 辅助函数：打印产品信息
// ============================================================================
// 说明: 查询并打印产品的所有信息，包括版本号
func printProduct(id int64) {
	// 查询指定 ID 的产品记录
	record, _ := dbkit.Table("products").Where("id = ?", id).FindFirst()
	if record != nil {
		fmt.Printf("   Product: name=%s, price=%.2f, stock=%d, version=%d\n",
			record.GetString("name"),
			record.GetFloat("price"),
			record.GetInt("stock"),
			record.GetInt("version"))
	}
}

// ============================================================================
// 辅助函数：打印订单信息
// ============================================================================
// 说明: 查询并打印订单的所有信息，包括自定义版本字段
func printOrder(id int64) {
	// 查询指定 ID 的订单记录
	record, _ := dbkit.Table("orders").Where("id = ?", id).FindFirst()
	if record != nil {
		fmt.Printf("   Order: customer=%s, total=%.2f, revision=%d\n",
			record.GetString("customer"),
			record.GetFloat("total"),
			record.GetInt("revision"))
	}
}
