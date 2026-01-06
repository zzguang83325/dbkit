package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/sqlserver/models"
)

func main() {
	dsn := "sqlserver://sa:123456@192.168.10.44:1433?database=test"
	err := dbkit.OpenDatabaseWithDBName("sqlserver", dbkit.SQLServer, dsn, 25)
	if err != nil {
		log.Fatalf("SQL Server数据库连接失败: %v", err)
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
	IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'demo')
	CREATE TABLE demo (
		id INT IDENTITY(1,1) PRIMARY KEY,
		name NVARCHAR(100),
		age INT,
		salary DECIMAL(10, 2),
		is_active BIT,
		birthday DATE,
		created_at DATETIME DEFAULT GETDATE(),
		metadata NVARCHAR(MAX)
	)`
	dbkit.Use("sqlserver").Exec(sql)
}

func prepareData() {
	count, _ := dbkit.Use("sqlserver").Count("demo", "")
	if count >= 100 {
		fmt.Printf("SQL Server: Already has %d rows, skipping data preparation.\n", count)
		return
	}
	fmt.Println("SQL Server: Inserting 110 rows of data...")
	records := make([]*dbkit.Record, 0, 110)
	for i := 1; i <= 110; i++ {
		record := dbkit.NewRecord().
			Set("name", fmt.Sprintf("MSSQL_User_%d", i)).
			Set("age", 18+rand.Intn(40)).
			Set("salary", 3000.0+float64(i)).
			Set("is_active", true).
			Set("birthday", time.Now()).
			Set("metadata", "MSSQL Meta")
		records = append(records, record)
	}
	dbkit.Use("sqlserver").BatchInsert("demo", records, 100)
	fmt.Println("SQL Server: Data preparation complete.")
}

func demoRecordOperations() {
	fmt.Println("\n--- MSSQL Record Operations ---")
	records, err := dbkit.Use("sqlserver").Query("SELECT TOP 5 * FROM demo WHERE age > ?", 30)
	if err != nil {
		log.Printf("MSSQL Query failed: %v", err)
		return
	}
	fmt.Printf("Query returned %d records\n", len(records))
}

func demoDbModelOperations() {
	fmt.Println("\n--- SQL Server DbModel CRUD Operations ---")
	model := &models.Demo{}

	// 1. Insert
	newUser := &models.Demo{
		Name:      "New_MSSQL_User",
		Age:       33,
		Salary:    9500.5,
		IsActive:  1,
		Birthday:  time.Now(),
		CreatedAt: time.Now(),
		Metadata:  "MSSQL DbModel Meta",
	}
	id, err := newUser.Insert()
	if err != nil {
		log.Printf("SQL Server DbModel Insert failed: %v", err)
		return
	}
	fmt.Printf("SQL Server DbModel Insert: ID = %d\n", id)
	newUser.ID = id

	// 2. FindFirst (Read)
	foundUser, err := model.FindFirst("name = ?", "New_MSSQL_User")
	if err != nil {
		log.Printf("SQL Server DbModel FindFirst failed: %v", err)
	} else if foundUser != nil {
		fmt.Printf("SQL Server DbModel FindFirst: Found user %s, Age: %d\n", foundUser.Name, foundUser.Age)
	}

	// 3. Update
	foundUser.Age = 38
	foundUser.Salary = 10500.75
	affected, err := foundUser.Update()
	if err != nil {
		log.Printf("SQL Server DbModel Update failed: %v", err)
	} else {
		fmt.Printf("SQL Server DbModel Update: %d rows affected\n", affected)
	}

	// 4. Find (Read)
	results, err := model.Find("age >= ?", "id DESC", 30)
	if err != nil {
		log.Printf("SQL Server DbModel Find failed: %v", err)
	} else {
		fmt.Printf("SQL Server DbModel Find: %d results, first user: %s\n", len(results), results[0].Name)
	}

	// 5. Paginate (Read)
	page, err := model.Paginate(1, 10, "select * from demo where age > ? order by id ASC", 20)
	if err != nil {
		log.Printf("SQL Server DbModel Paginate failed: %v", err)
	} else {
		fmt.Printf("SQL Server DbModel Paginate: Total %d rows, current page size %d\n", page.TotalRow, len(page.List))
	}

	// 6. Delete
	affected, err = foundUser.Delete()
	if err != nil {
		log.Printf("SQL Server DbModel Delete failed: %v", err)
	} else {
		fmt.Printf("SQL Server DbModel Delete: %d rows affected\n", affected)
	}
}

func demoChainOperations() {
	fmt.Println("\n--- MSSQL Chain Operations ---")
	page, err := dbkit.Use("sqlserver").Table("demo").Where("age > ?", 20).OrderBy("id").Paginate(1, 10)
	if err != nil {
		log.Printf("MSSQL Chain Paginate failed: %v", err)
		return
	}
	fmt.Printf("MSSQL Chain Paginate: Total %d rows, current page size %d\n", page.TotalRow, len(page.List))
}

func demoCacheOperations() {
	fmt.Println("\n--- SQL Server Cache Operations ---")
	var results []models.Demo
	// First call - should hit DB and save to cache
	start := time.Now()
	err := dbkit.Use("sqlserver").Cache("mssql_demo_cache", 60).Table("demo").Where("age > ?", 35).FindToDbModel(&results)
	if err != nil {
		log.Printf("SQL Server Cache Find (1st) failed: %v", err)
	} else {
		fmt.Printf("SQL Server Cache Find (1st): %d results, took %v\n", len(results), time.Since(start))
	}

	// Second call - should hit cache
	start = time.Now()
	err = dbkit.Use("sqlserver").Cache("mssql_demo_cache", 60).Table("demo").Where("age > ?", 35).FindToDbModel(&results)
	if err != nil {
		log.Printf("SQL Server Cache Find (2nd) failed: %v", err)
	} else {
		fmt.Printf("SQL Server Cache Find (2nd): %d results, took %v (from cache)\n", len(results), time.Since(start))
	}
}

func demoUpdateDeleteOperations() {
	fmt.Println("\n--- MSSQL Update/Delete Operations ---")
	// Update
	affected, err := dbkit.Use("sqlserver").Table("demo").Where("name = ?", "MSSQL_User_1").Update(dbkit.NewRecord().Set("age", 99))
	if err != nil {
		log.Printf("MSSQL Update failed: %v", err)
	} else {
		fmt.Printf("MSSQL Update: %d rows affected\n", affected)
	}

	// Delete
	affected, err = dbkit.Use("sqlserver").Table("demo").Where("name = ?", "MSSQL_User_2").Delete()
	if err != nil {
		log.Printf("MSSQL Delete failed: %v", err)
	} else {
		fmt.Printf("MSSQL Delete: %d rows affected\n", affected)
	}
}
