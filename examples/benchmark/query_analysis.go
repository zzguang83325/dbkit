package main

import (
	"fmt"
	"time"

	"github.com/zzguang83325/dbkit"
)

// 分析单条查询性能差异
func analyzeQueryPerformance() {
	fmt.Println("\n" + repeat("=", 70))
	fmt.Println("  单条查询性能分析")
	fmt.Println(repeat("=", 70))

	// 准备数据
	prepareTestData()

	iterations := 1000

	// 测试 1: DBKit QueryFirst with WHERE
	fmt.Println("\n[测试 1] DBKit QueryFirst (WHERE id = ?)")
	start := time.Now()
	for i := 0; i < iterations; i++ {
		dbkit.QueryFirst("SELECT * FROM "+DbkitTable+" WHERE id = ?", i%InsertCount+1)
	}
	dbkitWhereTime := time.Since(start)
	fmt.Printf("  时间: %v (%.0f ops/s)\n", dbkitWhereTime, float64(iterations)/dbkitWhereTime.Seconds())

	// 测试 2: DBKit Table().Where().FindFirst()
	fmt.Println("\n[测试 2] DBKit Table().Where().FindFirst()")
	start = time.Now()
	for i := 0; i < iterations; i++ {
		dbkit.Table(DbkitTable).Where("id = ?", i%InsertCount+1).FindFirst()
	}
	dbkitChainTime := time.Since(start)
	fmt.Printf("  时间: %v (%.0f ops/s)\n", dbkitChainTime, float64(iterations)/dbkitChainTime.Seconds())

	// 测试 3: GORM First with primary key
	fmt.Println("\n[测试 3] GORM First(&user, id) - 主键查询")
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var user User
		gormDB.First(&user, i%InsertCount+1)
	}
	gormPKTime := time.Since(start)
	fmt.Printf("  时间: %v (%.0f ops/s)\n", gormPKTime, float64(iterations)/gormPKTime.Seconds())

	// 测试 4: GORM First with WHERE
	fmt.Println("\n[测试 4] GORM Where(\"id = ?\", id).First(&user)")
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var user User
		gormDB.Where("id = ?", i%InsertCount+1).First(&user)
	}
	gormWhereTime := time.Since(start)
	fmt.Printf("  时间: %v (%.0f ops/s)\n", gormWhereTime, float64(iterations)/gormWhereTime.Seconds())

	// 分析结果
	fmt.Println("\n" + repeat("-", 70))
	fmt.Println("分析结果:")
	fmt.Println(repeat("-", 70))

	fmt.Printf("DBKit QueryFirst vs GORM First(pk): ")
	if dbkitWhereTime < gormPKTime {
		pct := float64(gormPKTime-dbkitWhereTime) / float64(gormPKTime) * 100
		fmt.Printf("DBKit 快 %.1f%%\n", pct)
	} else {
		pct := float64(dbkitWhereTime-gormPKTime) / float64(dbkitWhereTime) * 100
		fmt.Printf("GORM 快 %.1f%%\n", pct)
	}

	fmt.Printf("DBKit QueryFirst vs GORM Where().First(): ")
	if dbkitWhereTime < gormWhereTime {
		pct := float64(gormWhereTime-dbkitWhereTime) / float64(gormWhereTime) * 100
		fmt.Printf("DBKit 快 %.1f%%\n", pct)
	} else {
		pct := float64(dbkitWhereTime-gormWhereTime) / float64(dbkitWhereTime) * 100
		fmt.Printf("GORM 快 %.1f%%\n", pct)
	}

	fmt.Printf("\nGORM First(pk) vs GORM Where().First(): ")
	if gormPKTime < gormWhereTime {
		pct := float64(gormWhereTime-gormPKTime) / float64(gormWhereTime) * 100
		fmt.Printf("First(pk) 快 %.1f%% (主键查询优化)\n", pct)
	} else {
		fmt.Printf("性能相同\n")
	}

	fmt.Println("\n结论:")
	fmt.Println("- GORM 的 First(&user, id) 使用主键查询优化，可能有特殊处理")
	fmt.Println("- 如果使用相同的 WHERE 条件，DBKit 和 GORM 性能应该更接近")
	fmt.Println("- DBKit 的链式查询可能有额外的构建开销")
}
