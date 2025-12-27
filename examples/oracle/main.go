package main

import (
	"fmt"
	"log"
	"path/filepath"

	"dbkit"
)

func main() {
	fmt.Println("=== DBKit Oracle 示例 ===")

	// 初始化日志记录器（debug模式）
	logFilePath := filepath.Join(".", "log.log")
	dbkit.InitLoggerWithFile("debug", logFilePath)

	// 初始化数据库连接
	fmt.Println("1. 初始化数据库连接...")
	// Oracle 连接字符串格式: oracle://username:password@host:port/service_name
	// 示例: oracle://system:oracle@localhost:1521/XE
	// 或者使用 TNS 格式: oracle://username:password@tns_name
	dsn := "oracle://test:123456@192.168.10.44:1521/orcl"
	dbkit.OpenDatabase(dbkit.Oracle, dsn, 25)
	defer dbkit.Close()

	// 检查连接
	if err := dbkit.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("✓ 数据库连接成功")

	// 创建表
	fmt.Println("\n2. 创建示例表...")
	// 先删除旧表（Oracle 使用 DROP TABLE ... PURGE 彻底删除）
	dbkit.Exec("BEGIN EXECUTE IMMEDIATE 'DROP TABLE order_items PURGE'; EXCEPTION WHEN OTHERS THEN NULL; END;")
	dbkit.Exec("BEGIN EXECUTE IMMEDIATE 'DROP TABLE orders PURGE'; EXCEPTION WHEN OTHERS THEN NULL; END;")
	dbkit.Exec("BEGIN EXECUTE IMMEDIATE 'DROP TABLE products PURGE'; EXCEPTION WHEN OTHERS THEN NULL; END;")
	dbkit.Exec("BEGIN EXECUTE IMMEDIATE 'DROP TABLE users PURGE'; EXCEPTION WHEN OTHERS THEN NULL; END;")
	dbkit.Exec("BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE users_seq'; EXCEPTION WHEN OTHERS THEN NULL; END;")
	dbkit.Exec("BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE orders_seq'; EXCEPTION WHEN OTHERS THEN NULL; END;")
	dbkit.Exec("BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE products_seq'; EXCEPTION WHEN OTHERS THEN NULL; END;")
	dbkit.Exec("BEGIN EXECUTE IMMEDIATE 'DROP SEQUENCE order_items_seq'; EXCEPTION WHEN OTHERS THEN NULL; END;")

	// 创建序列
	dbkit.Exec("CREATE SEQUENCE users_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE")
	dbkit.Exec("CREATE SEQUENCE orders_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE")
	dbkit.Exec("CREATE SEQUENCE products_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE")
	dbkit.Exec("CREATE SEQUENCE order_items_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE")

	_, err := dbkit.Exec(`
		CREATE TABLE users (
			id NUMBER PRIMARY KEY,
			name VARCHAR2(100) NOT NULL,
			email VARCHAR2(200) NOT NULL,
			age NUMBER DEFAULT 0,
			salary NUMBER(10,2) DEFAULT 0,
			department VARCHAR2(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("创建表失败: %v", err)
	} else {
		// 创建触发器
		dbkit.Exec(`
			CREATE OR REPLACE TRIGGER users_bir
			BEFORE INSERT ON users
			FOR EACH ROW
			BEGIN
				IF :NEW.id IS NULL THEN
					SELECT users_seq.NEXTVAL INTO :NEW.id FROM dual;
				END IF;
			END;
		`)
		fmt.Println("✓ users 表创建成功")
	}

	// 创建订单表
	_, err = dbkit.Exec(`
		CREATE TABLE orders (
			id NUMBER PRIMARY KEY,
			user_id NUMBER NOT NULL,
			order_date DATE NOT NULL,
			total_amount NUMBER(10,2) NOT NULL,
			status VARCHAR2(20) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_orders_user_id FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		log.Printf("创建订单表失败: %v", err)
	} else {
		// 创建触发器
		dbkit.Exec(`
			CREATE OR REPLACE TRIGGER orders_bir
			BEFORE INSERT ON orders
			FOR EACH ROW
			BEGIN
				IF :NEW.id IS NULL THEN
					SELECT orders_seq.NEXTVAL INTO :NEW.id FROM dual;
				END IF;
			END;
		`)
		fmt.Println("✓ orders 表创建成功")
	}

	// 创建产品表
	_, err = dbkit.Exec(`
		CREATE TABLE products (
			id NUMBER PRIMARY KEY,
			name VARCHAR2(200) NOT NULL,
			category VARCHAR2(50),
			price NUMBER(10,2) NOT NULL,
			stock NUMBER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Printf("创建产品表失败: %v", err)
	} else {
		// 创建触发器
		dbkit.Exec(`
			CREATE OR REPLACE TRIGGER products_bir
			BEFORE INSERT ON products
			FOR EACH ROW
			BEGIN
				IF :NEW.id IS NULL THEN
					SELECT products_seq.NEXTVAL INTO :NEW.id FROM dual;
				END IF;
			END;
		`)
		fmt.Println("✓ products 表创建成功")
	}

	// 创建订单明细表
	_, err = dbkit.Exec(`
		CREATE TABLE order_items (
			id NUMBER PRIMARY KEY,
			order_id NUMBER NOT NULL,
			product_id NUMBER NOT NULL,
			quantity NUMBER NOT NULL,
			price NUMBER(10,2) NOT NULL,
			CONSTRAINT fk_order_items_order_id FOREIGN KEY (order_id) REFERENCES orders(id),
			CONSTRAINT fk_order_items_product_id FOREIGN KEY (product_id) REFERENCES products(id)
		)
	`)
	if err != nil {
		log.Printf("创建订单明细表失败: %v", err)
	} else {
		// 创建触发器
		dbkit.Exec(`
			CREATE OR REPLACE TRIGGER order_items_bir
			BEFORE INSERT ON order_items
			FOR EACH ROW
			BEGIN
				IF :NEW.id IS NULL THEN
					SELECT order_items_seq.NEXTVAL INTO :NEW.id FROM dual;
				END IF;
			END;
		`)
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
				u.GetString("NAME"),
				u.GetString("EMAIL"),
				u.GetInt("AGE"))
		}
	}

	// 条件查询
	fmt.Println("\n  5.2 查询年龄大于25的用户:")
	adultUsers, err := dbkit.Query("SELECT * FROM users WHERE age > ?", 25)
	if err != nil {
		log.Printf("查询年龄大于25的用户失败: %v", err)
	} else {
		for _, u := range adultUsers {
			fmt.Printf("    - %s, 年龄: %d\n", u.GetString("NAME"), u.GetInt("AGE"))
		}
	}

	// 查询单条记录
	fmt.Println("\n  5.3 查询单个用户:")
	firstUser, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)
	if err != nil {
		log.Printf("查询单个用户失败: %v", err)
	} else if firstUser != nil {
		fmt.Printf("    第一个用户: %s, 邮箱: %s\n",
			firstUser.GetString("NAME"),
			firstUser.GetString("EMAIL"))
	}

	// 按ID查询
	fmt.Println("\n  5.4 按ID查询:")
	userById, err := dbkit.QueryFirst("SELECT * FROM users WHERE id = ?", 1)
	if err != nil {
		log.Printf("按ID查询用户失败: %v", err)
	} else if userById != nil {
		fmt.Printf("    ID=1的用户: %s\n", userById.GetString("NAME"))
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

	// 分页查询
	fmt.Println("\n8. 分页查询...")

	// 8.1 使用 dbkit.Paginate (自动处理 Oracle ROWNUM 逻辑)
	fmt.Println("\n  8.1 使用 dbkit.Paginate:")
	page1, totalUsers, err := dbkit.Paginate(1, 2, "SELECT *", "users", "", "id DESC")
	if err != nil {
		log.Printf("分页查询失败: %v", err)
	} else {
		fmt.Printf("    第1页（每页2条），总用户数: %d\n", totalUsers)
		for i, u := range page1 {
			fmt.Printf("      %d. %s (ID: %d)\n", i+1, u.GetString("NAME"), u.GetInt("ID"))
		}
	}

	// 8.2 使用原生 SQL 进行 ROWNUM 分页测试:
	fmt.Println("\n  8.2 使用原生 SQL 进行 ROWNUM 分页测试:")
	rawSql := "SELECT a.* FROM (SELECT a.*, ROWNUM rn FROM (SELECT * FROM users ORDER BY id DESC) a WHERE ROWNUM <= 2) a WHERE rn > 0"
	fmt.Printf("    执行原生 SQL: %s\n", rawSql)
	rawPage, err := dbkit.Query(rawSql)
	if err != nil {
		log.Printf("原生分页查询失败: %v", err)
	} else {
		fmt.Printf("    查询结果数: %d\n", len(rawPage))
		for _, u := range rawPage {
			fmt.Printf("      - %s (RN: %d)\n", u.GetString("NAME"), u.GetInt("RN"))
		}
	}

	// 事务操作示例...
	fmt.Println("\n10. 事务操作示例...")
	err = dbkit.WithTransaction(func(tx *dbkit.Tx) error {
		// 插入数据
		userTx := dbkit.NewRecord()
		userTx.Set("name", "事务用户")
		userTx.Set("email", "tx@example.com")
		userTx.Set("age", 40)

		_, err := tx.Save("users", userTx)
		if err != nil {
			return err
		}

		// 更新数据
		_, err = tx.Exec("UPDATE users SET age = ? WHERE name = ?", 41, "事务用户")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf("事务执行失败: %v", err)
	}
	fmt.Println("✓ 事务执行成功")

	txUser, err := dbkit.QueryFirst("SELECT * FROM users WHERE name = ?", "事务用户")
	if err == nil && txUser != nil {
		fmt.Printf("  事务插入的用户: %s, 年龄: %d\n", txUser.GetString("NAME"), txUser.GetInt("AGE"))
	}

	// 连接信息
	fmt.Println("\n11. 数据库连接信息...")
	fmt.Printf("  当前数据库: %s\n", dbkit.GetCurrentDBName())
	fmt.Printf("  支持的驱动: %v\n", dbkit.SupportedDrivers())

	fmt.Println("\n=== 示例完成 ===")
}
