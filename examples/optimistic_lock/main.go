package main

import (
	"errors"
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

	// Create test table with version field
	_, err = dbkit.Exec(`
		CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price REAL,
			stock INTEGER,
			version INTEGER DEFAULT 1
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Configure optimistic lock for products table
	dbkit.ConfigOptimisticLock("products")

	fmt.Println("=== Optimistic Lock Demo ===")
	fmt.Println()

	// 1. Insert - version will be auto-initialized to 1
	fmt.Println("1. Inserting a new product (version will be auto-initialized to 1)...")
	record := dbkit.NewRecord()
	record.Set("name", "Laptop")
	record.Set("price", 999.99)
	record.Set("stock", 100)
	id, _ := dbkit.Insert("products", record)
	fmt.Printf("   Inserted product with ID: %d\n", id)
	printProduct(id)

	// 2. Normal update with correct version
	fmt.Println("\n2. Updating product with correct version...")
	updateRecord := dbkit.NewRecord()
	updateRecord.Set("version", int64(1)) // Current version
	updateRecord.Set("price", 899.99)
	updateRecord.Set("stock", 95)
	rows, err := dbkit.Update("products", updateRecord, "id = ?", id)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Updated %d row(s)\n", rows)
	}
	printProduct(id)

	// 3. Simulate concurrent update - use stale version
	fmt.Println("\n3. Simulating concurrent update with stale version (version=1, but current is 2)...")
	staleRecord := dbkit.NewRecord()
	staleRecord.Set("version", int64(1)) // Stale version!
	staleRecord.Set("price", 799.99)
	rows, err = dbkit.Update("products", staleRecord, "id = ?", id)
	if errors.Is(err, dbkit.ErrVersionMismatch) {
		fmt.Println("   ✓ Detected version mismatch! Concurrent modification prevented.")
		fmt.Printf("   Error: %v\n", err)
	} else if err != nil {
		fmt.Printf("   Unexpected error: %v\n", err)
	}
	printProduct(id) // Price should still be 899.99

	// 4. Correct way to handle concurrent update - read latest version first
	fmt.Println("\n4. Correct way: Read latest version, then update...")
	latestRecord, _ := dbkit.Table("products").Where("id = ?", id).FindFirst()
	if latestRecord != nil {
		currentVersion := latestRecord.GetInt("version")
		fmt.Printf("   Current version: %d\n", currentVersion)

		updateRecord2 := dbkit.NewRecord()
		updateRecord2.Set("version", currentVersion)
		updateRecord2.Set("price", 799.99)
		rows, err = dbkit.Update("products", updateRecord2, "id = ?", id)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
		} else {
			fmt.Printf("   Updated %d row(s)\n", rows)
		}
	}
	printProduct(id)

	// 5. Update without version field - no version check
	fmt.Println("\n5. Update without version field (no version check)...")
	noVersionRecord := dbkit.NewRecord()
	noVersionRecord.Set("stock", 90)
	rows, err = dbkit.Update("products", noVersionRecord, "id = ?", id)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Updated %d row(s) (no version check)\n", rows)
	}
	printProduct(id) // Note: version unchanged because we didn't include it

	// 6. Custom version field name
	fmt.Println("\n6. Demo with custom version field name...")
	_, err = dbkit.Exec(`
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer TEXT NOT NULL,
			total REAL,
			revision INTEGER DEFAULT 1
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Configure with custom field name
	dbkit.ConfigOptimisticLockWithField("orders", "revision")

	orderRecord := dbkit.NewRecord()
	orderRecord.Set("customer", "John Doe")
	orderRecord.Set("total", 150.00)
	orderId, _ := dbkit.Insert("orders", orderRecord)
	fmt.Printf("   Inserted order with ID: %d\n", orderId)
	printOrder(orderId)

	// Update order
	orderUpdate := dbkit.NewRecord()
	orderUpdate.Set("revision", int64(1))
	orderUpdate.Set("total", 175.00)
	rows, _ = dbkit.Update("orders", orderUpdate, "id = ?", orderId)
	fmt.Printf("   Updated %d row(s)\n", rows)
	printOrder(orderId)

	// 7. Using UpdateRecord (auto extracts PK from record)
	fmt.Println("\n7. Using UpdateRecord with optimistic lock...")
	product, _ := dbkit.Table("products").Where("id = ?", id).FindFirst()
	if product != nil {
		product.Set("name", "Gaming Laptop")
		// version is already in the record from query
		rows, err = dbkit.Use("default").UpdateRecord("products", product)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
		} else {
			fmt.Printf("   Updated %d row(s)\n", rows)
		}
	}
	printProduct(id)

	// 8. Transaction with optimistic lock
	fmt.Println("\n8. Transaction with optimistic lock...")
	err = dbkit.Transaction(func(tx *dbkit.Tx) error {
		// Read current product
		rec, err := tx.Table("products").Where("id = ?", id).FindFirst()
		if err != nil {
			return err
		}

		currentVersion := rec.GetInt("version")
		fmt.Printf("   In transaction: current version = %d\n", currentVersion)

		// Update with version check
		updateRec := dbkit.NewRecord()
		updateRec.Set("version", currentVersion)
		updateRec.Set("stock", 80)
		_, err = tx.Update("products", updateRec, "id = ?", id)
		return err
	})
	if err != nil {
		fmt.Printf("   Transaction error: %v\n", err)
	} else {
		fmt.Println("   Transaction committed successfully")
	}
	printProduct(id)

	fmt.Println("\n=== Demo Complete ===")
}

func printProduct(id int64) {
	record, _ := dbkit.Table("products").Where("id = ?", id).FindFirst()
	if record != nil {
		fmt.Printf("   Product: name=%s, price=%.2f, stock=%d, version=%d\n",
			record.GetString("name"),
			record.GetFloat("price"),
			record.GetInt("stock"),
			record.GetInt("version"))
	}
}

func printOrder(id int64) {
	record, _ := dbkit.Table("orders").Where("id = ?", id).FindFirst()
	if record != nil {
		fmt.Printf("   Order: customer=%s, total=%.2f, revision=%d\n",
			record.GetString("customer"),
			record.GetFloat("total"),
			record.GetInt("revision"))
	}
}
