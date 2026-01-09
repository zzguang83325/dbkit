package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples2/dbmodel/models"
)

// DbModel 使用示例 - 展示如何使用生成的 DbModel 进行数据库操作
func main() {
	// 初始化日志
	//dbkit.InitLogger("debug")

	fmt.Println("\n===========================================")
	fmt.Println("   DBKit DbModel 使用示例")
	fmt.Println("===========================================")

	// 1. 连接数据库
	fmt.Println("\n[Step 1] 连接 MySQL 数据库...")
	dsn := "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabaseWithDBName("mysql", dbkit.MySQL, dsn, 10)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	defer dbkit.Close()
	fmt.Println("✓ 连接成功")

	// 2. 配置全局特性
	fmt.Println("\n[Step 2] 配置全局特性 (软删除, 乐观锁, 自动时间戳)...")
	dbkit.InitLocalCache(time.Minute * 10)
	dbkit.EnableSoftDelete()
	dbkit.EnableOptimisticLock()
	dbkit.EnableTimestamps()

	// 配置自动时间戳 (使用默认字段名 created_at, updated_at)
	dbkit.ConfigTimestamps("users")
	dbkit.ConfigTimestamps("products")
	dbkit.ConfigTimestamps("orders")

	// 配置乐观锁 (使用默认字段名 version)
	dbkit.ConfigOptimisticLock("users")

	// 配置软删除 (使用默认字段名 deleted_at)
	dbkit.ConfigSoftDelete("products")
	dbkit.ConfigSoftDelete("users")

	fmt.Println("✓ 配置完成")

	list, err := dbkit.Query("select * from users")
	fmt.Println(len(list))

	list, err = dbkit.QueryWithOutTrashed("select * from users where id>? and username like ? ", 1, "%david%")
	fmt.Println(len(list))

	// 3. 基础 CRUD 操作
	testBasicCRUD()

	// 4. 查询操作
	testQuery()

	// 5. 分页查询
	testPagination()

	// 6. 缓存功能
	testCache()

	// 7. 软删除功能
	testSoftDelete()

	// 8. 乐观锁功能
	testOptimisticLock()

	// 9. 复杂业务场景
	testBusinessScenario()

	fmt.Println("\n===========================================")
	fmt.Println("   测试完成!")
	fmt.Println("===========================================")
}

// Helper functions for pointers
func ptrString(s string) *string    { return &s }
func ptrInt64(i int64) *int64       { return &i }
func ptrFloat64(f float64) *float64 { return &f }

// testBasicCRUD 测试基础 CRUD 操作
func testBasicCRUD() {
	fmt.Println("\n[Step 3] 测试基础 CRUD 操作...")

	uuid := uuid.New().String()
	// 1. Insert - 创建新用户
	user := &models.User{
		Username: "david" + uuid,
		Email:    "david@example.com",
		Password: "hashed_password_4",
		Balance:  ptrFloat64(1500.00),
		Status:   ptrInt64(1),
		//CreatedAt: time.Now(),
	}

	id, err := user.Insert()
	if err != nil {
		fmt.Printf("⚠️ Insert 失败 (可能已存在): %v\n", err)
		// 如果用户已存在,查询获取
		existUser := &models.User{}
		existUser, err = existUser.FindFirst("username = ?", "david")
		if err == nil && existUser.ID > 0 {
			id = existUser.ID
			user = existUser
		}
	} else {
		user.ID = id
		fmt.Printf("✓ Insert 成功, ID: %d\n", id)
	}

	// 2. FindFirst - 查询单条记录
	foundUser := &models.User{}
	foundUser, err = foundUser.FindFirst("id = ?", id)
	if err != nil {
		log.Fatalf("❌ FindFirst 失败: %v", err)
	}

	balance := 0.0
	if foundUser.Balance != nil {
		balance = *foundUser.Balance
	}
	fmt.Printf("✓ FindFirst 成功: %s (Balance: %.2f)\n", foundUser.Username, balance)

	// 3. Update - 更新记录
	foundUser.Balance = ptrFloat64(2000.00)
	foundUser.Status = ptrInt64(1)
	affected, err := foundUser.Update()
	if err != nil {
		log.Fatalf("❌ Update 失败: %v", err)
	}
	fmt.Printf("✓ Update 成功, 影响行数: %d\n", affected)

	// 4. Save - Upsert 操作
	newProduct := &models.Product{
		Name:        "iPad Pro",
		Description: ptrString("专业平板电脑"),
		Price:       6999.00,
		Stock:       ptrInt64(20),
		Category:    ptrString("电子产品"),
		IsActive:    ptrInt64(1),
	}
	savedID, err := newProduct.Save()
	if err != nil {
		log.Printf("⚠️ Save 失败: %v", err)
	} else {
		fmt.Printf("✓ Save 成功, ID: %d\n", savedID)
	}
}

// testQuery 测试查询操作
func testQuery() {
	fmt.Println("\n[Step 3] 测试查询操作...")

	// 1. Find - 查询多条记录
	userModel := &models.User{}
	users, err := userModel.Find("balance > ?", "id DESC", 500.00)
	if err != nil {
		log.Fatalf("❌ Find 失败: %v", err)
	}
	fmt.Printf("✓ Find 成功, 查询到 %d 个用户:\n", len(users))
	for _, u := range users {
		balance := 0.0
		if u.Balance != nil {
			balance = *u.Balance
		}
		fmt.Printf("  - %s (Balance: %.2f)\n", u.Username, balance)
	}

	// 2. 查询商品
	productModel := &models.Product{}
	products, err := productModel.Find("category = ? AND is_active = ?", "price DESC", "电子产品", 1)
	if err != nil {
		log.Fatalf("❌ Find Products 失败: %v", err)
	}
	fmt.Printf("✓ 查询到 %d 个电子产品:\n", len(products))
	for _, p := range products {
		stock := int64(0)
		if p.Stock != nil {
			stock = *p.Stock
		}
		fmt.Printf("  - %s (Price: %.2f, Stock: %d)\n", p.Name, p.Price, stock)
	}

	// 3. 查询订单及明细
	orderModel := &models.Order{}
	orders, err := orderModel.Find("status >= ?", "created_at DESC", 0)
	if err != nil {
		log.Fatalf("❌ Find Orders 失败: %v", err)
	}
	fmt.Printf("✓ 查询到 %d 个订单:\n", len(orders))
	for _, o := range orders {
		status := int64(0)
		if o.Status != nil {
			status = *o.Status
		}
		fmt.Printf("  - 订单号: %s, 金额: %.2f, 状态: %d\n", o.OrderNo, o.TotalAmount, status)
	}
}

// testPagination 测试分页查询
func testPagination() {
	fmt.Println("\n[Step 4] 测试分页查询...")

	// 1. 使用 PaginateBuilder (传统方式)
	userModel := &models.User{}
	page1, err := userModel.PaginateBuilder(1, 5, "balance > ?", "id DESC", 0)
	if err != nil {
		log.Fatalf("❌ PaginateBuilder 失败: %v", err)
	}
	fmt.Printf("✓ PaginateBuilder: 第 %d/%d 页, 总记录: %d\n", page1.PageNumber, page1.TotalPage, page1.TotalRow)
	for _, u := range page1.List {
		balance := 0.0
		if u.Balance != nil {
			balance = *u.Balance
		}
		fmt.Printf("  - %s (Balance: %.2f)\n", u.Username, balance)
	}

	// 2. 使用 Paginate (完整 SQL 方式 - 推荐)
	productModel := &models.Product{}
	fullSQL := "SELECT * FROM products WHERE is_active = ? ORDER BY price DESC"
	page2, err := productModel.Paginate(1, 3, fullSQL, 1)
	if err != nil {
		log.Fatalf("❌ Paginate 失败: %v", err)
	}
	fmt.Printf("✓ Paginate (完整SQL): 第 %d/%d 页, 总记录: %d\n", page2.PageNumber, page2.TotalPage, page2.TotalRow)
	for _, p := range page2.List {
		fmt.Printf("  - %s (Price: %.2f)\n", p.Name, p.Price)
	}
}

// testCache 测试缓存功能
func testCache() {
	fmt.Println("\n[Step 5] 测试缓存功能...")

	userModel := &models.User{}

	// 第一次查询 (Cache Miss)
	start := time.Now()
	user1, err := userModel.Cache("user_cache", time.Minute*5).FindFirst("username = ?", "alice")
	if err != nil {
		log.Fatalf("❌ Cache 第一次查询失败: %v", err)
	}
	duration1 := time.Since(start)
	fmt.Printf("✓ 第一次查询 (Cache Miss): %s, 耗时: %v\n", user1.Username, duration1)

	// 第二次查询 (Cache Hit)
	start = time.Now()
	user2, err := userModel.Cache("user_cache", time.Minute*5).FindFirst("username = ?", "alice")
	if err != nil {
		log.Fatalf("❌ Cache 第二次查询失败: %v", err)
	}
	duration2 := time.Since(start)
	fmt.Printf("✓ 第二次查询 (Cache Hit): %s, 耗时: %v\n", user2.Username, duration2)

	if duration2 < duration1 {
		fmt.Printf("⚡ 缓存加速: %.1fx 倍\n", float64(duration1)/float64(duration2))
	}

	// 测试分页缓存
	productModel := &models.Product{}
	start = time.Now()
	page1, _ := productModel.Cache("product_page", time.Minute*5).
		WithCountCache(time.Minute*10).
		PaginateBuilder(1, 5, "is_active = ?", "id DESC", 1)
	duration1 = time.Since(start)
	fmt.Printf("✓ 分页第一次查询: %d 条记录, 耗时: %v\n", len(page1.List), duration1)

	start = time.Now()
	page2, _ := productModel.Cache("product_page", time.Minute*5).
		WithCountCache(time.Minute*10).
		PaginateBuilder(1, 5, "is_active = ?", "id DESC", 1)
	duration2 = time.Since(start)
	fmt.Printf("✓ 分页第二次查询 (Cache Hit): %d 条记录, 耗时: %v\n", len(page2.List), duration2)
}

// testSoftDelete 测试软删除功能
func testSoftDelete() {
	fmt.Println("\n[Step 6] 测试软删除功能...")

	// 启用软删除检查
	dbkit.EnableSoftDelete()
	dbkit.ConfigSoftDelete("products")

	// 创建测试商品
	product := &models.Product{
		Name:     "测试商品-软删除",
		Price:    99.99,
		Stock:    ptrInt64(10),
		Category: ptrString("测试"),
		IsActive: ptrInt64(1),
	}
	id, _ := product.Save()
	product.ID = id

	// 执行软删除
	affected, err := product.Delete()
	if err != nil {
		log.Fatalf("❌ 软删除失败: %v", err)
	}
	fmt.Printf("✓ 软删除成功, 影响行数: %d\n", affected)

	// 普通查询应该查不到
	productModel := &models.Product{}
	found, err := productModel.FindFirst("id = ?", id)
	if err != nil {
		fmt.Println("✓ 普通查询查不到已软删除的记录")
	} else if found.ID == 0 {
		fmt.Println("✓ 普通查询查不到已软删除的记录 (返回空记录)")
	} else {
		fmt.Printf("⚠️ 意外：普通查询仍能查到软删除记录: %s\n", found.Name)
	}
	// 使用 FindWithTrashed 可以查到
	trashedProducts, _ := productModel.FindWithTrashed("id = ?", "", id)
	if len(trashedProducts) > 0 {
		fmt.Printf("✓ FindWithTrashed 查询到软删除记录: %s\n", trashedProducts[0].Name)
	}

	// 使用 FindOnlyTrashed 只查询已删除的记录
	onlyTrashedProducts, _ := productModel.FindOnlyTrashed("id = ?", "", id)
	if len(onlyTrashedProducts) > 0 {
		fmt.Printf("✓ FindOnlyTrashed 只查询到已删除记录: %s\n", onlyTrashedProducts[0].Name)
	}

	// 恢复软删除的记录
	if len(trashedProducts) > 0 {
		restored, err := trashedProducts[0].Restore()
		if err != nil {
			log.Printf("⚠️ 恢复失败: %v", err)
		} else {
			fmt.Printf("✓ 恢复成功, 影响行数: %d\n", restored)
		}
	}

	// 物理删除
	if len(trashedProducts) > 0 {
		forceDeleted, err := trashedProducts[0].ForceDelete()
		if err != nil {
			log.Printf("⚠️ 物理删除失败: %v", err)
		} else {
			fmt.Printf("✓ 物理删除成功, 影响行数: %d\n", forceDeleted)
		}
	}
}

// testOptimisticLock 测试乐观锁功能
func testOptimisticLock() {
	fmt.Println("\n[Step 7] 测试乐观锁功能...")

	// 启用乐观锁检查
	dbkit.EnableOptimisticLock()

	// 查询用户
	userModel := &models.User{}
	user, err := userModel.FindFirst("username = ?", "alice")
	if err != nil {
		log.Printf("⚠️ 查询失败: %v", err)
		return
	}
	if user.ID == 0 {
		log.Printf("⚠️ 用户不存在: alice")
		return
	}

	balance := 0.0
	if user.Balance != nil {
		balance = *user.Balance
	}
	fmt.Printf("✓ 当前用户: %s, Version: %d, Balance: %.2f\n", user.Username, user.Version, balance)

	// 模拟并发更新
	originalVersion := user.Version
	newBalance := balance + 100.00
	user.Balance = &newBalance

	affected, err := user.Update()
	if err != nil {
		log.Printf("❌ 更新失败 (版本冲突): %v", err)
	} else {
		fmt.Printf("✓ 更新成功, 影响行数: %d, 新 Version: %d\n", affected, originalVersion+1)
	}
}

// testBusinessScenario 测试复杂业务场景
func testBusinessScenario() {
	fmt.Println("\n[Step 8] 测试复杂业务场景 - 创建订单...")

	// 场景: 用户下单购买商品
	// 1. 查询用户
	userModel := &models.User{}
	user, err := userModel.FindFirst("username = ?", "bob")
	if err != nil {
		log.Printf("⚠️ 查询用户失败: %v", err)
		return
	}
	if user.ID == 0 {
		log.Printf("⚠️ 用户不存在: bob")
		return
	}
	fmt.Printf("✓ 用户: %s, 余额: %.2f\n", user.Username, user.Balance)

	// 2. 查询商品
	productModel := &models.Product{}
	product, err := productModel.FindFirst("name = ?", "AirPods Pro")
	if err != nil {
		log.Printf("⚠️ 查询商品失败: %v", err)
		return
	}
	if product.ID == 0 {
		log.Printf("⚠️ 商品不存在: AirPods Pro")
		return
	}
	fmt.Printf("✓ 商品: %s, 价格: %.2f, 库存: %d\n", product.Name, product.Price, product.Stock)

	// 3. 创建订单 (使用事务)
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		// 创建订单
		order := &models.Order{
			UserID:          user.ID,
			OrderNo:         fmt.Sprintf("ORD%d", time.Now().Unix()),
			TotalAmount:     product.Price,
			Status:          ptrInt64(0),
			PaymentMethod:   ptrString("余额支付"),
			ShippingAddress: ptrString("测试地址"),
		}
		orderID, err := dbkit.InsertDbModel(order)
		if err != nil {
			return fmt.Errorf("创建订单失败: %v", err)
		}
		order.ID = orderID
		fmt.Printf("  ✓ 创建订单: %s\n", order.OrderNo)

		// 创建订单明细
		orderItem := &models.OrderItem{
			OrderID:     order.ID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Price:       product.Price,
			Quantity:    1,
			Subtotal:    product.Price,
		}
		_, err = dbkit.InsertDbModel(orderItem)
		if err != nil {
			return fmt.Errorf("创建订单明细失败: %v", err)
		}
		fmt.Println("  ✓ 创建订单明细")

		// 扣减库存
		_, err = tx.Exec("UPDATE products SET stock = stock - 1 WHERE id = ? AND stock > 0", product.ID)
		if err != nil {
			return fmt.Errorf("扣减库存失败: %v", err)
		}
		fmt.Println("  ✓ 扣减库存")

		// 扣减余额
		_, err = tx.Exec("UPDATE users SET balance = balance - ? WHERE id = ? AND balance >= ?",
			product.Price, user.ID, product.Price)
		if err != nil {
			return fmt.Errorf("扣减余额失败: %v", err)
		}
		fmt.Println("  ✓ 扣减余额")

		return nil
	})

	if err != nil {
		fmt.Printf("❌ 下单失败: %v\n", err)
	} else {
		fmt.Println("✓ 下单成功!")
	}

	// 4. 查询用户的所有订单
	orderModel := &models.Order{}
	userOrders, err := orderModel.Find("user_id = ?", "created_at DESC", user.ID)
	if err != nil {
		log.Printf("⚠️ 查询订单失败: %v", err)
	} else {
		fmt.Printf("✓ 用户 %s 的订单列表 (%d 个):\n", user.Username, len(userOrders))
		for _, o := range userOrders {
			status := int64(0)
			if o.Status != nil {
				status = *o.Status
			}
			fmt.Printf("  - %s: %.2f 元 (状态: %d)\n", o.OrderNo, o.TotalAmount, status)
		}
	}
}
