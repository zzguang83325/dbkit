package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/sqlite/models"
)

func main() {
	dsn := "file:test_multi.db?cache=shared&mode=rwc"
	err := dbkit.OpenDatabaseWithDBName("sqlite", dbkit.SQLite3, dsn, 10)
	if err != nil {
		log.Fatalf("SQLite数据库连接失败: %v", err)
	}
	dbkit.SetDebugMode(true)

	setupTable()
	prepareData()
	demoRecordOperations()
	demoDbModelOperations()
	demoChainOperations()
	demoCacheOperations()
	demoUpdateDeleteOperations()
}

func setupTable() {
	sql := `
	CREATE TABLE IF NOT EXISTS demo (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		age INTEGER,
		salary REAL,
		is_active INTEGER,
		birthday TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT
	)`
	dbkit.Use("sqlite").Exec(sql)
}

func prepareData() {
	count, _ := dbkit.Count("demo", "")
	if count >= 100 {
		fmt.Printf("SQLite: Already has %d rows, skipping data preparation.\n", count)
		return
	}
	fmt.Println("SQLite: Inserting 110 rows of data...")
	records := make([]*dbkit.Record, 0, 110)
	for i := 1; i <= 110; i++ {
		record := dbkit.NewRecord().
			Set("name", fmt.Sprintf("SQLite_User_%d", i)).
			Set("age", 18+rand.Intn(40)).
			Set("salary", 3000.0+float64(i)).
			Set("is_active", i%2 == 0).
			Set("birthday", time.Now().Format("2006-01-02")).
			Set("metadata", "SQLite Meta")
		records = append(records, record)
	}
	dbkit.Use("sqlite").BatchInsert("demo", records, 100)
	fmt.Println("SQLite: Data preparation complete.")
}

func demoRecordOperations() {
	fmt.Println("\n--- SQLite Record Operations ---")
	records, err := dbkit.Query("SELECT * FROM demo WHERE age > ? LIMIT 5", 30)
	if err != nil {
		log.Printf("SQLite Query failed: %v", err)
		return
	}
	fmt.Printf("Query returned %d records\n", len(records))
}

func demoDbModelOperations() {
	fmt.Println("\n--- SQLite DbModel CRUD Operations ---")
	model := &models.Demo{}

	// 1. Insert
	newUser := &models.Demo{
		Name:      "New_DbModel_User",
		Age:       25,
		Salary:    5000.5,
		IsActive:  1,
		Birthday:  time.Now().Format("2006-01-02"),
		CreatedAt: time.Now(),
		Metadata:  "DbModel Meta",
	}
	id, err := newUser.Insert()
	if err != nil {
		log.Printf("SQLite DbModel Insert failed: %v", err)
		return
	}
	fmt.Printf("SQLite DbModel Insert: ID = %d\n", id)
	newUser.ID = id

	// 2. FindFirst (Read)
	foundUser, err := model.FindFirst("name = ?", "New_DbModel_User")
	if err != nil {
		log.Printf("SQLite DbModel FindFirst failed: %v", err)
	} else if foundUser != nil {
		fmt.Printf("SQLite DbModel FindFirst: Found user %s, Age: %d\n", foundUser.Name, foundUser.Age)
	}

	// 3. Update
	foundUser.Age = 30
	foundUser.Salary = 6000.7
	affected, err := foundUser.Update()
	if err != nil {
		log.Printf("SQLite DbModel Update failed: %v", err)
	} else {
		fmt.Printf("SQLite DbModel Update: %d rows affected\n", affected)
	}

	// 4. Find (Read)
	results, err := model.Find("age >= ?", "id DESC", 25)
	if err != nil {
		log.Printf("SQLite DbModel Find failed: %v", err)
	} else {
		fmt.Printf("SQLite DbModel Find: %d results, first user: %s\n", len(results), results[0].Name)
	}

	// 5. Paginate (Read)
	page, err := model.PaginateBuilder(1, 10, "age > ?", "id ASC", 20)
	if err != nil {
		log.Printf("SQLite DbModel Paginate failed: %v", err)
	} else {
		fmt.Printf("SQLite DbModel Paginate: Total %d rows, current page size %d\n", page.TotalRow, len(page.List))
	}

	// 6. Delete
	affected, err = foundUser.Delete()
	if err != nil {
		log.Printf("SQLite DbModel Delete failed: %v", err)
	} else {
		fmt.Printf("SQLite DbModel Delete: %d rows affected\n", affected)
	}
}

func demoChainOperations() {
	fmt.Println("\n--- SQLite Chain Operations ---")
	page, err := dbkit.Use("sqlite").Table("demo").Where("age > ?", 20).Paginate(1, 10)
	if err != nil {
		log.Printf("SQLite Chain Paginate failed: %v", err)
		return
	}
	fmt.Printf("SQLite Chain Paginate: Total %d rows, current page size %d\n", page.TotalRow, len(page.List))
}

func demoCacheOperations() {
	fmt.Println("\n--- SQLite Cache Operations ---")
	var results []*models.Demo
	// First call - should hit DB and save to cache
	start := time.Now()
	err := dbkit.Use("sqlite").Cache("sqlite_demo_cache", 60).Table("demo").Where("age > ?", 35).FindToDbModel(&results)

	if err != nil {
		log.Printf("SQLite Cache Find (1st) failed: %v", err)
	} else {
		fmt.Printf("SQLite Cache Find (1st): %d results, took %v\n", len(results), time.Since(start))
	}

	// Second call - should hit cache
	start = time.Now()
	err = dbkit.Use("sqlite").Cache("sqlite_demo_cache", 60).Table("demo").Where("age > ?", 35).FindToDbModel(&results)
	if err != nil {
		log.Printf("SQLite Cache Find (2nd) failed: %v", err)
	} else {
		fmt.Printf("SQLite Cache Find (2nd): %d results, took %v (from cache)\n", len(results), time.Since(start))
	}

	// Test Paginate cache
	fmt.Println("\n--- SQLite Paginate Cache Operations ---")
	start = time.Now()
	page, err := dbkit.Use("sqlite").Cache("sqlite_page_cache", 60).Table("demo").Where("age > ?", 30).Paginate(1, 10)
	if err != nil {
		log.Printf("SQLite Paginate Cache (1st) failed: %v", err)
	} else {
		fmt.Printf("SQLite Paginate Cache (1st): %d results, took %v\n", len(page.List), time.Since(start))
	}

	start = time.Now()
	page, err = dbkit.Use("sqlite").Cache("sqlite_page_cache", 60).Table("demo").Where("age > ?", 30).Paginate(1, 10)
	if err != nil {
		log.Printf("SQLite Paginate Cache (2nd) failed: %v", err)
	} else {
		fmt.Printf("SQLite Paginate Cache (2nd): %d results, took %v (from cache)\n", len(page.List), time.Since(start))
	}

	// Test Count cache
	fmt.Println("\n--- SQLite Count Cache Operations ---")
	start = time.Now()
	count, err := dbkit.Use("sqlite").Cache("sqlite_count_cache", 60).Table("demo").Where("age > ?", 30).Count()
	if err != nil {
		log.Printf("SQLite Count Cache (1st) failed: %v", err)
	} else {
		fmt.Printf("SQLite Count Cache (1st): %d, took %v\n", count, time.Since(start))
	}

	start = time.Now()
	count, err = dbkit.Use("sqlite").Cache("sqlite_count_cache", 60).Table("demo").Where("age > ?", 30).Count()
	if err != nil {
		log.Printf("SQLite Count Cache (2nd) failed: %v", err)
	} else {
		fmt.Printf("SQLite Count Cache (2nd): %d, took %v (from cache)\n", count, time.Since(start))
	}
}

func demoUpdateDeleteOperations() {
	fmt.Println("\n--- SQLite Update/Delete Operations ---")
	// Update
	affected, err := dbkit.Use("sqlite").Table("demo").Where("name = ?", "SQLite_User_1").Update(dbkit.NewRecord().Set("age", 99))
	if err != nil {
		log.Printf("SQLite Update failed: %v", err)
	} else {
		fmt.Printf("SQLite Update: %d rows affected\n", affected)
	}

	// Delete
	affected, err = dbkit.Use("sqlite").Table("demo").Where("name = ?", "SQLite_User_2").Delete()
	if err != nil {
		log.Printf("SQLite Delete failed: %v", err)
	} else {
		fmt.Printf("SQLite Delete: %d rows affected\n", affected)
	}
}
