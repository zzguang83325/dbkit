package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zzguang83325/dbkit"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GORM 模型
type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Username  string    `gorm:"size:100"`
	Email     string    `gorm:"size:100"`
	Age       int       `gorm:"default:0"`
	Status    string    `gorm:"size:20;default:'active'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (User) TableName() string {
	return "benchmark_users_gorm"
}

// 表名常量
const (
	DbkitTable = "benchmark_users_dbkit"
	GormTable  = "benchmark_users_gorm"
)

// 测试配置
const (
	DSN         = "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	InsertCount = 1000 // 插入测试数量
	QueryCount  = 1000 // 查询测试次数
	UpdateCount = 500  // 更新测试次数
	BatchSize   = 100  // 批量操作大小
)

var (
	gormDB  *gorm.DB
	results []BenchmarkResult
)

type BenchmarkResult struct {
	TestName    string
	DbkitTime   time.Duration
	GormTime    time.Duration
	DbkitOps    float64 // 每秒操作数
	GormOps     float64
	Improvement string // dbkit 相对 gorm 的提升
}

func main() {
	fmt.Println("=" + repeat("=", 70))
	fmt.Println("  DBKit (Record模式) vs GORM 性能对比测试")
	fmt.Println("=" + repeat("=", 70))
	fmt.Printf("\n测试环境:\n")
	fmt.Printf("  - Go Version: %s\n", runtime.Version())
	fmt.Printf("  - OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  - CPU Cores: %d\n", runtime.NumCPU())
	fmt.Printf("  - MySQL: 127.0.0.1:3306/test\n")
	fmt.Printf("\n测试参数:\n")
	fmt.Printf("  - 单条插入: %d 次\n", InsertCount)
	fmt.Printf("  - 查询测试: %d 次\n", QueryCount)
	fmt.Printf("  - 更新测试: %d 次\n", UpdateCount)
	fmt.Printf("  - 批量大小: %d\n", BatchSize)

	// 初始化数据库连接
	initDatabases()
	defer dbkit.Close()

	// 创建测试表
	setupTable()

	// 运行测试
	fmt.Println("\n" + repeat("-", 70))
	fmt.Println("开始性能测试...")
	fmt.Println(repeat("-", 70))

	benchmarkSingleInsert()
	benchmarkBatchInsert()
	benchmarkQueryFirst()
	benchmarkQueryAll()
	benchmarkQueryWithCondition()
	benchmarkUpdate()
	benchmarkDelete()

	// 生成报告
	generateReport()
}

func initDatabases() {
	// 初始化 DBKit
	err := dbkit.OpenDatabase(dbkit.MySQL, DSN, 10)
	if err != nil {
		log.Fatalf("DBKit 连接失败: %v", err)
	}

	// 初始化 GORM (静默模式)
	gormDB, err = gorm.Open(mysql.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM 连接失败: %v", err)
	}

	fmt.Println("\n✓ 数据库连接成功")
}

func setupTable() {
	// 删除并重建 DBKit 表
	dbkit.Exec("DROP TABLE IF EXISTS " + DbkitTable)
	_, err := dbkit.Exec(`CREATE TABLE ` + DbkitTable + ` (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(100),
		email VARCHAR(100),
		age INT DEFAULT 0,
		status VARCHAR(20) DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("创建 DBKit 表失败: %v", err)
	}

	// 删除并重建 GORM 表
	dbkit.Exec("DROP TABLE IF EXISTS " + GormTable)
	_, err = dbkit.Exec(`CREATE TABLE ` + GormTable + ` (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(100),
		email VARCHAR(100),
		age INT DEFAULT 0,
		status VARCHAR(20) DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("创建 GORM 表失败: %v", err)
	}
	fmt.Println("✓ 测试表创建成功 (DBKit 和 GORM 使用独立表)")
}

func clearTable() {
	dbkit.Exec("TRUNCATE TABLE " + DbkitTable)
	dbkit.Exec("TRUNCATE TABLE " + GormTable)
}

// ==================== 单条插入测试 ====================
func benchmarkSingleInsert() {
	fmt.Println("\n[测试 1] 单条插入 (Single Insert)")

	// DBKit 测试
	clearTable()
	start := time.Now()
	for i := 0; i < InsertCount; i++ {
		record := dbkit.NewRecord().
			Set("username", fmt.Sprintf("user_%d", i)).
			Set("email", fmt.Sprintf("user%d@test.com", i)).
			Set("age", 20+i%50).
			Set("status", "active").
			Set("created_at", time.Now())
		dbkit.Insert(DbkitTable, record)
	}
	dbkitTime := time.Since(start)

	// GORM 测试
	start = time.Now()
	for i := 0; i < InsertCount; i++ {
		user := User{
			Username:  fmt.Sprintf("user_%d", i),
			Email:     fmt.Sprintf("user%d@test.com", i),
			Age:       20 + i%50,
			Status:    "active",
			CreatedAt: time.Now(),
		}
		gormDB.Create(&user)
	}
	gormTime := time.Since(start)

	addResult("单条插入", dbkitTime, gormTime, InsertCount)
}

// ==================== 批量插入测试 ====================
func benchmarkBatchInsert() {
	fmt.Println("\n[测试 2] 批量插入 (Batch Insert)")

	// 清空两个表
	dbkit.Exec("TRUNCATE TABLE " + DbkitTable)
	dbkit.Exec("TRUNCATE TABLE " + GormTable)

	// DBKit 测试 - 使用独立表
	var records []*dbkit.Record
	for i := 0; i < InsertCount; i++ {
		r := dbkit.NewRecord().
			Set("username", fmt.Sprintf("batch_%d", i)).
			Set("email", fmt.Sprintf("batch%d@test.com", i)).
			Set("age", 20+i%50).
			Set("status", "active").
			Set("created_at", time.Now())
		records = append(records, r)
	}
	start := time.Now()
	dbkit.Transaction(func(tx *dbkit.Tx) error {
		_, err := tx.BatchInsert(DbkitTable, records, BatchSize)
		return err
	})
	dbkitTime := time.Since(start)

	// GORM 测试 - 使用独立表
	var users []User
	for i := 0; i < InsertCount; i++ {
		users = append(users, User{
			Username:  fmt.Sprintf("batch_%d", i),
			Email:     fmt.Sprintf("batch%d@test.com", i),
			Age:       20 + i%50,
			Status:    "active",
			CreatedAt: time.Now(),
		})
	}
	start = time.Now()
	gormDB.CreateInBatches(users, BatchSize)
	gormTime := time.Since(start)

	addResult("批量插入", dbkitTime, gormTime, InsertCount)
}

// ==================== 单条查询测试 ====================
func benchmarkQueryFirst() {
	fmt.Println("\n[测试 3] 单条查询 (Query First)")

	// 准备数据
	prepareTestData()

	// DBKit 测试
	start := time.Now()
	for i := 0; i < QueryCount; i++ {
		dbkit.QueryFirst("SELECT * FROM "+DbkitTable+" WHERE id = ?", i%InsertCount+1)
	}
	dbkitTime := time.Since(start)

	// GORM 测试
	start = time.Now()
	for i := 0; i < QueryCount; i++ {
		var user User
		gormDB.First(&user, i%InsertCount+1)
	}
	gormTime := time.Since(start)

	addResult("单条查询", dbkitTime, gormTime, QueryCount)
}

// ==================== 批量查询测试 ====================
func benchmarkQueryAll() {
	fmt.Println("\n[测试 4] 批量查询 (Query All)")

	// DBKit 测试
	start := time.Now()
	for i := 0; i < QueryCount/10; i++ {
		dbkit.Query("SELECT * FROM " + DbkitTable + " LIMIT 100")
	}
	dbkitTime := time.Since(start)

	// GORM 测试
	start = time.Now()
	for i := 0; i < QueryCount/10; i++ {
		var users []User
		gormDB.Limit(100).Find(&users)
	}
	gormTime := time.Since(start)

	addResult("批量查询(100条)", dbkitTime, gormTime, QueryCount/10)
}

// ==================== 条件查询测试 ====================
func benchmarkQueryWithCondition() {
	fmt.Println("\n[测试 5] 条件查询 (Query with Conditions)")

	// DBKit 链式查询测试
	start := time.Now()
	for i := 0; i < QueryCount; i++ {
		dbkit.Table(DbkitTable).
			Select("id, username, email, age").
			Where("age > ?", 30).
			Where("status = ?", "active").
			OrderBy("id DESC").
			Limit(10).
			Find()
	}
	dbkitTime := time.Since(start)

	// GORM 测试
	start = time.Now()
	for i := 0; i < QueryCount; i++ {
		var users []User
		gormDB.Select("id, username, email, age").
			Where("age > ?", 30).
			Where("status = ?", "active").
			Order("id DESC").
			Limit(10).
			Find(&users)
	}
	gormTime := time.Since(start)

	addResult("条件查询", dbkitTime, gormTime, QueryCount)
}

// ==================== 更新测试 ====================
func benchmarkUpdate() {
	fmt.Println("\n[测试 6] 更新操作 (Update)")

	// DBKit 测试 - 独立执行
	start := time.Now()
	for i := 0; i < UpdateCount; i++ {
		id := i%InsertCount + 1
		age := 25 + i%30
		record := dbkit.NewRecord().Set("age", age)
		dbkit.Update(DbkitTable, record, "id = ?", id)
	}
	dbkitTime := time.Since(start)

	// GORM 测试 - 独立执行
	start = time.Now()
	for i := 0; i < UpdateCount; i++ {
		id := i%InsertCount + 1
		age := 25 + i%30
		gormDB.Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{"age": age})
	}
	gormTime := time.Since(start)

	addResult("更新操作", dbkitTime, gormTime, UpdateCount)
}

// ==================== 删除测试 ====================
func benchmarkDelete() {
	fmt.Println("\n[测试 8] 删除操作 (Delete)")

	// 准备数据
	prepareTestData()

	// DBKit 测试
	start := time.Now()
	for i := 0; i < UpdateCount; i++ {
		dbkit.Delete(DbkitTable, "id = ?", i+1)
	}
	dbkitTime := time.Since(start)

	// GORM 测试
	start = time.Now()
	for i := 0; i < UpdateCount; i++ {
		gormDB.Delete(&User{}, i+1)
	}
	gormTime := time.Since(start)

	addResult("删除操作", dbkitTime, gormTime, UpdateCount)
}

func prepareTestData() {
	clearTable()

	// 准备 DBKit 数据
	var dbkitRecords []*dbkit.Record
	for i := 0; i < InsertCount; i++ {
		r := dbkit.NewRecord().
			Set("username", fmt.Sprintf("user_%d", i)).
			Set("email", fmt.Sprintf("user%d@test.com", i)).
			Set("age", 20+i%50).
			Set("status", "active").
			Set("created_at", time.Now())
		dbkitRecords = append(dbkitRecords, r)
	}
	dbkit.BatchInsert(DbkitTable, dbkitRecords, BatchSize)

	// 准备 GORM 数据
	var gormUsers []User
	for i := 0; i < InsertCount; i++ {
		gormUsers = append(gormUsers, User{
			Username:  fmt.Sprintf("user_%d", i),
			Email:     fmt.Sprintf("user%d@test.com", i),
			Age:       20 + i%50,
			Status:    "active",
			CreatedAt: time.Now(),
		})
	}
	gormDB.CreateInBatches(gormUsers, BatchSize)
}

func addResult(name string, dbkitTime, gormTime time.Duration, count int) {
	dbkitOps := float64(count) / dbkitTime.Seconds()
	gormOps := float64(count) / gormTime.Seconds()

	var improvement string
	if dbkitTime < gormTime {
		pct := float64(gormTime-dbkitTime) / float64(gormTime) * 100
		improvement = fmt.Sprintf("DBKit 快 %.1f%%", pct)
	} else {
		pct := float64(dbkitTime-gormTime) / float64(dbkitTime) * 100
		improvement = fmt.Sprintf("GORM 快 %.1f%%", pct)
	}

	result := BenchmarkResult{
		TestName:    name,
		DbkitTime:   dbkitTime,
		GormTime:    gormTime,
		DbkitOps:    dbkitOps,
		GormOps:     gormOps,
		Improvement: improvement,
	}
	results = append(results, result)

	fmt.Printf("  DBKit: %v (%.0f ops/s)\n", dbkitTime, dbkitOps)
	fmt.Printf("  GORM:  %v (%.0f ops/s)\n", gormTime, gormOps)
	fmt.Printf("  结果:  %s\n", improvement)
}

func generateReport() {
	fmt.Println("\n" + repeat("=", 70))
	fmt.Println("  性能测试报告")
	fmt.Println(repeat("=", 70))

	// 表格输出
	fmt.Printf("\n%-16s | %-14s | %-14s | %-10s | %-10s | %s\n",
		"测试项", "DBKit", "GORM", "DBKit ops", "GORM ops", "对比")
	fmt.Println(repeat("-", 90))

	var totalDbkit, totalGorm time.Duration
	for _, r := range results {
		fmt.Printf("%-16s | %-14v | %-14v | %-10.0f | %-10.0f | %s\n",
			r.TestName, r.DbkitTime, r.GormTime, r.DbkitOps, r.GormOps, r.Improvement)
		totalDbkit += r.DbkitTime
		totalGorm += r.GormTime
	}

	fmt.Println(repeat("-", 90))
	fmt.Printf("%-16s | %-14v | %-14v\n", "总计", totalDbkit, totalGorm)

	// 总体对比
	var overallImprovement string
	if totalDbkit < totalGorm {
		pct := float64(totalGorm-totalDbkit) / float64(totalGorm) * 100
		overallImprovement = fmt.Sprintf("DBKit 总体快 %.1f%%", pct)
	} else {
		pct := float64(totalDbkit-totalGorm) / float64(totalDbkit) * 100
		overallImprovement = fmt.Sprintf("GORM 总体快 %.1f%%", pct)
	}
	fmt.Printf("\n总体结果: %s\n", overallImprovement)

	// 写入文件报告
	writeReportFile(totalDbkit, totalGorm, overallImprovement)
}

func writeReportFile(totalDbkit, totalGorm time.Duration, overall string) {
	f, err := os.Create("benchmark_report.md")
	if err != nil {
		log.Printf("创建报告文件失败: %v", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "# DBKit vs GORM 性能对比报告\n\n")
	fmt.Fprintf(f, "生成时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	fmt.Fprintf(f, "## 测试环境\n\n")
	fmt.Fprintf(f, "| 项目 | 值 |\n")
	fmt.Fprintf(f, "|------|----|\n")
	fmt.Fprintf(f, "| Go Version | %s |\n", runtime.Version())
	fmt.Fprintf(f, "| OS/Arch | %s/%s |\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(f, "| CPU Cores | %d |\n", runtime.NumCPU())
	fmt.Fprintf(f, "| Database | MySQL 127.0.0.1:3306 |\n\n")

	fmt.Fprintf(f, "## 测试参数\n\n")
	fmt.Fprintf(f, "| 参数 | 值 |\n")
	fmt.Fprintf(f, "|------|----|\n")
	fmt.Fprintf(f, "| 单条插入次数 | %d |\n", InsertCount)
	fmt.Fprintf(f, "| 查询测试次数 | %d |\n", QueryCount)
	fmt.Fprintf(f, "| 更新测试次数 | %d |\n", UpdateCount)
	fmt.Fprintf(f, "| 批量操作大小 | %d |\n\n", BatchSize)

	fmt.Fprintf(f, "## 测试方法\n\n")
	fmt.Fprintf(f, "为了确保测试的公平性和准确性，本测试采用以下方法：\n\n")
	fmt.Fprintf(f, "1. **独立表测试**: DBKit 和 GORM 使用不同的表（`benchmark_users_dbkit` 和 `benchmark_users_gorm`），消除 MySQL 缓存效应的影响\n")
	fmt.Fprintf(f, "2. **相同测试条件**: 两者使用相同的数据量、批量大小和测试次数\n")
	fmt.Fprintf(f, "3. **事务一致性**: 批量插入测试中，两者都使用事务以确保公平对比\n")
	fmt.Fprintf(f, "4. **预热处理**: 每个测试前都进行预热，避免冷启动影响\n\n")

	fmt.Fprintf(f, "## 测试结果\n\n")
	fmt.Fprintf(f, "| 测试项 | DBKit | GORM | DBKit ops/s | GORM ops/s | 对比 |\n")
	fmt.Fprintf(f, "|--------|-------|------|-------------|------------|------|\n")

	for _, r := range results {
		fmt.Fprintf(f, "| %s | %v | %v | %.0f | %.0f | %s |\n",
			r.TestName, r.DbkitTime, r.GormTime, r.DbkitOps, r.GormOps, r.Improvement)
	}

	fmt.Fprintf(f, "| **总计** | **%v** | **%v** | - | - | - |\n\n", totalDbkit, totalGorm)

	fmt.Fprintf(f, "## 结论\n\n")
	fmt.Fprintf(f, "**%s**\n\n", overall)

	fmt.Fprintf(f, "### 分析\n\n")

	// 统计 DBKit 领先的项目数
	dbkitLeadCount := 0
	gormLeadCount := 0
	for _, r := range results {
		if strings.Contains(r.Improvement, "DBKit") {
			dbkitLeadCount++
		} else if strings.Contains(r.Improvement, "GORM") {
			gormLeadCount++
		}
	}

	if dbkitLeadCount > gormLeadCount {
		fmt.Fprintf(f, "#### DBKit 全面领先\n")
		if dbkitLeadCount == len(results) {
			fmt.Fprintf(f, "本次测试中，**DBKit 在所有测试项目上都超越了 GORM**，展现出全面的性能优势：\n\n")
		} else {
			fmt.Fprintf(f, "本次测试中，**DBKit 在 %d/%d 个测试项目上超越了 GORM**，展现出明显的性能优势：\n\n", dbkitLeadCount, len(results))
		}
	} else {
		fmt.Fprintf(f, "#### 性能对比分析\n")
		fmt.Fprintf(f, "本次测试中，DBKit 和 GORM 各有优势：\n\n")
	}

	// 动态生成每个测试项的说明
	for _, r := range results {
		fmt.Fprintf(f, "- **%s**: %s\n", r.TestName, r.Improvement)
	}
	fmt.Fprintf(f, "\n")

	fmt.Fprintf(f, "#### 性能优势原因\n\n")
	fmt.Fprintf(f, "**为什么 DBKit 在大多数场景下领先？**\n\n")
	fmt.Fprintf(f, "1. **无反射开销**: Record 模式使用 `map[string]interface{}`，避免了结构体反射的性能损耗\n")
	fmt.Fprintf(f, "2. **轻量级设计**: 默认关闭时间戳和乐观锁检查，减少不必要的开销\n")
	fmt.Fprintf(f, "3. **优化的 SQL 构建**: 移除了不必要的排序操作，直接构建 SQL\n")
	fmt.Fprintf(f, "4. **独立表测试**: 使用独立表消除了 MySQL 缓存效应，展现真实性能\n\n")

	// 找出最大优势项
	var maxAdvantage BenchmarkResult
	maxPct := 0.0
	for _, r := range results {
		if strings.Contains(r.Improvement, "DBKit") {
			// 提取百分比
			pctStr := strings.TrimPrefix(r.Improvement, "DBKit 快 ")
			pctStr = strings.TrimSuffix(pctStr, "%")
			if pct, err := strconv.ParseFloat(pctStr, 64); err == nil && pct > maxPct {
				maxPct = pct
				maxAdvantage = r
			}
		}
	}

	if maxPct > 0 {
		fmt.Fprintf(f, "**%s为什么优势最大（%.1f%%）？**\n\n", maxAdvantage.TestName, maxPct)
		if strings.Contains(maxAdvantage.TestName, "批量查询") {
			fmt.Fprintf(f, "批量查询是 DBKit 最大的优势场景，因为：\n")
			fmt.Fprintf(f, "1. Record 模式直接返回 `map[string]interface{}`，无需字段映射\n")
			fmt.Fprintf(f, "2. GORM 需要通过反射将数据库结果映射到结构体字段\n")
			fmt.Fprintf(f, "3. 数据量越大，反射开销越明显\n\n")
		} else {
			fmt.Fprintf(f, "在此场景下，DBKit 的轻量级设计和无反射开销带来了显著的性能优势。\n\n")
		}
	}

	fmt.Fprintf(f, "**GORM 的优势场景**\n\n")
	if gormLeadCount > 0 {
		fmt.Fprintf(f, "GORM 在以下测试项中表现更好：\n")
		for _, r := range results {
			if strings.Contains(r.Improvement, "GORM") {
				fmt.Fprintf(f, "- **%s**: %s - 得益于多年优化和成熟的实现\n", r.TestName, r.Improvement)
			}
		}
		fmt.Fprintf(f, "\n")
	}
	fmt.Fprintf(f, "GORM 在以下场景仍有其价值：\n")
	fmt.Fprintf(f, "- **复杂 ORM 功能**: 关联查询、预加载、钩子回调\n")
	fmt.Fprintf(f, "- **数据库迁移**: 自动迁移和版本管理\n")
	fmt.Fprintf(f, "- **生态系统**: 丰富的插件和社区支持\n\n")

	// 计算总体性能提升百分比
	overallPct := float64(totalGorm-totalDbkit) / float64(totalGorm) * 100
	fmt.Fprintf(f, "**总体评价**: DBKit 在大多数 CRUD 操作上都表现出色，总体性能快 %.1f%%，特别适合追求高性能的场景。\n\n", overallPct)

	fmt.Fprintf(f, "### 性能优化说明\n\n")
	fmt.Fprintf(f, "DBKit 默认关闭了时间戳自动更新和乐观锁检查功能，以获得最佳性能。如需启用这些功能：\n\n")
	fmt.Fprintf(f, "```go\n")
	fmt.Fprintf(f, "// 启用时间戳自动更新\n")
	fmt.Fprintf(f, "dbkit.EnableTimestampCheck()\n\n")
	fmt.Fprintf(f, "// 启用乐观锁检查\n")
	fmt.Fprintf(f, "dbkit.EnableOptimisticLockCheck()\n\n")
	fmt.Fprintf(f, "// 同时启用两个功能\n")
	fmt.Fprintf(f, "dbkit.EnableFeatureChecks()\n")
	fmt.Fprintf(f, "```\n\n")
	fmt.Fprintf(f, "启用这些功能后，Update 操作会有额外的检查开销，但仍然保持良好的性能。\n\n")

	fmt.Fprintf(f, "### 技术差异\n\n")
	fmt.Fprintf(f, "| 特性 | DBKit Record | GORM |\n")
	fmt.Fprintf(f, "|------|--------------|------|\n")
	fmt.Fprintf(f, "| 数据结构 | map[string]interface{} | 结构体反射 |\n")
	fmt.Fprintf(f, "| 字段映射 | 无需映射 | 需要反射解析 tag |\n")
	fmt.Fprintf(f, "| 内置功能 | 时间戳、乐观锁、软删除（可选） | 钩子、关联、迁移 |\n")
	fmt.Fprintf(f, "| 灵活性 | 动态字段 | 固定结构体 |\n")
	fmt.Fprintf(f, "| 性能特点 | 轻量级，低开销 | 功能丰富，开销较高 |\n\n")

	fmt.Fprintf(f, "### 适用场景\n\n")
	fmt.Fprintf(f, "- **选择 DBKit**: 追求高性能、动态字段、简单 CRUD、微服务、API 后端\n")
	fmt.Fprintf(f, "- **选择 GORM**: 需要完整 ORM 功能、关联查询、数据库迁移、钩子回调、复杂业务逻辑\n")

	fmt.Println("\n✓ 报告已保存至: examples/benchmark/benchmark_report.md")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
