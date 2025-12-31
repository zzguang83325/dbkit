package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/comprehensive/models"
)

func main() {
	// 1. 初始化数据库与日志
	// 使用 SQLite 内存模式
	dbPath := "./comprehensive.db"
	err := dbkit.OpenDatabase(dbkit.SQLite3, dbPath, 10)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer dbkit.Close()

	// 2. 初始化环境 (创建表)
	setupTables()

	// 3. 准备测试数据
	prepareData()
	// 开启 Debug 模式输出 SQL
	dbkit.SetDebugMode(true)
	fmt.Println("\n" + stringsRepeat("=", 50))
	fmt.Println("DBKit 综合示例演示")
	fmt.Println(stringsRepeat("=", 50))

	// --- 演示 1: 链式查询 (QueryBuilder) ---
	fmt.Println("\n[演示 1: 链式查询]")
	users, err := dbkit.Table("users").
		Select("username, age").
		Where("age > ?", 20).
		OrderBy("age DESC").
		Limit(5).
		Find()
	if err == nil {
		for i, u := range users {
			fmt.Printf("  %d. 用户: %-10s 年龄: %d\n", i+1, u.Str("username"), u.Int("age"))
		}
	}

	// --- 演示 2: 复杂 JOIN 查询 ---
	fmt.Println("\n[演示 2: 复杂 JOIN 查询 (User + Order)]")
	// 查询消费金额大于 100 的用户订单详情
	joinSQL := `
		SELECT u.username, o.amount, o.status, o.created_at
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		WHERE o.amount > ?
		ORDER BY o.amount DESC limit 5
	`
	type UserOrderDTO struct {
		Username    string    `column:"username"`   // 来自 users 表
		OrderAmount float64   `column:"amount"`     // 来自 orders 表
		OrderStatus string    `column:"status"`     // 来自 orders 表
		OrderTime   time.Time `column:"created_at"` // 来自 orders 表
	}

	var details []UserOrderDTO
	err = dbkit.QueryToDbModel(&details, joinSQL, 100.0)
	if err == nil {
		for _, d := range details {
			fmt.Printf("  用户: %-10s 金额: %8.2f 状态: %-8s 时间: %s\n",
				d.Username, d.OrderAmount, d.OrderStatus, d.OrderTime.Format("2006-01-02 15:04"))
		}
	}

	// --- 演示 3: DbModel ActiveRecord 风格操作 ---
	fmt.Println("\n[演示 3: DbModel 操作]")
	newUser := &models.User{
		Username:  "NewUser",
		Age:       25,
		CreatedAt: time.Now(),
	}
	// 插入
	id, _ := dbkit.InsertDbModel(newUser)
	fmt.Printf("  插入新用户成功，ID: %d\n", id)

	// 查询刚刚插入的用户
	s := &models.User{}
	u, err := s.FindFirst("id = ?", id)
	if err == nil {
		fmt.Printf("  查找到用户: %s, 年龄: %d\n", u.Username, u.Age)
	}

	objPage, err := s.Paginate(1, 5, "id > ?", " id desc ", 1)
	if err == nil {
		fmt.Printf("  第 %d 页，共 %d 页，共 %d 条记录\n", objPage.PageNumber, objPage.TotalPage, objPage.TotalRow)
		for _, u := range objPage.List {
			fmt.Printf("  用户: %-10s 年龄: %d\n", u.Username, u.Age)
		}
	}

	// --- 演示 4: 缓存演示 (Cache) ---
	fmt.Println("\n[演示 4: 缓存演示]")
	cacheDbName := "top_users_cache"
	// 第一次查询，会查数据库并存入缓存
	fmt.Println("  --- 第一次查询 (查数据库) ---")
	startTime := time.Now()
	users1, _ := dbkit.Table("users").
		Where("age >= ?", 30).
		Cache(cacheDbName, 10*time.Second).
		Find()
	fmt.Printf("  查询到 %d 人，耗时: %v\n", len(users1), time.Since(startTime))

	// 第二次查询，应从缓存获取
	fmt.Println("  --- 第二次查询 (应命中缓存) ---")
	startTime = time.Now()
	users2, err := dbkit.Table("users").
		Where("age >= ?", 30).
		Cache(cacheDbName, 10*time.Second).
		Find()
	if err == nil {
		fmt.Printf("  查询到 %d 人，耗时: %v (通常极快)\n", len(users2), time.Since(startTime))
	}

	// --- 演示 5: 事务处理 (Transaction) ---
	fmt.Println("\n[演示 5: 事务处理]")
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		// 1. 创建新用户
		u := dbkit.NewRecord().Set("username", "TransUser").Set("age", 28).Set("created_at", time.Now())
		uid, err := tx.Insert("users", u)
		if err != nil {
			return err
		}

		// 2. 为该用户创建订单
		o := dbkit.NewRecord().Set("user_id", uid).Set("amount", 999.9).Set("status", "PAID").Set("created_at", time.Now())
		_, err = tx.Insert("orders", o)
		return err
	})
	if err == nil {
		fmt.Println("  事务执行成功：用户与订单已同步创建")
	}

	// --- 演示 6: 分页查询 (Paginate) ---
	fmt.Println("\n[演示 6: 分页查询]")
	page, err := dbkit.Table("orders").
		OrderBy("id ASC").
		Paginate(1, 5) // 第 1 页，每页 5 条
	if err == nil {
		fmt.Printf("  当前第 %d 页 / 总共 %d 页 (总记录数: %d)\n", page.PageNumber, page.TotalPage, page.TotalRow)
		for _, r := range page.List {
			fmt.Printf("    订单ID: %d, 金额: %.2f\n", r.GetInt("id"), r.GetFloat("amount"))
		}
	}

	fmt.Println("\n" + stringsRepeat("=", 50))
	fmt.Println("综合演示结束")
	fmt.Println(stringsRepeat("=", 50))
}

func setupTables() {
	// 用户表
	userTable := `CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		age INTEGER,
		created_at DATETIME
	)`
	// 订单表
	orderTable := `CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		amount DECIMAL(10,2),
		status TEXT,
		created_at DATETIME
	)`
	dbkit.Exec(userTable)
	dbkit.Exec(orderTable)
}

func prepareData() {
	// 批量插入一些用户
	names := []string{"Alice", "Bob", "Charlie", "David", "Eve", "Frank"}
	for i, name := range names {
		// 使用 dbkit.Insert 插入，返回 int64 ID
		u := dbkit.NewRecord().Set("username", name).Set("age", 20+i*5).Set("created_at", time.Now())
		uid, err := dbkit.Insert("users", u)
		if err != nil {
			log.Printf("插入用户失败: %v", err)
			continue
		}

		// 为每个用户生成两个订单
		o1 := dbkit.NewRecord().Set("user_id", uid).Set("amount", float64(100+i*50)).Set("status", "COMPLETED").Set("created_at", time.Now())
		dbkit.Insert("orders", o1)

		o2 := dbkit.NewRecord().Set("user_id", uid).Set("amount", float64(50+i*20)).Set("status", "PENDING").Set("created_at", time.Now())
		dbkit.Insert("orders", o2)
	}
}

// 辅助函数
func stringsRepeat(s string, count int) string {
	res := ""
	for i := 0; i < count; i++ {
		res += s
	}
	return res
}
