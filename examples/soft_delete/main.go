package main

import (
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
)

func main() {
	// Initialize SQLite database
	err := dbkit.OpenDatabase(dbkit.SQLite3, ":memory:", 10)
	if err != nil {
		log.Fatal(err)
	}

	// Create test table with soft delete field
	_, err = dbkit.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			deleted_at DATETIME
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Enable soft delete check first
	dbkit.EnableSoftDeleteCheck()

	// Configure soft delete for users table
	dbkit.ConfigSoftDelete("users", "deleted_at")

	fmt.Println("=== Soft Delete Demo ===\n")

	// Insert test data
	fmt.Println("1. Inserting test users...")
	for i := 1; i <= 5; i++ {
		record := dbkit.NewRecord()
		record.Set("name", fmt.Sprintf("User%d", i))
		record.Set("email", fmt.Sprintf("user%d@example.com", i))
		dbkit.Insert("users", record)
	}
	printUsers("After insert")

	// Soft delete user with id=2
	fmt.Println("\n2. Soft deleting user with id=2...")
	dbkit.Delete("users", "id = ?", 2)
	printUsers("After soft delete (normal query)")

	// Query with trashed (include deleted)
	fmt.Println("\n3. Query with WithTrashed() - includes deleted records...")
	records, _ := dbkit.Table("users").WithTrashed().Find()
	fmt.Printf("   Found %d users (including deleted)\n", len(records))
	for _, r := range records {
		deletedAt := r.Get("deleted_at")
		status := "active"
		if deletedAt != nil {
			status = "deleted"
		}
		fmt.Printf("   - ID: %d, Name: %s, Status: %s\n", r.GetInt("id"), r.GetString("name"), status)
	}

	// Query only trashed
	fmt.Println("\n4. Query with OnlyTrashed() - only deleted records...")
	records, _ = dbkit.Table("users").OnlyTrashed().Find()
	fmt.Printf("   Found %d deleted users\n", len(records))
	for _, r := range records {
		fmt.Printf("   - ID: %d, Name: %s\n", r.GetInt("id"), r.GetString("name"))
	}

	// Restore deleted user
	fmt.Println("\n5. Restoring user with id=2...")
	dbkit.Restore("users", "id = ?", 2)
	printUsers("After restore")

	// Soft delete again and then force delete
	fmt.Println("\n6. Soft deleting user with id=3, then force deleting...")
	dbkit.Delete("users", "id = ?", 3)
	fmt.Println("   After soft delete:")
	records, _ = dbkit.Table("users").WithTrashed().Find()
	fmt.Printf("   Total users (with trashed): %d\n", len(records))

	dbkit.ForceDelete("users", "id = ?", 3)
	fmt.Println("   After force delete:")
	records, _ = dbkit.Table("users").WithTrashed().Find()
	fmt.Printf("   Total users (with trashed): %d\n", len(records))

	// Using QueryBuilder chain
	fmt.Println("\n7. Using QueryBuilder chain methods...")
	dbkit.Table("users").Where("id = ?", 4).Delete()
	fmt.Println("   Soft deleted user 4 via QueryBuilder")

	count, _ := dbkit.Table("users").Count()
	fmt.Printf("   Active users count: %d\n", count)

	count, _ = dbkit.Table("users").WithTrashed().Count()
	fmt.Printf("   All users count (with trashed): %d\n", count)

	// Restore via QueryBuilder
	dbkit.Table("users").Where("id = ?", 4).Restore()
	fmt.Println("   Restored user 4 via QueryBuilder")
	printUsers("Final state")

	fmt.Println("\n=== Demo Complete ===")
}

func printUsers(label string) {
	records, _ := dbkit.Table("users").Find()
	fmt.Printf("   %s: Found %d active users\n", label, len(records))
	for _, r := range records {
		fmt.Printf("   - ID: %d, Name: %s, Email: %s\n",
			r.GetInt("id"), r.GetString("name"), r.GetString("email"))
	}
}
