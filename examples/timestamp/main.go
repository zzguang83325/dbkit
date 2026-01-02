package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
)

func main() {
	// Initialize SQLite database
	err := dbkit.OpenDatabase(dbkit.SQLite3, ":memory:", 10)
	if err != nil {
		log.Fatal(err)
	}

	// Create test table with timestamp fields
	_, err = dbkit.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Enable timestamp first
	dbkit.EnableTimestamps()

	// Configure auto timestamps for users table
	dbkit.ConfigTimestamps("users")

	fmt.Println("=== Auto Timestamps Demo ===")
	fmt.Println()

	// 1. Insert - created_at will be auto-filled
	fmt.Println("1. Inserting a new user (created_at will be auto-filled)...")
	record := dbkit.NewRecord()
	record.Set("name", "John Doe")
	record.Set("email", "john@example.com")
	id, _ := dbkit.Insert("users", record)
	fmt.Printf("   Inserted user with ID: %d\n", id)
	printUser(id)

	// Wait a bit to see timestamp difference
	time.Sleep(time.Second)

	// 2. Update - updated_at will be auto-filled
	fmt.Println("\n2. Updating user (updated_at will be auto-filled)...")
	updateRecord := dbkit.NewRecord()
	updateRecord.Set("name", "John Updated")
	dbkit.Update("users", updateRecord, "id = ?", id)
	printUser(id)

	// 3. Insert with custom timestamp (won't be overwritten)
	fmt.Println("\n3. Inserting with custom created_at (won't be overwritten)...")
	customTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	record2 := dbkit.NewRecord()
	record2.Set("name", "Jane Doe")
	record2.Set("email", "jane@example.com")
	record2.Set("created_at", customTime)
	id2, _ := dbkit.Insert("users", record2)
	fmt.Printf("   Inserted user with ID: %d\n", id2)
	printUser(id2)

	// 4. Update using QueryBuilder with WithoutTimestamps
	fmt.Println("\n4. Updating with WithoutTimestamps() (updated_at won't change)...")
	time.Sleep(time.Second)
	updateRecord2 := dbkit.NewRecord()
	updateRecord2.Set("email", "jane.new@example.com")
	dbkit.Table("users").Where("id = ?", id2).WithoutTimestamps().Update(updateRecord2)
	printUser(id2)

	// 5. Custom field names
	fmt.Println("\n5. Demo with custom field names...")
	_, err = dbkit.Exec(`
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product TEXT NOT NULL,
			create_time DATETIME,
			update_time DATETIME
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Configure with custom field names
	dbkit.ConfigTimestampsWithFields("orders", "create_time", "update_time")

	orderRecord := dbkit.NewRecord()
	orderRecord.Set("product", "Laptop")
	orderId, _ := dbkit.Insert("orders", orderRecord)
	fmt.Printf("   Inserted order with ID: %d\n", orderId)
	printOrder(orderId)

	// 6. Only created_at field
	fmt.Println("\n6. Demo with only created_at field...")
	_, err = dbkit.Exec(`
		CREATE TABLE logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message TEXT NOT NULL,
			log_time DATETIME
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Configure only created_at
	dbkit.ConfigCreatedAt("logs", "log_time")

	logRecord := dbkit.NewRecord()
	logRecord.Set("message", "System started")
	logId, _ := dbkit.Insert("logs", logRecord)
	fmt.Printf("   Inserted log with ID: %d\n", logId)
	printLog(logId)

	fmt.Println("\n=== Demo Complete ===")
}

func printUser(id int64) {
	record, _ := dbkit.Table("users").Where("id = ?", id).FindFirst()
	if record != nil {
		fmt.Printf("   User: name=%s, email=%s, created_at=%v, updated_at=%v\n",
			record.GetString("name"),
			record.GetString("email"),
			record.Get("created_at"),
			record.Get("updated_at"))
	}
}

func printOrder(id int64) {
	record, _ := dbkit.Table("orders").Where("id = ?", id).FindFirst()
	if record != nil {
		fmt.Printf("   Order: product=%s, create_time=%v, update_time=%v\n",
			record.GetString("product"),
			record.Get("create_time"),
			record.Get("update_time"))
	}
}

func printLog(id int64) {
	record, _ := dbkit.Table("logs").Where("id = ?", id).FindFirst()
	if record != nil {
		fmt.Printf("   Log: message=%s, log_time=%v\n",
			record.GetString("message"),
			record.Get("log_time"))
	}
}
