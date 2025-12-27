package main

import (
	"fmt"
	"log"
	"path/filepath"

	"dbkit"
)

func main() {
	fmt.Println("=== DBKit SQL Server 示例 ===")

	// 初始化日志记录器（debug模式）
	logFilePath := filepath.Join(".", "log.log")
	dbkit.InitLoggerWithFile("debug", logFilePath)

	// 初始化数据库连接
	fmt.Println("1. 初始化数据库连接...")
	// SQL Server 连接字符串格式: sqlserver://username:password@host:port?database=dbname
	// 示例: sqlserver://sa:password@localhost:1433?database=test
	// 或者使用 ODBC 格式: server=localhost;port=1433;user id=sa;password=password;database=test
	dsn := "sqlserver://sa:123456@192.168.10.44:1433?database=test"
	dbkit.OpenDatabase(dbkit.SQLServer, dsn, 25)
	defer dbkit.Close()

	// 检查连接
	if err := dbkit.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("✓ 数据库连接成功")

	// 创建表
	fmt.Println("\n2. 创建示例表...")
	// 先删除旧表
	dbkit.Exec("IF OBJECT_ID('order_items', 'U') IS NOT NULL DROP TABLE order_items")
	dbkit.Exec("IF OBJECT_ID('orders', 'U') IS NOT NULL DROP TABLE orders")
	dbkit.Exec("IF OBJECT_ID('products', 'U') IS NOT NULL DROP TABLE products")
	dbkit.Exec("IF OBJECT_ID('users', 'U') IS NOT NULL DROP TABLE users")

	_, err := dbkit.Exec(`
		CREATE TABLE users (
			id INT IDENTITY(1,1) PRIMARY KEY,
			name NVARCHAR(100) NOT NULL,
			email NVARCHAR(200) NOT NULL,
			age INT DEFAULT 0,
			salary DECIMAL(10,2) DEFAULT 0,
			department NVARCHAR(50),
			created_at DATETIME DEFAULT GETDATE()
		)
	`)
	if err != nil {
		log.Printf("创建表失败: %v", err)
	} else {
		fmt.Println("✓ users 表创建成功")
	}

	// 创建订单表
	_, err = dbkit.Exec(`
		CREATE TABLE orders (
			id INT IDENTITY(1,1) PRIMARY KEY,
			user_id INT NOT NULL,
			order_date DATE NOT NULL,
			total_amount DECIMAL(10,2) NOT NULL,
			status NVARCHAR(20) DEFAULT 'pending',
			created_at DATETIME DEFAULT GETDATE(),
			CONSTRAINT fk_orders_user_id FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		log.Printf("创建订单表失败: %v", err)
	} else {
		fmt.Println("✓ orders 表创建成功")
	}

	// 创建产品表
	_, err = dbkit.Exec(`
		CREATE TABLE products (
			id INT IDENTITY(1,1) PRIMARY KEY,
			name NVARCHAR(200) NOT NULL,
			category NVARCHAR(50),
			price DECIMAL(10,2) NOT NULL,
			stock INT DEFAULT 0,
			created_at DATETIME DEFAULT GETDATE()
		)
	`)
	if err != nil {
		log.Printf("创建产品表失败: %v", err)
	} else {
		fmt.Println("✓ products 表创建成功")
	}

	// 创建订单明细表
	_, err = dbkit.Exec(`
		CREATE TABLE order_items (
			id INT IDENTITY(1,1) PRIMARY KEY,
			order_id INT NOT NULL,
			product_id INT NOT NULL,
			quantity INT NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			CONSTRAINT fk_order_items_order_id FOREIGN KEY (order_id) REFERENCES orders(id),
			CONSTRAINT fk_order_items_product_id FOREIGN KEY (product_id) REFERENCES products(id)
		)
	`)
	if err != nil {
		log.Printf("创建订单明细表失败: %v", err)
	} else {
		fmt.Println("✓ order_items 表创建成功")
	}

	// 插入数据
	fmt.Println("\n3. 插入数据...")
	user := dbkit.NewRecord()
	user.Set("name", "张三")
	user.Set("email", "zhangsan@example.com")
	user.Set("age", 25)

	id, err := dbkit.Save("users", user)
	if err != nil {
		log.Fatalf("插入失败: %v", err)
	}
	fmt.Printf("✓ 插入成功，ID: %d\n", id)

	// 批量插入
	fmt.Println("\n4. 批量插入数据...")
	user1 := dbkit.NewRecord()
	user1.Set("name", "李四")
	user1.Set("email", "lisi@example.com")
	user1.Set("age", 30)

	user2 := dbkit.NewRecord()
	user2.Set("name", "王五")
	user2.Set("email", "wangwu@example.com")
	user2.Set("age", 28)

	user3 := dbkit.NewRecord()
	user3.Set("name", "赵六")
	user3.Set("email", "zhaoliu@example.com")
	user3.Set("age", 35)

	users := []*dbkit.Record{user1, user2, user3}

	total, err := dbkit.BatchInsertDefault("users", users)
	if err != nil {
		log.Fatalf("批量插入失败: %v", err)
	}
	fmt.Printf("✓ 批量插入成功，影响行数: %d\n", total)

	// 查询数据
	fmt.Println("\n5. 查询数据...")

	// 查询所有用户
	fmt.Println("\n  5.1 查询所有用户:")
	allUsers, err := dbkit.Query("SELECT * FROM users")
	if err != nil {
		log.Printf("查询所有用户失败: %v", err)
	} else {
		for i, u := range allUsers {
			fmt.Printf("    用户%d: %s, %s, 年龄: %d\n",
				i+1,
				u.GetString("name"),
				u.GetString("email"),
				u.GetInt("age"))
		}
	}

	// 条件查询
	fmt.Println("\n  5.2 查询年龄大于25的用户:")
	adultUsers, err := dbkit.Query("SELECT * FROM users WHERE age > ?", 25)
	if err != nil {
		log.Printf("条件查询失败: %v", err)
	} else {
		for _, u := range adultUsers {
			fmt.Printf("    - %s, 年龄: %d\n", u.GetString("name"), u.GetInt("age"))
		}
	}

	// 查询单条记录
	fmt.Println("\n  5.3 查询单个用户:")
	firstUser, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)
	if err != nil {
		log.Printf("查询单个用户失败: %v", err)
	} else if firstUser != nil {
		fmt.Printf("    第一个用户: %s, 邮箱: %s\n",
			firstUser.GetString("name"),
			firstUser.GetString("email"))
	}

	// 按ID查询
	fmt.Println("\n  5.4 按ID查询:")
	userById, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)
	if err != nil {
		log.Printf("按ID查询失败: %v", err)
	} else if userById != nil {
		fmt.Printf("    ID=1的用户: %s\n", userById.GetString("name"))
	}

	// 更新数据
	fmt.Println("\n6. 更新数据...")
	userToUpdate := dbkit.NewRecord()
	userToUpdate.Set("name", "张三三")
	userToUpdate.Set("age", 26)

	affected, err := dbkit.Update("users", userToUpdate, "id = ?", 1)
	if err != nil {
		log.Fatalf("更新失败: %v", err)
	}
	fmt.Printf("✓ 更新成功，影响行数: %d\n", affected)

	// 统计查询
	fmt.Println("\n7. 统计查询...")
	count, _ := dbkit.Count("users", "")
	fmt.Printf("  用户总数: %d\n", count)

	countAdult, _ := dbkit.Count("users", "age > ?", 25)
	fmt.Printf("  年龄大于25的用户数: %d\n", countAdult)

	// 8.1 使用 dbkit.Paginate (自动处理 SQL Server OFFSET/FETCH 逻辑)
	fmt.Println("\n  8.1 使用 dbkit.Paginate:")
	page1, totalUsers, err := dbkit.Paginate(1, 2, "SELECT *", "users", "", "id DESC")
	if err != nil {
		log.Printf("分页查询失败: %v", err)
	} else {
		fmt.Printf("    第1页（每页2条），总用户数: %d\n", totalUsers)
		for i, u := range page1 {
			fmt.Printf("      %d. %s (ID: %d)\n", i+1, u.GetString("name"), u.GetInt("id"))
		}
	}

	// 删除数据
	fmt.Println("\n9. 删除数据...")
	deleted, err := dbkit.Delete("users", "id = ?", 1)
	if err != nil {
		log.Fatalf("删除失败: %v", err)
	}
	fmt.Printf("✓ 删除成功，影响行数: %d\n", deleted)

	// 事务操作
	fmt.Println("\n10. 事务操作示例...")
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		// 在事务中插入数据
		txUser := dbkit.NewRecord()
		txUser.Set("name", "事务用户")
		txUser.Set("email", "tx@example.com")
		txUser.Set("age", 40)

		_, err := tx.Save("users", txUser)
		if err != nil {
			return err
		}

		// 在事务中更新数据
		txUpdate := dbkit.NewRecord()
		txUpdate.Set("age", 41)
		_, err = tx.Update("users", txUpdate, "name = ?", "事务用户")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("事务失败: %v", err)
	} else {
		fmt.Println("✓ 事务执行成功")
	}

	// 查询事务插入的数据
	txUsers, err := dbkit.Query("SELECT * FROM users WHERE name = ?", "事务用户")
	if err != nil {
		log.Printf("查询事务插入数据失败: %v", err)
	} else if len(txUsers) > 0 {
		fmt.Printf("  事务插入的用户: %s, 年龄: %d\n",
			txUsers[0].GetString("name"),
			txUsers[0].GetInt("age"))
	}

	// 连接信息
	fmt.Println("\n11. 数据库连接信息...")
	fmt.Printf("  当前数据库: %s\n", dbkit.GetCurrentDBName())
	fmt.Printf("  支持的驱动: %v\n", dbkit.SupportedDrivers())

	fmt.Println("\n=== 示例完成 ===")
}
