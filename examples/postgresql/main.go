package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/lib/pq"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/postgresql/models"
)

func main() {
	// 1. 连接 PostgreSQL 数据库
	dsn := "user=test password=123456 host=192.168.10.220 port=5432 dbname=postgres sslmode=disable"
	err := dbkit.OpenDatabaseWithDBName("postgresql", dbkit.PostgreSQL, dsn, 25)
	if err != nil {
		log.Fatalf("PostgreSQL数据库连接失败: %v", err)
	}
	dbkit.SetDebugMode(true)

	setupTable()
	prepareData()
	demoRecordOperations()
	demoDbModelOperations()
	demoChainOperations()
	demoCacheOperations()
}

func setupTable() {
	sql := `
	CREATE TABLE IF NOT EXISTS demo (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100),
		age INT,
		salary DECIMAL(10, 2),
		is_active BOOLEAN,
		birthday DATE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		metadata JSONB
	)`
	_, err := dbkit.Use("postgresql").Exec(sql)
	if err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
	fmt.Println("PostgreSQL: Table 'demo' ensured.")
}

func prepareData() {
	count, _ := dbkit.Use("postgresql").Count("demo", "")
	if count >= 100 {
		return
	}
	fmt.Println("PostgreSQL: Inserting 110 rows of data...")
	records := make([]*dbkit.Record, 0, 110)
	for i := 1; i <= 110; i++ {
		record := dbkit.NewRecord().
			Set("name", fmt.Sprintf("PG_User_%d", i)).
			Set("age", 18+rand.Intn(40)).
			Set("salary", 3000.0+rand.Float64()*7000.0).
			Set("is_active", i%2 == 0).
			Set("birthday", time.Now().AddDate(-20-rand.Intn(10), 0, 0)).
			Set("metadata", fmt.Sprintf(`{"tag": "pg_%d"}`, i))
		records = append(records, record)
	}
	dbkit.Use("postgresql").BatchInsert("demo", records, 100)
	fmt.Println("PostgreSQL: Data preparation complete.")
}

func demoRecordOperations() {
	fmt.Println("\n--- PG Record Operations ---")
	records, _ := dbkit.Use("postgresql").Query("SELECT * FROM demo WHERE age > $1 LIMIT 5", 30)
	fmt.Printf("Query returned %d records\n", len(records))
}

func demoDbModelOperations() {
	fmt.Println("\n--- PostgreSQL DbModel CRUD Operations ---")
	model := &models.Demo{}

	// 1. Insert
	newUser := &models.Demo{
		Name:      "New_PG_User",
		Age:       22,
		Salary:    7777.77,
		IsActive:  true,
		Birthday:  time.Now(),
		CreatedAt: time.Now(),
		Metadata:  `{"info": "PG Meta Data"}`,
	}
	id, err := newUser.Insert()
	if err != nil {
		log.Printf("PostgreSQL DbModel Insert failed: %v", err)
		return
	}
	fmt.Printf("PostgreSQL DbModel Insert: ID = %d\n", id)
	newUser.ID = id

	// 2. FindFirst (Read)
	foundUser, err := model.FindFirst("name = ?", "New_PG_User")
	if err != nil {
		log.Printf("PostgreSQL DbModel FindFirst failed: %v", err)
	} else if foundUser != nil {
		fmt.Printf("PostgreSQL DbModel FindFirst: Found user %s, Age: %d\n", foundUser.Name, foundUser.Age)
	}

	// 3. Update
	foundUser.Age = 28
	foundUser.Salary = 8888.88
	affected, err := foundUser.Update()
	if err != nil {
		log.Printf("PostgreSQL DbModel Update failed: %v", err)
	} else {
		fmt.Printf("PostgreSQL DbModel Update: %d rows affected\n", affected)
	}

	// 4. Find (Read)
	results, err := model.Find("age >= ?", "id DESC", 20)
	if err != nil {
		log.Printf("PostgreSQL DbModel Find failed: %v", err)
	} else {
		fmt.Printf("PostgreSQL DbModel Find: %d results, first user: %s\n", len(results), results[0].Name)
	}

	// 5. Paginate (Read)
	page, err := model.Paginate(1, 10, "age > ?", "id ASC", 18)
	if err != nil {
		log.Printf("PostgreSQL DbModel Paginate failed: %v", err)
	} else {
		fmt.Printf("PostgreSQL DbModel Paginate: Total %d rows, current page size %d\n", page.TotalRow, len(page.List))
	}

	// 6. Delete
	affected, err = foundUser.Delete()
	if err != nil {
		log.Printf("PostgreSQL DbModel Delete failed: %v", err)
	} else {
		fmt.Printf("PostgreSQL DbModel Delete: %d rows affected\n", affected)
	}
}

func demoChainOperations() {
	fmt.Println("\n--- PG Chain Operations ---")
	page, _ := dbkit.Use("postgresql").Table("demo").Where("age > ?", 20).Paginate(1, 10)
	fmt.Printf("Chain Paginate: Total %d rows\n", page.TotalRow)
}

func demoCacheOperations() {
	fmt.Println("\n--- PostgreSQL Cache Operations ---")
	var results []models.Demo
	// First call - should hit DB and save to cache
	start := time.Now()
	err := dbkit.Use("postgresql").Cache("pg_demo_cache", 60).Table("demo").Where("age > ?", 35).FindToDbModel(&results)
	if err != nil {
		log.Printf("PostgreSQL Cache Find (1st) failed: %v", err)
	} else {
		fmt.Printf("PostgreSQL Cache Find (1st): %d results, took %v\n", len(results), time.Since(start))
	}

	// Second call - should hit cache
	start = time.Now()
	err = dbkit.Use("postgresql").Cache("pg_demo_cache", 60).Table("demo").Where("age > ?", 35).FindToDbModel(&results)
	if err != nil {
		log.Printf("PostgreSQL Cache Find (2nd) failed: %v", err)
	} else {
		fmt.Printf("PostgreSQL Cache Find (2nd): %d results, took %v (from cache)\n", len(results), time.Since(start))
	}
}
