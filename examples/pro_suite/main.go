package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/zzguang83325/dbkit"
	"github.com/zzguang83325/dbkit/examples/pro_suite/models"
)

func main() {
	fmt.Println("================================================================")
	fmt.Println("   DBKit Expert Suite - Advanced Multi-DB & Real-world Scenarios")
	fmt.Println("================================================================")

	// 1. Setup Multiple Databases
	setupDatabases()
	defer dbkit.Close()

	// 2. Migrate Schema
	migrateSchema()

	// 3. Initialize SQL Templates
	initializeSqlTemplates()

	// 4. Seed Initial Data
	seedData()

	// 5. Run Advanced Scenarios
	testAdvancedQueryBuilder()
	testRealWorldTransactions()
	testOptimisticLockStress()
	testCrossDatabaseMigration()
	testSqlTemplatePower()

	fmt.Println("\n================================================================")
	fmt.Println("   Expert Suite Completed Successfully!")
	fmt.Println("================================================================")
}

func setupDatabases() {
	fmt.Println("\n[Phase 1] Setting up multiple database connections...")

	// Primary MySQL Database
	mysqlDSN := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	err := dbkit.OpenDatabaseWithDBName("default", dbkit.MySQL, mysqlDSN, 20)
	if err != nil {
		log.Fatalf("MySQL connection failed: %v", err)
	}

	// Archival/Secondary SQLite Database
	sqlitePath := "./archival.db"
	err = dbkit.OpenDatabaseWithDBName("archival", dbkit.SQLite3, sqlitePath, 5)
	if err != nil {
		log.Fatalf("SQLite connection failed: %v", err)
	}

	dbkit.SetDebugMode(true)
	fmt.Println("✓ Primary (MySQL) and Archival (SQLite) databases connected.")
}

func migrateSchema() {
	fmt.Println("\n[Phase 2] Migrating advanced schemas...")

	// Create MySQL Tables
	mysqlTables := []string{
		"DROP TABLE IF EXISTS pro_article_categories",
		"DROP TABLE IF EXISTS pro_comments",
		"DROP TABLE IF EXISTS pro_categories",
		"DROP TABLE IF EXISTS pro_articles",
		"DROP TABLE IF EXISTS pro_users",
		`CREATE TABLE pro_users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			email VARCHAR(100) NOT NULL,
			role VARCHAR(20) DEFAULT 'user',
			settings TEXT,
			credits DECIMAL(10,2) DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE pro_articles (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			author_id BIGINT NOT NULL,
			title VARCHAR(255) NOT NULL,
			content LONGTEXT,
			status VARCHAR(20) DEFAULT 'draft',
			version BIGINT DEFAULT 1,
			deleted_at DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE pro_categories (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50) NOT NULL UNIQUE
		)`,
		`CREATE TABLE pro_article_categories (
			article_id BIGINT NOT NULL,
			category_id BIGINT NOT NULL,
			PRIMARY KEY (article_id, category_id)
		)`,
		`CREATE TABLE pro_comments (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			article_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			parent_id BIGINT DEFAULT 0,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, sql := range mysqlTables {
		_, err := dbkit.Use("default").Exec(sql)
		if err != nil {
			log.Fatalf("MySQL Migration failed: %v", err)
		}
	}

	// Create SQLite Table for Archival
	sqliteSQL := `CREATE TABLE IF NOT EXISTS archived_articles (
		id INTEGER PRIMARY KEY,
		original_id BIGINT,
		title TEXT,
		author_name TEXT,
		archived_at DATETIME
	)`
	_, err := dbkit.Use("archival").Exec(sqliteSQL)
	if err != nil {
		log.Fatalf("SQLite Migration failed: %v", err)
	}

	// Configure DBKit Features
	dbkit.EnableTimestamps()
	dbkit.ConfigTimestamps("pro_users")
	dbkit.ConfigTimestamps("pro_articles")
	
	dbkit.EnableSoftDelete()
	dbkit.ConfigSoftDelete("pro_articles", "deleted_at")
	
	dbkit.EnableOptimisticLock()
	dbkit.ConfigOptimisticLock("pro_articles")

	fmt.Println("✓ Schema migration and feature configuration completed.")
}

func initializeSqlTemplates() {
	dbkit.LoadSqlConfig("./config/user_service.json")
	dbkit.LoadSqlConfig("./config/article_service.json")
}

func seedData() {
	fmt.Println("\n[Phase 3] Seeding initial data...")

	// 1. Create Users
	users := []*models.User{
		{Username: "admin", Email: "admin@pro.com", Role: "admin", Credits: 1000, Settings: `{"theme":"dark"}`},
		{Username: "writer1", Email: "writer1@pro.com", Role: "user", Credits: 100},
		{Username: "reader1", Email: "reader1@pro.com", Role: "user", Credits: 50},
	}
	for _, u := range users {
		u.Insert()
	}

	// 2. Create Categories
	cats := []string{"Tech", "Lifestyle", "Finance"}
	for _, c := range cats {
		(&models.Category{Name: c}).Insert()
	}

	// 3. Create Articles
	article := &models.Article{
		AuthorID: 2,
		Title:    "Mastering DBKit in 2026",
		Content:  "Comprehensive guide to ActiveRecord in Go...",
		Status:   "published",
	}
	aid, _ := article.Insert()

	// 4. Link Categories
	(&models.ArticleCategory{ArticleID: aid, CategoryID: 1}).Insert()
	(&models.ArticleCategory{ArticleID: aid, CategoryID: 3}).Insert()

	// 5. Add Comments
	comments := []*models.Comment{
		{ArticleID: aid, UserID: 3, Content: "Great article!"},
		{ArticleID: aid, UserID: 1, Content: "Very helpful, thanks."},
	}
	for _, c := range comments {
		c.Insert()
	}

	fmt.Println("✓ Initial data seeded.")
}

func testAdvancedQueryBuilder() {
	fmt.Println("\n[Phase 4] Testing Advanced QueryBuilder...")

	// Scenario: Find articles with complex filtering and subqueries
	// We want articles that:
	// - Are published
	// - Belong to 'Tech' or 'Finance' categories
	// - Have more than 1 comment
	
	subqueryCats := dbkit.NewSubquery().
		Table("pro_categories").
		Select("id").
		Where("name IN (?, ?)", "Tech", "Finance")

	subqueryArticleIDs := dbkit.NewSubquery().
		Table("pro_article_categories").
		Select("article_id").
		Where("category_id IN (SELECT id FROM pro_categories WHERE name IN (?, ?))", "Tech", "Finance")

	results, err := dbkit.Table("pro_articles").
		Select("pro_articles.*, pro_users.username as author_name").
		InnerJoin("pro_users", "pro_articles.author_id = pro_users.id").
		WhereIn("pro_articles.id", subqueryArticleIDs).
		WhereGroup(func(qb *dbkit.QueryBuilder) *dbkit.QueryBuilder {
			return qb.Where("status = ?", "published").OrWhere("credits > ?", 500)
		}).
		OrderBy("created_at DESC").
		Find()

	if err != nil {
		log.Printf("Advanced Query failed: %v", err)
	} else {
		fmt.Printf("✓ Found %d articles matching complex criteria.\n", len(results))
		for _, r := range results {
			fmt.Printf("  - [%s] by %s\n", r.Str("title"), r.Str("author_name"))
		}
	}
	_ = subqueryCats // Avoid unused warning
}

func testRealWorldTransactions() {
	fmt.Println("\n[Phase 5] Testing Real-world Business Transaction (Buy Premium Article)...")

	// Logic: Reader1 wants to buy a premium article from Writer1
	// 1. Check Reader balance
	// 2. Deduct balance from Reader
	// 3. Add balance to Writer
	// 4. Record the transaction (simulated)
	
	err := dbkit.Transaction(func(tx *dbkit.Tx) error {
		reader, _ := tx.Table("pro_users").Where("username = ?", "reader1").FindFirst()
		writer, _ := tx.Table("pro_users").Where("username = ?", "writer1").FindFirst()
		
		price := 20.0
		if reader.Float("credits") < price {
			return fmt.Errorf("insufficient balance")
		}

		// Deduct from reader
		reader.Set("credits", reader.Float("credits")-price)
		tx.Update("pro_users", reader, "id = ?", reader.Int64("id"))

		// Add to writer
		writer.Set("credits", writer.Float("credits")+price)
		tx.Update("pro_users", writer, "id = ?", writer.Int64("id"))

		fmt.Println("  [TX] Balances updated: Reader(-20), Writer(+20)")
		return nil
	})

	if err != nil {
		fmt.Printf("Transaction failed: %v\n", err)
	} else {
		fmt.Println("✓ Transaction committed successfully.")
	}
}

func testOptimisticLockStress() {
	fmt.Println("\n[Phase 6] Testing Optimistic Locking under Concurrency...")

	article, _ := (&models.Article{}).FindFirst("id = ?", 1)
	fmt.Printf("  Initial Version: %d, Title: %s\n", article.Version, article.Title)

	var wg sync.WaitGroup
	workers := 5
	wg.Add(workers)

	results := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			// Simulate concurrent update
			// We need to fetch the LATEST version inside the loop or try-catch
			localArticle, _ := (&models.Article{}).FindFirst("id = ?", 1)
			localArticle.Title = fmt.Sprintf("Updated by Worker %d", id)
			
			_, err := localArticle.Update()
			results <- err
		}(i)
	}

	wg.Wait()
	close(results)

	success := 0
	conflicts := 0
	for err := range results {
		if err == nil {
			success++
		} else if err == dbkit.ErrVersionMismatch {
			conflicts++
		}
	}

	final, _ := (&models.Article{}).FindFirst("id = ?", 1)
	fmt.Printf("✓ Stress Test Results: %d Success, %d Conflicts (Expected). Final Version: %d\n", success, conflicts, final.Version)
}

func testCrossDatabaseMigration() {
	fmt.Println("\n[Phase 7] Testing Cross-Database Data Archival (MySQL -> SQLite)...")

	// Fetch data from MySQL
	articles, _ := dbkit.Use("default").Table("pro_articles").
		Select("pro_articles.id, pro_articles.title, pro_users.username").
		InnerJoin("pro_users", "pro_articles.author_id = pro_users.id").
		Find()

	fmt.Printf("  Archiving %d articles from MySQL to SQLite...\n", len(articles))

	// Batch Insert into SQLite
	var archivalRecords []*dbkit.Record
	for _, a := range articles {
		rec := dbkit.NewRecord().
			Set("original_id", a.Int64("id")).
			Set("title", a.Str("title")).
			Set("author_name", a.Str("username")).
			Set("archived_at", time.Now())
		archivalRecords = append(archivalRecords, rec)
	}

	affected, err := dbkit.Use("archival").BatchInsertDefault("archived_articles", archivalRecords)
	if err != nil {
		log.Printf("Archival failed: %v", err)
	} else {
		fmt.Printf("✓ Successfully archived %d records into SQLite archival database.\n", affected)
		
		// Verify
		count, _ := dbkit.Use("archival").Count("archived_articles", "1=1")
		fmt.Printf("  Verification: SQLite now has %d archived records.\n", count)
	}
}

func testSqlTemplatePower() {
	fmt.Println("\n[Phase 8] Testing SQL Template Power & Custom Transformation...")

	// 1. Get Profile with Article Count (Subquery in Template)
	profile, err := dbkit.SqlTemplate("user_service.getProfile", 2).QueryFirst()
	if err == nil && profile != nil {
		fmt.Printf("✓ Profile for writer1: %s, Article Count: %d\n", 
			profile.Str("username"), profile.Int("article_count"))
	}

	// 2. Get Popular Articles with Comment Counts
	popular, err := dbkit.SqlTemplate("article_service.getPopularArticles", 1).Query()
	if err == nil {
		fmt.Printf("✓ Popular Articles (>= 1 comment):\n")
		for _, p := range popular {
			fmt.Printf("  - %s (%d comments)\n", p.Str("title"), p.Int("comment_count"))
		}
	}
}
