package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	_ "github.com/sijms/go-ora/v2"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/oracle/models"
)

func main() {
	dsn := "oracle://test:123456@192.168.10.44:1521/orcl"
	err := dbkit.OpenDatabaseWithDBName("oracle", dbkit.Oracle, dsn, 25)
	if err != nil {
		log.Fatalf("Oracle数据库连接失败: %v", err)
	}
	dbkit.SetDebugMode(true)

	// setupTable()
	prepareData()
	demoRecordOperations()
	demoDbModelOperations()
	demoChainOperations()
	demoCacheOperations()
	demoUpdateDeleteOperations()
}

func setupTable() {
	// 尝试直接创建表，忽略已存在的错误
	sql := `CREATE TABLE DEMO (
			ID NUMBER PRIMARY KEY,
			NAME VARCHAR2(100),
			AGE NUMBER,
			SALARY NUMBER(10,2),
			IS_ACTIVE NUMBER(1),
			BIRTHDAY DATE,
			CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			METADATA CLOB
		)`
	_, err := dbkit.Use("oracle").Exec(sql)
	if err != nil {
		if !strings.Contains(err.Error(), "ORA-00955") {
			log.Printf("Oracle table setup warning: %v", err)
		}
	}
}

func prepareData() {
	count, err := dbkit.Use("oracle").Count("DEMO", "")
	if err != nil {
		log.Printf("Oracle Count failed: %v", err)
		return
	}
	if count >= 100 {
		return
	}
	fmt.Println("Oracle: Inserting 110 rows of data...")
	records := make([]*dbkit.Record, 0, 110)
	for i := 1; i <= 110; i++ {
		record := dbkit.NewRecord().
			Set("ID", i).
			Set("NAME", fmt.Sprintf("Ora_User_%d", i)).
			Set("AGE", 18+rand.Intn(40)).
			Set("SALARY", 3000.0+rand.Float64()*7000.0).
			Set("IS_ACTIVE", i%2).
			Set("BIRTHDAY", time.Now().AddDate(-20, 0, 0)).
			Set("METADATA", "Oracle CLOB data")
		records = append(records, record)
	}
	dbkit.Use("oracle").BatchInsert("DEMO", records, 100)
	fmt.Println("Oracle: Data preparation complete.")
}

func demoRecordOperations() {
	fmt.Println("\n--- Oracle Record Operations ---")
	records, err := dbkit.Use("oracle").Query("SELECT * FROM DEMO WHERE AGE > :1 AND ROWNUM <= 5", 30)
	if err != nil {
		log.Printf("Oracle Record Query failed: %v", err)
		return
	}
	fmt.Printf("Query returned %d records\n", len(records))
}

func demoDbModelOperations() {
	fmt.Println("\n--- Oracle DbModel CRUD Operations ---")
	model := &models.Demo{}

	// 1. Insert
	newUser := &models.Demo{
		Name:      "New_Oracle_User",
		Age:       40,
		Salary:    12000.5,
		IsActive:  1,
		Birthday:  time.Now(),
		CreatedAt: time.Now(),
		Metadata:  "Oracle LOB Meta",
	}
	id, err := newUser.Insert()
	if err != nil {
		log.Printf("Oracle DbModel Insert failed: %v", err)
		return
	}
	fmt.Printf("Oracle DbModel Insert: ID = %d\n", id)
	newUser.ID = float64(id)

	// 2. FindFirst (Read)
	foundUser, err := model.FindFirst("NAME = ?", "New_Oracle_User")
	if err != nil {
		log.Printf("Oracle DbModel FindFirst failed: %v", err)
	} else if foundUser != nil {
		fmt.Printf("Oracle DbModel FindFirst: Found user %s, Age: %d\n", foundUser.Name, foundUser.Age)
	}

	// 3. Update
	foundUser.Age = 45
	foundUser.Salary = 15000.75
	affected, err := foundUser.Update()
	if err != nil {
		log.Printf("Oracle DbModel Update failed: %v", err)
	} else {
		fmt.Printf("Oracle DbModel Update: %d rows affected\n", affected)
	}

	// 4. Find (Read)
	results, err := model.Find("AGE >= ?", "ID DESC", 30)
	if err != nil {
		log.Printf("Oracle DbModel Find failed: %v", err)
	} else {
		fmt.Printf("Oracle DbModel Find: %d results, first user: %s\n", len(results), results[0].Name)
	}

	// 5. Paginate (Read)
	page, err := model.PaginateBuilder(1, 10, "AGE > ?", "ID ASC", 20)
	if err != nil {
		log.Printf("Oracle DbModel Paginate failed: %v", err)
	} else {
		fmt.Printf("Oracle DbModel Paginate: Total %d rows, current page size %d\n", page.TotalRow, len(page.List))
	}

	// 6. Delete
	affected, err = foundUser.Delete()
	if err != nil {
		log.Printf("Oracle DbModel Delete failed: %v", err)
	} else {
		fmt.Printf("Oracle DbModel Delete: %d rows affected\n", affected)
	}
}

func demoChainOperations() {
	page, err := dbkit.Use("oracle").Table("DEMO").Where("AGE > ?", 20).Paginate(1, 10)

	if err != nil {
		log.Printf("Oracle Chain Paginate failed: %v", err)
		return
	}
	fmt.Printf("Oracle Chain Paginate: Total %d rows\n", page.TotalRow)
}

func demoCacheOperations() {
	fmt.Println("\n--- Oracle Cache Operations ---")
	// 使用缓存查询 (新样式: Use().Cache().Table().Find())
	var results []models.Demo
	start := time.Now()
	err := dbkit.Use("oracle").Cache("oracle_demo_cache", 60).Table("DEMO").Where("AGE > ?", 30).FindToDbModel(&results)
	if err != nil {
		log.Printf("Oracle Cache Find (1st) failed: %v", err)
	} else {
		fmt.Printf("Oracle Cache Find (1st): %d results, took %v\n", len(results), time.Since(start))
	}

	// Second call - should hit cache
	start = time.Now()
	err = dbkit.Use("oracle").Cache("oracle_demo_cache", 60).Table("DEMO").Where("AGE > ?", 30).FindToDbModel(&results)
	if err != nil {
		log.Printf("Oracle Cache Find (2nd) failed: %v", err)
	} else {
		fmt.Printf("Oracle Cache Find (2nd): %d results, took %v (from cache)\n", len(results), time.Since(start))
	}

	// Test Paginate cache
	fmt.Println("\n--- Oracle Paginate Cache Operations ---")
	start = time.Now()
	page, err := dbkit.Use("oracle").Cache("oracle_page_cache", 60).Table("DEMO").Where("AGE > ?", 30).Paginate(1, 10)
	if err != nil {
		log.Printf("Oracle Paginate Cache (1st) failed: %v", err)
	} else {
		fmt.Printf("Oracle Paginate Cache (1st): %d results, took %v\n", len(page.List), time.Since(start))
	}

	start = time.Now()
	page, err = dbkit.Use("oracle").Cache("oracle_page_cache", 60).Table("DEMO").Where("AGE > ?", 30).Paginate(1, 10)
	if err != nil {
		log.Printf("Oracle Paginate Cache (2nd) failed: %v", err)
	} else {
		fmt.Printf("Oracle Paginate Cache (2nd): %d results, took %v (from cache)\n", len(page.List), time.Since(start))
	}

	// Test Count cache
	fmt.Println("\n--- Oracle Count Cache Operations ---")
	start = time.Now()
	count, err := dbkit.Use("oracle").Cache("oracle_count_cache", 60).Table("DEMO").Where("AGE > ?", 30).Count()
	if err != nil {
		log.Printf("Oracle Count Cache (1st) failed: %v", err)
	} else {
		fmt.Printf("Oracle Count Cache (1st): %d, took %v\n", count, time.Since(start))
	}

	start = time.Now()
	count, err = dbkit.Use("oracle").Cache("oracle_count_cache", 60).Table("DEMO").Where("AGE > ?", 30).Count()
	if err != nil {
		log.Printf("Oracle Count Cache (2nd) failed: %v", err)
	} else {
		fmt.Printf("Oracle Count Cache (2nd): %d, took %v (from cache)\n", count, time.Since(start))
	}
}

func demoUpdateDeleteOperations() {
	fmt.Println("\n--- Oracle Update/Delete Operations ---")
	// 更新一条记录
	affected, err := dbkit.Use("oracle").Table("DEMO").Where("ID = ?", 1).Update(dbkit.NewRecord().Set("NAME", "Updated_Name"))
	if err != nil {
		log.Printf("Oracle Update failed: %v", err)
	} else {
		fmt.Printf("Oracle Update affected %d rows\n", affected)
	}

	// 删除一条记录
	affected, err = dbkit.Use("oracle").Table("DEMO").Where("ID = ?", 110).Delete()
	if err != nil {
		log.Printf("Oracle Delete failed: %v", err)
	} else {
		fmt.Printf("Oracle Delete affected %d rows\n", affected)
	}
}
