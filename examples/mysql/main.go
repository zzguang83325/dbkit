package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/mysql/models"
)

func main() {
	// 1. 连接 MySQL 数据库
	dsn := "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabaseWithDBName("mysql", dbkit.MySQL, dsn, 25)
	if err != nil {
		log.Fatalf("MySQL数据库连接失败: %v", err)
	}
	dbkit.SetDebugMode(true)

	// 2. 初始化环境: 创建表
	setupTable()

	// 3. 准备数据: 插入 100 条以上的数据
	prepareData()

	// 4. Record 操作演示
	demoRecordOperations()

	// 5. DbModel 操作演示
	demoDbModelOperations()

	// 6. 链式调用演示
	demoChainOperations()

	// 7. 缓存使用演示
	demoCacheOperations()
}

func setupTable() {
	sql := `
	CREATE TABLE IF NOT EXISTS demo (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100),
		age INT,
		salary DECIMAL(10, 2),
		is_active BOOLEAN,
		birthday DATE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT
	)`
	_, err := dbkit.Use("mysql").Exec(sql)
	if err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
	fmt.Println("MySQL: Table 'demo' ensured.")
}

func prepareData() {
	count, _ := dbkit.Use("mysql").Count("demo", "")
	if count >= 100 {
		fmt.Printf("MySQL: Already has %d rows, skipping data preparation.\n", count)
		return
	}

	fmt.Println("MySQL: Inserting 110 rows of data...")
	records := make([]*dbkit.Record, 0, 110)
	for i := 1; i <= 110; i++ {
		record := dbkit.NewRecord().
			Set("name", fmt.Sprintf("User_%d", i)).
			Set("age", 18+rand.Intn(40)).
			Set("salary", 3000.0+rand.Float64()*7000.0).
			Set("is_active", i%2 == 0).
			Set("birthday", time.Now().AddDate(-20-rand.Intn(10), 0, 0)).
			Set("metadata", fmt.Sprintf(`{"tag": "tag_%d", "info": "info_%d"}`, i, i))
		records = append(records, record)
	}
	dbkit.Use("mysql").BatchInsert("demo", records, 100)
	fmt.Println("MySQL: Data preparation complete.")
}

func demoRecordOperations() {
	fmt.Println("\n--- Record Operations ---")

	// 插入
	newRec := dbkit.NewRecord().Set("name", "RecordUser").Set("age", 25)
	id, _ := dbkit.Use("mysql").Insert("demo", newRec)
	fmt.Printf("Inserted Record ID: %d\n", id)

	// 多条件查询
	records, _ := dbkit.Use("mysql").Query("SELECT * FROM demo WHERE age > ? AND is_active = ?", 30, true)
	fmt.Printf("Query returned %d records\n", len(records))

	// 分页查询
	page, _ := dbkit.Use("mysql").Paginate(1, 10, "*", "demo", "age > ?", "id DESC", 20)
	fmt.Printf("Paginate: Page %d/%d, TotalRows: %d, ListSize: %d\n", page.PageNumber, page.TotalPage, page.TotalRow, len(page.List))

	// 更新
	updateRec := dbkit.NewRecord().Set("salary", 9999.99)
	rows, _ := dbkit.Use("mysql").Update("demo", updateRec, "id = ?", id)
	fmt.Printf("Updated %d rows\n", rows)

	// 删除
	dbkit.Use("mysql").Delete("demo", "id = ?", id)
}

func demoDbModelOperations() {
	fmt.Println("\n--- MySQL DbModel CRUD Operations ---")
	model := &models.Demo{}

	// 1. Insert
	newUser := &models.Demo{
		Name:      "New_MySQL_User",
		Age:       28,
		Salary:    8888.88,
		IsActive:  1,
		Birthday:  time.Now(),
		CreatedAt: time.Now(),
		Metadata:  `{"tag": "new", "info": "model_test"}`,
	}
	id, err := newUser.Insert()
	if err != nil {
		log.Printf("MySQL DbModel Insert failed: %v", err)
		return
	}
	fmt.Printf("MySQL DbModel Insert: ID = %d\n", id)
	newUser.ID = id

	// 2. FindFirst (Read)
	foundUser, err := model.FindFirst("name = ?", "New_MySQL_User")
	if err != nil {
		log.Printf("MySQL DbModel FindFirst failed: %v", err)
	} else if foundUser != nil {
		fmt.Printf("MySQL DbModel FindFirst: Found user %s, Age: %d\n", foundUser.Name, foundUser.Age)
	}

	// 3. Update
	foundUser.Age = 35
	foundUser.Salary = 9999.99
	affected, err := foundUser.Update()
	if err != nil {
		log.Printf("MySQL DbModel Update failed: %v", err)
	} else {
		fmt.Printf("MySQL DbModel Update: %d rows affected\n", affected)
	}

	// 4. Find (Read)
	results, err := model.Find("age >= ?", "id DESC", 30)
	if err != nil {
		log.Printf("MySQL DbModel Find failed: %v", err)
	} else {
		fmt.Printf("MySQL DbModel Find: %d results, first user: %s\n", len(results), results[0].Name)
	}

	// 5. Paginate (Read)
	page, err := model.Paginate(1, 10, "age > ?", "id ASC", 20)
	if err != nil {
		log.Printf("MySQL DbModel Paginate failed: %v", err)
	} else {
		fmt.Printf("MySQL DbModel Paginate: Total %d rows, current page size %d\n", page.TotalRow, len(page.List))
	}

	// 6. Delete
	affected, err = foundUser.Delete()
	if err != nil {
		log.Printf("MySQL DbModel Delete failed: %v", err)
	} else {
		fmt.Printf("MySQL DbModel Delete: %d rows affected\n", affected)
	}
}

func demoChainOperations() {
	fmt.Println("\n--- Chain Operations (QueryBuilder) ---")

	// 链式查询多条件 + 排序 + 限制
	records, _ := dbkit.Use("mysql").Table("demo").
		Where("age >= ?", 20).
		Where("salary > ?", 5000).
		OrderBy("age DESC, salary ASC").
		Limit(5).
		Find()

	fmt.Printf("Chain Query: %d results\n", len(records))
	for i, r := range records {
		fmt.Printf("  [%d] %s (Age: %v, Salary: %v)\n", i, r.Get("name"), r.Get("age"), r.Get("salary"))
	}

	// 链式分页
	page, _ := dbkit.Use("mysql").Table("demo").
		Where("is_active = ?", true).
		OrderBy("created_at DESC").
		Paginate(1, 10)
	fmt.Printf("Chain Paginate: Total %d rows\n", page.TotalRow)
}

func demoCacheOperations() {
	fmt.Println("\n--- MySQL Cache Operations ---")
	var results []models.Demo
	// First call - should hit DB and save to cache
	start := time.Now()
	err := dbkit.Use("mysql").Cache("mysql_demo_cache", 60).Table("demo").Where("age > ?", 35).FindToDbModel(&results)
	if err != nil {
		log.Printf("MySQL Cache Find (1st) failed: %v", err)
	} else {
		fmt.Printf("MySQL Cache Find (1st): %d results, took %v\n", len(results), time.Since(start))
	}

	// Second call - should hit cache
	start = time.Now()
	err = dbkit.Use("mysql").Cache("mysql_demo_cache", 60).Table("demo").Where("age > ?", 35).FindToDbModel(&results)
	if err != nil {
		log.Printf("MySQL Cache Find (2nd) failed: %v", err)
	} else {
		fmt.Printf("MySQL Cache Find (2nd): %d results, took %v (from cache)\n", len(results), time.Since(start))
	}

	// Test Paginate cache
	fmt.Println("\n--- MySQL Paginate Cache Operations ---")
	start = time.Now()
	page, err := dbkit.Use("mysql").Cache("mysql_page_cache", 60).Table("demo").Where("age > ?", 30).Paginate(1, 10)
	if err != nil {
		log.Printf("MySQL Paginate Cache (1st) failed: %v", err)
	} else {
		fmt.Printf("MySQL Paginate Cache (1st): %d results, took %v\n", len(page.List), time.Since(start))
	}

	start = time.Now()
	page, err = dbkit.Use("mysql").Cache("mysql_page_cache", 60).Table("demo").Where("age > ?", 30).Paginate(1, 10)
	if err != nil {
		log.Printf("MySQL Paginate Cache (2nd) failed: %v", err)
	} else {
		fmt.Printf("MySQL Paginate Cache (2nd): %d results, took %v (from cache)\n", len(page.List), time.Since(start))
	}

	// Test Count cache
	fmt.Println("\n--- MySQL Count Cache Operations ---")
	start = time.Now()
	count, err := dbkit.Use("mysql").Cache("mysql_count_cache", 60).Table("demo").Where("age > ?", 30).Count()
	if err != nil {
		log.Printf("MySQL Count Cache (1st) failed: %v", err)
	} else {
		fmt.Printf("MySQL Count Cache (1st): %d, took %v\n", count, time.Since(start))
	}

	start = time.Now()
	count, err = dbkit.Use("mysql").Cache("mysql_count_cache", 60).Table("demo").Where("age > ?", 30).Count()
	if err != nil {
		log.Printf("MySQL Count Cache (2nd) failed: %v", err)
	} else {
		fmt.Printf("MySQL Count Cache (2nd): %d, took %v (from cache)\n", count, time.Since(start))
	}
}
