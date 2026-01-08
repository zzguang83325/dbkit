package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
)

// Expert Example: Library Management System
// Demonstrates advanced DBKit features in a real-world scenario.

func main() {
	fmt.Println("======================================================")
	fmt.Println("   DBKit Expert Comprehensive Example: Library System")
	fmt.Println("======================================================")

	// 1. Setup Database
	dbPath := "./expert_library.db"
	os.Remove(dbPath) // Start fresh
	
	err := dbkit.OpenDatabase(dbkit.SQLite3, dbPath, 10)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbkit.Close()

	// Enable features
	dbkit.SetDebugMode(true)
	dbkit.EnableSoftDelete()
	dbkit.EnableOptimisticLock()
	dbkit.EnableTimestamps()
	dbkit.InitLocalCache(1 * time.Minute)

	// 2. Initialize Schema
	initSchema()

	// 3. Populate Data
	populateInitialData()

	// 4. Advanced Tests
	testComplexJoins()
	testAdvancedSubqueries()
	testOptimisticLockingAndTransactions()
	testSoftDeleteAndRecovery()
	testAdvancedCaching()
	testPaginationAndFiltering()

	fmt.Println("\n======================================================")
	fmt.Println("   Expert Example Completed Successfully!")
	fmt.Println("======================================================")
}

func initSchema() {
	fmt.Println("\n[1/6] Initializing Schema...")

	// expert_authors
	dbkit.Exec(`CREATE TABLE expert_authors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		bio TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	// expert_categories
	dbkit.Exec(`CREATE TABLE expert_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	)`)

	// expert_books
	dbkit.Exec(`CREATE TABLE expert_books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		author_id INTEGER,
		category_id INTEGER,
		title TEXT NOT NULL,
		price DECIMAL(10,2),
		stock INTEGER DEFAULT 0,
		version INTEGER DEFAULT 1,
		deleted_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	// expert_loans
	dbkit.Exec(`CREATE TABLE expert_loans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		book_id INTEGER,
		user_name TEXT,
		loan_date DATETIME,
		return_date DATETIME,
		status TEXT DEFAULT 'ACTIVE'
	)`)

	// Configure DBKit Features for these tables
	dbkit.ConfigTimestamps("expert_authors")
	dbkit.ConfigTimestamps("expert_books")
	dbkit.ConfigSoftDelete("expert_books", "deleted_at")
	dbkit.ConfigOptimisticLock("expert_books")
}

func populateInitialData() {
	fmt.Println("\n[2/6] Populating Initial Data...")

	// Authors
	authorID1, _ := dbkit.Insert("expert_authors", dbkit.NewRecord().Set("name", "J.K. Rowling").Set("bio", "Author of Harry Potter"))
	authorID2, _ := dbkit.Insert("expert_authors", dbkit.NewRecord().Set("name", "George R.R. Martin").Set("bio", "Author of ASOIAF"))

	// Categories
	catID1, _ := dbkit.Insert("expert_categories", dbkit.NewRecord().Set("name", "Fantasy"))
	catID2, _ := dbkit.Insert("expert_categories", dbkit.NewRecord().Set("name", "Fiction"))
	_ = catID2 // Avoid unused variable error

	// Books
	dbkit.Insert("expert_books", dbkit.NewRecord().
		Set("author_id", authorID1).
		Set("category_id", catID1).
		Set("title", "Harry Potter and the Sorcerer's Stone").
		Set("price", 19.99).
		Set("stock", 10))

	dbkit.Insert("expert_books", dbkit.NewRecord().
		Set("author_id", authorID1).
		Set("category_id", catID1).
		Set("title", "Harry Potter and the Chamber of Secrets").
		Set("price", 21.99).
		Set("stock", 5))

	dbkit.Insert("expert_books", dbkit.NewRecord().
		Set("author_id", authorID2).
		Set("category_id", catID1).
		Set("title", "A Game of Thrones").
		Set("price", 25.00).
		Set("stock", 3))
}

func testComplexJoins() {
	fmt.Println("\n[3/6] Testing Complex Joins...")

	// Join Books with Authors and Categories
	records, err := dbkit.Table("expert_books").
		Select("expert_books.title, expert_authors.name as author_name, expert_categories.name as category_name").
		InnerJoin("expert_authors", "expert_books.author_id = expert_authors.id").
		LeftJoin("expert_categories", "expert_books.category_id = expert_categories.id").
		OrderBy("expert_books.title ASC").
		Find()

	if err != nil {
		log.Printf("Join failed: %v", err)
	} else {
		for _, r := range records {
			fmt.Printf("  - Book: %s | Author: %s | Category: %s\n", 
				r.Str("title"), r.Str("author_name"), r.Str("category_name"))
		}
	}
}

func testAdvancedSubqueries() {
	fmt.Println("\n[4/6] Testing Advanced Subqueries...")

	// Subquery 1: Books whose price is above average
	avgPriceSub := dbkit.NewSubquery().
		Table("expert_books").
		Select("AVG(price)")
	
	subSQL, subArgs := avgPriceSub.ToSQL()
	expensiveBooks, _ := dbkit.Table("expert_books").
		Where("price > ("+subSQL+")", subArgs...). 
		Find()
	
	fmt.Printf("  - Books above average price: %d\n", len(expensiveBooks))

	// Subquery 2: Author name in a column (Select Subquery)
	authorNameSub := dbkit.NewSubquery().
		Table("expert_authors").
		Select("name").
		Where("id = expert_books.author_id")

	booksWithAuthorColumn, _ := dbkit.Table("expert_books").
		Select("title").
		SelectSubquery(authorNameSub, "author_name").
		Limit(2).
		Find()

	for _, b := range booksWithAuthorColumn {
		fmt.Printf("  - %s (by %s)\n", b.Str("title"), b.Str("author_name"))
	}
}

func testOptimisticLockingAndTransactions() {
	fmt.Println("\n[5/6] Testing Optimistic Locking & Transactions...")

	// Scenario: Loan a book. Must check stock, update stock (dec), and create loan record.
	book, _ := dbkit.Table("expert_books").Where("title LIKE ?", "%Game of Thrones%").FindFirst()
	if book == nil {
		log.Println("Book not found")
		return
	}

	currentStock := book.Int("stock")
	currentVersion := book.Int64("version")
	bookID := book.Get("id")

	fmt.Printf("  - Initial Stock: %d, Version: %d\n", currentStock, currentVersion)

	// Transactional Loan
	err := dbkit.Transaction(func(tx *dbkit.Tx) error {
		// 1. Update stock with version check
		updateRec := dbkit.NewRecord().
			Set("stock", currentStock-1).
			Set("version", currentVersion) // Trigger Optimistic Lock
		
		affected, err := tx.Update("expert_books", updateRec, "id = ?", bookID)
		if err != nil {
			return err
		}
		if affected == 0 {
			return dbkit.ErrVersionMismatch
		}

		// 2. Create Loan Record
		loanRec := dbkit.NewRecord().
			Set("book_id", bookID).
			Set("user_name", "Expert Tester").
			Set("loan_date", time.Now()).
			Set("status", "ACTIVE")
		_, err = tx.Insert("expert_loans", loanRec)
		return err
	})

	if err != nil {
		fmt.Printf("  - Transaction Failed: %v\n", err)
	} else {
		fmt.Println("  - Loan successful! Stock decremented and Version incremented.")
	}

	// Verify update
	updatedBook, _ := dbkit.Table("expert_books").Where("id = ?", bookID).FindFirst()
	fmt.Printf("  - Updated Stock: %d, Version: %d\n", updatedBook.Int("stock"), updatedBook.Int64("version"))

	// Test Conflict
	fmt.Println("  - Testing Conflict (Simulating stale update)...")
	staleRec := dbkit.NewRecord().Set("stock", 0).Set("version", currentVersion) // Old version
	_, err = dbkit.Update("expert_books", staleRec, "id = ?", bookID)
	if err == dbkit.ErrVersionMismatch {
		fmt.Println("  - ✓ Successfully detected version mismatch conflict.")
	} else {
		fmt.Println("  - ✗ Version mismatch NOT detected!")
	}
}

func testSoftDeleteAndRecovery() {
	fmt.Println("\n[6/6] Testing Soft Delete & Recovery...")

	bookTitle := "Harry Potter and the Chamber of Secrets"
	
	// 1. Soft Delete
	affected, _ := dbkit.Table("expert_books").Where("title = ?", bookTitle).Delete()
	fmt.Printf("  - Soft Deleted '%s', affected: %d\n", bookTitle, affected)

	// 2. Verify normal query excludes it
	foundNormal, _ := dbkit.Table("expert_books").Where("title = ?", bookTitle).FindFirst()
	if foundNormal == nil {
		fmt.Println("  - ✓ Normal query correctly filtered out the soft-deleted book.")
	}

	// 3. Verify WithTrashed includes it
	foundWithTrashed, _ := dbkit.Table("expert_books").WithTrashed().Where("title = ?", bookTitle).FindFirst()
	if foundWithTrashed != nil {
		fmt.Println("  - ✓ WithTrashed query correctly found the deleted book.")
	}

	// 4. Restore
	affected, _ = dbkit.Table("expert_books").Where("title = ?", bookTitle).Restore()
	fmt.Printf("  - Restored book, affected: %d\n", affected)

	// 5. Verify it's back
	foundAgain, _ := dbkit.Table("expert_books").Where("title = ?", bookTitle).FindFirst()
	if foundAgain != nil {
		fmt.Println("  - ✓ Book is back in normal results.")
	}
}

func testAdvancedCaching() {
	fmt.Println("\n[Expert] Testing Advanced Caching...")

	// Query with 5s cache
	cacheRepo := "expert_cat_cache"
	start := time.Now()
	dbkit.Table("expert_categories").Cache(cacheRepo, 5*time.Second).Find()
	fmt.Printf("  - First query (DB): %v\n", time.Since(start))

	start = time.Now()
	dbkit.Table("expert_categories").Cache(cacheRepo, 5*time.Second).Find()
	fmt.Printf("  - Second query (Cache): %v (Expected to be near zero)\n", time.Since(start))
}

func testPaginationAndFiltering() {
	fmt.Println("\n[Expert] Testing Complex Pagination & Filtering...")

	// Complex Where Grouping: (Fantasy AND price > 20) OR (Stock < 5)
	page, err := dbkit.Table("expert_books").
		Select("title, price, stock").
		WhereGroup(func(qb *dbkit.QueryBuilder) *dbkit.QueryBuilder {
			return qb.Where("price > ?", 20)
		}).
		OrWhere("stock < ?", 5).
		Paginate(1, 10)

	if err != nil {
		log.Printf("Pagination failed: %v", err)
	} else {
		fmt.Printf("  - Found %d records on page %d\n", len(page.List), page.PageNumber)
		for _, b := range page.List {
			fmt.Printf("    - %s (Price: %.2f, Stock: %d)\n", b.Str("title"), b.Float("price"), b.Int("stock"))
		}
	}
}
