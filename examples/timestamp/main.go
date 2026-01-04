package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
)

// ============================================================================
// DBKit Auto Timestamp Functionality Demo
// DBKit 自动时间戳功能演示
// ============================================================================
// Features / 功能说明:
//   - Auto-fill created_at field on insert / 自动在插入时填充 created_at 字段
//   - Auto-fill updated_at field on update / 自动在更新时填充 updated_at 字段
//   - Support custom field names / 支持自定义字段名称
//   - Support configuring only created_at or updated_at / 支持只配置 created_at 或 updated_at
//   - Support disabling timestamp in specific operations / 支持在特定操作中禁用时间戳
//
// Use Cases / 使用场景:
//   - Record data creation time / 记录数据创建时间
//   - Record data last modification time / 记录数据最后修改时间
//   - Audit logs / 审计日志
//   - Data version control / 数据版本控制
// ============================================================================

func main() {
	// 1. Initialize database connection / 初始化数据库连接
	// Use SQLite in-memory database with connection pool size of 10
	// 使用 SQLite 内存数据库，连接池大小为 10
	err := dbkit.OpenDatabase(dbkit.SQLite3, ":memory:", 10)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Create test table with timestamp fields / 创建测试表，包含时间戳字段
	_, err = dbkit.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			created_at DATETIME,      -- Record creation time / 记录创建时间
			updated_at DATETIME       -- Record last update time / 记录最后更新时间
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Enable timestamp functionality (global) / 启用时间戳功能（全局）
	// Must be called before configuring tables / 必须在配置表之前调用
	dbkit.EnableTimestamps()

	// 4. Configure auto timestamps for users table / 为 users 表配置自动时间戳
	// Default uses created_at and updated_at field names / 默认使用 created_at 和 updated_at 字段名
	dbkit.ConfigTimestamps("users")

	fmt.Println("=== Auto Timestamps Demo ===")
	fmt.Println()

	// ========================================================================
	// Example 1: Auto-fill created_at on insert
	// 示例 1: 插入时自动填充 created_at
	// ========================================================================
	// Explanation: When inserting a new record, dbkit automatically sets
	// created_at to current time
	// 说明: 当插入新记录时，dbkit 会自动设置 created_at 为当前时间
	// Note: If created_at already has a value, it won't be overwritten
	// 注意: 如果记录中已有 created_at 值，则不会被覆盖
	fmt.Println("1. Inserting a new user (created_at will be auto-filled)...")
	record := dbkit.NewRecord()
	record.Set("name", "John Doe")
	record.Set("email", "john@example.com")
	id, _ := dbkit.Insert("users", record)
	fmt.Printf("   Inserted user with ID: %d\n", id)
	printUser(id)

	// ========================================================================
	// Example 2: Auto-fill updated_at on update
	// 示例 2: 更新时自动填充 updated_at
	// ========================================================================
	// Explanation: When updating a record, dbkit automatically sets
	// updated_at to current time
	// 说明: 当更新记录时，dbkit 会自动设置 updated_at 为当前时间
	// Note: created_at will not be modified
	// 注意: created_at 不会被修改
	fmt.Println("\n2. Updating user (updated_at will be auto-filled)...")
	time.Sleep(time.Second) // Wait 1 second to see time difference / 等待 1 秒，以便看到时间差异
	updateRecord := dbkit.NewRecord()
	updateRecord.Set("name", "John Updated")
	dbkit.Update("users", updateRecord, "id = ?", id)
	printUser(id)

	// ========================================================================
	// Example 3: Insert with custom created_at (won't be overwritten)
	// 示例 3: 插入时使用自定义 created_at（不会被覆盖）
	// ========================================================================
	// Explanation: If created_at is explicitly set during insert,
	// dbkit will preserve that value
	// 说明: 如果在插入时明确指定了 created_at，dbkit 会保留该值
	// Use case: Importing historical data while preserving original creation time
	// 用途: 导入历史数据时保持原始创建时间
	fmt.Println("\n3. Inserting with custom created_at (won't be overwritten)...")
	customTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	record2 := dbkit.NewRecord()
	record2.Set("name", "Jane Doe")
	record2.Set("email", "jane@example.com")
	record2.Set("created_at", customTime) // Set custom time / 指定自定义时间
	id2, _ := dbkit.Insert("users", record2)
	fmt.Printf("   Inserted user with ID: %d\n", id2)
	printUser(id2)

	// ========================================================================
	// Example 4: Use WithoutTimestamps() to disable timestamp update
	// 示例 4: 使用 WithoutTimestamps() 禁用时间戳更新
	// ========================================================================
	// Explanation: In some cases, you may want to update a record
	// without changing updated_at
	// 说明: 在某些情况下，你可能想更新记录但不改变 updated_at
	// Use case: Background tasks, data synchronization, etc.
	// 用途: 后台任务更新、数据同步等场景
	fmt.Println("\n4. Updating with WithoutTimestamps() (updated_at won't change)...")
	time.Sleep(time.Second)
	updateRecord2 := dbkit.NewRecord()
	updateRecord2.Set("email", "jane.new@example.com")
	// WithoutTimestamps() disables auto timestamp update / WithoutTimestamps() 会禁用自动时间戳更新
	dbkit.Table("users").Where("id = ?", id2).WithoutTimestamps().Update(updateRecord2)
	printUser(id2)

	// ========================================================================
	// Example 5: Custom field names
	// 示例 5: 自定义字段名称
	// ========================================================================
	// Explanation: If your table uses different field names,
	// you can configure them via ConfigTimestampsWithFields
	// 说明: 如果你的表使用不同的字段名，可以通过 ConfigTimestampsWithFields 配置
	// Use case: Compatible with different database design standards
	// 用途: 兼容不同的数据库设计规范
	fmt.Println("\n5. Demo with custom field names...")
	_, err = dbkit.Exec(`
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product TEXT NOT NULL,
			create_time DATETIME,    -- Custom creation time field name / 自定义创建时间字段名
			update_time DATETIME     -- Custom update time field name / 自定义更新时间字段名
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Configure timestamps with custom field names / 使用自定义字段名配置时间戳
	// Parameters: table name, created_at field name, updated_at field name
	// 参数: 表名, created_at 字段名, updated_at 字段名
	dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")

	orderRecord := dbkit.NewRecord()
	orderRecord.Set("product", "Laptop")
	orderId, _ := dbkit.Insert("orders", orderRecord)
	fmt.Printf("   Inserted order with ID: %d\n", orderId)
	printOrder(orderId)

	// ========================================================================
	// Example 6: Configure only created_at field
	// 示例 6: 只配置 created_at 字段
	// ========================================================================
	// Explanation: Some tables may only need creation time, not update time
	// 说明: 某些表可能只需要记录创建时间，不需要更新时间
	// Use case: Log tables, audit tables, or append-only tables
	// 用途: 日志表、审计表等只读或追加的表
	fmt.Println("\n6. Demo with only created_at field...")
	_, err = dbkit.Exec(`
		CREATE TABLE logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message TEXT NOT NULL,
			log_time DATETIME        -- Only log time / 只有日志时间
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Configure only created_at, don't configure updated_at / 只配置 created_at，不配置 updated_at
	// Parameters: table name, created_at field name / 参数: 表名, created_at 字段名
	dbkit.ConfigCreatedAt("logs", "log_time")

	logRecord := dbkit.NewRecord()
	logRecord.Set("message", "System started")
	logId, _ := dbkit.Insert("logs", logRecord)
	fmt.Printf("   Inserted log with ID: %d\n", logId)
	printLog(logId)

	fmt.Println("\n=== Demo Complete ===")
}

// ============================================================================
// Helper Function: Print user information
// 辅助函数：打印用户信息
// ============================================================================
func printUser(id int64) {
	// Query user record by ID / 查询指定 ID 的用户记录
	record, _ := dbkit.Table("users").Where("id = ?", id).FindFirst()
	if record != nil {
		fmt.Printf("   User: name=%s, email=%s, created_at=%v, updated_at=%v\n",
			record.GetString("name"),
			record.GetString("email"),
			record.Get("created_at"),
			record.Get("updated_at"))
	}
}

// ============================================================================
// Helper Function: Print order information
// 辅助函数：打印订单信息
// ============================================================================
func printOrder(id int64) {
	// Query order record by ID / 查询指定 ID 的订单记录
	record, _ := dbkit.Table("orders").Where("id = ?", id).FindFirst()
	if record != nil {
		fmt.Printf("   Order: product=%s, create_time=%v, update_time=%v\n",
			record.GetString("product"),
			record.Get("create_time"),
			record.Get("update_time"))
	}
}

// ============================================================================
// Helper Function: Print log information
// 辅助函数：打印日志信息
// ============================================================================
func printLog(id int64) {
	// Query log record by ID / 查询指定 ID 的日志记录
	record, _ := dbkit.Table("logs").Where("id = ?", id).FindFirst()
	if record != nil {
		fmt.Printf("   Log: message=%s, log_time=%v\n",
			record.GetString("message"),
			record.Get("log_time"))
	}
}
