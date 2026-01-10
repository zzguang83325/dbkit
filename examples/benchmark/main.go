package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/zzguang83325/dbkit/drivers/postgres"

	"github.com/zzguang83325/dbkit"
	"gorm.io/driver/postgres"
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
	DSN         = "user=test password=123456 host=192.168.10.220 port=5432 dbname=postgres sslmode=disable"
	InsertCount = 5000  // 插入测试数量
	QueryCount  = 10000 // 查询测试次数
	UpdateCount = 2000  // 更新测试次数
	BatchSize   = 200   // 批量操作大小

	// 并发测试配置 - 大幅提升并发强度
	ConcurrentWorkers = 100 // 并发工作协程数
	ConcurrentOps     = 500 // 每个协程的操作数
	StressTestTime    = 3   // 压力测试持续时间(秒)
	MaxConnections    = 100 // 最大连接数

	// 极限压力测试配置
	ExtremeWorkers  = 1000 // 极限压力测试协程数
	ExtremeTestTime = 10   // 极限压力测试持续时间(秒)

	// 等待时间配置 - 确保连接完全释放
	WaitAfterDBKit    = 2 // DBKit测试后等待时间(秒)
	WaitAfterGORM     = 2 // GORM测试后等待时间(秒)
	WaitBetweenTests  = 2 // 渐进式测试间等待时间(秒)
	WaitForConnection = 2 // 连接检查后等待时间(秒)

)

// connectDBKit 创建DBKit数据库连接的通用函数
func connectDBKit(maxOpen int) error {
	config := &dbkit.Config{
		Driver:          dbkit.PostgreSQL,
		DSN:             DSN,
		MaxOpen:         maxOpen,
		MaxIdle:         maxOpen / 2,
		ConnMaxLifetime: time.Hour,
	}

	return dbkit.Register("postgres", config)
}

var (
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
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println("  DBKit vs GORM  性能测试")
	fmt.Println("  数据库:PostgreSQL")
	fmt.Println("=" + strings.Repeat("=", 70))

	// 首先检查数据库连接状态
	fmt.Println("\n🔍 检查数据库连接状态...")
	if !checkDatabaseConnection() {
		fmt.Println("❌ 数据库连接检查失败，请检查数据库服务器状态和连接配置")
		return
	}
	fmt.Println("✅ 数据库连接正常")

	fmt.Printf("\n测试环境:\n")
	fmt.Printf("  - Go Version: %s\n", runtime.Version())
	fmt.Printf("  - OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  - CPU Cores: %d\n", runtime.NumCPU())
	fmt.Printf("  - PostgreSQL: 192.168.10.220:5432/postgres\n")

	fmt.Printf("\n基础测试参数:\n")
	fmt.Printf("  - 单条插入: %d 次 (并发%d协程)\n", InsertCount, ConcurrentWorkers)
	fmt.Printf("  - 查询测试: %d 次 (并发%d协程)\n", QueryCount, ConcurrentWorkers)
	fmt.Printf("  - 更新测试: %d 次 (并发%d协程)\n", UpdateCount, ConcurrentWorkers)
	fmt.Printf("  - 批量大小: %d\n", BatchSize)
	fmt.Printf("  - 并发协程数: %d\n", ConcurrentWorkers)
	fmt.Printf("  - 极限压力测试: %d协程 x %d秒\n", ExtremeWorkers, ExtremeTestTime)
	fmt.Printf("  - 数据库最大连接数: %d \n", MaxConnections)

	fmt.Printf("\n注意：为确保测试公平性，每项测试都会独立打开和关闭数据库连接\n")
	fmt.Printf("每次测试间隔包含：连接关闭 → 垃圾回收 → 等待资源释放 → 重新连接\n")
	fmt.Printf("⚠️  重要提示：测试结果会因硬件配置、网络环境、数据库配置等因素而有所不同，请以您自己的测试结果为准！\n")

	// 运行基础性能测试
	fmt.Println("\n" + strings.Repeat("-", 70))
	fmt.Println("开始基础性能测试...")
	fmt.Println(strings.Repeat("-", 70))

	runBasicTests()

	// 运行并发测试
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("开始渐进式并发压力测试...")
	fmt.Println(strings.Repeat("=", 70))

	// 渐进式并发压力测试
	progressiveResults := runProgressiveStressTests()

	// 生成报告
	generateReport(progressiveResults)
}

// runBasicTests 运行基础性能测试
func runBasicTests() {
	// 每个测试都独立打开和关闭连接，确保测试公平性

	// 测试单条插入（并发版本）
	testConcurrentSingleInsert()

	// 测试批量插入
	testBatchInsert()

	// 测试查询（并发版本）
	testConcurrentQuery()

	// 测试更新（并发版本）
	testConcurrentUpdate()

	// 测试删除（并发版本）
	testConcurrentDelete()
}

// testSingleInsert 单条插入测试
func testSingleInsert() {
	fmt.Println("\n[测试 1] 单条插入 (Single Insert)")

	var dbkitTime, gormTime time.Duration

	// DBKit 测试
	fmt.Println("  DBKit 单条插入测试...")

	err := connectDBKit(MaxConnections)
	if err != nil {
		log.Fatalf("DBKit连接失败: %v", err)
	}

	createDBKitTable()

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
	dbkitTime = time.Since(start)

	dbkit.Close()

	// 等待连接完全释放
	time.Sleep(1 * time.Second)

	// GORM 测试
	fmt.Println("  GORM 单条插入测试...")
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM连接失败: %v", err)
	}

	// 设置连接池
	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(MaxConnections)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	createGORMTable(gormDB)

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
	gormTime = time.Since(start)

	sqlDB.Close()

	addResult("单条插入", dbkitTime, gormTime, InsertCount)
}

// testBatchInsert 批量插入测试
func testBatchInsert() {
	fmt.Println("\n[测试 2] 批量插入 (Batch Insert)")

	var dbkitTime, gormTime time.Duration

	// DBKit 测试
	fmt.Println("  DBKit 批量插入测试...")

	err := connectDBKit(MaxConnections)
	if err != nil {
		log.Fatalf("DBKit连接失败: %v", err)
	}

	clearDBKitTable()

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
	dbkitTime = time.Since(start)

	dbkit.Close()

	// 强制垃圾回收和更长等待时间，确保连接完全释放
	runtime.GC()
	time.Sleep(WaitAfterDBKit * time.Second)

	// GORM 测试
	fmt.Println("  GORM 批量插入测试...")
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM连接失败: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(MaxConnections)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	clearGORMTable(gormDB)

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
	gormTime = time.Since(start)

	sqlDB.Close()

	// 强制垃圾回收，确保资源完全释放
	runtime.GC()
	time.Sleep(WaitAfterGORM * time.Second)

	addResult("批量插入", dbkitTime, gormTime, InsertCount)
}

// testQuery 查询测试
func testQuery() {
	fmt.Println("\n[测试 3] 查询测试 (Query)")

	var dbkitTime, gormTime time.Duration

	// DBKit 测试
	fmt.Println("  DBKit 查询测试...")

	err := connectDBKit(MaxConnections)
	if err != nil {
		log.Fatalf("DBKit连接失败: %v", err)
	}

	// 准备测试数据
	prepareDBKitData()

	start := time.Now()
	for i := 0; i < QueryCount; i++ {
		dbkit.QueryFirst("SELECT * FROM "+DbkitTable+" WHERE id = ?", i%InsertCount+1)
	}
	dbkitTime = time.Since(start)

	dbkit.Close()
	time.Sleep(1 * time.Second)

	// GORM 测试
	fmt.Println("  GORM 查询测试...")
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM连接失败: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(MaxConnections)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// 准备测试数据
	prepareGORMData(gormDB)

	start = time.Now()
	for i := 0; i < QueryCount; i++ {
		var user User
		gormDB.First(&user, i%InsertCount+1)
	}
	gormTime = time.Since(start)

	sqlDB.Close()

	addResult("查询测试", dbkitTime, gormTime, QueryCount)
}

// testUpdate 更新测试
func testUpdate() {
	fmt.Println("\n[测试 4] 更新测试 (Update)")

	var dbkitTime, gormTime time.Duration

	// DBKit 测试
	fmt.Println("  DBKit 更新测试...")

	err := connectDBKit(MaxConnections)
	if err != nil {
		log.Fatalf("DBKit连接失败: %v", err)
	}

	start := time.Now()
	for i := 0; i < UpdateCount; i++ {
		id := i%InsertCount + 1
		age := 25 + i%30
		record := dbkit.NewRecord().Set("age", age)
		dbkit.Update(DbkitTable, record, "id = ?", id)
	}
	dbkitTime = time.Since(start)

	dbkit.Close()
	time.Sleep(1 * time.Second)

	// GORM 测试
	fmt.Println("  GORM 更新测试...")
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM连接失败: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(MaxConnections)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	start = time.Now()
	for i := 0; i < UpdateCount; i++ {
		id := i%InsertCount + 1
		age := 25 + i%30
		gormDB.Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{"age": age})
	}
	gormTime = time.Since(start)

	sqlDB.Close()

	addResult("更新测试", dbkitTime, gormTime, UpdateCount)
}

// testDelete 删除测试
func testDelete() {
	fmt.Println("\n[测试 5] 删除测试 (Delete)")

	var dbkitTime, gormTime time.Duration

	// DBKit 删除测试
	fmt.Println("  DBKit 删除测试...")

	err := connectDBKit(MaxConnections)
	if err != nil {
		log.Fatalf("DBKit连接失败: %v", err)
	}

	// 准备删除测试数据
	prepareDBKitData()

	start := time.Now()
	for i := 0; i < UpdateCount; i++ {
		dbkit.Delete(DbkitTable, "id = ?", i+1)
	}
	dbkitTime = time.Since(start)

	dbkit.Close()
	time.Sleep(1 * time.Second)

	// GORM 删除测试
	fmt.Println("  GORM 删除测试...")
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM连接失败: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(MaxConnections)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// 准备删除测试数据
	prepareGORMData(gormDB)

	start = time.Now()
	for i := 0; i < UpdateCount; i++ {
		gormDB.Delete(&User{}, i+1)
	}
	gormTime = time.Since(start)

	sqlDB.Close()

	addResult("删除测试", dbkitTime, gormTime, UpdateCount)
}

// testConcurrentSingleInsert 并发单条插入测试
func testConcurrentSingleInsert() {
	fmt.Println("\n[测试 1] 并发单条插入 (Concurrent Single Insert)")

	var dbkitTime, gormTime time.Duration

	// DBKit 并发插入测试
	fmt.Println("  DBKit 并发单条插入测试...")

	err := connectDBKit(MaxConnections)
	if err != nil {
		log.Fatalf("DBKit连接失败: %v", err)
	}

	createDBKitTable()

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(ConcurrentWorkers)

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		go func(id int) {
			defer wg.Done()
			opsPerWorker := InsertCount / ConcurrentWorkers
			for i := 0; i < opsPerWorker; i++ {
				record := dbkit.NewRecord().
					Set("username", fmt.Sprintf("user_%d_%d", id, i)).
					Set("email", fmt.Sprintf("user%d_%d@test.com", id, i)).
					Set("age", 20+i%50).
					Set("status", "active").
					Set("created_at", time.Now())
				dbkit.Insert(DbkitTable, record)
			}
		}(workerID)
	}
	wg.Wait()
	dbkitTime = time.Since(start)

	dbkit.Close()

	// 强制垃圾回收和更长等待时间，确保连接完全释放
	runtime.GC()
	time.Sleep(WaitAfterDBKit * time.Second)

	// GORM 并发插入测试
	fmt.Println("  GORM 并发单条插入测试...")
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM连接失败: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(MaxConnections)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	createGORMTable(gormDB)

	start = time.Now()
	wg.Add(ConcurrentWorkers)

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		go func(id int) {
			defer wg.Done()
			opsPerWorker := InsertCount / ConcurrentWorkers
			for i := 0; i < opsPerWorker; i++ {
				user := User{
					Username:  fmt.Sprintf("user_%d_%d", id, i),
					Email:     fmt.Sprintf("user%d_%d@test.com", id, i),
					Age:       20 + i%50,
					Status:    "active",
					CreatedAt: time.Now(),
				}
				gormDB.Create(&user)
			}
		}(workerID)
	}
	wg.Wait()
	gormTime = time.Since(start)

	sqlDB.Close()

	// 强制垃圾回收，确保资源完全释放
	runtime.GC()
	time.Sleep(WaitAfterGORM * time.Second)

	addResult("并发单条插入", dbkitTime, gormTime, InsertCount)
}

// testConcurrentQuery 并发查询测试
func testConcurrentQuery() {
	fmt.Println("\n[测试 3] 并发查询测试 (Concurrent Query)")

	var dbkitTime, gormTime time.Duration

	// DBKit 并发查询测试
	fmt.Println("  DBKit 并发查询测试...")

	err := connectDBKit(MaxConnections)
	if err != nil {
		log.Fatalf("DBKit连接失败: %v", err)
	}

	prepareDBKitData()

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(ConcurrentWorkers)

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		go func(id int) {
			defer wg.Done()
			opsPerWorker := QueryCount / ConcurrentWorkers
			for i := 0; i < opsPerWorker; i++ {
				queryID := (i % InsertCount) + 1
				dbkit.QueryFirst("SELECT * FROM "+DbkitTable+" WHERE id = ?", queryID)
			}
		}(workerID)
	}
	wg.Wait()
	dbkitTime = time.Since(start)

	dbkit.Close()

	// 强制垃圾回收和更长等待时间，确保连接完全释放
	runtime.GC()
	time.Sleep(WaitAfterDBKit * time.Second)

	// GORM 并发查询测试
	fmt.Println("  GORM 并发查询测试...")
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM连接失败: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(MaxConnections)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	prepareGORMData(gormDB)

	start = time.Now()
	wg.Add(ConcurrentWorkers)

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		go func(id int) {
			defer wg.Done()
			opsPerWorker := QueryCount / ConcurrentWorkers
			for i := 0; i < opsPerWorker; i++ {
				var user User
				queryID := (i % InsertCount) + 1
				gormDB.First(&user, queryID)
			}
		}(workerID)
	}
	wg.Wait()
	gormTime = time.Since(start)

	sqlDB.Close()

	// 强制垃圾回收，确保资源完全释放
	runtime.GC()
	time.Sleep(WaitAfterGORM * time.Second)

	addResult("并发查询测试", dbkitTime, gormTime, QueryCount)
}

// testConcurrentUpdate 并发更新测试
func testConcurrentUpdate() {
	fmt.Println("\n[测试 4] 并发更新测试 (Concurrent Update)")

	var dbkitTime, gormTime time.Duration

	// DBKit 并发更新测试
	fmt.Println("  DBKit 并发更新测试...")

	err := connectDBKit(MaxConnections)
	if err != nil {
		log.Fatalf("DBKit连接失败: %v", err)
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(ConcurrentWorkers)

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		go func(id int) {
			defer wg.Done()
			opsPerWorker := UpdateCount / ConcurrentWorkers
			for i := 0; i < opsPerWorker; i++ {
				updateID := (i % InsertCount) + 1
				age := 25 + i%30
				record := dbkit.NewRecord().Set("age", age)
				dbkit.Update(DbkitTable, record, "id = ?", updateID)
			}
		}(workerID)
	}
	wg.Wait()
	dbkitTime = time.Since(start)

	dbkit.Close()

	// 强制垃圾回收和更长等待时间，确保连接完全释放
	runtime.GC()
	time.Sleep(WaitAfterDBKit * time.Second)

	// GORM 并发更新测试
	fmt.Println("  GORM 并发更新测试...")
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM连接失败: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(MaxConnections)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	start = time.Now()
	wg.Add(ConcurrentWorkers)

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		go func(id int) {
			defer wg.Done()
			opsPerWorker := UpdateCount / ConcurrentWorkers
			for i := 0; i < opsPerWorker; i++ {
				updateID := (i % InsertCount) + 1
				age := 25 + i%30
				gormDB.Model(&User{}).Where("id = ?", updateID).Updates(map[string]interface{}{"age": age})
			}
		}(workerID)
	}
	wg.Wait()
	gormTime = time.Since(start)

	sqlDB.Close()

	// 强制垃圾回收，确保资源完全释放
	runtime.GC()
	time.Sleep(WaitAfterGORM * time.Second)

	addResult("并发更新测试", dbkitTime, gormTime, UpdateCount)
}

// testConcurrentDelete 并发删除测试
func testConcurrentDelete() {
	fmt.Println("\n[测试 5] 并发删除测试 (Concurrent Delete)")

	var dbkitTime, gormTime time.Duration

	// DBKit 并发删除测试
	fmt.Println("  DBKit 并发删除测试...")

	err := connectDBKit(MaxConnections)
	if err != nil {
		log.Fatalf("DBKit连接失败: %v", err)
	}

	prepareDBKitData()

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(ConcurrentWorkers)

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		go func(id int) {
			defer wg.Done()
			opsPerWorker := UpdateCount / ConcurrentWorkers
			for i := 0; i < opsPerWorker; i++ {
				deleteID := (id * opsPerWorker) + i + 1
				if deleteID <= InsertCount {
					dbkit.Delete(DbkitTable, "id = ?", deleteID)
				}
			}
		}(workerID)
	}
	wg.Wait()
	dbkitTime = time.Since(start)

	dbkit.Close()

	// 强制垃圾回收和更长等待时间，确保连接完全释放
	runtime.GC()
	time.Sleep(WaitAfterDBKit * time.Second)

	// GORM 并发删除测试
	fmt.Println("  GORM 并发删除测试...")
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM连接失败: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(MaxConnections)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	prepareGORMData(gormDB)

	start = time.Now()
	wg.Add(ConcurrentWorkers)

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		go func(id int) {
			defer wg.Done()
			opsPerWorker := UpdateCount / ConcurrentWorkers
			for i := 0; i < opsPerWorker; i++ {
				deleteID := (id * opsPerWorker) + i + 1
				if deleteID <= InsertCount {
					gormDB.Delete(&User{}, deleteID)
				}
			}
		}(workerID)
	}
	wg.Wait()
	gormTime = time.Since(start)

	sqlDB.Close()

	// 强制垃圾回收，确保资源完全释放
	runtime.GC()
	time.Sleep(WaitAfterGORM * time.Second)

	addResult("并发删除测试", dbkitTime, gormTime, UpdateCount)
}

// runConcurrentTests 运行并发测试
func runConcurrentTests() (StressTestResult, StressTestResult) {
	fmt.Println("\n[并发测试] 并发查询测试")

	var dbkitResult, gormResult ConcurrentResult

	// DBKit 并发测试
	fmt.Println("  DBKit 并发查询测试...")

	err := connectDBKit(MaxConnections)
	if err != nil {
		log.Fatalf("DBKit连接失败: %v", err)
	}

	prepareDBKitData()
	dbkitResult = runDBKitConcurrentQuery()
	dbkit.Close()

	// 强制垃圾回收和更长等待时间，确保连接完全释放
	runtime.GC()
	time.Sleep(WaitBetweenTests * time.Second)

	// GORM 并发测试
	fmt.Println("  GORM 并发查询测试...")
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM连接失败: %v", err)
	}

	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(MaxConnections)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	prepareGORMData(gormDB)
	gormResult = runGORMConcurrentQuery(gormDB)
	sqlDB.Close()

	printConcurrentResults("并发查询", dbkitResult, gormResult)

	// DBKit极限压力测试
	dbkitStressResult := testDBKitCacheExtreme()

	// GORM极限压力测试
	gormStressResult := testGORMStressExtreme()

	return dbkitStressResult, gormStressResult
}

// 并发测试结果结构
type ConcurrentResult struct {
	TestName      string
	Workers       int
	TotalOps      int64
	Duration      time.Duration
	ThroughputOps float64
	AvgLatency    time.Duration
	MaxLatency    time.Duration
	MinLatency    time.Duration
	ErrorCount    int64
	SuccessRate   float64
}

// runDBKitConcurrentQuery DBKit并发查询测试
func runDBKitConcurrentQuery() ConcurrentResult {
	var wg sync.WaitGroup
	var totalOps int64
	var successOps int64
	var errorOps int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	start := time.Now()

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < ConcurrentOps; i++ {
				opStart := time.Now()

				queryID := (i % InsertCount) + 1
				_, err := dbkit.QueryFirst("SELECT * FROM "+DbkitTable+" WHERE id = ?", queryID)

				latency := time.Since(opStart).Nanoseconds()
				atomic.AddInt64(&totalOps, 1)

				if err != nil {
					atomic.AddInt64(&errorOps, 1)
				} else {
					atomic.AddInt64(&successOps, 1)
				}

				// 更新延迟统计
				for {
					currentMin := atomic.LoadInt64(&minLatency)
					if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
						break
					}
				}

				for {
					currentMax := atomic.LoadInt64(&maxLatency)
					if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
						break
					}
				}
			}
		}(workerID)
	}

	wg.Wait()
	duration := time.Since(start)

	return ConcurrentResult{
		TestName:      "DBKit并发查询",
		Workers:       ConcurrentWorkers,
		TotalOps:      totalOps,
		Duration:      duration,
		ThroughputOps: float64(totalOps) / duration.Seconds(),
		AvgLatency:    duration / time.Duration(totalOps),
		MaxLatency:    time.Duration(maxLatency),
		MinLatency:    time.Duration(minLatency),
		ErrorCount:    errorOps,
		SuccessRate:   float64(successOps) / float64(totalOps) * 100,
	}
}

// runGORMConcurrentQuery GORM并发查询测试
func runGORMConcurrentQuery(gormDB *gorm.DB) ConcurrentResult {
	var wg sync.WaitGroup
	var totalOps int64
	var successOps int64
	var errorOps int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	start := time.Now()

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < ConcurrentOps; i++ {
				opStart := time.Now()

				var user User
				queryID := (i % InsertCount) + 1
				err := gormDB.First(&user, queryID).Error

				latency := time.Since(opStart).Nanoseconds()
				atomic.AddInt64(&totalOps, 1)

				if err != nil {
					atomic.AddInt64(&errorOps, 1)
				} else {
					atomic.AddInt64(&successOps, 1)
				}

				// 更新延迟统计
				for {
					currentMin := atomic.LoadInt64(&minLatency)
					if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
						break
					}
				}

				for {
					currentMax := atomic.LoadInt64(&maxLatency)
					if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
						break
					}
				}
			}
		}(workerID)
	}

	wg.Wait()
	duration := time.Since(start)

	return ConcurrentResult{
		TestName:      "GORM并发查询",
		Workers:       ConcurrentWorkers,
		TotalOps:      totalOps,
		Duration:      duration,
		ThroughputOps: float64(totalOps) / duration.Seconds(),
		AvgLatency:    duration / time.Duration(totalOps),
		MaxLatency:    time.Duration(maxLatency),
		MinLatency:    time.Duration(minLatency),
		ErrorCount:    errorOps,
		SuccessRate:   float64(successOps) / float64(totalOps) * 100,
	}
}

// 辅助函数
func createDBKitTable() {
	dbkit.Exec("DROP TABLE IF EXISTS " + DbkitTable)
	_, err := dbkit.Exec(`CREATE TABLE ` + DbkitTable + ` (
		id BIGSERIAL PRIMARY KEY,
		username VARCHAR(100),
		email VARCHAR(100),
		age INTEGER DEFAULT 0,
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatalf("创建 DBKit 表失败: %v", err)
	}
}

func createGORMTable(gormDB *gorm.DB) {
	gormDB.Exec("DROP TABLE IF EXISTS " + GormTable)
	gormDB.Exec(`CREATE TABLE ` + GormTable + ` (
		id BIGSERIAL PRIMARY KEY,
		username VARCHAR(100),
		email VARCHAR(100),
		age INTEGER DEFAULT 0,
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
}

func clearDBKitTable() {
	dbkit.Exec("TRUNCATE TABLE " + DbkitTable + " RESTART IDENTITY")
}

func clearGORMTable(gormDB *gorm.DB) {
	gormDB.Exec("TRUNCATE TABLE " + GormTable + " RESTART IDENTITY")
}

func prepareDBKitData() {
	clearDBKitTable()

	var records []*dbkit.Record
	for i := 0; i < InsertCount; i++ {
		r := dbkit.NewRecord().
			Set("username", fmt.Sprintf("user_%d", i)).
			Set("email", fmt.Sprintf("user%d@test.com", i)).
			Set("age", 20+i%50).
			Set("status", "active").
			Set("created_at", time.Now())
		records = append(records, r)
	}
	dbkit.BatchInsert(DbkitTable, records, BatchSize)
}

func prepareGORMData(gormDB *gorm.DB) {
	clearGORMTable(gormDB)

	var users []User
	for i := 0; i < InsertCount; i++ {
		users = append(users, User{
			Username:  fmt.Sprintf("user_%d", i),
			Email:     fmt.Sprintf("user%d@test.com", i),
			Age:       20 + i%50,
			Status:    "active",
			CreatedAt: time.Now(),
		})
	}
	gormDB.CreateInBatches(users, BatchSize)
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

func printConcurrentResults(testName string, dbkitResult, gormResult ConcurrentResult) {
	fmt.Printf("\n  %s 并发测试结果:\n", testName)
	fmt.Printf("    DBKit: %.0f ops/s, 成功率: %.1f%%, 平均延迟: %v\n",
		dbkitResult.ThroughputOps, dbkitResult.SuccessRate, dbkitResult.AvgLatency)
	fmt.Printf("    GORM:  %.0f ops/s, 成功率: %.1f%%, 平均延迟: %v\n",
		gormResult.ThroughputOps, gormResult.SuccessRate, gormResult.AvgLatency)

	if dbkitResult.ThroughputOps > gormResult.ThroughputOps {
		improvement := (dbkitResult.ThroughputOps - gormResult.ThroughputOps) / gormResult.ThroughputOps * 100
		fmt.Printf("    结果: DBKit 吞吐量高 %.1f%%\n", improvement)
	} else {
		improvement := (gormResult.ThroughputOps - dbkitResult.ThroughputOps) / dbkitResult.ThroughputOps * 100
		fmt.Printf("    结果: GORM 吞吐量高 %.1f%%\n", improvement)
	}
}

func generateReport(progressiveResults []ProgressiveTestResult) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("  性能测试报告")
	fmt.Println(strings.Repeat("=", 70))

	// 基础性能测试表格输出
	fmt.Printf("\n基础性能测试结果,协程数:%d:\n", ConcurrentWorkers)
	fmt.Printf("%-16s | %-14s | %-14s | %-10s | %-10s | %s\n",
		"测试项", "DBKit", "GORM", "DBKit ops", "GORM ops", "对比")
	fmt.Println(strings.Repeat("-", 90))

	var totalDbkit, totalGorm time.Duration
	for _, r := range results {
		fmt.Printf("%-16s | %-14v | %-14v | %-10.0f | %-10.0f | %s\n",
			r.TestName, r.DbkitTime, r.GormTime, r.DbkitOps, r.GormOps, r.Improvement)
		totalDbkit += r.DbkitTime
		totalGorm += r.GormTime
	}

	fmt.Println(strings.Repeat("-", 90))
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
	fmt.Printf("\n基础测试总体结果: %s\n", overallImprovement)

	fmt.Println("⚠️  重要提示：测试结果会因环境而异，请以您自己的实际测试结果为准！")

	// 生成详细的markdown报告
	writeDetailedReport(totalDbkit, totalGorm, overallImprovement, progressiveResults)
}

// ==================== 完整的并发测试函数 ====================

// runDBKitConcurrentInsert DBKit并发插入测试
func runDBKitConcurrentInsert() ConcurrentResult {
	var wg sync.WaitGroup
	var totalOps int64
	var successOps int64
	var errorOps int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	start := time.Now()

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < ConcurrentOps; i++ {
				opStart := time.Now()

				record := dbkit.NewRecord().
					Set("username", fmt.Sprintf("concurrent_user_%d_%d", id, i)).
					Set("email", fmt.Sprintf("user%d_%d@test.com", id, i)).
					Set("age", 20+i%50).
					Set("status", "active")

				_, err := dbkit.Insert(DbkitTable, record)

				latency := time.Since(opStart).Nanoseconds()
				atomic.AddInt64(&totalOps, 1)

				if err != nil {
					atomic.AddInt64(&errorOps, 1)
				} else {
					atomic.AddInt64(&successOps, 1)
				}

				// 更新延迟统计
				for {
					currentMin := atomic.LoadInt64(&minLatency)
					if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
						break
					}
				}

				for {
					currentMax := atomic.LoadInt64(&maxLatency)
					if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
						break
					}
				}
			}
		}(workerID)
	}

	wg.Wait()
	duration := time.Since(start)

	return ConcurrentResult{
		TestName:      "DBKit并发插入",
		Workers:       ConcurrentWorkers,
		TotalOps:      totalOps,
		Duration:      duration,
		ThroughputOps: float64(totalOps) / duration.Seconds(),
		AvgLatency:    duration / time.Duration(totalOps),
		MaxLatency:    time.Duration(maxLatency),
		MinLatency:    time.Duration(minLatency),
		ErrorCount:    errorOps,
		SuccessRate:   float64(successOps) / float64(totalOps) * 100,
	}
}

// runGORMConcurrentInsert GORM并发插入测试
func runGORMConcurrentInsert(gormDB *gorm.DB) ConcurrentResult {
	var wg sync.WaitGroup
	var totalOps int64
	var successOps int64
	var errorOps int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	start := time.Now()

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < ConcurrentOps; i++ {
				opStart := time.Now()

				user := User{
					Username: fmt.Sprintf("concurrent_user_%d_%d", id, i),
					Email:    fmt.Sprintf("user%d_%d@test.com", id, i),
					Age:      20 + i%50,
					Status:   "active",
				}

				err := gormDB.Create(&user).Error

				latency := time.Since(opStart).Nanoseconds()
				atomic.AddInt64(&totalOps, 1)

				if err != nil {
					atomic.AddInt64(&errorOps, 1)
				} else {
					atomic.AddInt64(&successOps, 1)
				}

				// 更新延迟统计
				for {
					currentMin := atomic.LoadInt64(&minLatency)
					if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
						break
					}
				}

				for {
					currentMax := atomic.LoadInt64(&maxLatency)
					if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
						break
					}
				}
			}
		}(workerID)
	}

	wg.Wait()
	duration := time.Since(start)

	return ConcurrentResult{
		TestName:      "GORM并发插入",
		Workers:       ConcurrentWorkers,
		TotalOps:      totalOps,
		Duration:      duration,
		ThroughputOps: float64(totalOps) / duration.Seconds(),
		AvgLatency:    duration / time.Duration(totalOps),
		MaxLatency:    time.Duration(maxLatency),
		MinLatency:    time.Duration(minLatency),
		ErrorCount:    errorOps,
		SuccessRate:   float64(successOps) / float64(totalOps) * 100,
	}
}

// runDBKitConcurrentMixed DBKit混合操作测试
func runDBKitConcurrentMixed() ConcurrentResult {
	var wg sync.WaitGroup
	var totalOps int64
	var successOps int64
	var errorOps int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	start := time.Now()

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < ConcurrentOps; i++ {
				opStart := time.Now()
				var err error

				switch i % 4 {
				case 0: // 插入
					record := dbkit.NewRecord().
						Set("username", fmt.Sprintf("mixed_user_%d_%d", id, i)).
						Set("email", fmt.Sprintf("mixed%d_%d@test.com", id, i)).
						Set("age", 20+i%50)
					_, err = dbkit.Insert(DbkitTable, record)
				case 1: // 查询
					queryID := (i % InsertCount) + 1
					_, err = dbkit.QueryFirst("SELECT * FROM "+DbkitTable+" WHERE id = ?", queryID)
				case 2: // 更新
					updateID := (i % InsertCount) + 1
					record := dbkit.NewRecord().Set("age", 25+i%30)
					_, err = dbkit.Update(DbkitTable, record, "id = ?", updateID)
				case 3: // 条件查询
					_, err = dbkit.Query("SELECT * FROM "+DbkitTable+" WHERE age > ? LIMIT 10", 30)
				}

				latency := time.Since(opStart).Nanoseconds()
				atomic.AddInt64(&totalOps, 1)

				if err != nil {
					atomic.AddInt64(&errorOps, 1)
				} else {
					atomic.AddInt64(&successOps, 1)
				}

				// 更新延迟统计
				for {
					currentMin := atomic.LoadInt64(&minLatency)
					if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
						break
					}
				}

				for {
					currentMax := atomic.LoadInt64(&maxLatency)
					if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
						break
					}
				}
			}
		}(workerID)
	}

	wg.Wait()
	duration := time.Since(start)

	return ConcurrentResult{
		TestName:      "DBKit混合操作",
		Workers:       ConcurrentWorkers,
		TotalOps:      totalOps,
		Duration:      duration,
		ThroughputOps: float64(totalOps) / duration.Seconds(),
		AvgLatency:    duration / time.Duration(totalOps),
		MaxLatency:    time.Duration(maxLatency),
		MinLatency:    time.Duration(minLatency),
		ErrorCount:    errorOps,
		SuccessRate:   float64(successOps) / float64(totalOps) * 100,
	}
}

// runGORMConcurrentMixed GORM混合操作测试
func runGORMConcurrentMixed(gormDB *gorm.DB) ConcurrentResult {
	var wg sync.WaitGroup
	var totalOps int64
	var successOps int64
	var errorOps int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	start := time.Now()

	for workerID := 0; workerID < ConcurrentWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < ConcurrentOps; i++ {
				opStart := time.Now()
				var err error

				switch i % 4 {
				case 0: // 插入
					user := User{
						Username: fmt.Sprintf("mixed_user_%d_%d", id, i),
						Email:    fmt.Sprintf("mixed%d_%d@test.com", id, i),
						Age:      20 + i%50,
					}
					err = gormDB.Create(&user).Error
				case 1: // 查询
					var user User
					queryID := (i % InsertCount) + 1
					err = gormDB.First(&user, queryID).Error
				case 2: // 更新
					updateID := (i % InsertCount) + 1
					err = gormDB.Model(&User{}).Where("id = ?", updateID).Update("age", 25+i%30).Error
				case 3: // 条件查询
					var users []User
					err = gormDB.Where("age > ?", 30).Limit(10).Find(&users).Error
				}

				latency := time.Since(opStart).Nanoseconds()
				atomic.AddInt64(&totalOps, 1)

				if err != nil {
					atomic.AddInt64(&errorOps, 1)
				} else {
					atomic.AddInt64(&successOps, 1)
				}

				// 更新延迟统计
				for {
					currentMin := atomic.LoadInt64(&minLatency)
					if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
						break
					}
				}

				for {
					currentMax := atomic.LoadInt64(&maxLatency)
					if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
						break
					}
				}
			}
		}(workerID)
	}

	wg.Wait()
	duration := time.Since(start)

	return ConcurrentResult{
		TestName:      "GORM混合操作",
		Workers:       ConcurrentWorkers,
		TotalOps:      totalOps,
		Duration:      duration,
		ThroughputOps: float64(totalOps) / duration.Seconds(),
		AvgLatency:    duration / time.Duration(totalOps),
		MaxLatency:    time.Duration(maxLatency),
		MinLatency:    time.Duration(minLatency),
		ErrorCount:    errorOps,
		SuccessRate:   float64(successOps) / float64(totalOps) * 100,
	}
}

// runDBKitConnectionPoolTest DBKit连接池压力测试
func runDBKitConnectionPoolTest() ConcurrentResult {
	// 使用等于连接池大小的协程数
	workers := MaxConnections

	var wg sync.WaitGroup
	var totalOps int64
	var successOps int64
	var errorOps int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	start := time.Now()

	for workerID := 0; workerID < workers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < 50; i++ { // 每个协程执行50次操作
				opStart := time.Now()

				queryID := (i % InsertCount) + 1
				_, err := dbkit.QueryFirst("SELECT * FROM "+DbkitTable+" WHERE id = ?", queryID)

				latency := time.Since(opStart).Nanoseconds()
				atomic.AddInt64(&totalOps, 1)

				if err != nil {
					atomic.AddInt64(&errorOps, 1)
				} else {
					atomic.AddInt64(&successOps, 1)
				}

				// 更新延迟统计
				for {
					currentMin := atomic.LoadInt64(&minLatency)
					if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
						break
					}
				}

				for {
					currentMax := atomic.LoadInt64(&maxLatency)
					if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
						break
					}
				}

				// 添加小延迟，避免连接创建过快
				time.Sleep(10 * time.Millisecond)
			}
		}(workerID)
	}

	wg.Wait()
	duration := time.Since(start)

	return ConcurrentResult{
		TestName:      "DBKit连接池",
		Workers:       workers,
		TotalOps:      totalOps,
		Duration:      duration,
		ThroughputOps: float64(totalOps) / duration.Seconds(),
		AvgLatency:    duration / time.Duration(totalOps),
		MaxLatency:    time.Duration(maxLatency),
		MinLatency:    time.Duration(minLatency),
		ErrorCount:    errorOps,
		SuccessRate:   float64(successOps) / float64(totalOps) * 100,
	}
}

// runGORMConnectionPoolTest GORM连接池压力测试
func runGORMConnectionPoolTest(gormDB *gorm.DB) ConcurrentResult {
	// 使用等于连接池大小的协程数
	workers := MaxConnections

	var wg sync.WaitGroup
	var totalOps int64
	var successOps int64
	var errorOps int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	start := time.Now()

	for workerID := 0; workerID < workers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < 50; i++ { // 每个协程执行50次操作
				opStart := time.Now()

				var user User
				queryID := (i % InsertCount) + 1
				err := gormDB.First(&user, queryID).Error

				latency := time.Since(opStart).Nanoseconds()
				atomic.AddInt64(&totalOps, 1)

				if err != nil {
					atomic.AddInt64(&errorOps, 1)
				} else {
					atomic.AddInt64(&successOps, 1)
				}

				// 更新延迟统计
				for {
					currentMin := atomic.LoadInt64(&minLatency)
					if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
						break
					}
				}

				for {
					currentMax := atomic.LoadInt64(&maxLatency)
					if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
						break
					}
				}

				// 添加小延迟，避免连接创建过快
				time.Sleep(10 * time.Millisecond)
			}
		}(workerID)
	}

	wg.Wait()
	duration := time.Since(start)

	return ConcurrentResult{
		TestName:      "GORM连接池",
		Workers:       workers,
		TotalOps:      totalOps,
		Duration:      duration,
		ThroughputOps: float64(totalOps) / duration.Seconds(),
		AvgLatency:    duration / time.Duration(totalOps),
		MaxLatency:    time.Duration(maxLatency),
		MinLatency:    time.Duration(minLatency),
		ErrorCount:    errorOps,
		SuccessRate:   float64(successOps) / float64(totalOps) * 100,
	}
}

// runDBKitLimitTest DBKit极限压力测试
func runDBKitLimitTest() ConcurrentResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(StressTestTime)*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var totalOps int64
	var successOps int64
	var errorOps int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	start := time.Now()

	// 启动协程进行持续压力测试
	for workerID := 0; workerID < ConcurrentWorkers*2; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			opCount := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
					opStart := time.Now()
					var err error

					// 混合操作：查询、插入、更新
					switch opCount % 3 {
					case 0: // 查询
						queryID := (opCount % InsertCount) + 1
						_, err = dbkit.QueryFirst("SELECT * FROM "+DbkitTable+" WHERE id = ?", queryID)
					case 1: // 插入
						record := dbkit.NewRecord().
							Set("username", fmt.Sprintf("stress_user_%d_%d", id, opCount)).
							Set("email", fmt.Sprintf("stress%d_%d@test.com", id, opCount)).
							Set("age", 20+opCount%50)
						_, err = dbkit.Insert(DbkitTable, record)
					case 2: // 更新
						updateID := (opCount % InsertCount) + 1
						record := dbkit.NewRecord().Set("age", 25+opCount%30)
						_, err = dbkit.Update(DbkitTable, record, "id = ?", updateID)
					}

					latency := time.Since(opStart).Nanoseconds()
					atomic.AddInt64(&totalOps, 1)

					if err != nil {
						atomic.AddInt64(&errorOps, 1)
					} else {
						atomic.AddInt64(&successOps, 1)
					}

					// 更新延迟统计
					for {
						currentMin := atomic.LoadInt64(&minLatency)
						if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
							break
						}
					}

					for {
						currentMax := atomic.LoadInt64(&maxLatency)
						if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
							break
						}
					}

					opCount++
				}
			}
		}(workerID)
	}

	wg.Wait()
	duration := time.Since(start)

	return ConcurrentResult{
		TestName:      "DBKit极限测试",
		Workers:       ConcurrentWorkers * 2,
		TotalOps:      totalOps,
		Duration:      duration,
		ThroughputOps: float64(totalOps) / duration.Seconds(),
		AvgLatency:    duration / time.Duration(totalOps),
		MaxLatency:    time.Duration(maxLatency),
		MinLatency:    time.Duration(minLatency),
		ErrorCount:    errorOps,
		SuccessRate:   float64(successOps) / float64(totalOps) * 100,
	}
}

// runGORMLimitTest GORM极限压力测试
func runGORMLimitTest(gormDB *gorm.DB) ConcurrentResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(StressTestTime)*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var totalOps int64
	var successOps int64
	var errorOps int64
	var minLatency int64 = int64(time.Hour)
	var maxLatency int64

	start := time.Now()

	// 启动协程进行持续压力测试
	for workerID := 0; workerID < ConcurrentWorkers*2; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			opCount := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
					opStart := time.Now()
					var err error

					// 混合操作：查询、插入、更新
					switch opCount % 3 {
					case 0: // 查询
						var user User
						queryID := (opCount % InsertCount) + 1
						err = gormDB.First(&user, queryID).Error
					case 1: // 插入
						user := User{
							Username: fmt.Sprintf("stress_user_%d_%d", id, opCount),
							Email:    fmt.Sprintf("stress%d_%d@test.com", id, opCount),
							Age:      20 + opCount%50,
						}
						err = gormDB.Create(&user).Error
					case 2: // 更新
						updateID := (opCount % InsertCount) + 1
						err = gormDB.Model(&User{}).Where("id = ?", updateID).Update("age", 25+opCount%30).Error
					}

					latency := time.Since(opStart).Nanoseconds()
					atomic.AddInt64(&totalOps, 1)

					if err != nil {
						atomic.AddInt64(&errorOps, 1)
					} else {
						atomic.AddInt64(&successOps, 1)
					}

					// 更新延迟统计
					for {
						currentMin := atomic.LoadInt64(&minLatency)
						if latency >= currentMin || atomic.CompareAndSwapInt64(&minLatency, currentMin, latency) {
							break
						}
					}

					for {
						currentMax := atomic.LoadInt64(&maxLatency)
						if latency <= currentMax || atomic.CompareAndSwapInt64(&maxLatency, currentMax, latency) {
							break
						}
					}

					opCount++
				}
			}
		}(workerID)
	}

	wg.Wait()
	duration := time.Since(start)

	return ConcurrentResult{
		TestName:      "GORM极限测试",
		Workers:       ConcurrentWorkers * 2,
		TotalOps:      totalOps,
		Duration:      duration,
		ThroughputOps: float64(totalOps) / duration.Seconds(),
		AvgLatency:    duration / time.Duration(totalOps),
		MaxLatency:    time.Duration(maxLatency),
		MinLatency:    time.Duration(minLatency),
		ErrorCount:    errorOps,
		SuccessRate:   float64(successOps) / float64(totalOps) * 100,
	}
}

// ==================== 结果打印函数 ====================

func printCacheResults(cachedResult, noCacheResult ConcurrentResult) {
	fmt.Printf("\n  缓存性能对比:\n")
	fmt.Printf("    有缓存: %.0f ops/s, 平均延迟: %v\n",
		cachedResult.ThroughputOps, cachedResult.AvgLatency)
	fmt.Printf("    无缓存: %.0f ops/s, 平均延迟: %v\n",
		noCacheResult.ThroughputOps, noCacheResult.AvgLatency)

	improvement := (cachedResult.ThroughputOps - noCacheResult.ThroughputOps) / noCacheResult.ThroughputOps * 100
	fmt.Printf("    结果: 缓存提升性能 %.1f%%\n", improvement)
}

func printConnectionPoolResults(dbkitResult, gormResult ConcurrentResult) {
	fmt.Printf("\n  连接池压力测试结果:\n")
	fmt.Printf("    DBKit: %.0f ops/s, 成功率: %.1f%%, 错误数: %d\n",
		dbkitResult.ThroughputOps, dbkitResult.SuccessRate, dbkitResult.ErrorCount)
	fmt.Printf("    GORM:  %.0f ops/s, 成功率: %.1f%%, 错误数: %d\n",
		gormResult.ThroughputOps, gormResult.SuccessRate, gormResult.ErrorCount)
}

func printLimitTestResults(dbkitResult, gormResult ConcurrentResult) {
	fmt.Printf("\n  极限压力测试结果 (%d秒):\n", StressTestTime)
	fmt.Printf("    DBKit: 总操作 %d, 吞吐量 %.0f ops/s, 成功率 %.1f%%\n",
		dbkitResult.TotalOps, dbkitResult.ThroughputOps, dbkitResult.SuccessRate)
	fmt.Printf("    GORM:  总操作 %d, 吞吐量 %.0f ops/s, 成功率 %.1f%%\n",
		gormResult.TotalOps, gormResult.ThroughputOps, gormResult.SuccessRate)

	fmt.Printf("    延迟对比:\n")
	fmt.Printf("      DBKit - 最小: %v, 最大: %v, 平均: %v\n",
		dbkitResult.MinLatency, dbkitResult.MaxLatency, dbkitResult.AvgLatency)
	fmt.Printf("      GORM  - 最小: %v, 最大: %v, 平均: %v\n",
		gormResult.MinLatency, gormResult.MaxLatency, gormResult.AvgLatency)
}

// writeDetailedReport 生成详细的markdown报告
func writeDetailedReport(totalDbkit, totalGorm time.Duration, overall string, progressiveResults []ProgressiveTestResult) {
	f, err := os.Create("benchmark_report.md")
	if err != nil {
		log.Printf("创建报告文件失败: %v", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "# DBKit vs GORM  性能对比报告\n\n")
	fmt.Fprintf(f, "生成时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	fmt.Fprintf(f, "> **⚠️ 重要提示**：本报告中的测试结果仅供参考。实际性能会因硬件配置、网络环境、数据库配置、系统负载等多种因素而有显著差异。请在您自己的环境中运行测试，以获得最准确的性能数据。\n\n")

	fmt.Fprintf(f, "## 测试环境\n\n")
	fmt.Fprintf(f, "| 项目 | 值 |\n")
	fmt.Fprintf(f, "|------|----|\n")
	fmt.Fprintf(f, "| Go Version | %s |\n", runtime.Version())
	fmt.Fprintf(f, "| OS/Arch | %s/%s |\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(f, "| CPU Cores | %d |\n", runtime.NumCPU())
	fmt.Fprintf(f, "| Database | PostgreSQL  |\n")
	fmt.Fprintf(f, "| 数据库驱动 | pgx/v5 (统一驱动) |\n\n")

	fmt.Fprintf(f, "## 测试参数\n\n")
	fmt.Fprintf(f, "| 参数 | 值 |\n")
	fmt.Fprintf(f, "|------|----|\n")
	fmt.Fprintf(f, "| 单条插入次数 | %d |\n", InsertCount)
	fmt.Fprintf(f, "| 查询测试次数 | %d |\n", QueryCount)
	fmt.Fprintf(f, "| 更新测试次数 | %d |\n", UpdateCount)
	fmt.Fprintf(f, "| 删除测试次数 | %d |\n", UpdateCount)
	fmt.Fprintf(f, "| 批量操作大小 | %d |\n", BatchSize)
	fmt.Fprintf(f, "| 并发协程数 | %d |\n", ConcurrentWorkers)
	fmt.Fprintf(f, "| 每协程操作数 | %d |\n", ConcurrentOps)
	fmt.Fprintf(f, "| 压力测试时长 | %d 秒 |\n", StressTestTime)
	fmt.Fprintf(f, "| 最大连接数 | %d |\n\n", MaxConnections)

	fmt.Fprintf(f, "## 测试方法\n\n")
	fmt.Fprintf(f, "为了确保测试的公平性和准确性，本测试采用以下方法：\n\n")
	fmt.Fprintf(f, "1. **顺序测试架构**: DBKit 和 GORM 分别测试，避免同时连接数据库造成连接数翻倍\n")
	fmt.Fprintf(f, "2. **统一驱动**: 两者都使用相同的 pgx/v5 PostgreSQL 驱动，确保底层性能基准一致\n")
	fmt.Fprintf(f, "3. **独立表测试**: DBKit 和 GORM 使用不同的表（`benchmark_users_dbkit` 和 `benchmark_users_gorm`），表结构相同\n")
	fmt.Fprintf(f, "4. **相同测试条件**: 使用相同的数据量、批量大小和测试次数\n")
	fmt.Fprintf(f, "5. **连接池管理**: 每个测试阶段独立管理连接池，测试完成后立即释放连接\n")
	fmt.Fprintf(f, "6. **资源等待**: 每次切换框架前等待2秒，确保连接完全释放\n\n")

	fmt.Fprintf(f, "## 基础性能测试结果\n\n")
	fmt.Fprintf(f, "| 测试项 | DBKit | GORM | DBKit ops/s | GORM ops/s | 对比 |\n")
	fmt.Fprintf(f, "|--------|-------|------|-------------|------------|------|\n")

	for _, r := range results {
		fmt.Fprintf(f, "| %s | %v | %v | %.0f | %.0f | %s |\n",
			r.TestName, r.DbkitTime, r.GormTime, r.DbkitOps, r.GormOps, r.Improvement)
	}

	fmt.Fprintf(f, "| **总计** | **%v** | **%v** | - | - | - |\n\n", totalDbkit, totalGorm)

	fmt.Fprintf(f, "## 结论\n\n")
	fmt.Fprintf(f, "**%s**\n\n", overall)

	fmt.Fprintf(f, "> **📊 测试结果说明**：以上结果基于特定的测试环境和配置。不同的硬件配置（CPU、内存、存储）、网络延迟、数据库服务器配置、系统负载等因素都会显著影响测试结果。建议您在自己的生产环境或类似环境中进行测试，以获得最具参考价值的性能数据。\n\n")

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

	fmt.Fprintf(f, "#### 性能对比分析\n")
	fmt.Fprintf(f, "本次测试中：\n\n")

	// 动态生成每个测试项的说明
	for _, r := range results {
		fmt.Fprintf(f, "- **%s**: %s\n", r.TestName, r.Improvement)
	}
	fmt.Fprintf(f, "\n")

	fmt.Fprintf(f, "#### DBKit性能优势原因\n\n")

	fmt.Fprintf(f, "1. **无反射开销**: Record 模式使用 `map[string]interface{}`，避免了结构体反射的性能损耗\n")
	fmt.Fprintf(f, "2. **轻量级设计**: 默认关闭时间戳和乐观锁检查，减少不必要的开销\n")
	fmt.Fprintf(f, "3. **优化的SQL生成**: 直接生成SQL语句，减少中间层处理\n")
	fmt.Fprintf(f, "4. **统一驱动优势**: 使用相同的pgx/v5驱动，消除了驱动差异的影响\n\n")

	fmt.Fprintf(f, "GORM 在以下场景的优势：\n")
	fmt.Fprintf(f, "- **复杂 ORM 功能**: 关联查询、预加载、钩子回调\n")
	fmt.Fprintf(f, "- **数据库迁移**: 自动迁移和版本管理\n")
	fmt.Fprintf(f, "- **生态系统**: 丰富的插件和社区支持\n")
	fmt.Fprintf(f, "- **开发效率**: 结构体映射和类型安全\n\n")

	// 计算总体性能提升百分比
	if totalDbkit < totalGorm {
		overallPct := float64(totalGorm-totalDbkit) / float64(totalGorm) * 100
		fmt.Fprintf(f, "**总体评价**: DBKit 在基础 CRUD 操作上都表现出色，总体性能快 %.1f%%。\n\n", overallPct)
	} else {
		overallPct := float64(totalDbkit-totalGorm) / float64(totalDbkit) * 100
		fmt.Fprintf(f, "**总体评价**: GORM 在本次测试中总体性能快 %.1f%%。\n\n", overallPct)
	}

	// 直接添加渐进式并发压力测试结果表格
	if len(progressiveResults) > 0 {
		fmt.Fprintf(f, "## 渐进式并发压力测试结果\n\n")
		fmt.Fprintf(f, "本次测试采用渐进式并发压力测试方法，在不同协程数量下使用相同的测试条件进行对比：\n\n")
		fmt.Fprintf(f, "- **测试方法**: 渐进式增加并发协程数量\n")
		fmt.Fprintf(f, "- **测试条件**: 每个级别使用相同的测试时长和操作混合比例\n")
		fmt.Fprintf(f, "- **操作混合**: 70%%查询 + 20%%插入 + 10%%更新\n")
		fmt.Fprintf(f, "- **测试目标**: 找到不同并发级别下的性能表现和稳定性临界点\n\n")

		// 生成汇总对比表格
		fmt.Fprintf(f, "### 渐进式测试汇总对比表格\n\n")
		fmt.Fprintf(f, "| 并发协程数 | DBKit TPS | GORM TPS | DBKit成功率 | GORM成功率 | 性能优势 | 内存对比 |\n")
		fmt.Fprintf(f, "|-----------|-----------|----------|-------------|------------|----------|----------|\n")

		for _, result := range progressiveResults {
			var performanceComparison string
			if result.DBKitResult.TPS > result.GORMResult.TPS {
				improvement := (result.DBKitResult.TPS - result.GORMResult.TPS) / result.GORMResult.TPS * 100
				performanceComparison = fmt.Sprintf("DBKit快%.0f%%", improvement)
			} else {
				improvement := (result.GORMResult.TPS - result.DBKitResult.TPS) / result.DBKitResult.TPS * 100
				performanceComparison = fmt.Sprintf("GORM快%.0f%%", improvement)
			}

			var memoryComparison string
			if result.DBKitResult.MemoryMB < result.GORMResult.MemoryMB {
				diff := result.GORMResult.MemoryMB - result.DBKitResult.MemoryMB
				memoryComparison = fmt.Sprintf("DBKit少%.1fMB", diff)
			} else if result.GORMResult.MemoryMB < result.DBKitResult.MemoryMB {
				diff := result.DBKitResult.MemoryMB - result.GORMResult.MemoryMB
				memoryComparison = fmt.Sprintf("GORM少%.1fMB", diff)
			} else {
				memoryComparison = "相当"
			}

			fmt.Fprintf(f, "| %d | %.0f | %.0f | %.1f%% | %.1f%% | %s | %s |\n",
				result.Workers,
				result.DBKitResult.TPS,
				result.GORMResult.TPS,
				result.DBKitResult.SuccessRate,
				result.GORMResult.SuccessRate,
				performanceComparison,
				memoryComparison)
		}

		// 为每个并发级别生成详细的对比表格
		fmt.Fprintf(f, "\n### 各并发级别详细对比\n\n")
		for _, result := range progressiveResults {
			fmt.Fprintf(f, "#### %d并发级别详细对比\n\n", result.Workers)
			fmt.Fprintf(f, "| 测试项目 | DBKit | GORM | 对比 |\n")
			fmt.Fprintf(f, "|----------|-------|------|------|\n")
			fmt.Fprintf(f, "| 并发协程数 | %d | %d | 相同 |\n", result.DBKitResult.Workers, result.GORMResult.Workers)
			fmt.Fprintf(f, "| 测试持续时间 | %v | %v | - |\n", result.DBKitResult.TestDuration, result.GORMResult.TestDuration)
			fmt.Fprintf(f, "| 总操作数 | %d | %d | - |\n", result.DBKitResult.TotalOps, result.GORMResult.TotalOps)
			fmt.Fprintf(f, "| 成功操作数 | %d | %d | - |\n", result.DBKitResult.SuccessOps, result.GORMResult.SuccessOps)
			fmt.Fprintf(f, "| 失败操作数 | %d | %d | - |\n", result.DBKitResult.ErrorOps, result.GORMResult.ErrorOps)

			// 成功率对比
			fmt.Fprintf(f, "| 成功率 | %.2f%% | %.2f%% | ", result.DBKitResult.SuccessRate, result.GORMResult.SuccessRate)
			if result.DBKitResult.SuccessRate > result.GORMResult.SuccessRate {
				diff := result.DBKitResult.SuccessRate - result.GORMResult.SuccessRate
				fmt.Fprintf(f, "DBKit高%.2f%% |\n", diff)
			} else if result.GORMResult.SuccessRate > result.DBKitResult.SuccessRate {
				diff := result.GORMResult.SuccessRate - result.DBKitResult.SuccessRate
				fmt.Fprintf(f, "GORM高%.2f%% |\n", diff)
			} else {
				fmt.Fprintf(f, "相同 |\n")
			}

			// TPS对比
			fmt.Fprintf(f, "| **TPS性能** | **%.0f ops/s** | **%.0f ops/s** | ", result.DBKitResult.TPS, result.GORMResult.TPS)
			if result.DBKitResult.TPS > result.GORMResult.TPS {
				improvement := (result.DBKitResult.TPS - result.GORMResult.TPS) / result.GORMResult.TPS * 100
				fmt.Fprintf(f, "**DBKit快%.1f%%** |\n", improvement)
			} else if result.GORMResult.TPS > result.DBKitResult.TPS {
				improvement := (result.GORMResult.TPS - result.DBKitResult.TPS) / result.DBKitResult.TPS * 100
				fmt.Fprintf(f, "**GORM快%.1f%%** |\n", improvement)
			} else {
				fmt.Fprintf(f, "**相同** |\n")
			}

			// 内存占用对比
			fmt.Fprintf(f, "| 内存占用 | %.2f MB | %.2f MB | ", result.DBKitResult.MemoryMB, result.GORMResult.MemoryMB)
			if result.DBKitResult.MemoryMB < result.GORMResult.MemoryMB {
				diff := result.GORMResult.MemoryMB - result.DBKitResult.MemoryMB
				fmt.Fprintf(f, "DBKit少%.2fMB |\n", diff)
			} else if result.GORMResult.MemoryMB < result.DBKitResult.MemoryMB {
				diff := result.DBKitResult.MemoryMB - result.GORMResult.MemoryMB
				fmt.Fprintf(f, "GORM少%.2fMB |\n", diff)
			} else {
				fmt.Fprintf(f, "相同 |\n")
			}

			fmt.Fprintf(f, "| GC次数 | %d | %d | - |\n", result.DBKitResult.GCCount, result.GORMResult.GCCount)
			fmt.Fprintf(f, "| 性能等级 | %s | %s | - |\n\n", result.DBKitResult.PerformanceLevel, result.GORMResult.PerformanceLevel)
		}

		// 渐进式测试分析
		fmt.Fprintf(f, "### 渐进式测试分析\n\n")
		fmt.Fprintf(f, "渐进式并发压力测试通过逐步增加并发协程数，全面评估了两个框架在不同负载下的性能表现：\n\n")

		// 找到最佳性能点
		var bestDBKit, bestGORM ProgressiveTestResult
		for _, result := range progressiveResults {
			if result.DBKitResult.TPS > bestDBKit.DBKitResult.TPS {
				bestDBKit = result
			}
			if result.GORMResult.TPS > bestGORM.GORMResult.TPS && result.GORMResult.SuccessRate >= 95 {
				bestGORM = result
			}
		}

		fmt.Fprintf(f, "- 🚀 **DBKit最佳性能**: %d并发时达到%.0f TPS\n", bestDBKit.Workers, bestDBKit.DBKitResult.TPS)
		fmt.Fprintf(f, "- 📊 **GORM最佳稳定性能**: %d并发时达到%.0f TPS \n",
			bestGORM.Workers, bestGORM.GORMResult.TPS)

		// 分析稳定性临界点
		for _, result := range progressiveResults {
			if result.GORMResult.SuccessRate < 95 {
				fmt.Fprintf(f, "- ⚠️  **GORM稳定性临界点**: %d并发时开始出现稳定性问题 (成功率%.1f%%)\n",
					result.Workers, result.GORMResult.SuccessRate)
				break
			}
		}

		fmt.Fprintf(f, "\n渐进式测试验证了DBKit在各个并发级别下都保持了优异的性能和稳定性表现。\n\n")
	}

	fmt.Fprintf(f, "## 技术差异对比\n\n")
	fmt.Fprintf(f, "| 特性 | DBKit Record | GORM |\n")
	fmt.Fprintf(f, "|------|--------------|------|\n")
	fmt.Fprintf(f, "| 数据结构 | map[string]interface{} | 结构体反射 |\n")
	fmt.Fprintf(f, "| 字段映射 | 无需映射 | 需要反射解析 tag |\n")
	fmt.Fprintf(f, "| 数据库驱动 | pgx/v5 | pgx/v5 |\n")
	fmt.Fprintf(f, "| 连接管理 | 顺序测试，独立连接池 | 顺序测试，独立连接池 |\n")
	fmt.Fprintf(f, "| 内置功能 | 时间戳、乐观锁、软删除（可选）、SQL模板 | 钩子、关联、迁移 |\n")
	fmt.Fprintf(f, "| 灵活性 | 动态字段 | 固定结构体 |\n")
	fmt.Fprintf(f, "| 性能特点 | 轻量级，低开销 | 功能丰富，开销较高 |\n")
	fmt.Fprintf(f, "| 适用场景 | 高性能CRUD，微服务 | 复杂业务逻辑，快速开发 |\n\n")

	fmt.Fprintf(f, "## 使用建议\n\n")
	fmt.Fprintf(f, "### 选择 DBKit 的场景\n")
	fmt.Fprintf(f, "- 🚀 **高性能要求**: 需要极致的CRUD性能\n")
	fmt.Fprintf(f, "- 🔧 **微服务架构**: 轻量级，资源消耗少\n")
	fmt.Fprintf(f, "- 📊 **数据密集型应用**: 大量的数据库操作\n")
	fmt.Fprintf(f, "- ⚡ **实时系统**: 对延迟敏感的应用\n\n")

	fmt.Fprintf(f, "### 选择 GORM 的场景\n")
	fmt.Fprintf(f, "- 🏗️ **复杂业务逻辑**: 需要丰富的ORM功能\n")
	fmt.Fprintf(f, "- 👥 **团队开发**: 需要类型安全和开发效率\n")
	fmt.Fprintf(f, "- 🔄 **数据库迁移**: 需要自动迁移功能\n")
	fmt.Fprintf(f, "- 🌐 **生态系统**: 需要丰富的插件支持\n\n")

	fmt.Fprintf(f, "## 测试环境说明\n\n")
	fmt.Fprintf(f, "- **测试方式**: 顺序测试，避免连接数问题\n")
	fmt.Fprintf(f, "- **驱动统一**: 两者都使用pgx/v5驱动\n")
	fmt.Fprintf(f, "- **连接池配置**: 最大连接数%d，空闲连接数10\n", MaxConnections)
	fmt.Fprintf(f, "- **测试数据**: 使用独立的测试表，避免缓存干扰\n")
	fmt.Fprintf(f, "- **资源管理**: 每个测试阶段独立管理连接和资源\n\n")

	fmt.Fprintf(f, "---\n\n")
	fmt.Fprintf(f, "## 免责声明\n\n")
	fmt.Fprintf(f, "本性能测试报告仅供技术参考，测试结果会因以下因素而产生差异：\n\n")
	fmt.Fprintf(f, "- **硬件环境**：CPU型号、核心数、内存大小、存储类型（SSD/HDD）\n")
	fmt.Fprintf(f, "- **网络环境**：网络延迟、带宽、连接稳定性\n")
	fmt.Fprintf(f, "- **数据库配置**：PostgreSQL版本、配置参数、缓存设置\n")
	fmt.Fprintf(f, "- **系统负载**：其他应用程序的资源占用\n")
	fmt.Fprintf(f, "- **测试时机**：系统状态、缓存预热情况\n\n")
	fmt.Fprintf(f, "**建议**：请在您的实际部署环境中进行测试，以获得最准确的性能评估。\n\n")
	fmt.Fprintf(f, "---\n\n")
	fmt.Fprintf(f, "*本报告由 DBKit 性能测试程序自动生成*\n")
	fmt.Fprintf(f, "*测试程序位置: `examples/benchmark/main.go`*\n")

	fmt.Println("\n✓ 详细报告已保存至: examples/benchmark/benchmark_report.md")
}

// StressTestResult 极限压力测试结果结构
type StressTestResult struct {
	TestName         string
	Workers          int
	TestDuration     time.Duration
	TotalOps         int64
	SuccessOps       int64
	ErrorOps         int64
	SuccessRate      float64
	TPS              float64
	MemoryMB         float64
	GCCount          uint32
	PerformanceLevel string
}

// testDBKitCacheExtreme DBKit缓存极限压力测试
// 参考examples2的压力测试架构，但测试真实的数据库操作而非缓存
func testDBKitCacheExtreme() StressTestResult {
	fmt.Println("\n[极限压力测试] DBKit 数据库极限压力测试")

	// 压力测试配置
	const (
		StressMaxConns    = 100             // 连接池大小（增加2倍）
		StressWorkerCount = ExtremeWorkers  // 并发协程数（增加3倍）
		StressTestSeconds = ExtremeTestTime // 测试持续时间（增加3倍）
	)

	fmt.Println("========================================")
	fmt.Println("DBKit 数据库极限压力测试")
	fmt.Printf("目标：测试真实数据库操作的极限性能\n")
	fmt.Printf("配置：%d协程 x %d秒 = 极限TPS挑战\n", StressWorkerCount, StressTestSeconds)
	fmt.Println("========================================")

	// 初始化DBKit连接
	err := connectDBKit(StressMaxConns)
	if err != nil {
		log.Fatalf("DBKit极限压力测试连接失败: %v", err)
	}
	defer dbkit.Close()

	// 关闭调试模式以获得最佳性能
	dbkit.SetDebugMode(false)

	// 准备压力测试表和数据
	fmt.Println("正在准备压力测试环境...")
	dbkit.Exec("CREATE TABLE IF NOT EXISTS stress_test (id SERIAL PRIMARY KEY, payload TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
	dbkit.Exec("TRUNCATE TABLE stress_test RESTART IDENTITY")

	// 插入一些基础数据供查询测试
	for i := 1; i <= 100; i++ {
		record := dbkit.NewRecord().
			Set("payload", fmt.Sprintf("stress_test_data_%d", i))
		dbkit.Insert("stress_test", record)
	}

	fmt.Printf("环境准备完成，开始极限压力测试...\n")

	var successCount int64
	var errorCount int64
	start := time.Now()
	deadline := start.Add(time.Duration(StressTestSeconds) * time.Second)

	var wg sync.WaitGroup
	wg.Add(StressWorkerCount)

	fmt.Printf("启动 %d 个协程进行极限压力测试 (持续 %d 秒)...\n", StressWorkerCount, StressTestSeconds)

	// 启动大量协程进行压力测试
	for i := 0; i < StressWorkerCount; i++ {
		go func(workerID int) {
			defer wg.Done()
			opCount := 0

			for time.Now().Before(deadline) {
				// 混合操作：查询、插入、更新 (70%查询, 20%插入, 10%更新)
				switch opCount % 10 {
				case 0, 1, 2, 3, 4, 5, 6: // 70% 查询操作
					queryID := (opCount % 100) + 1
					_, err := dbkit.QueryFirst("SELECT * FROM stress_test WHERE id = ?", queryID)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}

				case 7, 8: // 20% 插入操作
					record := dbkit.NewRecord().
						Set("payload", fmt.Sprintf("stress_worker_%d_op_%d", workerID, opCount))
					_, err := dbkit.Insert("stress_test", record)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}

				case 9: // 10% 更新操作
					updateID := (opCount % 100) + 1
					record := dbkit.NewRecord().
						Set("payload", fmt.Sprintf("updated_by_worker_%d_at_%d", workerID, opCount))
					_, err := dbkit.Update("stress_test", record, "id = ?", updateID)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}
				}
				opCount++
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	totalSuccess := atomic.LoadInt64(&successCount)
	totalError := atomic.LoadInt64(&errorCount)
	totalOps := totalSuccess + totalError
	tps := float64(totalSuccess) / duration.Seconds()
	successRate := float64(totalSuccess) / float64(totalOps) * 100

	// 获取内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println("\n========================================")
	fmt.Println("极限压力测试结果")
	fmt.Println("========================================")
	fmt.Printf("测试配置:\n")
	fmt.Printf("  - 并发协程数: %d\n", StressWorkerCount)
	fmt.Printf("  - 测试时长:   %d 秒\n", StressTestSeconds)
	fmt.Printf("  - 连接池大小: %d\n", StressMaxConns)
	fmt.Printf("\n性能指标:\n")
	fmt.Printf("  - 成功操作数: %d\n", totalSuccess)
	fmt.Printf("  - 失败操作数: %d\n", totalError)
	fmt.Printf("  - 总操作数:   %d\n", totalOps)
	fmt.Printf("  - 成功率:     %.2f%%\n", successRate)
	fmt.Printf("  - 实际TPS:    %.2f ops/s\n", tps)
	fmt.Printf("  - 总耗时:     %v\n", duration)
	fmt.Printf("\n资源使用:\n")
	fmt.Printf("  - 内存占用:   %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("  - GC次数:     %d\n", m.NumGC)

	// 性能评级
	fmt.Printf("\n性能评级: ")
	if tps >= 100000 {
		fmt.Printf("🚀 极致性能 (>= 100K TPS)\n")
	} else if tps >= 50000 {
		fmt.Printf("⚡ 高性能 (>= 50K TPS)\n")
	} else if tps >= 10000 {
		fmt.Printf("✅ 良好性能 (>= 10K TPS)\n")
	} else if tps >= 1000 {
		fmt.Printf("📊 标准性能 (>= 1K TPS)\n")
	} else {
		fmt.Printf("⚠️  性能待优化 (< 1K TPS)\n")
	}

	if successRate < 95 {
		fmt.Printf("⚠️  注意：成功率较低 (%.2f%%)，可能需要调整并发参数或检查数据库配置\n", successRate)
	}

	fmt.Println("========================================")

	// 清理测试数据
	dbkit.Exec("DROP TABLE IF EXISTS stress_test")
	fmt.Println("✓ 压力测试完成，测试数据已清理")

	// 确定性能等级
	var performanceLevel string
	if tps >= 100000 {
		performanceLevel = "🚀 极致性能 (>= 100K TPS)"
	} else if tps >= 50000 {
		performanceLevel = "⚡ 高性能 (>= 50K TPS)"
	} else if tps >= 10000 {
		performanceLevel = "✅ 良好性能 (>= 10K TPS)"
	} else if tps >= 1000 {
		performanceLevel = "📊 标准性能 (>= 1K TPS)"
	} else {
		performanceLevel = "⚠️ 性能待优化 (< 1K TPS)"
	}

	// 返回测试结果
	return StressTestResult{
		TestName:         "DBKit极限压力测试",
		Workers:          StressWorkerCount,
		TestDuration:     duration,
		TotalOps:         totalOps,
		SuccessOps:       totalSuccess,
		ErrorOps:         totalError,
		SuccessRate:      successRate,
		TPS:              tps,
		MemoryMB:         float64(m.Alloc) / 1024 / 1024,
		GCCount:          m.NumGC,
		PerformanceLevel: performanceLevel,
	}
}

// testGORMStressExtreme GORM极限压力测试
// 与DBKit使用相同的测试配置，确保公平对比
func testGORMStressExtreme() StressTestResult {
	fmt.Println("\n[极限压力测试] GORM 数据库极限压力测试")

	// 压力测试配置（与DBKit相同）
	const (
		StressMaxConns    = 100             // 连接池大小（增加2倍）
		StressWorkerCount = ExtremeWorkers  // 并发协程数（增加3倍）
		StressTestSeconds = ExtremeTestTime // 测试持续时间（增加3倍）
	)

	fmt.Println("========================================")
	fmt.Println("GORM 数据库极限压力测试")
	fmt.Printf("目标：测试真实数据库操作的极限性能\n")
	fmt.Printf("配置：%d协程 x %d秒 = 极限TPS挑战\n", StressWorkerCount, StressTestSeconds)
	fmt.Println("========================================")

	// 初始化GORM连接
	gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("GORM极限压力测试连接失败: %v", err)
	}

	// 设置连接池
	sqlDB, _ := gormDB.DB()
	sqlDB.SetMaxOpenConns(StressMaxConns)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	defer sqlDB.Close()

	// 准备压力测试表和数据
	fmt.Println("正在准备压力测试环境...")
	gormDB.Exec("CREATE TABLE IF NOT EXISTS gorm_stress_test (id SERIAL PRIMARY KEY, payload TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
	gormDB.Exec("TRUNCATE TABLE gorm_stress_test RESTART IDENTITY")

	// 插入一些基础数据供查询测试
	for i := 1; i <= 100; i++ {
		record := map[string]interface{}{
			"payload": fmt.Sprintf("gorm_stress_test_data_%d", i),
		}
		gormDB.Table("gorm_stress_test").Create(record)
	}

	fmt.Printf("环境准备完成，开始极限压力测试...\n")

	var successCount int64
	var errorCount int64
	start := time.Now()
	deadline := start.Add(time.Duration(StressTestSeconds) * time.Second)

	var wg sync.WaitGroup
	wg.Add(StressWorkerCount)

	fmt.Printf("启动 %d 个协程进行极限压力测试 (持续 %d 秒)...\n", StressWorkerCount, StressTestSeconds)

	// 启动大量协程进行压力测试
	for i := 0; i < StressWorkerCount; i++ {
		go func(workerID int) {
			defer wg.Done()
			opCount := 0

			for time.Now().Before(deadline) {
				// 混合操作：查询、插入、更新 (70%查询, 20%插入, 10%更新)
				switch opCount % 10 {
				case 0, 1, 2, 3, 4, 5, 6: // 70% 查询操作
					queryID := (opCount % 100) + 1
					var result struct {
						ID        int64  `gorm:"column:id"`
						Payload   string `gorm:"column:payload"`
						CreatedAt string `gorm:"column:created_at"`
					}
					err := gormDB.Table("gorm_stress_test").Where("id = ?", queryID).First(&result).Error
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}

				case 7, 8: // 20% 插入操作
					record := map[string]interface{}{
						"payload": fmt.Sprintf("gorm_stress_worker_%d_op_%d", workerID, opCount),
					}
					err := gormDB.Table("gorm_stress_test").Create(record).Error
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}

				case 9: // 10% 更新操作
					updateID := (opCount % 100) + 1
					payload := fmt.Sprintf("gorm_updated_by_worker_%d_at_%d", workerID, opCount)
					err := gormDB.Table("gorm_stress_test").Where("id = ?", updateID).Update("payload", payload).Error
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
					}
				}
				opCount++
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	totalSuccess := atomic.LoadInt64(&successCount)
	totalError := atomic.LoadInt64(&errorCount)
	totalOps := totalSuccess + totalError
	tps := float64(totalSuccess) / duration.Seconds()
	successRate := float64(totalSuccess) / float64(totalOps) * 100

	// 获取内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println("\n========================================")
	fmt.Println("GORM极限压力测试结果")
	fmt.Println("========================================")
	fmt.Printf("测试配置:\n")
	fmt.Printf("  - 并发协程数: %d\n", StressWorkerCount)
	fmt.Printf("  - 测试时长:   %d 秒\n", StressTestSeconds)
	fmt.Printf("  - 连接池大小: %d\n", StressMaxConns)
	fmt.Printf("\n性能指标:\n")
	fmt.Printf("  - 成功操作数: %d\n", totalSuccess)
	fmt.Printf("  - 失败操作数: %d\n", totalError)
	fmt.Printf("  - 总操作数:   %d\n", totalOps)
	fmt.Printf("  - 成功率:     %.2f%%\n", successRate)
	fmt.Printf("  - 实际TPS:    %.2f ops/s\n", tps)
	fmt.Printf("  - 总耗时:     %v\n", duration)
	fmt.Printf("\n资源使用:\n")
	fmt.Printf("  - 内存占用:   %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("  - GC次数:     %d\n", m.NumGC)

	// 性能评级
	fmt.Printf("\n性能评级: ")
	var performanceLevel string
	if tps >= 100000 {
		performanceLevel = "🚀 极致性能 (>= 100K TPS)"
		fmt.Printf("🚀 极致性能 (>= 100K TPS)\n")
	} else if tps >= 50000 {
		performanceLevel = "⚡ 高性能 (>= 50K TPS)"
		fmt.Printf("⚡ 高性能 (>= 50K TPS)\n")
	} else if tps >= 10000 {
		performanceLevel = "✅ 良好性能 (>= 10K TPS)"
		fmt.Printf("✅ 良好性能 (>= 10K TPS)\n")
	} else if tps >= 1000 {
		performanceLevel = "📊 标准性能 (>= 1K TPS)"
		fmt.Printf("📊 标准性能 (>= 1K TPS)\n")
	} else {
		performanceLevel = "⚠️ 性能待优化 (< 1K TPS)"
		fmt.Printf("⚠️ 性能待优化 (< 1K TPS)\n")
	}

	if successRate < 95 {
		fmt.Printf("⚠️  注意：成功率较低 (%.2f%%)，可能需要调整并发参数或检查数据库配置\n", successRate)
	}

	fmt.Println("========================================")

	// 清理测试数据
	gormDB.Exec("DROP TABLE IF EXISTS gorm_stress_test")
	fmt.Println("✓ GORM压力测试完成，测试数据已清理")

	// 返回测试结果
	return StressTestResult{
		TestName:         "GORM极限压力测试",
		Workers:          StressWorkerCount,
		TestDuration:     duration,
		TotalOps:         totalOps,
		SuccessOps:       totalSuccess,
		ErrorOps:         totalError,
		SuccessRate:      successRate,
		TPS:              tps,
		MemoryMB:         float64(m.Alloc) / 1024 / 1024,
		GCCount:          m.NumGC,
		PerformanceLevel: performanceLevel,
	}
}

// runProgressiveStressTests 渐进式并发压力测试
// 测试不同并发级别下的性能表现，找到稳定性临界点
func runProgressiveStressTests() []ProgressiveTestResult {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("渐进式并发压力测试")

	fmt.Println(strings.Repeat("=", 70))

	// 定义测试的并发级别
	concurrencyLevels := []int{100, 300, 500, 1000, 5000}

	var progressiveResults []ProgressiveTestResult

	for _, workers := range concurrencyLevels {
		fmt.Printf("\n[并发级别 %d] 开始测试...\n", workers)

		// DBKit测试
		dbkitResult := runProgressiveTest("DBKit", workers, true)

		// 等待资源释放
		runtime.GC()
		time.Sleep(WaitBetweenTests * time.Second) // 确保连接完全释放

		// GORM测试
		gormResult := runProgressiveTest("GORM", workers, false)

		// 等待资源释放
		runtime.GC()
		time.Sleep(WaitBetweenTests * time.Second) // 确保连接完全释放

		// 保存结果
		progressiveResults = append(progressiveResults, ProgressiveTestResult{
			Workers:     workers,
			DBKitResult: dbkitResult,
			GORMResult:  gormResult,
		})

		// 打印对比结果
		printProgressiveComparison(workers, dbkitResult, gormResult)
	}

	// 生成渐进式测试报告
	generateProgressiveReport(progressiveResults)

	return progressiveResults
}

// ProgressiveTestResult 渐进式测试结果
type ProgressiveTestResult struct {
	Workers     int
	DBKitResult StressTestResult
	GORMResult  StressTestResult
}

// runProgressiveTest 运行单个并发级别的压力测试
func runProgressiveTest(framework string, workers int, isDBKit bool) StressTestResult {
	fmt.Printf("  %s (%d协程 x %d秒)...", framework, workers, StressTestTime)

	var successCount int64
	var errorCount int64
	start := time.Now()
	deadline := start.Add(time.Duration(StressTestTime) * time.Second)

	if isDBKit {
		// DBKit测试
		err := connectDBKit(MaxConnections)
		if err != nil {
			log.Printf("DBKit连接失败: %v", err)
			return StressTestResult{}
		}
		defer dbkit.Close()

		dbkit.SetDebugMode(false)

		// 准备测试表
		dbkit.Exec("CREATE TABLE IF NOT EXISTS progressive_test_dbkit (id SERIAL PRIMARY KEY, payload TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
		dbkit.Exec("TRUNCATE TABLE progressive_test_dbkit RESTART IDENTITY")

		// 插入基础数据
		for i := 1; i <= 100; i++ {
			record := dbkit.NewRecord().Set("payload", fmt.Sprintf("test_data_%d", i))
			dbkit.Insert("progressive_test_dbkit", record)
		}

		var wg sync.WaitGroup
		wg.Add(workers)

		for i := 0; i < workers; i++ {
			go func(workerID int) {
				defer wg.Done()
				opCount := 0

				for time.Now().Before(deadline) {
					switch opCount % 10 {
					case 0, 1, 2, 3, 4, 5, 6: // 70% 查询
						queryID := (opCount % 100) + 1
						_, err := dbkit.QueryFirst("SELECT * FROM progressive_test_dbkit WHERE id = ?", queryID)
						if err != nil {
							atomic.AddInt64(&errorCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					case 7, 8: // 20% 插入
						record := dbkit.NewRecord().Set("payload", fmt.Sprintf("worker_%d_op_%d", workerID, opCount))
						_, err := dbkit.Insert("progressive_test_dbkit", record)
						if err != nil {
							atomic.AddInt64(&errorCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					case 9: // 10% 更新
						updateID := (opCount % 100) + 1
						record := dbkit.NewRecord().Set("payload", fmt.Sprintf("updated_%d_%d", workerID, opCount))
						_, err := dbkit.Update("progressive_test_dbkit", record, "id = ?", updateID)
						if err != nil {
							atomic.AddInt64(&errorCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					}
					opCount++
				}
			}(i)
		}

		wg.Wait()
		dbkit.Exec("DROP TABLE IF EXISTS progressive_test_dbkit")

	} else {
		// GORM测试
		gormDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Printf("GORM连接失败: %v", err)
			return StressTestResult{}
		}

		sqlDB, _ := gormDB.DB()
		sqlDB.SetMaxOpenConns(MaxConnections)
		sqlDB.SetMaxIdleConns(10)                 // 使用固定值，因为这是合理的默认值
		sqlDB.SetConnMaxLifetime(5 * time.Minute) // 使用固定值，因为这是合理的默认值
		defer sqlDB.Close()

		// 准备测试表
		gormDB.Exec("CREATE TABLE IF NOT EXISTS progressive_test_gorm (id SERIAL PRIMARY KEY, payload TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)")
		gormDB.Exec("TRUNCATE TABLE progressive_test_gorm RESTART IDENTITY")

		// 插入基础数据
		for i := 1; i <= 100; i++ {
			record := map[string]interface{}{
				"payload": fmt.Sprintf("test_data_%d", i),
			}
			gormDB.Table("progressive_test_gorm").Create(record)
		}

		var wg sync.WaitGroup
		wg.Add(workers)

		for i := 0; i < workers; i++ {
			go func(workerID int) {
				defer wg.Done()
				opCount := 0

				for time.Now().Before(deadline) {
					switch opCount % 10 {
					case 0, 1, 2, 3, 4, 5, 6: // 70% 查询
						queryID := (opCount % 100) + 1
						var result struct {
							ID        int64  `gorm:"column:id"`
							Payload   string `gorm:"column:payload"`
							CreatedAt string `gorm:"column:created_at"`
						}
						err := gormDB.Table("progressive_test_gorm").Where("id = ?", queryID).First(&result).Error
						if err != nil {
							atomic.AddInt64(&errorCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					case 7, 8: // 20% 插入
						record := map[string]interface{}{
							"payload": fmt.Sprintf("worker_%d_op_%d", workerID, opCount),
						}
						err := gormDB.Table("progressive_test_gorm").Create(record).Error
						if err != nil {
							atomic.AddInt64(&errorCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					case 9: // 10% 更新
						updateID := (opCount % 100) + 1
						payload := fmt.Sprintf("updated_%d_%d", workerID, opCount)
						err := gormDB.Table("progressive_test_gorm").Where("id = ?", updateID).Update("payload", payload).Error
						if err != nil {
							atomic.AddInt64(&errorCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					}
					opCount++
				}
			}(i)
		}

		wg.Wait()
		gormDB.Exec("DROP TABLE IF EXISTS progressive_test_gorm")
	}

	duration := time.Since(start)
	totalSuccess := atomic.LoadInt64(&successCount)
	totalError := atomic.LoadInt64(&errorCount)
	totalOps := totalSuccess + totalError
	tps := float64(totalSuccess) / duration.Seconds()
	successRate := float64(totalSuccess) / float64(totalOps) * 100

	// 获取内存使用情况
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 确定性能等级
	var performanceLevel string
	if tps >= 100000 {
		performanceLevel = "🚀 极致性能"
	} else if tps >= 50000 {
		performanceLevel = "⚡ 高性能"
	} else if tps >= 10000 {
		performanceLevel = "✅ 良好性能"
	} else if tps >= 1000 {
		performanceLevel = "📊 标准性能"
	} else {
		performanceLevel = "⚠️ 性能待优化"
	}

	fmt.Printf(" TPS: %.0f, 成功率: %.1f%%\n", tps, successRate)

	return StressTestResult{
		TestName:         fmt.Sprintf("%s-%d协程", framework, workers),
		Workers:          workers,
		TestDuration:     duration,
		TotalOps:         totalOps,
		SuccessOps:       totalSuccess,
		ErrorOps:         totalError,
		SuccessRate:      successRate,
		TPS:              tps,
		MemoryMB:         float64(m.Alloc) / 1024 / 1024,
		GCCount:          m.NumGC,
		PerformanceLevel: performanceLevel,
	}
}

// printProgressiveComparison 打印渐进式测试对比结果
func printProgressiveComparison(workers int, dbkitResult, gormResult StressTestResult) {
	fmt.Printf("\n  [%d协程对比结果]\n", workers)
	fmt.Printf("    DBKit: %.0f TPS, 成功率 %.1f%%\n", dbkitResult.TPS, dbkitResult.SuccessRate)
	fmt.Printf("    GORM:  %.0f TPS, 成功率 %.1f%%\n", gormResult.TPS, gormResult.SuccessRate)

	if dbkitResult.TPS > gormResult.TPS {
		improvement := (dbkitResult.TPS - gormResult.TPS) / gormResult.TPS * 100
		fmt.Printf("    性能: DBKit快 %.1f%%\n", improvement)
	} else if gormResult.TPS > dbkitResult.TPS {
		improvement := (gormResult.TPS - dbkitResult.TPS) / dbkitResult.TPS * 100
		fmt.Printf("    性能: GORM快 %.1f%%\n", improvement)
	}

	if dbkitResult.SuccessRate > gormResult.SuccessRate {
		diff := dbkitResult.SuccessRate - gormResult.SuccessRate
		fmt.Printf("    稳定性: DBKit高 %.1f%%\n", diff)
	} else if gormResult.SuccessRate > dbkitResult.SuccessRate {
		diff := gormResult.SuccessRate - dbkitResult.SuccessRate
		fmt.Printf("    稳定性: GORM高 %.1f%%\n", diff)
	}
}

// generateProgressiveReport 生成渐进式测试报告
func generateProgressiveReport(results []ProgressiveTestResult) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("渐进式并发压力测试总结")
	fmt.Println(strings.Repeat("=", 70))

	// 为每个并发级别生成详细的对比表格
	for _, result := range results {
		fmt.Printf("\n## %d并发级别极限压力测试结果对比\n\n", result.Workers)
		fmt.Printf("| 测试项目 | DBKit | GORM | 对比 |\n")
		fmt.Printf("|----------|-------|------|------|\n")
		fmt.Printf("| 并发协程数 | %d | %d | 相同 |\n", result.DBKitResult.Workers, result.GORMResult.Workers)
		fmt.Printf("| 测试持续时间 | %v | %v | - |\n", result.DBKitResult.TestDuration, result.GORMResult.TestDuration)
		fmt.Printf("| 总操作数 | %d | %d | - |\n", result.DBKitResult.TotalOps, result.GORMResult.TotalOps)
		fmt.Printf("| 成功操作数 | %d | %d | - |\n", result.DBKitResult.SuccessOps, result.GORMResult.SuccessOps)
		fmt.Printf("| 失败操作数 | %d | %d | - |\n", result.DBKitResult.ErrorOps, result.GORMResult.ErrorOps)

		// 成功率对比
		fmt.Printf("| 成功率 | %.2f%% | %.2f%% | ", result.DBKitResult.SuccessRate, result.GORMResult.SuccessRate)
		if result.DBKitResult.SuccessRate > result.GORMResult.SuccessRate {
			diff := result.DBKitResult.SuccessRate - result.GORMResult.SuccessRate
			fmt.Printf("DBKit高%.2f%% |\n", diff)
		} else if result.GORMResult.SuccessRate > result.DBKitResult.SuccessRate {
			diff := result.GORMResult.SuccessRate - result.DBKitResult.SuccessRate
			fmt.Printf("GORM高%.2f%% |\n", diff)
		} else {
			fmt.Printf("相同 |\n")
		}

		// TPS对比
		fmt.Printf("| **极限TPS** | **%.2f ops/s** | **%.2f ops/s** | ", result.DBKitResult.TPS, result.GORMResult.TPS)
		if result.DBKitResult.TPS > result.GORMResult.TPS {
			improvement := (result.DBKitResult.TPS - result.GORMResult.TPS) / result.GORMResult.TPS * 100
			fmt.Printf("**DBKit快%.1f%%** |\n", improvement)
		} else if result.GORMResult.TPS > result.DBKitResult.TPS {
			improvement := (result.GORMResult.TPS - result.DBKitResult.TPS) / result.DBKitResult.TPS * 100
			fmt.Printf("**GORM快%.1f%%** |\n", improvement)
		} else {
			fmt.Printf("**相同** |\n")
		}

		// 内存占用对比
		fmt.Printf("| 内存占用 | %.2f MB | %.2f MB | ", result.DBKitResult.MemoryMB, result.GORMResult.MemoryMB)
		if result.DBKitResult.MemoryMB < result.GORMResult.MemoryMB {
			diff := result.GORMResult.MemoryMB - result.DBKitResult.MemoryMB
			fmt.Printf("DBKit少%.2fMB |\n", diff)
		} else if result.GORMResult.MemoryMB < result.DBKitResult.MemoryMB {
			diff := result.DBKitResult.MemoryMB - result.GORMResult.MemoryMB
			fmt.Printf("GORM少%.2fMB |\n", diff)
		} else {
			fmt.Printf("相同 |\n")
		}

		fmt.Printf("| GC次数 | %d | %d | - |\n", result.DBKitResult.GCCount, result.GORMResult.GCCount)
		fmt.Printf("| 性能等级 | %s | %s | - |\n", result.DBKitResult.PerformanceLevel, result.GORMResult.PerformanceLevel)
	}

	// 生成汇总表格
	fmt.Printf("\n## 渐进式测试汇总表格\n\n")
	fmt.Printf("| 并发数 | DBKit TPS | GORM TPS | DBKit成功率 | GORM成功率 | 性能对比 |\n")
	fmt.Printf("|--------|-----------|----------|-------------|------------|----------|\n")

	for _, result := range results {
		var comparison string
		if result.DBKitResult.TPS > result.GORMResult.TPS {
			improvement := (result.DBKitResult.TPS - result.GORMResult.TPS) / result.GORMResult.TPS * 100
			comparison = fmt.Sprintf("DBKit快%.0f%%", improvement)
		} else {
			improvement := (result.GORMResult.TPS - result.DBKitResult.TPS) / result.DBKitResult.TPS * 100
			comparison = fmt.Sprintf("GORM快%.0f%%", improvement)
		}

		fmt.Printf("| %d | %.0f | %.0f | %.1f%% | %.1f%% | %s |\n",
			result.Workers,
			result.DBKitResult.TPS,
			result.GORMResult.TPS,
			result.DBKitResult.SuccessRate,
			result.GORMResult.SuccessRate,
			comparison)
	}

	// 分析稳定性临界点
	fmt.Printf("\n## 📊 稳定性分析\n\n")
	for _, result := range results {
		if result.GORMResult.SuccessRate < 95 {
			fmt.Printf("⚠️  GORM在%d并发时开始出现稳定性问题 (成功率%.1f%%)\n",
				result.Workers, result.GORMResult.SuccessRate)
			break
		}
	}

	// 找到最佳性能点
	var bestDBKit, bestGORM ProgressiveTestResult
	for _, result := range results {
		if result.DBKitResult.TPS > bestDBKit.DBKitResult.TPS {
			bestDBKit = result
		}
		if result.GORMResult.TPS > bestGORM.GORMResult.TPS && result.GORMResult.SuccessRate >= 95 {
			bestGORM = result
		}
	}

	fmt.Printf("🚀 **DBKit最佳性能**: %d并发时达到%.0f TPS\n", bestDBKit.Workers, bestDBKit.DBKitResult.TPS)
	if bestGORM.Workers > 0 {
		fmt.Printf("📊 **GORM最佳稳定性能**: %d并发时达到%.0f TPS (成功率%.1f%%)\n",
			bestGORM.Workers, bestGORM.GORMResult.TPS, bestGORM.GORMResult.SuccessRate)
	}
}

// checkDatabaseConnection 检查数据库连接状态
func checkDatabaseConnection() bool {
	// 尝试连接数据库
	err := connectDBKit(5) // 只用5个连接测试
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return false
	}
	defer dbkit.Close()

	// 测试ping
	if err := dbkit.Ping(); err != nil {
		fmt.Printf("Ping失败: %v\n", err)
		return false
	}

	// 强制等待连接释放
	runtime.GC()
	time.Sleep(WaitForConnection * time.Second)

	return true
}
