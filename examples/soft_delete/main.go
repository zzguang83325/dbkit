package main

import (
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
)

// ============================================================================
// DBKit Soft Delete Functionality Demo
// DBKit 软删除功能演示
// ============================================================================
// Features / 功能说明:
//   - Soft delete: Mark record as deleted but don't remove from database
//     软删除：标记记录为已删除，但不从数据库中移除
//   - Restore: Restore soft-deleted record to active state
//     恢复：将软删除的记录恢复为活跃状态
//   - Force delete: Permanently delete record
//     强制删除：永久删除记录
//   - Query control: Support querying active, deleted, or all records
//     查询控制：支持查询活跃、已删除或所有记录
//
// Use Cases / 使用场景:
//   - User deactivation but retain historical data / 用户注销但保留历史数据
//   - Order cancellation but retain audit logs / 订单取消但保留审计日志
//   - Data recovery and undo operations / 数据恢复和撤销操作
//   - Compliance requirements (retain deleted records) / 合规性要求（保留删除记录）
//
// How It Works / 工作原理:
//   - Use deleted_at field to mark deletion status / 使用 deleted_at 字段标记删除状态
//   - deleted_at = NULL means active record / deleted_at 为 NULL 表示活跃记录
//   - deleted_at has value means deleted record / deleted_at 有值表示已删除记录
// ============================================================================

func main() {
	// 1. Initialize database connection / 初始化数据库连接
	// Use SQLite in-memory database / 使用 SQLite 内存数据库
	err := dbkit.OpenDatabase(dbkit.SQLite3, ":memory:", 10)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Create test table with soft delete field / 创建测试表，包含软删除字段
	_, err = dbkit.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			deleted_at DATETIME      -- Soft delete marker field / 软删除标记字段
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Enable soft delete functionality (global) / 启用软删除功能（全局）
	// Must be called before configuring tables / 必须在配置表之前调用
	dbkit.EnableSoftDelete()

	// 4. Configure soft delete for users table / 为 users 表配置软删除
	// Parameters: table name, deleted_at field name / 参数: 表名, deleted_at 字段名
	dbkit.ConfigSoftDelete("users", "deleted_at")

	fmt.Println("=== Soft Delete Demo ===\n")

	// ========================================================================
	// 示例 1: 插入测试数据
	// ========================================================================
	// 说明: 插入 5 个用户记录，所有记录的 deleted_at 都为 NULL（活跃状态）
	fmt.Println("1. Inserting test users...")
	for i := 1; i <= 5; i++ {
		record := dbkit.NewRecord()
		record.Set("name", fmt.Sprintf("User%d", i))
		record.Set("email", fmt.Sprintf("user%d@example.com", i))
		dbkit.Insert("users", record)
	}
	printUsers("After insert")

	// ========================================================================
	// 示例 2: 软删除记录
	// ========================================================================
	// 说明: 软删除会将 deleted_at 设置为当前时间，但记录仍在数据库中
	// 效果: 普通查询不会返回已删除的记录
	fmt.Println("\n2. Soft deleting user with id=2...")
	dbkit.Delete("users", "id = ?", 2)
	printUsers("After soft delete (normal query)")

	// ========================================================================
	// 示例 3: 查询包含已删除的记录
	// ========================================================================
	// 说明: 使用 WithTrashed() 可以查询所有记录，包括已删除的
	// 用途: 审计、恢复、数据分析等
	fmt.Println("\n3. Query with WithTrashed() - includes deleted records...")
	records, _ := dbkit.Table("users").WithTrashed().Find()
	fmt.Printf("   Found %d users (including deleted)\n", len(records))
	for _, r := range records {
		deletedAt := r.Get("deleted_at")
		status := "active"
		if deletedAt != nil {
			status = "deleted"
		}
		fmt.Printf("   - ID: %d, Name: %s, Status: %s\n", r.GetInt("id"), r.GetString("name"), status)
	}

	// ========================================================================
	// 示例 4: 只查询已删除的记录
	// ========================================================================
	// 说明: 使用 OnlyTrashed() 只返回已删除的记录
	// 用途: 查看回收站、恢复操作等
	fmt.Println("\n4. Query with OnlyTrashed() - only deleted records...")
	records, _ = dbkit.Table("users").OnlyTrashed().Find()
	fmt.Printf("   Found %d deleted users\n", len(records))
	for _, r := range records {
		fmt.Printf("   - ID: %d, Name: %s\n", r.GetInt("id"), r.GetString("name"))
	}

	// ========================================================================
	// 示例 5: 恢复已删除的记录
	// ========================================================================
	// 说明: Restore() 会将 deleted_at 设置为 NULL，恢复记录为活跃状态
	// 用途: 撤销删除操作、恢复用户账户等
	fmt.Println("\n5. Restoring user with id=2...")
	dbkit.Restore("users", "id = ?", 2)
	printUsers("After restore")

	// ========================================================================
	// 示例 6: 软删除后强制删除
	// ========================================================================
	// 说明: ForceDelete() 会永久删除记录，无法恢复
	// 用途: 清理垃圾数据、GDPR 数据删除等
	fmt.Println("\n6. Soft deleting user with id=3, then force deleting...")
	dbkit.Delete("users", "id = ?", 3)
	fmt.Println("   After soft delete:")
	records, _ = dbkit.Table("users").WithTrashed().Find()
	fmt.Printf("   Total users (with trashed): %d\n", len(records))

	// 永久删除记录
	dbkit.ForceDelete("users", "id = ?", 3)
	fmt.Println("   After force delete:")
	records, _ = dbkit.Table("users").WithTrashed().Find()
	fmt.Printf("   Total users (with trashed): %d\n", len(records))

	// ========================================================================
	// 示例 7: 使用链式查询进行软删除操作
	// ========================================================================
	// 说明: 可以使用 QueryBuilder 链式调用进行软删除操作
	// 优点: 支持复杂的 WHERE 条件
	fmt.Println("\n7. Using QueryBuilder chain methods...")
	// 软删除 id=4 的用户
	dbkit.Table("users").Where("id = ?", 4).Delete()
	fmt.Println("   Soft deleted user 4 via QueryBuilder")

	// 查询活跃用户数量
	count, _ := dbkit.Table("users").Count()
	fmt.Printf("   Active users count: %d\n", count)

	// 查询所有用户数量（包括已删除）
	count, _ = dbkit.Table("users").WithTrashed().Count()
	fmt.Printf("   All users count (with trashed): %d\n", count)

	// 恢复 id=4 的用户
	dbkit.Table("users").Where("id = ?", 4).Restore()
	fmt.Println("   Restored user 4 via QueryBuilder")
	printUsers("Final state")

	fmt.Println("\n=== Demo Complete ===")
}

// ============================================================================
// 辅助函数：打印用户列表
// ============================================================================
// 说明: 查询并打印所有活跃用户（未删除的记录）
func printUsers(label string) {
	// 使用 Find() 查询所有活跃用户
	// 软删除配置后，Find() 会自动排除已删除的记录
	records, _ := dbkit.Table("users").Find()
	fmt.Printf("   %s: Found %d active users\n", label, len(records))
	for _, r := range records {
		fmt.Printf("   - ID: %d, Name: %s, Email: %s\n",
			r.GetInt("id"), r.GetString("name"), r.GetString("email"))
	}
}
